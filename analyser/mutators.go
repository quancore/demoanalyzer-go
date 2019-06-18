package analyser

import (
	"bufio"
	"time"

	"github.com/golang/geo/r3"
	dem "github.com/markus-wa/demoinfocs-golang"
	p_common "github.com/markus-wa/demoinfocs-golang/common"
	common "github.com/quancore/demoanalyzer-go/common"
	logging "github.com/sirupsen/logrus"
)

// Different mutators for demo analyser such as update, add players
// reset match variables etc.

// empirical equipment value limits for
// finding type of round for each team
const (
	ecoRoundLimit   = 2300
	forceRoundLimit = 11000
)

// ######## Initilizers and reset functions##########
// resetAnalyser reset state of analyser
func (analyser *Analyser) resetAnalyser() {
	newStream := bufio.NewReader(analyser.buf)
	parser := dem.NewParserWithConfig(newStream, analyser.cfg)
	analyser.parser = parser
	analyser.resetAnalyserVars()
}

// resetAnalyserVars reset analyser vars
func (analyser *Analyser) resetAnalyserVars() {
	// initilize maps for further use
	analyser.players = make(map[int64]*common.PPlayer)
	analyser.disconnectedPlayers = make(map[int64]*common.DisconnectedTuple)
	analyser.roundWinners = make(map[int]string)
	analyser.killPositions = nil
	analyser.NumOvertime = 6
	analyser.minPlayedRound = 5
	analyser.roundPlayed = 0
	analyser.inRound = false
	analyser.isSuccesfulAnalyzed = false
	analyser.lastCheckedTick = 0
}

// initilizeRoundMaps initilize map vars related a round with empty maps
func (analyser *Analyser) initilizeRoundMaps(teamT, teamCT []*p_common.Player, tick int) {
	analyser.resetAlivePlayers(teamT, teamCT)
	analyser.killedPlayers = make(map[int64][]*common.KillTuples)
	analyser.kastPlayers = make(map[int64]bool)
	analyser.droppedItems = make(map[int64]*common.ItemDrop)

}

// resetRoundVars reset round based variables
func (analyser *Analyser) resetRoundVars(teamT, teamCT []*p_common.Player, tick int) {
	// if we are not already in a round
	analyser.log.WithFields(logging.Fields{
		"tick":         tick,
		"round played": analyser.roundPlayed,
	}).Info("Resetting round vars")
	analyser.initilizeRoundMaps(teamT, teamCT, tick)
	analyser.isBombPlanted = false
	analyser.isBombDefusing = false
	analyser.isBombDefused = false
	analyser.currentTRoundType = common.NormalRound
	analyser.currentCTRoundType = common.NormalRound
	// analyser.currentRoundType = common.NormalRound
	analyser.isTPossibleClutch = false
	analyser.isCTPossibleClutch = false
	analyser.defuser = nil
	analyser.tClutchPlayer = nil
	analyser.ctClutchPlayer = nil
	analyser.inRound = true
	// will cancelled if true
	analyser.isCancelled = false
	analyser.isPlayerHurt = false
	analyser.isTFirstKill = false
	analyser.isCTFirstKill = false
	analyser.isEventHappened = false
	analyser.isPlayerWaiting = false
	analyser.isWeaponFired = false
	analyser.winnerTeam = p_common.TeamUnassigned

	// for second parse, register map occupancy event for each round
	if !analyser.isFirstParse {
		if analyser.navigator != nil {
			analyser.navigator.ResetNavigator()

			// calculate end of the periodic event
			eventEndTick := analyser.roundEnd - analyser.remaningTickCheck
			mapControlEvent := mapControl{eventCommon: eventCommon{analyser: analyser,
				offsetSec: analyser.periodOcccupancyCheck, isPeriodic: true, endTick: eventEndTick}}
			analyser.customScheduler.addEvent(tick, analyser.periodOcccupancyCheck, mapControlEvent)
		}
	}

}

