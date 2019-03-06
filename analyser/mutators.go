package analyser

import (
	"bufio"
	"math"

	dem "github.com/markus-wa/demoinfocs-golang"
	p_common "github.com/markus-wa/demoinfocs-golang/common"
	common "github.com/quancore/demoanalyzer-go/common"
	logging "github.com/sirupsen/logrus"
)

// Different mutators for demo analyser such as update, add players
// reset match variables etc.

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
	analyser.NumOvertime = 6
	analyser.minPlayedRound = 5
	analyser.roundPlayed = 0
}

// initilizeRoundMaps initilize map vars related a round with empty maps
func (analyser *Analyser) initilizeRoundMaps(teamT, teamCT []*p_common.Player, tick int) {
	analyser.resetAlivePlayers(teamT, teamCT)
	analyser.killedPlayers = make(map[int64][]*common.KillTuples)
	analyser.kastPlayers = make(map[int64]bool)
}

// resetRoundVars reset round based variables
func (analyser *Analyser) resetRoundVars(teamT, teamCT []*p_common.Player, tick int) {
	analyser.initilizeRoundMaps(teamT, teamCT, tick)
	analyser.isBombPlanted = false
	analyser.isBombDefusing = false
	analyser.isBombDefused = false
	analyser.currentRoundType = common.NormalRound
	analyser.isPossibleClutch = false
	analyser.defuser = nil
	analyser.clutchPlayer = nil
	analyser.inRound = true
	// will cancelled if true
	analyser.isCancelled = false
	analyser.isPlayerHurt = false
	analyser.isEventHappened = false

}

// resetRoundVars reset match based variables
func (analyser *Analyser) resetMatchVars() {
	// first check whether the match has been played for certain number of rounds
	if !analyser.checkIsMatchValid() {
		analyser.resetScore()
		analyser.resetPlayerStates()
		analyser.resetMatchFlags()
	} else {
		analyser.log.WithFields(logging.Fields{
			"tick":         analyser.getGameTick(),
			"round played": analyser.roundPlayed,
		}).Info("Certain number of match played for this match so it will not be reset")
	}

}

// resetScore reset match score and round played
func (analyser *Analyser) resetScore() {
	analyser.roundPlayed = 0
	analyser.ctScore = 0
	analyser.tScore = 0
	analyser.log.WithFields(logging.Fields{
		"tick": analyser.getGameTick(),
	}).Info("Score has been reset")
}

// resetMatchFlags reset match flags
func (analyser *Analyser) resetMatchFlags() {
	analyser.matchStarted = true
	analyser.matchEnded = false
	analyser.scoreSwapped = false
	analyser.lastScoreSwapped = 0
	analyser.lastMatchStartedCalled = 0
	analyser.lastRoundEndCalled = 0
	analyser.log.WithFields(logging.Fields{
		"tick":         analyser.getGameTick(),
		"round played": analyser.roundPlayed,
	}).Info("Match flags has been reset")
}

// ##################################################
// ######## Analyser state mutators ###########

// deleteAlivePlayer remove alive player from alive container
func (analyser *Analyser) deleteAlivePlayer(side p_common.Team, uid int64) bool {
	switch side {
	case p_common.TeamTerrorists:
		delete(analyser.tAlive, uid)
		// after deleteion check clutch situation
		analyser.checkClutchSituation()
		return true
	case p_common.TeamCounterTerrorists:
		delete(analyser.ctAlive, uid)
		// after deleteion check clutch situation
		analyser.checkClutchSituation()
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
				NewPPlayer = common.NewPPlayer(currPlayer)
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
				NewPPlayer = common.NewPPlayer(currPlayer)
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
	curValidEnd := analyser.roundEnd
	if analyser.roundOffEnd > 0 {
		curValidEnd = analyser.roundOffEnd
	}

	// current round is ongoing, the event is already valid
	if analyser.roundStart <= tick && curValidEnd >= tick {
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
func (analyser *Analyser) setRoundType() {
	var currentTEquipment, currentCTequipment int
	roundType := "NormalRound"
	tick := analyser.getGameTick()
	gs := analyser.parser.GameState()

	// get teams
	tTeam := gs.Participants().TeamMembers(p_common.TeamTerrorists)
	ctTeam := gs.Participants().TeamMembers(p_common.TeamCounterTerrorists)

	for _, currPlayer := range tTeam {
		if NewPPlayer, ok := analyser.getPlayerByID(currPlayer.SteamID, false); ok {
			currentTEquipment += NewPPlayer.GetCurrentEqValue()
		}
	}

	for _, currPlayer := range ctTeam {
		if NewPPlayer, ok := analyser.getPlayerByID(currPlayer.SteamID, false); ok {
			currentCTequipment += NewPPlayer.GetCurrentEqValue()
		}
	}

	// pistol round handling only normal time
	// first round of each halfs
	if analyser.roundPlayed <= 30 && analyser.roundPlayed%15 == 1 {
		roundType = "PistolRound"
		analyser.currentRoundType = common.PistolRound
	} else {
		// calculate percentage of current eq. value
		diffPercent := math.Abs(math.Round((((float64)(currentCTequipment - currentTEquipment)) / (((float64)(currentCTequipment + currentTEquipment)) / 2) * 100)))
		if diffPercent >= 75 {
			analyser.currentRoundType = common.EcoRound
			roundType = "EcoRound"
		} else if diffPercent >= 50 && diffPercent < 75 {
			analyser.currentRoundType = common.ForceBuyRound
			roundType = "ForceBuyRound"
		} //else {
		// 	analyser.currentRoundType = NormalRound
		// 	roundType = "NormalRound"
		//
		// }
	}

	analyser.log.WithFields(logging.Fields{
		"t team":             tTeam[0].TeamState.ClanName,
		"ct team":            ctTeam[0].TeamState.ClanName,
		"special round type": roundType,
		"tick":               tick,
		"round":              analyser.roundPlayed,
	}).Info("Playing round type:")
}

// ############################################
