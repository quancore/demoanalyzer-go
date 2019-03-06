package analyser

import (
	copier "github.com/jinzhu/copier"
	p_common "github.com/markus-wa/demoinfocs-golang/common"
	"github.com/markus-wa/demoinfocs-golang/events"
	common "github.com/quancore/demoanalyzer-go/common"
	logging "github.com/sirupsen/logrus"
)

// handlers for match events

// handleRoundEnd handle end of the round
// mainly called by score update or round end event
func (analyser *Analyser) handleRoundEnd(e interface{}) {
	tick := analyser.getGameTick()

	// check match has already started and not yet finished
	if analyser.isFirstParse && !analyser.checkMatchValidity() {
		analyser.log.WithFields(logging.Fields{
			"tick": tick,
		}).Error("Round end or score update called outside the match.")
		analyser.isCancelled = true
		return
	}

	var winnerTS, loserTS *p_common.TeamState
	var newTscore, newCTscore int
	var eventString string

	switch e.(type) {
	case events.ScoreUpdated:
		e := e.(events.ScoreUpdated)
		eventString = "scoreUpdate"
		// get team states
		winnerTS = e.TeamState
		loserTS = e.TeamState.Opponent

		if winnerTS == nil || loserTS == nil {
			analyser.log.WithFields(logging.Fields{
				"tick":  tick,
				"event": eventString,
			}).Error("Score update event with nill states")
			return
		}

		switch winnerTS.Team() {
		case p_common.TeamTerrorists:
			newTscore = winnerTS.Score
			newCTscore = loserTS.Score

		case p_common.TeamCounterTerrorists:
			newTscore = loserTS.Score
			newCTscore = winnerTS.Score

		default:
			// Probably match medic or something similar
			analyser.log.Info("No winner in this round ")
		}

		// this score update has been done if we are in round and score update has
		// been called
		// with this call we only update score without ending round because sometimes
		// there is a score update in the beginning of the round but after round start
		// so we should update score without ending this active round
		// so if round is already started but no one hurt, this score update will not end
		// a round
		// for second parse, score update event cannot end a round if it is not counted as
		// round end score update event for the first parsing.
		if analyser.isFirstParse && analyser.inRound && !analyser.isPlayerHurt {
			if !analyser.updateScore(newTscore, newCTscore, eventString) {
				analyser.log.WithFields(logging.Fields{
					"t score":  newTscore,
					"ct score": newCTscore,
					"tick":     tick,
					"event":    eventString,
				}).Error("Invalid score update without ending round.")
			} else {
				analyser.log.WithFields(logging.Fields{
					"t score":  newTscore,
					"ct score": newCTscore,
					"tick":     tick,
					"event":    eventString,
				}).Error("Score has been updated without ending round")
			}
			return
		}

	case events.RoundEnd:
		e := e.(events.RoundEnd)
		eventString = "roundEnd"

		// if first parse
		if analyser.isFirstParse {
			// if the event is round end event
			// check the end reason as well
			if !analyser.isCancelled && !analyser.checkFinishedRoundValidity(e) {
				analyser.log.WithFields(logging.Fields{
					"tick": tick,
				}).Error("Round end because of invalid round end reason")
				analyser.isCancelled = true
				analyser.inRound = false
				return
			}
			if !analyser.isCancelled && !analyser.isPlayerHurt {
				analyser.log.WithFields(logging.Fields{
					"tick": tick,
				}).Error("No one hurted in this round. Will cancelled.")
				analyser.isCancelled = true
			}

		} else { // second parse
			if tick != analyser.roundEnd {
				analyser.log.WithFields(logging.Fields{
					"tick":           tick,
					"round end tick": analyser.roundEnd,
					"round number":   analyser.roundPlayed,
					"event name":     eventString,
				}).Info("Invalid round end.")
				return
			}
		}

		// get team states
		winnerTS = e.WinnerState
		loserTS = e.LoserState

		if winnerTS == nil || loserTS == nil {
			analyser.log.WithFields(logging.Fields{
				"tick":  tick,
				"event": eventString,
			}).Error("Round end with nill states")
			analyser.isCancelled = true
			analyser.inRound = false
			return
		}
		// update score
		switch winnerTS.Team() {
		case p_common.TeamTerrorists:
			// Winner's score + 1 because it hasn't actually been updated yet
			newTscore = winnerTS.Score + 1
			newCTscore = loserTS.Score

		case p_common.TeamCounterTerrorists:
			newTscore = loserTS.Score
			newCTscore = winnerTS.Score + 1

		default:
			// Probably match medic or something similar
			analyser.log.Info("No winner in this round ")
			analyser.isCancelled = true
		}
	default:
	}

	// common handling for two events
	if !analyser.isFirstParse {
		// if this round is not already handled
		// or no round has been played
		// sometimes, an update event can update score in round so it
		// should not be considered as round end
		if analyser.roundEnd == tick && (analyser.roundPlayed > analyser.lastRoundEndCalled) {
			// get team members
			gs := analyser.parser.GameState()
			participants := gs.Participants()
			winnerTeam := participants.TeamMembers(winnerTS.Team())
			loserTeam := participants.TeamMembers(loserTS.Team())

			analyser.handleSpecialRound(winnerTeam, loserTeam)
			analyser.handleClutchSituation(winnerTS.Team(), loserTS.Team(), tick)

			// update scores
			analyser.tScore = analyser.curValidRound.TScore
			analyser.ctScore = analyser.curValidRound.CTScore

			analyser.log.WithFields(logging.Fields{
				"t score":      analyser.tScore,
				"ct score":     analyser.ctScore,
				"tick":         tick,
				"winner":       common.GetSideString(winnerTS.Team()),
				"event":        eventString,
				"round number": analyser.roundPlayed,
			}).Info("Round is ended.")

			// update last round called
			analyser.lastRoundEndCalled = analyser.roundPlayed

			// check match is ended if there is no official end for
			// this round and also handle KAST as well
			if analyser.roundOffEnd <= 0 {
				// notify kast to alive players
				for _, alive := range analyser.ctAlive {
					analyser.kastPlayers[alive.SteamID] = true
				}

				for _, alive := range analyser.tAlive {
					analyser.kastPlayers[alive.SteamID] = true
				}

				analyser.handleKAST()
				analyser.checkMatchContinuity()
			}

			// reset round end for duplicate calls in same tick
			analyser.roundEnd = 0

		}

	} else { //first parsing

		if analyser.inRound {
			// check score update is valid
			if !analyser.isCancelled && !analyser.updateScore(newTscore, newCTscore, eventString) {
				analyser.log.WithFields(logging.Fields{
					"t score":  newTscore,
					"ct score": newCTscore,
					"tick":     tick,
					"winner":   common.GetSideString(winnerTS.Team()),
					"event":    eventString,
				}).Error("Invalid score update.Will cancelled.")
				analyser.isCancelled = true
			}

			// reset money set flag for each half end because
			// sometimes money is not set for a half
			if analyser.checkHalfBreak(newTscore, newCTscore) {
				analyser.isMoneySet = false
			}

			// set this round is finished
			analyser.inRound = false
			if !analyser.isCancelled {
				// analyser.RoundPlayed++

				analyser.log.WithFields(logging.Fields{
					"t score":         newTscore,
					"ct score":        newCTscore,
					"tick":            tick,
					"winner":          common.GetSideString(winnerTS.Team()),
					"event":           eventString,
					"winner team":     winnerTS.ClanName,
					"number of round": analyser.roundPlayed,
				}).Info("Round has been finished")

				analyser.roundEnd = tick

				newValidRound := common.RoundTuples{StartTick: analyser.roundStart, EndTick: analyser.roundEnd,
					TScore: analyser.tScore, CTScore: analyser.ctScore}

				analyser.validRounds[analyser.roundPlayed] = &newValidRound
				analyser.log.WithFields(logging.Fields{
					"start tick":   analyser.validRounds[analyser.roundPlayed].StartTick,
					"end tick":     analyser.validRounds[analyser.roundPlayed].EndTick,
					"round number": analyser.roundPlayed,
				}).Info("New round has been added to list")

			} else {
				analyser.log.WithFields(logging.Fields{
					"tick":  tick,
					"event": eventString,
				}).Error("An invalid round end.")
			}

		} else {
			analyser.log.WithFields(logging.Fields{
				"t score":  newTscore,
				"ct score": newCTscore,
				"tick":     tick,
				"event":    eventString,
			}).Error("Round has already end or not start already")
		}

		// check match is over
		analyser.checkMatchContinuity()
	}

}