// resetRoundVars reset match based variables
func (analyser *Analyser) resetMatchVars(tick int) {
	// first check whether the match has been played for certain number of rounds
	if !analyser.checkIsMatchValid() {
		analyser.resetScore(tick)
		analyser.resetPlayerStates()
		analyser.resetMatchFlags(tick)
	} else {
		analyser.log.WithFields(logging.Fields{
			"tick":         tick,
			"round played": analyser.roundPlayed,
		}).Info("Certain number of match played for this match so it will not be reset")
	}

}

// resetScore reset match score and round played
func (analyser *Analyser) resetScore(tick int) {
	analyser.roundPlayed = 0
	analyser.ctScore = 0
	analyser.tScore = 0
	analyser.log.WithFields(logging.Fields{
		"tick": tick,
	}).Info("Score has been reset")
}

// resetMatchFlags reset match flags
func (analyser *Analyser) resetMatchFlags(tick int) {
	analyser.matchStarted = true
	analyser.matchEnded = false
	analyser.scoreSwapped = false
	analyser.lastScoreSwapped = 0
	analyser.lastMatchStartedCalled = 0
	analyser.lastRoundEndCalled = 0

	analyser.log.WithFields(logging.Fields{
		"tick":         tick,
		"round played": analyser.roundPlayed,
	}).Info("Match flags has been reset")
}

// ############################################
// ######## Cumulative event notifiers ########

// notifyRoundEnd notify all players to round end
func (analyser *Analyser) notifyRoundEnd(numRoundPlayed int, winnerTeam p_common.Team, roundDurationSecond time.Duration) {
	for _, pplayer := range analyser.players {
		if _, ok := analyser.tAlive[pplayer.GetSteamID()]; ok {
			analyser.kastPlayers[pplayer.GetSteamID()] = true
		} else if _, ok := analyser.ctAlive[pplayer.GetSteamID()]; ok {
			analyser.kastPlayers[pplayer.GetSteamID()] = true
		}
		pplayer.NotifyRoundEnd(numRoundPlayed, winnerTeam, roundDurationSecond)
	}
}

// notifyAllMatchEnd notify all players match has ended
func (analyser *Analyser) notifyAllMatchEnd(tScore, ctScore int) {
	for _, pplayer := range analyser.players {
		pplayer.NotifyMatchEnd(tScore, ctScore)
	}
}

// notifySquareMeter notify all players occupied square meter in map for this round
func (analyser *Analyser) notifySquareMeter(tArea, ctArea float32, tick int) {
	analyser.log.WithFields(logging.Fields{
		"tick":            tick,
		"round":           analyser.roundPlayed,
		"t square meter":  tArea,
		"ct square meter": ctArea,
	}).Info("Square meter occupied for the round")
	for _, pplayer := range analyser.players {
		if pplayer.Team == p_common.TeamTerrorists {
			pplayer.NotifyOccupiedArea(tArea)
		} else if pplayer.Team == p_common.TeamCounterTerrorists {
			pplayer.NotifyOccupiedArea(ctArea)
		}
	}
}

// notifyAliveTeamMembers notify all alive players
func (analyser *Analyser) notifyAliveTeamMembers(playerTeam p_common.Team, deathPosition r3.Vector) {
	var victimTeamMembers map[int64]*common.PPlayer
	if playerTeam == p_common.TeamTerrorists {
		victimTeamMembers = analyser.tAlive
	} else if playerTeam == p_common.TeamCounterTerrorists {
		victimTeamMembers = analyser.ctAlive
	} else {
		return
	}
	for _, pplayer := range victimTeamMembers {
		pplayer.NotifyTeamMemberDistance(deathPosition)
	}
}

// ######## Analyser state mutators ###########

// deleteAlivePlayer remove alive player from alive container
func (analyser *Analyser) deleteAlivePlayer(side p_common.Team, uid int64) bool {
	switch side {
	case p_common.TeamTerrorists:
		delete(analyser.tAlive, uid)
		if !analyser.isFirstParse {
			// after deletion check clutch situation
			analyser.checkClutchSituation()
		}
		return true
	case p_common.TeamCounterTerrorists:
		delete(analyser.ctAlive, uid)
		if !analyser.isFirstParse {
			// after deletion check clutch situation
			analyser.checkClutchSituation()
		}
		return true
	default:
		analyser.log.WithFields(logging.Fields{
			"user id": uid,
		}).Error("Player has no side: ")
		return false
	}
}

