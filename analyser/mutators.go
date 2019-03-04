package analyser

import (
	"bufio"

	dem "github.com/markus-wa/demoinfocs-golang"
	"github.com/markus-wa/demoinfocs-golang/common"
	"github.com/quancore/demoanalyzer-go/player"
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
	analyser.Players = make(map[int64]*player.PPlayer)
	analyser.DisconnectedPlayers = make(map[int64]*disconnectedTuple)
	analyser.pendingPlayers = make(map[int64]*disconnectedTuple)
	analyser.NumOvertime = 6
	analyser.minPlayedRound = 5
	analyser.RoundPlayed = 0
}

// initilizeRoundMaps initilize map vars related a round with empty maps
func (analyser *Analyser) initilizeRoundMaps(teamT, teamCT []*common.Player, tick int) {
	analyser.resetAlivePlayers(teamT, teamCT)
	analyser.KilledPlayers = make(map[int64][]*killTuples)
	analyser.KastPlayers = make(map[int64]bool)
}

// resetRoundVars reset round based variables
func (analyser *Analyser) resetRoundVars(teamT, teamCT []*common.Player, tick int) {
	analyser.initilizeRoundMaps(teamT, teamCT, tick)
	analyser.IsBombPlanted = false
	analyser.IsBombDefusing = false
	analyser.IsPossibleCLutch = false
	analyser.Defuser = nil
	analyser.ClutchPLayer = nil
	analyser.InRound = true
	// will cancelled if true
	analyser.IsCancelled = false
	analyser.IsPlayerHurt = false
	analyser.IsEventHappened = false

}

// resetRoundVars reset match based variables
func (analyser *Analyser) resetMatchVars() {
	// if we are not in half break
	if !analyser.checkIsMatchValid() {
		analyser.resetScore()
		analyser.resetPlayerStates()
		analyser.resetMatchFlags()
	} else {
		analyser.log.WithFields(logging.Fields{
			"tick":         analyser.getGameTick(),
			"round played": analyser.RoundPlayed,
		}).Info("Certain number of match played for this match so it will not be reset")
	}

}

// resetScore reset match score and round played
func (analyser *Analyser) resetScore() {
	analyser.RoundPlayed = 0
	analyser.CTscore = 0
	analyser.Tscore = 0
	analyser.log.WithFields(logging.Fields{
		"tick": analyser.getGameTick(),
	}).Info("Score has been reset")
}

// resetMatchFlags reset match flags
func (analyser *Analyser) resetMatchFlags() {
	analyser.MatchStarted = true
	analyser.MatchEnded = false
	analyser.ScoreSwapped = false
	analyser.lastScoreSwapped = 0
	analyser.lastMatchStartedCalled = 0
	analyser.lastRoundEndCalled = 0
	analyser.log.WithFields(logging.Fields{
		"tick":         analyser.getGameTick(),
		"round played": analyser.RoundPlayed,
	}).Info("Match flags has been reset")
}

// ##################################################
// ######## Analyser state mutators ###########

// deleteAlivePlayer remove alive player from alive container
func (analyser *Analyser) deleteAlivePlayer(side string, uid int64) bool {
	if side == "T" {
		delete(analyser.TAlive, uid)
		// after deleteion check clutch situation
		analyser.checkClutchSituation()
		return true

	} else if side == "CT" {
		delete(analyser.CtAlive, uid)
		// after deletion check clutch situation
		analyser.checkClutchSituation()
		return true

	} else {
		analyser.log.WithFields(logging.Fields{
			"user id": uid,
		}).Error("Player has no side: ")
		return false
	}
}

// resetAlivePlayers reset alive players per round
func (analyser *Analyser) resetAlivePlayers(teamT, teamCT []*common.Player) {
	analyser.CtAlive = make(map[int64]*player.PPlayer)
	analyser.TAlive = make(map[int64]*player.PPlayer)

	if teamT != nil && teamCT != nil {
		// for each terorist
		for _, currPlayer := range teamT {
			var NewPPlayer *player.PPlayer
			var ok bool
			uid := currPlayer.SteamID
			// add non exist players
			if NewPPlayer, ok = analyser.getPlayerByID(uid, true); !ok {
				NewPPlayer = player.NewPPlayer(currPlayer)
				// new player add all player list as well
				analyser.Players[uid] = NewPPlayer
			}
			if _, ok = NewPPlayer.GetSide(); ok {
				analyser.TAlive[uid] = NewPPlayer
				NewPPlayer.NotifyRoundStart()
			}

		}

		// for each ct
		for _, currPlayer := range teamCT {
			var NewPPlayer *player.PPlayer
			var ok bool
			uid := currPlayer.SteamID
			if NewPPlayer, ok = analyser.getPlayerByID(uid, true); !ok {
				NewPPlayer = player.NewPPlayer(currPlayer)
				analyser.Players[uid] = NewPPlayer
			}
			if _, ok = NewPPlayer.GetSide(); ok {
				analyser.CtAlive[uid] = NewPPlayer
				NewPPlayer.NotifyRoundStart()

			}
		}
	}
}