// handleMatchStart handle match start event
func (analyser *Analyser) handleMatchStart(eventName string) {
	tick := analyser.getGameTick()

	// we already called match start for this tick
	// at the first match start itcan start at tick 0
	if analyser.lastMatchStartedCalled == tick && analyser.lastMatchStartedCalled != 0 {
		return
	}

	analyser.lastMatchStartedCalled = tick

	var teamT, teamCT []*p_common.Player
	var ok bool

	// check participant validity
	if teamT, teamCT, ok = analyser.checkParticipantValidity(); !ok {
		analyser.log.WithFields(logging.Fields{
			"tick":       analyser.getGameTick(),
			"event name": eventName,
			"ct number":  len(teamCT),
			"t number":   len(teamT),
		}).Error("Participant number is not expected for a match start.Aborted.")
		analyser.isCancelled = true
		return
	}

	// first parse of match
	if analyser.isFirstParse {
		if !analyser.checkMoneyValidity() {
			analyser.log.WithFields(logging.Fields{
				"tick":         tick,
				"played round": analyser.roundPlayed,
				"event name":   eventName,
			}).Error("Money has been invalid for starting round on match start")
			analyser.isCancelled = true
			return
		}

		analyser.roundStart = tick

		if !analyser.matchStarted {
			analyser.log.WithFields(logging.Fields{
				"tick":       analyser.getGameTick(),
				"event name": eventName,
			}).Info("A new match has been started")

			// reset match based variables
			analyser.resetMatchVars()

		} else {
			analyser.log.WithFields(logging.Fields{
				"tick":       analyser.getGameTick(),
				"event name": eventName,
			}).Info("Match has already started.Count as round start if possible.")
		}

		// reset round based variables as well
		analyser.resetRoundVars(teamT, teamCT, tick)

	} else {
		if !analyser.setRoundStart(tick) {
			return
		}

		// After set score and round number using setRoundStart, we do not want
		// to reset score and round number so that we are not calling rset match vars
		// because match start has been counting as round start on second parsing
		analyser.resetRoundVars(teamT, teamCT, tick)

		analyser.log.WithFields(logging.Fields{
			"tick": tick,
		}).Info("A match has been started.Count as round start.")
	}

}