// resetAlivePlayers reset alive players per round
func (analyser *Analyser) resetAlivePlayers(teamT, teamCT []*p_common.Player) {
	analyser.ctAlive = make(map[int64]*common.PPlayer)
	analyser.tAlive = make(map[int64]*common.PPlayer)

	if teamT != nil && teamCT != nil {
		// for each terorist
		for _, currPlayer := range teamT {
			var NewPPlayer *common.PPlayer
			var ok bool
			uid := currPlayer.SteamID
			// add non exist players
			if NewPPlayer, ok = analyser.getPlayerByID(uid, true); !ok {
				NewPPlayer = common.NewPPlayer(currPlayer, analyser.log)
				// new player add all player list as well
				analyser.players[uid] = NewPPlayer
			}
			if _, ok = NewPPlayer.GetSide(); ok {
				analyser.tAlive[uid] = NewPPlayer
				NewPPlayer.NotifyRoundStart()
			}

		}

		// for each ct
		for _, currPlayer := range teamCT {
			var NewPPlayer *common.PPlayer
			var ok bool
			uid := currPlayer.SteamID
			if NewPPlayer, ok = analyser.getPlayerByID(uid, true); !ok {
				NewPPlayer = common.NewPPlayer(currPlayer, analyser.log)
				analyser.players[uid] = NewPPlayer
			}
			if _, ok = NewPPlayer.GetSide(); ok {
				analyser.ctAlive[uid] = NewPPlayer
				NewPPlayer.NotifyRoundStart()

			}
		}
	}
}

// resetPlayerStates reset player states
func (analyser *Analyser) resetPlayerStates() {
	// for each players
	for _, currPlayer := range analyser.players {
		analyser.log.WithFields(logging.Fields{
			"player name": currPlayer.Name,
		}).Info("Player status has been reset")
		currPlayer.ResetPlayerState()
	}
}

// updateScore update and or swap t and ct score
func (analyser *Analyser) updateScore(newTscore, newCTscore int, eventType string) bool {
	if newTscore < 0 || newCTscore < 0 {
		return false
	}
	// we are getting new round number smaller than swap round, so need to swap back
	newRoundPlayed := newTscore + newCTscore
	oldTScore := analyser.tScore
	oldCTScore := analyser.ctScore

	// we are directly emitting all kind of score update event
	// without checking, because it usually leafds to correct score
	// updates

	// // if there is incorrect update for round end return false
	// if eventType != "scoreUpdate" && !(utils.Abs(newTscore-oldTScore) < 2 && utils.Abs(newCTscore-oldCTScore) < 2) {
	// 	log.WithFields(log.Fields{
	// 		"new t":      newTscore,
	// 		"old t":      oldTScore,
	// 		"new ct":     newCTscore,
	// 		"old ct":     oldCTScore,
	// 		"event_type": eventType,
	// 	}).Error("Invalid score update")
	//
	// 	return false
	// }

	analyser.tScore = newTscore
	analyser.ctScore = newCTscore
	analyser.roundPlayed = newRoundPlayed

	analyser.log.WithFields(logging.Fields{
		"new t":      newTscore,
		"old t":      oldTScore,
		"new ct":     newCTscore,
		"old ct":     oldCTScore,
		"event_type": eventType,
	}).Info("Score has been updated")

	return true

}

func (analyser *Analyser) swapScore(newTscore, newCTscore int) {
	nROundsPlayed := newCTscore + newTscore
	mpOvertimeMaxrounds := analyser.NumOvertime
	nOvertimeHalf := mpOvertimeMaxrounds / 2
	nOvertimeRounds := nROundsPlayed - maxRounds
	if nROundsPlayed == 15 || nROundsPlayed == maxRounds {
		if nROundsPlayed > analyser.lastScoreSwapped {
			analyser.log.Info("Score has been swapped")
			analyser.tScore, analyser.ctScore = newCTscore, newTscore
			analyser.roundPlayed = nROundsPlayed
			analyser.lastScoreSwapped = nROundsPlayed
		} else {
			analyser.log.Info("Score has already been swapped")
		}
	} else if nOvertimeRounds > 0 && nOvertimeRounds%nOvertimeHalf == 0 {
		if nROundsPlayed > analyser.lastScoreSwapped {
			analyser.log.Info("Score has been swapped")
			analyser.tScore, analyser.ctScore = newCTscore, newTscore
			analyser.lastScoreSwapped = nROundsPlayed
		} else {
			analyser.log.Info("Score has already been swapped")
		}
	}
}