// resetPlayerStates reset player states
func (analyser *Analyser) resetPlayerStates() {
	// for each players
	for _, currPlayer := range analyser.Players {
		currPlayer.ResetPlayerState()
	}
}

// // updateScore update and or swap t and ct score
// func (analyser *Analyser) updateScore(newTscore, newCTscore int) bool {
// 	if newTscore < 0 || newCTscore < 0 {
// 		return false
// 	}
// 	// we are getting new round number smaller than swap round, so need to swap back
// 	newRoundPlayed := newTscore + newCTscore
// 	isScoreSwapped := newRoundPlayed < analyser.lastScoreSwapped
// 	oldTScore := analyser.Tscore
// 	oldCTScore := analyser.CTscore
//
// 	if !isScoreSwapped && (newTscore == analyser.CTscore && newCTscore == analyser.Tscore) {
// 		log.WithFields(log.Fields{
// 			"new t":  newTscore,
// 			"old t":  oldTScore,
// 			"new ct": newCTscore,
// 			"old ct": oldCTScore,
// 		}).Info("Score has already been swapped")
// 		return false
// 	}
//
// 	// we are not updating score when we are in half break because the expected
// 	// behaviour in this period is only swap score.
// 	// Sometimes there is a score update in half breaks which do the same thing with swap
// 	// so if we consider this update, we will swap double and it leads to wrong situation
// 	if utils.Abs(newTscore-oldTScore) < 2 && utils.Abs(newCTscore-oldCTScore) < 2 {
// 		if !analyser.checkHalfBreak(newTscore, newCTscore) {
// 			analyser.Tscore = newTscore
// 			analyser.CTscore = newCTscore
// 			analyser.RoundPlayed = newRoundPlayed
//
// 			log.WithFields(log.Fields{
// 				"new t":  newTscore,
// 				"old t":  oldTScore,
// 				"new ct": newCTscore,
// 				"old ct": oldCTScore,
// 			}).Info("Score has been updated")
//
// 		}
//
// 		analyser.swapScore(newTscore, newCTscore)
//
// 		// incase of rolling back a situation where score swap is needed.
// 		if isScoreSwapped {
// 			analyser.Tscore, analyser.CTscore = analyser.CTscore, analyser.Tscore
// 			analyser.lastScoreSwapped = 0
// 			log.WithFields(log.Fields{
// 				"new t":  newTscore,
// 				"old t":  oldTScore,
// 				"new ct": newCTscore,
// 				"old ct": oldCTScore,
// 			}).Info("Score has been swapped back.")
// 		}
//
// 		return true
// 	}
//
// 	return false
// }

// updateScore update and or swap t and ct score
func (analyser *Analyser) updateScore(newTscore, newCTscore int, eventType string) bool {
	if newTscore < 0 || newCTscore < 0 {
		return false
	}
	// we are getting new round number smaller than swap round, so need to swap back
	newRoundPlayed := newTscore + newCTscore
	oldTScore := analyser.Tscore
	oldCTScore := analyser.CTscore

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

	analyser.Tscore = newTscore
	analyser.CTscore = newCTscore
	analyser.RoundPlayed = newRoundPlayed

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
			analyser.Tscore, analyser.CTscore = newCTscore, newTscore
			analyser.RoundPlayed = nROundsPlayed
			analyser.lastScoreSwapped = nROundsPlayed
		} else {
			analyser.log.Info("Score has already been swapped")
		}
	} else if nOvertimeRounds > 0 && nOvertimeRounds%nOvertimeHalf == 0 {
		if nROundsPlayed > analyser.lastScoreSwapped {
			analyser.log.Info("Score has been swapped")
			analyser.Tscore, analyser.CTscore = newCTscore, newTscore
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
		validend := currRound.endTick
		// if there is an official end for this round, consider it
		if currRound.officialEndTick > 0 {
			validend = currRound.officialEndTick
		}

		if roundNumber > analyser.RoundPlayed && currRound.startTick <= tick && validend >= tick {
			analyser.roundStart = currRound.startTick
			analyser.roundEnd = currRound.endTick
			analyser.roundOffEnd = currRound.officialEndTick
			analyser.curValidRound = currRound
			analyser.RoundPlayed = roundNumber
			analyser.log.WithFields(logging.Fields{
				"tick":         tick,
				"round number": roundNumber,
				"start tick":   currRound.startTick,
				"end tick":     currRound.endTick,
				"official end": currRound.officialEndTick,
			}).Info("Round number has been set.")

			return true
		}
	}
	return false
}

// ############################################