// handlePlayerConnect handle player connection event
func (analyser *Analyser) handlePlayerConnect(e events.PlayerConnect) {
	NewPlayer := e.Player
	uid := NewPlayer.SteamID
	var NewPPlayer *common.PPlayer
	tick := analyser.getGameTick()

	// if non player is connecting, ignore it
	if !analyser.checkPlayerTeamValidity(NewPlayer) {
		analyser.log.WithFields(logging.Fields{
			"name": NewPlayer.Name,
			"team": NewPlayer.Team,
			"tick": tick,
		}).Info("A non-player is tring to connect.ignored. ")
		return
	}

	// reconnection case
	if disconnectedPTuple, ok := analyser.getDisconnectedPlayerByID(uid); ok {
		// create new player and append to the list
		NewPPlayer = disconnectedPTuple.Player
		// update parser player inside pplayer
		NewPPlayer.Player = NewPlayer

		analyser.log.WithFields(logging.Fields{
			"name":              disconnectedPTuple.Player.Name,
			"user id":           disconnectedPTuple.Player.SteamID,
			"connected tick":    tick,
			"disconnected tick": disconnectedPTuple.DisconnectedTick,
			"player team":       NewPlayer.Team,
		}).Info("Player has been reconnected: ")

	} else { //new connection
		if val, ok := analyser.getPlayerByID(uid, false); ok {
			analyser.log.WithFields(logging.Fields{
				"name":    val.Name,
				"user id": val.SteamID,
				"tick":    tick,
			}).Error("Player has already been connected: ")
			return
		}
		// create new player and append to the list
		NewPPlayer = common.NewPPlayer(NewPlayer)

		analyser.log.WithFields(logging.Fields{
			"name":        NewPlayer.Name,
			"user id":     uid,
			"tick":        tick,
			"player team": NewPlayer.Team,
		}).Info("Player has been connected: ")
	}

	if playerSide, ok := NewPPlayer.GetSide(); ok {
		analyser.players[uid] = NewPPlayer
		if analyser.tAlive != nil && analyser.ctAlive != nil {
			switch playerSide {
			case p_common.TeamTerrorists:
				analyser.tAlive[uid] = NewPPlayer
			case p_common.TeamCounterTerrorists:
				analyser.ctAlive[uid] = NewPPlayer
			}
		}

	}

	delete(analyser.disconnectedPlayers, uid)
}