// setRoundStart set correct round number and start, end tick for second time parsing
func (analyser *Analyser) setRoundStart(tick int) bool {
	// current round is ongoing, the event is already valid
	if analyser.checkRoundEventValid(tick) {
		return true
	}

	// if the start tick is smaller than current tick, advance
	for roundNumber, currRound := range analyser.validRounds {
		validend := currRound.EndTick
		// if there is an official end for this round, consider it
		if currRound.OfficialEndTick > 0 {
			validend = currRound.OfficialEndTick
		}

		if roundNumber > analyser.roundPlayed && currRound.StartTick <= tick && validend >= tick {
			analyser.roundStart = currRound.StartTick
			analyser.roundEnd = currRound.EndTick
			analyser.roundOffEnd = currRound.OfficialEndTick
			analyser.curValidRound = currRound
			analyser.roundPlayed = roundNumber
			// register scheduled event handler
			if roundNumber == 1 {
				analyser.registerScheduler()
			}
			analyser.log.WithFields(logging.Fields{
				"tick":         tick,
				"round number": roundNumber,
				"start tick":   currRound.StartTick,
				"end tick":     currRound.EndTick,
				"official end": currRound.OfficialEndTick,
			}).Info("Round number has been set.")

			return true
		}
	}
	return false
}

// setRoundType find the type of round
func (analyser *Analyser) setRoundType(tick int) {
	var startTEquipment, startCTEquipment, spentTMoney, spentCTMoney int
	// var currentTEquipment, currentCTequipment, totalTMoney, totalCTmoney, startT, startCT int
	var tRoundType, ctRoundType string
	// var RoundType string
	gs := analyser.parser.GameState()

	// get teams
	tTeam := gs.Participants().TeamMembers(p_common.TeamTerrorists)
	ctTeam := gs.Participants().TeamMembers(p_common.TeamCounterTerrorists)

	// first calculate equipment values for each team
	for _, currPlayer := range tTeam {
		if NewPPlayer, ok := analyser.getPlayerByID(currPlayer.SteamID, false); ok {
			// we subscribe 200 because this is the price of
			// default pistol.We only care total amount spend
			// for each team
			playerRoundStart := NewPPlayer.GetStartEqValue() - 200
			playerRoundSaved := NewPPlayer.GetMoney()
			NewPPlayer.SetSavedMoney(playerRoundSaved)
			playerSpentMoney := NewPPlayer.GetStartMoney() - playerRoundSaved
			startTEquipment += playerRoundStart
			spentTMoney += playerSpentMoney

			analyser.log.WithFields(logging.Fields{
				"curr. equipment val.":      NewPPlayer.GetCurrentEqValue(),
				"freeze time end eq. value": NewPPlayer.GetFreezetEqValue(),
				"name":                      NewPPlayer.Name,
				"round start eq. val":       playerRoundStart,
				"money":                     NewPPlayer.GetMoney(),
				"round start money":         NewPPlayer.GetStartMoney(),
				"spent":                     playerSpentMoney,
				"saved":                     playerRoundSaved,
			}).Info("t team")
		}
	}

	for _, currPlayer := range ctTeam {
		if NewPPlayer, ok := analyser.getPlayerByID(currPlayer.SteamID, false); ok {
			// we subscribe 200 because this is the price of
			// default pistol.We only care total amount spend
			// for each team
			playerRoundStart := NewPPlayer.GetStartEqValue() - 200
			playerRoundSaved := NewPPlayer.GetMoney()
			NewPPlayer.SetSavedMoney(playerRoundSaved)
			playerSpentMoney := NewPPlayer.GetStartMoney() - playerRoundSaved
			startCTEquipment += playerRoundStart
			spentCTMoney += playerSpentMoney

			analyser.log.WithFields(logging.Fields{
				"curr. equipment val.":      NewPPlayer.GetCurrentEqValue(),
				"freeze time end eq. value": NewPPlayer.GetFreezetEqValue(),
				"name":                      NewPPlayer.Name,
				"round start eq. val":       playerRoundStart,
				"money":                     NewPPlayer.GetMoney(),
				"round start money":         NewPPlayer.GetStartMoney(),
				"spent":                     (NewPPlayer.GetStartMoney() - NewPPlayer.GetMoney()),
				"saved":                     playerRoundSaved,
			}).Info("ct team")

		}
	}
	totalTEqValue := startTEquipment + spentTMoney
	totalCTEqValue := startCTEquipment + spentCTMoney

	// find type of round for each team
	analyser.currentTRoundType, tRoundType = analyser.findRoundTypeByMoney(totalTEqValue)
	analyser.currentCTRoundType, ctRoundType = analyser.findRoundTypeByMoney(totalCTEqValue)
	// analyser.currentRoundType, RoundType = analyser.findRoundTypeByMoney(currentTEquipment, currentCTequipment)

	analyser.log.WithFields(logging.Fields{
		"t team":            tTeam[0].TeamState.ClanName,
		"ct team":           ctTeam[0].TeamState.ClanName,
		"T total spent":     spentTMoney,
		"CT total spent":    spentCTMoney,
		"T start eq. val.":  startTEquipment,
		"CT start eq. val.": startCTEquipment,
		"T total eq. val.":  totalTEqValue,
		"CT total eq. val.": totalCTEqValue,
		"T round type":      tRoundType,
		"CT round type":     ctRoundType,
		// "round type": RoundType,
		"tick":  tick,
		"round": analyser.roundPlayed,
	}).Info("Playing round type:")
}