// handlePlayerDisconnect handle player disconnection event
func (analyser *Analyser) handlePlayerDisconnect(e events.PlayerDisconnected) {
	currentPlayer := e.Player
	playerSide := currentPlayer.Team
	currentPLayerID := currentPlayer.SteamID
	tick := analyser.getGameTick()

	if currentPPlayer, ok := analyser.getPlayerByID(currentPLayerID, false); ok {
		analyser.log.WithFields(logging.Fields{
			"name":    currentPlayer.Name,
			"user id": currentPLayerID,
			"tick":    tick,
			"team":    currentPPlayer.Player.Team,
		}).Info("Player has been disconnected: ")

		// add player to disconnected player list
		disconnected := &common.DisconnectedTuple{DisconnectedTick: tick, Player: currentPPlayer}
		analyser.disconnectedPlayers[currentPLayerID] = disconnected
		// }

		// remove players from connected player list
		delete(analyser.players, currentPLayerID)
		// delete alive player as well
		analyser.deleteAlivePlayer(playerSide, currentPLayerID)

	} else {
		analyser.log.WithFields(logging.Fields{
			"name":    currentPlayer.Name,
			"user id": currentPLayerID,
			"tick":    tick,
		}).Error("Non-exist player has been disconnected: ")
	}
}

// handleTeamChange handle player team change
// for now, it is mainly use for update player pointer of reconnected player
func (analyser *Analyser) handleTeamChange(e events.PlayerTeamChange) {
	changedPlayer := e.Player
	oldTeam := e.OldTeam
	oldTeamState := e.OldTeamState
	newTeam := e.NewTeam
	// uid := changedPlayer.SteamID
	tick := analyser.getGameTick()

	if (oldTeam == p_common.TeamSpectators || oldTeam == p_common.TeamUnassigned) &&
		(newTeam == p_common.TeamTerrorists || newTeam == p_common.TeamCounterTerrorists) {
		analyser.log.WithFields(logging.Fields{
			"tick":     tick,
			"name":     changedPlayer.Name,
			"old team": oldTeam,
			"new team": newTeam,
		}).Info("Unactive player become playing player")
		analyser.handlePlayerConnect(events.PlayerConnect{Player: changedPlayer})
	} else if (newTeam == p_common.TeamSpectators || newTeam == p_common.TeamUnassigned) &&
		(oldTeam == p_common.TeamTerrorists || oldTeam == p_common.TeamCounterTerrorists) {
		analyser.log.WithFields(logging.Fields{
			"tick":     tick,
			"name":     changedPlayer.Name,
			"old team": oldTeam,
			"new team": newTeam,
		}).Info("Playing player become unactive")
		// deep copy unactive player from original player for disceonnection
		var unactivePlayer p_common.Player
		copier.Copy(&unactivePlayer, changedPlayer)
		// change team and team state to old one, so that there is no
		// problem for disconnection handler
		unactivePlayer.Team = oldTeam
		unactivePlayer.TeamState = oldTeamState
		// send uncative player to be disconnected
		analyser.handlePlayerDisconnect(events.PlayerDisconnected{Player: &unactivePlayer})
	}
}

// // handleGamePhaseChange handle when game phase has been changed
// // used for getting game ends
// func (analyser *Analyser) handleGamePhaseChange(e events.GamePhaseChanged) {
// 	newGamePhase := e.NewGamePhase
// 	if newGamePhase == common.GamePhaseGameEnded {
// 		analyser.log.WithFields(logging.Fields{
// 			"tick": analyser.getGameTick(),
// 		}).Info("Game has been ended")
// 		analyser.MatchEnded = true
// 		// first finish already started match
//
// 		analyser.log.WithFields(logging.Fields{
// 			"tick": analyser.getGameTick(),
// 		}).Info("Finished a valid match.")
// 		analyser.printPlayers()
//
// 	} else {
// 		analyser.log.WithFields(logging.Fields{
// 			"tick": analyser.getGameTick(),
// 		}).Info("Finished an invalid match.Aborted")
//
// 	}
// }