// // findRoundTypeByMoney find the type of round by using money type
// func (analyser *Analyser) findRoundTypeByMoney(tEqValue, ctEqValue int) (common.RoundType, string) {
// 	var roundType common.RoundType
// 	var roundTypeStr string
// 	var diffPercent float64
//
// 	// pistol round handling only normal time
// 	// first round of each halfs
// 	if analyser.roundPlayed <= 30 && analyser.roundPlayed%15 == 1 {
// 		roundTypeStr = "PistolRound"
// 		roundType = common.PistolRound
// 	} else {
// 		// calculate percentage of current eq. value
// 		diffPercent = math.Abs(math.Round(((float64)(ctEqValue - tEqValue)) / (((float64)(ctEqValue + tEqValue)) / 2) * 100))
// 		if diffPercent >= 75 {
// 			roundType = common.EcoRound
// 			roundTypeStr = "EcoRound"
// 		} else if diffPercent >= 50 && diffPercent < 75 {
// 			roundType = common.ForceBuyRound
// 			roundTypeStr = "ForceBuyRound"
// 		} else {
// 			roundType = common.NormalRound
// 			roundTypeStr = "NormalRound"
//
// 		}
// 	}
//
// 	return roundType, roundTypeStr
// }

// findRoundTypeByMoney find the type of round by using money type
func (analyser *Analyser) findRoundTypeByMoney(eqValue int) (common.RoundType, string) {
	var roundType common.RoundType
	var roundTypeStr string

	// pistol round handling only normal time
	// first round of each halfs
	if analyser.roundPlayed <= 30 && analyser.roundPlayed%15 == 1 {
		roundTypeStr = "PistolRound"
		roundType = common.PistolRound
	} else {
		if eqValue <= ecoRoundLimit {
			roundType = common.EcoRound
			roundTypeStr = "EcoRound"
		} else if eqValue > ecoRoundLimit && eqValue <= forceRoundLimit {
			roundType = common.ForceBuyRound
			roundTypeStr = "ForceBuyRound"
		} else {
			roundType = common.NormalRound
			roundTypeStr = "NormalRound"

		}
	}

	return roundType, roundTypeStr
}

// ############################################