// handleRoundStart handle round start event
func (analyser *Analyser) handleRoundStart(e events.RoundStart) {
	// defer utils.RecoverPanic()
	tick := analyser.getGameTick()

	// check teams
	var teamT, teamCT []*p_common.Player
	// var teamOk bool

	teamT, teamCT, _ = analyser.checkParticipantValidity()

	// teamT, teamCT, teamOk = analyser.checkParticipantValidity()
	// if analyser.isFirstParse && !teamOk {
	// 	analyser.IsCancelled = true
	// 	log.WithFields(log.Fields{
	// 		"tick":         tick,
	// 		"played round": analyser.RoundPlayed,
	// 	}).Error("Round has been cancelled because of participant deficit")
	// 	return
	// }

	if analyser.isFirstParse {
		// check match is over
		analyser.checkMatchContinuity()

		// check money validity
		if !analyser.checkMoneyValidity() {
			analyser.log.WithFields(logging.Fields{
				"tick":         tick,
				"played round": analyser.roundPlayed,
				"player money": analyser.getAllPlayers()[0].Money,
			}).Error("Money has been invalid for round start")
			analyser.isCancelled = true
			return
		}

		// check match has already started and not yet finished
		if !analyser.checkMatchValidity() {
			analyser.log.WithFields(logging.Fields{
				"tick": tick,
			}).Error("Round start called outside the match.")
			analyser.isCancelled = true
			return
		}
		analyser.roundStart = tick

		if analyser.inRound {
			analyser.log.WithFields(logging.Fields{
				"tick": tick,
			}).Error("Round has already been started.Invalid round.")

		} else {
			analyser.log.WithFields(logging.Fields{
				"tick": tick,
			}).Info("New round has been started")
		}

	} else {
		if !analyser.setRoundStart(tick) || tick != analyser.roundStart {
			analyser.log.WithFields(logging.Fields{
				"tick":         tick,
				"round number": analyser.roundPlayed,
			}).Info("Invalid round start event")
			return
		}
		analyser.log.WithFields(logging.Fields{
			"tick":         tick,
			"round number": analyser.roundPlayed,
		}).Info("A new round has been started")
	}

	// reset round based variables
	analyser.resetRoundVars(teamT, teamCT, tick)

}

// handleRoundOfficiallyEnd handle round officially end event
func (analyser *Analyser) handleRoundOfficiallyEnd(e events.RoundEndOfficial) {
	tick := analyser.getGameTick()

	// second parsing
	if !analyser.isFirstParse {
		// check round official end is matching with previously collected
		if analyser.roundOffEnd != tick {
			return
		}
		analyser.log.WithFields(logging.Fields{
			"tick":         tick,
			"round number": analyser.roundPlayed,
		}).Info("Round has officially ended.")

		// notify kast to alive players
		for _, alive := range analyser.ctAlive {
			analyser.kastPlayers[alive.SteamID] = true
		}

		for _, alive := range analyser.tAlive {
			analyser.kastPlayers[alive.SteamID] = true
		}

		analyser.handleKAST()
		analyser.checkMatchContinuity()

		// reset roundoffend for duplicate calls
		analyser.roundOffEnd = 0

	} else {
		// check match has already started and not yet finished
		if !analyser.checkMatchValidity() {
			analyser.isCancelled = true
			analyser.log.WithFields(logging.Fields{
				"tick": tick,
			}).Error("Official round end called outside the match.")
			return
		}

		// if we are not in round (round ended)
		if !analyser.isCancelled && !analyser.inRound {
			analyser.log.WithFields(logging.Fields{
				"tick": tick,
				// roundplayed has already incremented by one
				"round": analyser.roundPlayed,
			}).Info("Round has officially ended")

			var ok bool
			// sometimes there is a wrong score update for ending round which can cause nil error
			if analyser.curValidRound, ok = analyser.validRounds[analyser.roundPlayed]; ok && (analyser.roundPlayed > 0) {
				// check this is the official end of correct round
				if !(analyser.roundStart == analyser.curValidRound.StartTick &&
					analyser.roundEnd == analyser.curValidRound.EndTick) {
					analyser.log.WithFields(logging.Fields{
						"tick": tick,
					}).Error("Round official end is not matching with round start and round end")
					return
				}
			} else {
				analyser.log.WithFields(logging.Fields{
					"tick": tick,
				}).Error("Round official end has been called before at least one round played")
				return
			}

			analyser.curValidRound.OfficialEndTick = tick

		} else {
			analyser.log.WithFields(logging.Fields{
				"tick": tick,
			}).Error("Round officially ended without proper round end")
			analyser.isCancelled = true
		}
	}

}
