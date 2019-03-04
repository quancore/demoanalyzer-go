// Package oldAnalyser package to handle events and update
// player statistics
package oldAnalyser

import (
	"bytes"
	"encoding/gob"
	"io"
	"strconv"
	"time"

	utils "github.com/Logitech/demoparser-go/common"
	player "github.com/Logitech/demoparser-go/player"
	"github.com/gogo/protobuf/proto"
	dem "github.com/markus-wa/demoinfocs-golang"
	common "github.com/markus-wa/demoinfocs-golang/common"
	events "github.com/markus-wa/demoinfocs-golang/events"
	"github.com/markus-wa/demoinfocs-golang/msg"
	log "github.com/sirupsen/logrus"
)

// csgo match constants
const (
	maxRounds           = 30
	normalTimeWinRounds = maxRounds/2 + 1
)

// The parser used in is a program with state so each tick state
// of parser is changing.Because we delay event processing, we need to
// store this event in a struct so that we will not depend on the state of
// parser.

// ************* event structs **************
// base struct is used for events that have data not depend on
// state of parser or not used fields.
type baseEvent struct {
	tick  int
	event interface{}
}

// this events have several data depend on state of parser and this
// data is used in program
type flashEvent struct {
	baseEvent
	// attacker      *player.PPlayer
	// defender      *player.PPlayer
	flashDuration time.Duration
}

// ******************************************
// tuple struct to store kill event with tick
type killTuples struct {
	tick   int
	player *player.PPlayer
}

// tuple struct to store events in a round with tick
type roundTuple struct {
	tick int
	// store events per round beforehand and then dispath it
	collectedEvents []interface{}
}

// NewAnalyser constructer for getting an analyser
func newRoundTuple(tick int) *roundTuple {
	collectedEvents := make([]interface{}, 0)
	return &roundTuple{tick: tick, collectedEvents: collectedEvents}

}

// tuple struct to store kill event with tick
type disconnectedTuple struct {
	disconnectedTick int
	reconnectedTick  int //reconnection of disconnected playre if exist
	player           *player.PPlayer
}

//MatchVars struct to store match related vars.
// mainly seperated for serialization purposes
type MatchVars struct {
	// **********match related variables *********
	// main parserrer to parse demofile
	Parser *dem.Parser
	// header related information
	MapName string
	// map container to store currently connected Players
	Players map[int64]*player.PPlayer
	// map container to store currently disconnected players
	// it is usefull for reconnection of a player
	DisconnectedPlayers map[int64]*disconnectedTuple
	// scores
	Tscore      int
	CTscore     int
	RoundPlayed int
	// number of rounds will be played as overtime
	NumOvertime int
	// flags
	// match started flag
	MatchStarted bool
	// match started flag
	MatchEnded bool
	// flag indicating overtime is playing
	IsOvertime bool
	// flag indicating score is swapped
	ScoreSwapped bool
}

// Analyser main struct to analyse events
type Analyser struct {
	MatchVars
	// **********round related variables*******
	// kill map to store pointers to killed players for each player in this round
	// killer id : {killed tick, pointer to victim}
	// // TODO: check consistency
	killedPlayers map[int64][]*killTuples
	// map container to store alive ct players for each round
	ctAlive map[int64]*player.PPlayer
	// map container to store alive t players for each round
	tAlive map[int64]*player.PPlayer
	// keep round based player did kast
	kastPlayers map[int64]bool
	//store player get disconnected during a round
	// usefull to dispatch events belongs to this players
	// disconnectedLastRound map[int64]*disconnectedTuple

	// store last round event with tick number
	lastRound *roundTuple
	// player pointer possible to make a clutch
	clutchPLayer *player.PPlayer
	// player currently defusing the bomb
	defuser *player.PPlayer
	// flag for bomb defused
	isBombDefused bool
	// flag for bomb planted
	isBombPlanted bool
	// flag for bomb planted
	isBombDefusing bool
	// clutch situation flag
	isPossibleCLutch bool
	// variables to check validity of a round
	//Truth value for whether we're currently in a round
	inRound bool
	// Truth value whether this round is valid
	isCancelled bool
	// Truth value whether a player has been hurt
	isPlayerHurt bool
	// ***********Serialization**************
	// buffer for serialization
	b bytes.Buffer
	// flag indicating whether there is a serialized match
	matchEncoded bool
}

// ######## public interface #######################

// NewAnalyser constructer for getting an analyser
func NewAnalyser(demostream io.Reader) *Analyser {
	log.Info("Analyser has been created")
	// parser := dem.NewParser(demostream)
	// Configure parsing of ConVar net-message (id=6)
	cfg := dem.DefaultParserConfig
	cfg.AdditionalNetMessageCreators = map[int]dem.NetMessageCreator{
		6: func() proto.Message {
			return new(msg.CNETMsg_SetConVar)
		},
	}

	parser := dem.NewParserWithConfig(demostream, cfg)

	analyser := &Analyser{MatchVars: MatchVars{Parser: parser}}

	// initilize maps for further use
	analyser.Players = make(map[int64]*player.PPlayer)
	analyser.DisconnectedPlayers = make(map[int64]*disconnectedTuple)
	analyser.NumOvertime = 6
	// register net message handlers
	analyser.registerNetMessageHandlers()

	return analyser

}

// HandleHeader handle header information an initilize related variables
func (analyser *Analyser) HandleHeader() {
	log.Info("Parsing header of demo file")
	// Parse header
	header, err := analyser.Parser.ParseHeader()
	utils.CheckError(err)
	analyser.MapName = header.MapName

	log.WithFields(log.Fields{
		"server name": header.ServerName,
		"client name": header.ClientName,
		"map name":    header.MapName,
	}).Info("Several fields of header: ")
}

// registerNetMessageHandlers register net message handlers
func (analyser *Analyser) registerNetMessageHandlers() {
	// Register handler for ConVar updates
	analyser.Parser.RegisterNetMessageHandler(func(m *msg.CNETMsg_SetConVar) {
		for _, cvar := range m.Convars.Cvars {
			if cvar.Name == "mp_overtime_maxrounds" {
				analyser.NumOvertime, _ = strconv.Atoi(cvar.Value)
			}
			log.WithFields(log.Fields{
				"cvar name":  cvar.Name,
				"cvar value": cvar.Value,
			}).Info("Cvars ")
		}
	})
}

// RegisterEventHandlers register handlers for each needed events
func (analyser *Analyser) RegisterEventHandlers() {
	// defer utils.RecoverPanic()
	// Register handler on match start
	analyser.Parser.RegisterEventHandler(func(e events.MatchStart) { analyser.handleMatchStart() })

	// Register handler on match start.Sometimes, match start event is not called
	analyser.Parser.RegisterEventHandler(func(e events.MatchStartedChanged) { analyser.handleMatchStart() })

	// Register handler on game phase changed. Useful for match end
	analyser.Parser.RegisterEventHandler(func(e events.GamePhaseChanged) {

		newGamePhase := e.NewGamePhase
		if newGamePhase == common.GamePhaseGameEnded {
			log.WithFields(log.Fields{
				"tick": analyser.getGameTick(),
			}).Info("Game has been ended")
			analyser.MatchEnded = true
			// first finish already started match
			if analyser.checkFinishedMatchValidity() {
				log.WithFields(log.Fields{
					"tick": analyser.getGameTick(),
				}).Info("Finished a valid match.")
				analyser.printPlayers()

			} else {
				log.WithFields(log.Fields{
					"tick": analyser.getGameTick(),
				}).Info("Finished an invalid match match.Aborted")
			}
		}
	})

	// Register handler on round start
	analyser.Parser.RegisterEventHandler(func(e events.RoundStart) {
		tick := analyser.getGameTick()

		// check match is over
		analyser.checkMatchContinuity()

		// check match has already started and not yet finished
		if !analyser.checkMatchValidity() {
			analyser.isCancelled = true
			// log.Error("invalid event")
			return
		}

		// check teams
		teamT, teamCT, teamOk := analyser.checkParticipantValidity()

		if !teamOk {
			analyser.isCancelled = true
			return
		}

		// we are already in a round, do not dispatch events
		if analyser.inRound {
			canacelledTick := analyser.lastRound.tick
			log.WithFields(log.Fields{
				"tick": canacelledTick,
			}).Error("Round has already been started.Cancelled round.")

		} else {
			log.WithFields(log.Fields{
				"tick": tick,
			}).Info("New round has been started")

		}
		// reset round based variables
		analyser.resetRoundVars(teamT, teamCT, tick)

	})
	// Register handler on player connected
	analyser.Parser.RegisterEventHandler(func(e events.PlayerConnect) {

		// check match has already started and not yet finished
		if !analyser.checkMatchValidity() {
			analyser.isCancelled = true
			// log.Error("invalid event")
			return
		}

		NewPlayer := e.Player
		uid := NewPlayer.SteamID
		var NewPPlayer *player.PPlayer

		// reconnection case
		if val, ok := analyser.getDisconnectedPlayerByID(uid); ok {
			log.WithFields(log.Fields{
				"name":              val.player.Name,
				"user id":           val.player.SteamID,
				"tick":              analyser.getGameTick(),
				"disconnected tick": val.disconnectedTick,
			}).Info("Player is reconnecting: ")
			// create new player and append to the list
			NewPPlayer = val.player

		} else { //new connection
			if val, ok := analyser.getPlayerByID(uid, false); ok {
				log.WithFields(log.Fields{
					"name":    val.Name,
					"user id": val.SteamID,
					"tick":    analyser.getGameTick(),
				}).Error("Player has already been connected: ")
				return
			}
			// create new player and append to the list
			NewPPlayer = player.NewPPlayer(NewPlayer)
		}
		// append new value to mapNewPPlayer := player.NewPPlayer(NewPlayer)

		if playerSide, ok := NewPPlayer.GetSide(); ok {
			analyser.Players[uid] = NewPPlayer

			switch playerSide {
			case "T":
				analyser.tAlive[uid] = NewPPlayer
			case "CT":
				analyser.ctAlive[uid] = NewPPlayer
			}

		}

		log.WithFields(log.Fields{
			"name":    NewPlayer.Name,
			"user id": uid,
			"tick":    analyser.getGameTick(),
		}).Info("Player has been connected: ")

		// analyser.players = append(analyser.players, NewPPlayer)
	})
	// game handling events
	// Register handler on player disconnected
	analyser.Parser.RegisterEventHandler(func(e events.PlayerDisconnected) {
		// check match has already started and not yer finished
		// if !analyser.checkMatchValidity() {
		// 	analyser.isCancelled = true
		// 	// log.Error("invalid event")
		// 	return
		// }
		currentPlayer := e.Player
		playerSide := analyser.getSideString(currentPlayer.Team)
		currentPLayerID := currentPlayer.SteamID
		tick := analyser.getGameTick()

		if currentPPlayer, ok := analyser.getPlayerByID(currentPLayerID, false); ok {
			log.WithFields(log.Fields{
				"name":    currentPlayer.Name,
				"user id": currentPLayerID,
				"tick":    tick,
			}).Info("Player has been disconnected: ")
			// first remove players from connected player and alive player list
			delete(analyser.Players, currentPLayerID)
			analyser.deleteAlivePlayer(playerSide, currentPLayerID)
			// then add player to disconnected player list
			disconnected := &disconnectedTuple{disconnectedTick: tick, player: currentPPlayer}
			analyser.DisconnectedPlayers[currentPLayerID] = disconnected
			// last disconnected player to current round disconnected players
			// analyser.disconnectedLastRound[currentPLayerID] = disconnected

		} else {
			log.WithFields(log.Fields{
				"name":    currentPlayer.Name,
				"user id": currentPLayerID,
				"tick":    tick,
			}).Error("Non-exist player has been disconnected: ")
		}
	})

	// Register handler on score updated event
	analyser.Parser.RegisterEventHandler(func(e events.ScoreUpdated) {

		// gs := analyser.parser.GameState()
		tick := analyser.getGameTick()
		// get team states
		winnerTS := e.TeamState
		loserTS := e.TeamState.Opponent
		// get scores
		winnerScore := winnerTS.Score
		loserScore := loserTS.Score
		updatedScore := e.NewScore
		var newTscore, newCTscore int

		switch winnerTS.Team() {
		case common.TeamTerrorists:
			// Winner's score + 1 because it hasn't actually been updated yet
			newTscore = winnerScore
			newCTscore = loserScore

		case common.TeamCounterTerrorists:
			newTscore = loserScore
			newCTscore = winnerScore

		default:
			// Probably match medic or something similar
			log.Info("No winner in this round ")
			analyser.isCancelled = true
		}

		if !analyser.isPlayerHurt || !analyser.updateScore(newTscore, newCTscore) {
			log.WithFields(log.Fields{
				"tick": analyser.getGameTick(),
			}).Error("Invalid score update or invalid round")
			analyser.isCancelled = true
		}

		if analyser.inRound {
			analyser.inRound = false

			// check match has already started and not yet finished
			if !analyser.checkMatchValidity() {
				analyser.isCancelled = true
				// log.Error("invalid event")
				return
			}

			// // check teams
			// if _, _, teamOk := analyser.checkParticipantValidity(); !teamOk {
			// 	analyser.isCancelled = true
			// 	return
			// }

			if !analyser.isCancelled {
				analyser.dispatchPlayerEvents()

				log.WithFields(log.Fields{
					"winner score":      winnerScore,
					"loser score":       loserScore,
					"updated score":     updatedScore,
					"updated team name": winnerTS.ClanName,
					"tick":              analyser.getGameTick(),
				}).Info("Round end with score updated event")

				clutchPLayer := analyser.clutchPLayer
				analyser.RoundPlayed++

				// get team members
				gs := analyser.Parser.GameState()
				participants := gs.Participants()
				winnerTeam := participants.TeamMembers(winnerTS.Team())
				loserTeam := participants.TeamMembers(loserTS.Team())

				analyser.handleSpecialRound(winnerTeam, loserTeam)

				// check whether player did a clutch
				// any kind of 1 to n winning count as clutch
				// we are checking clutch player is not dead as well
				if analyser.isPossibleCLutch && clutchPLayer != nil {
					clutchPlayerSide := clutchPLayer.Team
					if clutchPlayerSide == winnerTS.Team() {
						switch clutchPlayerSide {
						case common.TeamTerrorists:
							opponentAliveNum := len(analyser.ctAlive)
							if opponentAliveNum == 0 {
								clutchPLayer.NotifyClutchWon()
							}
						case common.TeamCounterTerrorists:
							opponentAliveNum := len(analyser.tAlive)
							if opponentAliveNum == 0 {
								clutchPLayer.NotifyClutchWon()
							}
						default:
							log.WithFields(log.Fields{
								"name":    clutchPLayer.Name,
								"user id": clutchPLayer.SteamID,
								"tick":    tick,
							}).Error("Player has no side for clutch situation: ")
						}
					}
				}
			} else {
				log.WithFields(log.Fields{
					"tick": tick,
				}).Error("Round is invalid.Cancelled round.")
			}
		} else {
			log.WithFields(log.Fields{
				"winner score":      winnerScore,
				"loser score":       loserScore,
				"updated score":     updatedScore,
				"updated team name": winnerTS.ClanName,
				"tick":              analyser.getGameTick(),
			}).Error("Round end has already been called or not started round has been end")
		}

		return

	})

	// Register handler on round end event
	analyser.Parser.RegisterEventHandler(func(e events.RoundEnd) {

		// gs := analyser.parser.GameState()
		tick := analyser.getGameTick()
		// get team states
		winnerTS := e.WinnerState
		loserTS := e.LoserState

		// get team members
		gs := analyser.Parser.GameState()
		participants := gs.Participants()
		winnerTeam := participants.TeamMembers(winnerTS.Team())
		loserTeam := participants.TeamMembers(loserTS.Team())

		var newTscore, newCTscore int

		if analyser.inRound {
			analyser.inRound = false

			// check match has already started and not yet finished
			if !analyser.checkMatchValidity() {
				analyser.isCancelled = true
				// log.Error("invalid event")
				return
			}

			if !analyser.checkFinishedRoundValidity(e) {
				log.WithFields(log.Fields{
					"tick": tick,
				}).Error("Round end because of invalid reason")
				analyser.isCancelled = true
				// log.Error("invalid event")
				return
			}

			// // check teams
			// if _, _, teamOk := analyser.checkParticipantValidity(); !teamOk {
			// 	analyser.isCancelled = true
			// 	return
			// }

			switch winnerTS.Team() {
			case common.TeamTerrorists:
				// Winner's score + 1 because it hasn't actually been updated yet
				newTscore = winnerTS.Score + 1
				newCTscore = loserTS.Score

			case common.TeamCounterTerrorists:
				newTscore = loserTS.Score
				newCTscore = winnerTS.Score + 1

			default:
				// Probably match medic or something similar
				log.Info("No winner in this round ")
				analyser.isCancelled = true
			}

			// no player hurts and invalid score update
			if analyser.isPlayerHurt {
				if !analyser.updateScore(newTscore, newCTscore) {
					log.WithFields(log.Fields{
						"t score":  newTscore,
						"ct score": newCTscore,
						"tick":     tick,
						"winner":   analyser.getSideString(winnerTS.Team()),
					}).Error("Invalid score update on round.Will cancelled.")
					analyser.isCancelled = true
				}
			} else {
				log.WithFields(log.Fields{
					"tick": tick,
				}).Error("No one hurted in this round. Will cancelled.")
				analyser.isCancelled = true
			}

			if !analyser.isCancelled {

				analyser.dispatchPlayerEvents()

				clutchPLayer := analyser.clutchPLayer
				analyser.RoundPlayed++
				analyser.handleSpecialRound(winnerTeam, loserTeam)

				log.WithFields(log.Fields{
					"t score":  newTscore,
					"ct score": newCTscore,
					"tick":     tick,
					"winner":   analyser.getSideString(winnerTS.Team()),
				}).Info("Round has been finished")

				// check whether player did a clutch
				// any kind of 1 to n winning count as clutch
				if analyser.isPossibleCLutch && clutchPLayer != nil {
					clutchPlayerSide := clutchPLayer.Team
					if clutchPlayerSide == winnerTS.Team() {
						switch clutchPlayerSide {
						case common.TeamTerrorists:
							opponentAliveNum := len(analyser.ctAlive)
							if opponentAliveNum == 0 {
								clutchPLayer.NotifyClutchWon()
							}
						case common.TeamCounterTerrorists:
							opponentAliveNum := len(analyser.tAlive)
							if opponentAliveNum == 0 {
								clutchPLayer.NotifyClutchWon()
							}
						default:
							log.WithFields(log.Fields{
								"name":    clutchPLayer.Name,
								"user id": clutchPLayer.SteamID,
								"tick":    tick,
							}).Error("Player has no side for clutch situation: ")
						}
					}
				}
			} else {
				log.WithFields(log.Fields{
					"tick": tick,
				}).Error("Round is invalid.Cancelled round.")
			}
		} else {
			log.WithFields(log.Fields{
				"t score":  newTscore,
				"ct score": newCTscore,
				"tick":     tick,
			}).Error("Round end has already been called or not started round has been end")
		}

	})

	// Register handler on round end official event
	analyser.Parser.RegisterEventHandler(func(e events.RoundEndOfficial) {
		// check match has already started and not yet finished
		if !analyser.checkMatchValidity() {
			analyser.isCancelled = true
			log.Error("Official round end called outside the match")
			return
		}
		// if we are not in round (round ended)
		if !analyser.inRound {

			log.WithFields(log.Fields{
				"tick": analyser.getGameTick(),
			}).Info("Round has officially ended")
			// notify kast to alive players
			for _, alive := range analyser.ctAlive {
				analyser.kastPlayers[alive.SteamID] = true
			}

			for _, alive := range analyser.tAlive {
				analyser.kastPlayers[alive.SteamID] = true
			}

			analyser.handleKAST()

		} else {
			log.WithFields(log.Fields{
				"tick": analyser.getGameTick(),
			}).Error("Round officiallly ended without proper round end")
			analyser.isCancelled = true
			// analyser.inRound = false
		}
	})

	// Register handler on round end official event
	analyser.Parser.RegisterEventHandler(func(e events.TeamSideSwitch) {
		// switch scores
		// first swap scores if needed
		// first swap scores if needed
		mpOvertimeMaxrounds := analyser.NumOvertime
		nOvertimeRounds := analyser.RoundPlayed - maxRounds
		if analyser.RoundPlayed == 15 || analyser.RoundPlayed == maxRounds {
			if !analyser.ScoreSwapped {
				log.Info("Score has been swapped with team switch")
				analyser.Tscore, analyser.CTscore = analyser.CTscore, analyser.Tscore
				analyser.ScoreSwapped = true
			}
		} else if nOvertimeRounds > 0 && nOvertimeRounds%mpOvertimeMaxrounds == 0 {
			if !analyser.ScoreSwapped {
				log.Info("Score has been swapped with team switch")
				analyser.Tscore, analyser.CTscore = analyser.CTscore, analyser.Tscore
				analyser.ScoreSwapped = true
			}
		} else { //once score swapped and after swap situation we reset flag
			analyser.ScoreSwapped = false
		}
	})

	// Register handler on round end official event
	analyser.Parser.RegisterEventHandler(func(e events.IsWarmupPeriodChanged) {
		log.WithFields(log.Fields{
			"tick":        analyser.getGameTick(),
			"old warm up": e.OldIsWarmupPeriod,
			"new warm up": e.NewIsWarmupPeriod,
		}).Info("Warm up period")
	})

	analyser.registerFakeEventHandlers()
}

// Parse parse demofile
func (analyser *Analyser) Parse() {
	for hasMoreFrames, err := true, error(nil); hasMoreFrames; hasMoreFrames, err = analyser.Parser.ParseNextFrame() {
		utils.CheckError(err)
	}

	// err := analyser.Parser.ParseToEnd()
	// utils.CheckError(err)
}

// #################################################
// ######## Initilizers and reset functions##########

// initilizeRoundMaps initilize round maps with empty maps
func (analyser *Analyser) initilizeRoundMaps(teamT, teamCT []*common.Player, tick int) {
	analyser.resetAlivePlayers(teamT, teamCT)
	analyser.killedPlayers = make(map[int64][]*killTuples)
	analyser.lastRound = newRoundTuple(tick)
	analyser.kastPlayers = make(map[int64]bool)
}

// resetRoundVars reset round based variables
func (analyser *Analyser) resetRoundVars(teamT, teamCT []*common.Player, tick int) {
	analyser.initilizeRoundMaps(teamT, teamCT, tick)
	analyser.isBombPlanted = false
	analyser.isBombDefusing = false
	analyser.isPossibleCLutch = false
	analyser.defuser = nil
	analyser.clutchPLayer = nil
	analyser.inRound = true
	// will cancelled if true
	analyser.isCancelled = false
	analyser.isPlayerHurt = false

}

// resetRoundVars reset match based variables
func (analyser *Analyser) resetMatchVars(teamT, teamCT []*common.Player, tick int) {
	analyser.resetRoundVars(teamT, teamCT, tick)
	analyser.RoundPlayed = 0
	analyser.CTscore = 0
	analyser.Tscore = 0
	analyser.resetPlayerStates()
	analyser.MatchStarted = true
	analyser.MatchEnded = false
	analyser.ScoreSwapped = false

}

// ##########################################
// ########## Internal getters ##############

// getGameTick get tick of game
func (analyser *Analyser) getGameTick() int {
	tick := analyser.Parser.GameState().IngameTick()
	if tick < 0 {
		log.WithFields(log.Fields{
			"tick": tick,
		}).Error("Negative tick number")
	}
	return tick
}

// getPlayerByID get the pointer to player by player id
func (analyser *Analyser) getPlayerByID(uid int64, allPlayer bool) (*player.PPlayer, bool) {
	var player *player.PPlayer
	var ok bool

	if player, ok = analyser.Players[uid]; !ok {
		if playerDisconnected, discok := analyser.DisconnectedPlayers[uid]; discok {
			if allPlayer {
				return playerDisconnected.player, discok
			}
		}
	}

	return player, ok
}

// getPlayerByID get the pointer to disconnected player by player id
func (analyser *Analyser) getDisconnectedPlayerByID(uid int64) (*disconnectedTuple, bool) {

	player, ok := analyser.DisconnectedPlayers[uid]

	return player, ok
}

// getPlayerId get player id from players sent in an event
func (analyser *Analyser) getPlayerID(victim, killer *common.Player, eventName string) (int64, int64, bool, bool) {
	var victimID, killerID int64
	var victimOK, killerOK bool

	if killer == nil || victim == nil {
		log.WithFields(log.Fields{
			"event":  eventName,
			"tick":   analyser.getGameTick(),
			"victim": victim,
			"killer": killer,
		}).Error("Victim or killer is nill for event: ")
	}

	if victim != nil {
		victimID = victim.SteamID
		victimOK = true
	}
	if killer != nil {
		killerID = killer.SteamID
		killerOK = true
	}

	return victimID, killerID, victimOK, killerOK
}

// getSideString get side string like T or CT using team pointer
func (analyser *Analyser) getSideString(playerSide common.Team) string {
	switch playerSide {
	case common.TeamTerrorists:
		return "T"
	case common.TeamCounterTerrorists:
		return "CT"
	}

	return ""
}

// #############################
// ######## Internal checkers###
// checkPlayerValidity check whether a player is connected during
// event collection
func (analyser *Analyser) checkPlayerValidity(e interface{}) bool {
	switch e.(type) {
	case events.Kill:
		killEvent := e.(events.Kill)
		// get ids
		victimID, killerID, victimOK, killerOK := analyser.getPlayerID(killEvent.Victim, killEvent.Killer, "Kill")

		// check if victim and attacker exist in the event
		if !victimOK || !killerOK {
			return false
		}
		// get player pointers
		_, _, victimOK, killerOK = analyser.checkEventValidity(victimID, killerID, "Kill", false)

		// check if victim and attacker exist
		if !victimOK || !killerOK {
			return false
		}

		if killEvent.Assister != nil {
			if _, assisterOK := analyser.getPlayerByID(killEvent.Assister.SteamID, false); !assisterOK {
				return false
			}
		}

	case events.PlayerHurt:
		playerHurtEvent := e.(events.PlayerHurt)

		// get ids
		victimID, attackerID, victimOK, killerOK := analyser.getPlayerID(playerHurtEvent.Player, playerHurtEvent.Attacker, "playerHurt")

		// check if victim and attacker exist in the event
		if !victimOK || !killerOK {
			return false
		}
		// get player pointers
		_, _, victimOK, attackerOK := analyser.checkEventValidity(victimID, attackerID, "playerHurt", false)

		if !victimOK || !attackerOK {
			return false
		}

	case events.WeaponFire:
		weaponFireEvent := e.(events.WeaponFire)
		if _, ok := analyser.getPlayerByID(weaponFireEvent.Shooter.SteamID, false); !ok {
			return false
		}
	case events.BombDefuseStart:
		bombDefuseStartEvent := e.(events.BombDefuseStart)

		if _, ok := analyser.getPlayerByID(bombDefuseStartEvent.Player.SteamID, false); !ok {
			return false
		}

	case events.BombDefused:
		bombDefusedEvent := e.(events.BombDefused)

		if _, ok := analyser.getPlayerByID(bombDefusedEvent.Player.SteamID, false); !ok {
			return false
		}

	case events.BombPlanted:
		bombPlantedEvent := e.(events.BombPlanted)
		if _, ok := analyser.getPlayerByID(bombPlantedEvent.Player.SteamID, false); !ok {
			return false
		}

	case events.PlayerFlashed:

		playerFlashedEvent := e.(events.PlayerFlashed)

		// get ids
		victimID, killerID, victimOK, killerOK := analyser.getPlayerID(playerFlashedEvent.Player, playerFlashedEvent.Attacker, "Flash")

		// check if victim and attacker exist in the event
		if !victimOK || !killerOK || playerFlashedEvent.FlashDuration() <= 0 {
			return false
		}
		// get player pointers
		_, _, victimOK, killerOK = analyser.checkEventValidity(victimID, killerID, "Flash", false)

		// check if victim and attacker exist
		if !victimOK || !killerOK {
			return false
		}
	}

	return true
}

// checkEventValidity check validity of ids of given attacker and victim IDs
// return player pointers
func (analyser *Analyser) checkEventValidity(victimID, killerID int64, eventName string, allPlayer bool) (*player.PPlayer, *player.PPlayer, bool, bool) {
	var victimOK, killerOK bool
	var victimP, killerP *player.PPlayer

	// get player pointers
	victimP, victimOK = analyser.getPlayerByID(victimID, allPlayer)
	killerP, killerOK = analyser.getPlayerByID(killerID, allPlayer)

	// check if victim and attacker exist
	if !victimOK || !killerOK {
		log.WithFields(log.Fields{
			"event":       eventName,
			"tick":        analyser.getGameTick(),
			"victim ok":   victimOK,
			"attacker ok": killerOK,
		}).Error("Victim or attacker is undefined in map for an event: ")

		return nil, nil, victimOK, killerOK
	}

	return victimP, killerP, victimOK, killerOK

}

// checkMatchValidity return true value if match started and not yet finished
func (analyser *Analyser) checkMatchValidity() bool {
	tick := analyser.getGameTick()

	if analyser.MatchEnded || !analyser.MatchStarted || tick < 0 {
		return false
	}
	return true
}

// checkFinishedMatchValidity return true value if finished match is a valid one
func (analyser *Analyser) checkFinishedMatchValidity() bool {
	if analyser.RoundPlayed >= 15 {
		return true
	}

	return false
}

// checkFinishedMatchValidity return true value if finished match is a valid one
func (analyser *Analyser) checkFinishedRoundValidity(e events.RoundEnd) bool {
	reason := e.Reason
	if reason == events.RoundEndReasonCTSurrender || reason == events.RoundEndReasonDraw {
		return false
	}
	return true
}

// checkMatchContinuity check whether match is continuing with overtime
func (analyser *Analyser) checkMatchContinuity() bool {
	ctScore := analyser.CTscore
	tScore := analyser.Tscore
	mpOvertimeMaxrounds := analyser.NumOvertime

	nOvertimeRounds := ctScore + tScore - maxRounds
	if ((ctScore == normalTimeWinRounds) != (tScore == normalTimeWinRounds)) || nOvertimeRounds >= 0 {
		// a team won in normal time or at least 30 rounds have been played
		absDiff := utils.Abs(ctScore - tScore)
		x := nOvertimeRounds % mpOvertimeMaxrounds
		nRoundsOfHalf := mpOvertimeMaxrounds / 2
		if nOvertimeRounds < 0 || ((x == 0 && absDiff == 2) || (x > nRoundsOfHalf && absDiff >= nRoundsOfHalf)) {
			// match over
			// 	if ctScore > tScore {
			// 		fmt.Println("team A won")
			// 	} else if tScore > ctScore {
			// 		fmt.Println("team B won")
			// 	}
			// 	fmt.Println("it's a draw")
			// }
			if !analyser.MatchEnded {
				log.WithFields(log.Fields{
					"total round played": analyser.RoundPlayed,
					"tick":               analyser.getGameTick(),
					"team terrorist":     tScore,
					"team ct terrorist":  ctScore,
				}).Info("Match is over. ")
				analyser.printPlayers()
			}
			analyser.IsOvertime = false
			analyser.MatchEnded = true
		}
		analyser.IsOvertime = true
	}

	return analyser.IsOvertime
}

// checkClutchSituation check alive players to a clutch situation
func (analyser *Analyser) checkClutchSituation() {
	countALiveT := len(analyser.tAlive)
	countALiveCT := len(analyser.ctAlive)

	if !analyser.isPossibleCLutch {
		// possible clutch for ct
		if countALiveT > 1 && countALiveCT == 1 {
			analyser.isPossibleCLutch = true
			for _, playerPtr := range analyser.ctAlive {
				analyser.clutchPLayer = playerPtr
				log.WithFields(log.Fields{
					"name": playerPtr.Name,
				}).Error("Possible clutch player ")
			}

		} else if countALiveCT > 1 && countALiveT == 1 { // possible clutch for t
			analyser.isPossibleCLutch = true
			for _, playerPtr := range analyser.tAlive {
				analyser.clutchPLayer = playerPtr
				log.WithFields(log.Fields{
					"name": playerPtr.Name,
				}).Error("Possible clutch player ")
			}
		}
	}

}

// checkTeamSideValidity check validity of players with respest to their teams
// return player pointers
func (analyser *Analyser) checkTeamSideValidity(victim, killer *player.PPlayer) (string, string, bool) {
	// get side of players
	victimSide, vSideOK := victim.GetSide()
	killerSide, KSideOK := killer.GetSide()

	if !vSideOK || !KSideOK {
		log.WithFields(log.Fields{
			"victim side":   victimSide,
			"attacker side": killerSide,
			"tick":          analyser.getGameTick(),
		}).Error("Victim or attacker has no side: ")

		return victimSide, killerSide, false
	} else if victimSide == killerSide {
		log.WithFields(log.Fields{
			"victim side":   victimSide,
			"attacker side": killerSide,
			"tick":          analyser.getGameTick(),
		}).Error("Victim and attacker is the same side: ")
		return victimSide, killerSide, false
	}

	return victimSide, killerSide, true

}
func (analyser *Analyser) checkParticipantValidity() ([]*common.Player, []*common.Player, bool) {
	// first get players
	nTerrorists, nCTs := 5, 5
	gs := analyser.Parser.GameState()
	participants := gs.Participants()
	teamTerrorist := participants.TeamMembers(common.TeamTerrorists)
	teamCT := participants.TeamMembers(common.TeamCounterTerrorists)

	// all := participants.All()
	// players := participants.Playing()

	// check participants number etc
	if nTerrorists != len(teamTerrorist) || nCTs != len(teamCT) {
		// We know there should be 5 terrorists at match start in the default demo
		return nil, nil, false
	}

	return teamTerrorist, teamCT, true
}

// ######## Object modifiers ###########

// deleteAlivePlayer remove alive player from alive container
func (analyser *Analyser) deleteAlivePlayer(side string, uid int64) bool {
	if side == "T" {
		delete(analyser.tAlive, uid)
		// after deleteion check clutch situation
		analyser.checkClutchSituation()
		return true
	} else if side == "CT" {
		delete(analyser.ctAlive, uid)
		// after deletion check clutch situation
		analyser.checkClutchSituation()
		return true

	} else {
		log.WithFields(log.Fields{
			"user id": uid,
		}).Error("Player has no side: ")
		return false
	}
}

// resetAlivePlayers reset alive players per round
func (analyser *Analyser) resetAlivePlayers(teamT, teamCT []*common.Player) {
	analyser.ctAlive = make(map[int64]*player.PPlayer)
	analyser.tAlive = make(map[int64]*player.PPlayer)

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
			analyser.tAlive[uid] = NewPPlayer
		}

	}

	// for each c terorist
	for _, currPlayer := range teamCT {
		var NewPPlayer *player.PPlayer
		var ok bool
		uid := currPlayer.SteamID
		if NewPPlayer, ok = analyser.getPlayerByID(uid, true); !ok {
			NewPPlayer = player.NewPPlayer(currPlayer)
			analyser.Players[uid] = NewPPlayer
		}
		if _, ok = NewPPlayer.GetSide(); ok {
			analyser.ctAlive[uid] = NewPPlayer
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

// fakeEventHandler fake handler to store events
func (analyser *Analyser) fakeEventHandler(regEvent interface{}) {

	// check match has already started and not yet finished
	if !analyser.checkMatchValidity() {
		// log.Error("invalid event")
		return
	}
	// check event is valid in terms of players
	if !analyser.checkPlayerValidity(regEvent) {
		return
	}

	if analyser.inRound {
		// tick := analyser.getGameTick()
		// eventType := reflect.TypeOf(regEvent).String()
		// log.WithFields(log.Fields{
		// 	"event type": eventType,
		// 	"tick":       tick,
		// }).Info("Collection of event")

		eventList := &analyser.lastRound.collectedEvents
		*eventList = append(*eventList, regEvent)

		if baseevent, ok := regEvent.(baseEvent); ok {
			// playerhurt is base event type
			switch baseevent.event.(type) {
			// set player hurt event is received i thiss round
			case events.PlayerHurt:
				analyser.isPlayerHurt = true
			}
		}
	}
}

// createBaseEvent handle a base event
func (analyser *Analyser) createBaseEvent(e interface{}) baseEvent {
	tick := analyser.getGameTick()
	event := baseEvent{event: e, tick: tick}
	return event
}

// registerFakeEventHandlers register fake handler to player events
func (analyser *Analyser) registerFakeEventHandlers() {
	// in this function we register same handler for all function
	// because all events have been treaded in the same way in fake handler
	analyser.Parser.RegisterEventHandler(func(e events.Kill) {
		baseevent := analyser.createBaseEvent(e)
		analyser.fakeEventHandler(baseevent)
	})
	analyser.Parser.RegisterEventHandler(func(e events.PlayerHurt) {
		baseevent := analyser.createBaseEvent(e)
		analyser.fakeEventHandler(baseevent)
	})
	analyser.Parser.RegisterEventHandler(func(e events.WeaponFire) {
		baseevent := analyser.createBaseEvent(e)
		analyser.fakeEventHandler(baseevent)
	})
	analyser.Parser.RegisterEventHandler(func(e events.BombDefuseStart) {
		baseevent := analyser.createBaseEvent(e)
		analyser.fakeEventHandler(baseevent)
	})
	analyser.Parser.RegisterEventHandler(func(e events.BombDefused) {
		baseevent := analyser.createBaseEvent(e)
		analyser.fakeEventHandler(baseevent)
	})
	analyser.Parser.RegisterEventHandler(func(e events.BombPlanted) {
		baseevent := analyser.createBaseEvent(e)
		analyser.fakeEventHandler(baseevent)
	})
	analyser.Parser.RegisterEventHandler(func(e events.PlayerFlashed) {
		flashDuration := e.FlashDuration()
		baseevent := analyser.createBaseEvent(e)
		flashevent := flashEvent{baseEvent: baseevent, flashDuration: flashDuration}
		analyser.fakeEventHandler(flashevent)
	})
}

// dispatchPlayerEvents dispatch real events for valid round
func (analyser *Analyser) dispatchPlayerEvents() {
	lastRounTuple := analyser.lastRound
	tick := lastRounTuple.tick
	log.WithFields(log.Fields{
		"tick": tick,
	}).Info("Dispatching events for valid round")

	collectedEvents := lastRounTuple.collectedEvents
	for _, event := range collectedEvents {
		// first check whether is a base event or special event
		switch event.(type) {
		case baseEvent:
			baseevent := event.(baseEvent)
			originalevent := baseevent.event
			switch originalevent.(type) {
			case events.Kill:
				analyser.handleKill(baseevent)
			case events.PlayerHurt:
				analyser.handleHurt(baseevent)
			case events.WeaponFire:
				analyser.handleWeaponFire(baseevent)
			case events.BombDefuseStart:
				analyser.handleDefuseStart(baseevent)
			case events.BombDefused:
				analyser.handleBombDefused(baseevent)
			case events.BombPlanted:
				analyser.handleBombPlanted(baseevent)
			}
		case flashEvent:
			flashevent := event.(flashEvent)
			// originalevent := flashevent.event
			analyser.handlePlayerFlashed(flashevent)

		}

	}
	// // Register handler on bomb defuse abort
	// analyser.parser.RegisterEventHandler(func(e events.BombDefuseStart) {
	// 	defuserID = e.Player.SteamID
	//
	// 	if defuser, ok := analyser.GetPlayerByID(defuserID); ok {
	// 		analyser.isBombDefusing = true
	// 		analyser.defuser = defuser
	// 		log.WithFields(log.Fields{
	// 			"name":    e.defuser.Name,
	// 			"user id": defuserID,
	// 		}).Info("PLayer started to defuse bomb: ")
	// 	}
	//
	// })

	// // Register handler on bomb defuse abort
	// analyser.parser.RegisterEventHandler(func(e events.bomba) {
	// 	defuserID = e.Player.SteamID
	//
	// 	if defuser, ok := analyser.GetPlayerByID(defuserID); ok {
	// 		analyser.isBombDefusing = true
	// 		analyser.defuser = defuser
	// 		log.WithFields(log.Fields{
	// 			"name":    e.defuser.Name,
	// 			"user id": defuserID,
	// 		}).Info("PLayer started to defuse bomb: ")
	// 	}
	//
	// })
}

func (analyser *Analyser) updateScore(newTscore, newCTscore int) bool {
	if newTscore < 0 || newCTscore < 0 {
		return false
	}
	// first swap scores if needed
	mpOvertimeMaxrounds := analyser.NumOvertime
	nOvertimeRounds := analyser.RoundPlayed - maxRounds
	if analyser.RoundPlayed == 15 || analyser.RoundPlayed == maxRounds {
		if !analyser.ScoreSwapped {
			log.Info("Score has been swapped")
			analyser.Tscore, analyser.CTscore = analyser.CTscore, analyser.Tscore
			analyser.ScoreSwapped = true
		}
	} else if nOvertimeRounds > 0 && nOvertimeRounds%mpOvertimeMaxrounds == 0 {
		if !analyser.ScoreSwapped {
			log.Info("Score has been swapped")
			analyser.Tscore, analyser.CTscore = analyser.CTscore, analyser.Tscore
			analyser.ScoreSwapped = true
		}
	} else { //once score swapped and after swap situation we reset flag
		analyser.ScoreSwapped = false
	}

	oldTScore := analyser.Tscore
	oldCTScore := analyser.CTscore

	log.WithFields(log.Fields{
		"new t":  newTscore,
		"old t":  oldTScore,
		"new ct": newCTscore,
		"old ct": oldCTScore,
	}).Info("scores: ")

	if ((newTscore-oldTScore) == 1 && (newCTscore == oldCTScore)) ||
		((newCTscore-oldCTScore) == 1 && (newTscore == oldTScore)) {
		analyser.Tscore = newTscore
		analyser.CTscore = newCTscore
		return true

	}
	return false

}

// ##############################################
// ############# Seperate event handlers ########
func (analyser *Analyser) handleKill(baseEv baseEvent) {
	defer utils.RecoverPanic()

	// declare variables
	var ok bool
	var isVictimBlinded bool

	e := baseEv.event.(events.Kill)
	tick := baseEv.tick

	// get ids
	victimID, killerID, _, _ := analyser.getPlayerID(e.Victim, e.Killer, "Kill")

	// get player pointers
	// we are checking players disconnected as well
	// since event validity already confirmed
	victim, killer, victimOK, killerOK := analyser.checkEventValidity(victimID, killerID, "Kill", true)

	// check if victim and attacker exist
	if !victimOK || !killerOK {
		return
	}

	// get side of players
	victimSide, _, ok := analyser.checkTeamSideValidity(victim, killer)

	if !ok {
		return
	}

	// handle victim
	log.WithFields(log.Fields{
		"tick":   tick,
		"victim": victim.Name,
		"killer": e.Killer.Name,
		// "user id": victimID,
	}).Info("Player has been killed: ")

	victim.NotifyDeath(tick)
	analyser.deleteAlivePlayer(victimSide, victimID)
	// isVictimBlinded = victim.IsBlinded()
	// check trader - tradee relationship
	// if the victim killed someone not long ago we consider it's a trade, can be refined with
	// position as well
	killedByVictim, ok := analyser.killedPlayers[victimID]
	if ok {
		for _, victimKilled := range killedByVictim {
			if (tick - victimKilled.tick) <= 384 { // 3s
				victimKilledPlayer := victimKilled.player
				killer.NotifyTrader()
				victimKilledPlayer.NotifyTradee()
				// analyser.kastPlayers[killer.SteamID] = true
				analyser.kastPlayers[victimKilledPlayer.SteamID] = true
				break
			}
		}
	}

	// handle assister if exist
	if e.Assister != nil {
		assisterID := e.Assister.SteamID
		assister, assisterOK := analyser.getPlayerByID(assisterID, true)

		if assisterOK {
			log.WithFields(log.Fields{
				"tick":    tick,
				"name":    e.Assister.Name,
				"user id": assisterID,
			}).Info("Player did an assist for killing: ")
			assister.NotifyAssist()
			analyser.kastPlayers[assister.SteamID] = true
			// we will check is there a flash assist
			// first get the player that lastly flashed victim
			lastFlashedID := victim.GetLastFlashedBy()
			// if there is a player flashed victim
			if lastFlashedID > 0 {
				if lastFlashedPlayer, ok := analyser.getPlayerByID(lastFlashedID, true); ok {
					// if victim is blinded while killed and flash bomb has not thrown by killer
					// or players in the same team, then there is a flash assist
					lastFlashedPlayerSide, _ := lastFlashedPlayer.GetSide()
					if victim.IsBlinded() && lastFlashedID != killerID &&
						lastFlashedPlayerSide != victimSide {
						lastFlashedPlayer.NotifyFlashAssist()
					}
				}
			} //else {
			// 	log.WithFields(log.Fields{
			// 		"user id": lastFlashedID,
			// 	}).Error("Undefined flashed by player: ")
			// }
		}
	}

	IsHeadshot := e.IsHeadshot
	killer.NotifyKill(IsHeadshot, isVictimBlinded)
	analyser.kastPlayers[killer.SteamID] = true

	// update kill matrix
	newVictim := &killTuples{tick, victim}
	analyser.killedPlayers[killerID] = append(analyser.killedPlayers[killerID], newVictim)
}

func (analyser *Analyser) handleHurt(baseEv baseEvent) {
	e := baseEv.event.(events.PlayerHurt)
	tick := baseEv.tick
	// defer utils.RecoverPanic()

	// get entities in the event and game state variables
	damage := e.HealthDamage
	weaponType := e.Weapon.Class()

	// get ids
	victimID, attackerID, victimOK, killerOK := analyser.getPlayerID(e.Player, e.Attacker, "playerHurt")

	// check if victim and attacker exist in the event
	if !victimOK || !killerOK {
		return
	}
	// get player pointers
	_, attacker, victimOK, attackerOK := analyser.checkEventValidity(victimID, attackerID, "playerHurt", true)

	if !victimOK || !attackerOK {
		return
	}

	// handle victim
	log.WithFields(log.Fields{
		"tick":     tick,
		"victim":   e.Player.Name,
		"attacker": e.Attacker.Name,
		// "user id": victimID,
		"damage": damage,
	}).Info("Player has been hurt: ")

	// handle killer
	victimHealth := e.Health

	// granade damage
	if weaponType == common.EqClassGrenade {
		attacker.NotifyGranadeDamage(uint(damage))
	} else if e.Weapon.Weapon == common.EqIncendiary { //fire damage incendiary
		attacker.NotifyFireDamage(uint(damage))
	}

	// check if this is an headshot
	isHeadshot := false
	if e.HitGroup == events.HitGroupHead {
		isHeadshot = true
	}

	// if damage given by a weapon
	if weaponType != common.EqClassUnknown {
		attacker.NotifyDamageGiven(damage, victimHealth, isHeadshot, tick)
		attacker.NotifyDamageTaken(damage)
	}

}

func (analyser *Analyser) handleWeaponFire(baseEv baseEvent) {
	e := baseEv.event.(events.WeaponFire)
	tick := baseEv.tick

	shooterID := e.Shooter.SteamID

	if shooter, ok := analyser.getPlayerByID(shooterID, true); ok {
		shooter.NotifyWeaponFire(tick)
		log.WithFields(log.Fields{
			"tick":    tick,
			"name":    e.Shooter.Name,
			"user id": shooterID,
		}).Info("Player fired a weapon: ")
	} else {
		log.WithFields(log.Fields{
			"tick":    tick,
			"name":    e.Shooter.Name,
			"user id": shooterID,
		}).Error("Non exist player fired a weapon: ")
	}
}

func (analyser *Analyser) handleBombDefused(baseEv baseEvent) {
	e := baseEv.event.(events.BombDefused)
	tick := baseEv.tick

	defuserID := e.Player.SteamID

	if defuser, ok := analyser.getPlayerByID(defuserID, true); ok {
		defuser.NotifyBombDefused()
		analyser.isBombDefused = true
		log.WithFields(log.Fields{
			"tick":    tick,
			"defuser": defuser.Name,
			"user id": defuserID,
		}).Info("Bomb has been defused: ")
	} else {
		log.WithFields(log.Fields{
			"tick":    tick,
			"name":    defuser.Name,
			"user id": defuserID,
		}).Error("Bomb has been defused by non-exist player: ")
	}
}

func (analyser *Analyser) handleBombPlanted(baseEv baseEvent) {
	e := baseEv.event.(events.BombPlanted)
	tick := baseEv.tick

	planterID := e.Player.SteamID

	if planter, ok := analyser.getPlayerByID(planterID, true); ok {
		planter.NotifyBombPlanted()
		analyser.isBombPlanted = true
		log.WithFields(log.Fields{
			"tick":    tick,
			"planter": planter.Name,
			"user id": planterID,
		}).Info("Bomb has been planted: ")
	} else {
		log.WithFields(log.Fields{
			"tick":    tick,
			"name":    planter.Name,
			"user id": planterID,
		}).Error("Bomb has been planted by non exist player: ")
	}
}

func (analyser *Analyser) handleDefuseStart(baseEv baseEvent) {
	e := baseEv.event.(events.BombDefuseStart)
	tick := baseEv.tick

	defuserID := e.Player.SteamID

	if defuser, ok := analyser.getPlayerByID(defuserID, true); ok {
		analyser.isBombDefusing = true
		analyser.defuser = defuser
		log.WithFields(log.Fields{
			"tick":    tick,
			"defuser": defuser.Name,
			"user id": defuserID,
		}).Info("Player started to defuse bomb: ")
	} else {
		log.WithFields(log.Fields{
			"tick":    tick,
			"defuser": defuser.Name,
			"user id": defuserID,
		}).Error("Non exist player started to defuse bomb: ")
	}
}

// handleKAST notify players who did kast for this round
func (analyser *Analyser) handleKAST() {
	for currPlayerID, kastbool := range analyser.kastPlayers {
		if kastbool {
			player, isOK := analyser.getPlayerByID(currPlayerID, false)
			if isOK {
				player.NotifyKAST()
			}
		}
	}
}

func (analyser *Analyser) handlePlayerFlashed(flashedEv flashEvent) {
	e := flashedEv.event.(events.PlayerFlashed)
	tick := flashedEv.tick

	// get ids
	flashedPlayerID, attackerID, _, _ := analyser.getPlayerID(e.Player, e.Attacker, "Kill")

	// get player pointers
	// we are checking players disconnected as well
	// since event validity already confirmed
	flashed, killer, flashedOK, killerOK := analyser.checkEventValidity(flashedPlayerID, attackerID, "Flashed", true)

	// check if victim and attacker exist
	if !flashedOK || !killerOK {
		return
	}

	// get side of players
	_, _, ok := analyser.checkTeamSideValidity(flashed, killer)

	if !ok {
		return
	}

	duration := flashedEv.flashDuration
	if duration <= 0 {
		log.WithFields(log.Fields{
			"flashed":        flashed.Name,
			"attacker":       killer.Name,
			"flash duration": duration,
			"tick":           tick,
		}).Error("Player flashed error: ")
		return
	}
	log.WithFields(log.Fields{
		"player name":    flashed.Name,
		"tick":           tick,
		"attacker":       killer.Name,
		"flash duration": duration,
	}).Info("Player flashed: ")
	flashed.SetLastFlashedBy(killer.SteamID)
	if flashed.Team != killer.Team {
		flashed.NotifyBlindDuration(duration)
	}
}

func (analyser *Analyser) handleMatchStart() {
	// // TODO: modify depending on serialization
	if analyser.checkFinishedMatchValidity() {
		log.WithFields(log.Fields{
			"tick": analyser.getGameTick(),
		}).Info("A valid match is running")
		return
	}

	// first check whether we are in overtime
	if analyser.checkMatchContinuity() {
		log.WithFields(log.Fields{
			"tick": analyser.getGameTick(),
		}).Info("Overtime is playing for this match")
		return
	}

	var teamT, teamCT []*common.Player
	var ok bool
	tick := analyser.getGameTick()

	if !analyser.MatchStarted {

		log.WithFields(log.Fields{
			"tick": analyser.getGameTick(),
		}).Info("A new match has been started")

	} else {
		log.WithFields(log.Fields{
			"tick": analyser.getGameTick(),
		}).Error("Match has already started.Checking old match validity.")
		// first serialize the match
		// analyser.matchEncode()

		// second check already started match
		if analyser.checkFinishedMatchValidity() {
			log.WithFields(log.Fields{
				"tick": analyser.getGameTick(),
			}).Info("A valid match has been interrupted or finished.")

		} else {
			log.WithFields(log.Fields{
				"tick": analyser.getGameTick(),
			}).Error("Finished an invalid match match.Aborted")
			// analyser.matchDecode()
		}
	}

	// check participant validity
	if teamT, teamCT, ok = analyser.checkParticipantValidity(); !ok {
		log.WithFields(log.Fields{
			"tick": analyser.getGameTick(),
		}).Info("Not enough participant for a match start.Aborted.")
		analyser.isCancelled = true
		log.WithFields(log.Fields{
			"tick": analyser.getGameTick(),
		}).Error("Finished an invalid match match.Aborted")
		return
	}

	// reset match based variables
	analyser.resetMatchVars(teamT, teamCT, tick)
}

// handleSpecialRound handle special round won and loss
func (analyser *Analyser) handleSpecialRound(Winner, Loser []*common.Player) {
	// pistol round handling
	// only normal time
	if analyser.RoundPlayed <= 30 {
		// first round of each halfs
		if analyser.RoundPlayed%15 == 1 {
			// winner team
			for _, currPlayer := range Winner {
				if NewPPlayer, ok := analyser.getPlayerByID(currPlayer.SteamID, false); ok {
					NewPPlayer.NotifySpecialRoundWon("pistol_round")
				}
			}
			// loser team
			for _, currPlayer := range Loser {
				if NewPPlayer, ok := analyser.getPlayerByID(currPlayer.SteamID, false); ok {
					NewPPlayer.NotifySpecialRoundLost("pistol_round")
				}
			}
		}
	}
}

// ##############################################
// ############## Printer / writers #############
// printPlayers print player stats after finished the match
func (analyser *Analyser) printPlayers() {
	gs := analyser.Parser.GameState()
	log.Info("#########################################")

	log.WithFields(log.Fields{
		"t score":      analyser.Tscore,
		"ct score":     analyser.CTscore,
		"played round": analyser.RoundPlayed,
	}).Info("Match has been finished: ")
	for _, currPlayer := range analyser.Players {
		log.Info("**************************************")
		teamState := gs.Team(currPlayer.Team)
		log.WithFields(log.Fields{
			"name":                  currPlayer.Name,
			"team":                  teamState.ClanName,
			"kill":                  currPlayer.GetNumKills(),
			"blind kill":            currPlayer.GetBlindKills(),
			"blinded player killed": currPlayer.GetPlayerBlindedKills(),
			"hs kll":                currPlayer.GetNumHSKills(),
			"assist":                currPlayer.GetNumAssists(),
			"flash assist":          currPlayer.GetFlashAssist(),
			"death":                 currPlayer.GetNumDeaths(),
			"clutch won":            currPlayer.GetClutchWon(),
			"pistol won":            currPlayer.GetPistolRoundWon(),
			"granade damage":        currPlayer.GetGranadeDamage(),
			"fire damage":           currPlayer.GetFireDamage(),
			"time flashing":         currPlayer.GetTimeFlashing(),
			"kast":                  currPlayer.GetKAST(),
			"num trader":            currPlayer.GetNumTrader(),
			"num tradee":            currPlayer.GetNumTradee(),
		}).Info("Player: ")
	}
}

// ##############################################
// ############# Serialization function #########
// matchEncode encode a match
func (analyser *Analyser) matchEncode() {
	log.WithFields(log.Fields{
		"tick": analyser.getGameTick(),
	}).Info("A match has been encoded")

	encodedMatchVars := analyser.MatchVars
	e := gob.NewEncoder(&analyser.b)
	if err := e.Encode(encodedMatchVars); err != nil {
		utils.CheckError(err)
	}
	analyser.matchEncoded = true
}

// matchDecode decode a match
func (analyser *Analyser) matchDecode() {
	log.WithFields(log.Fields{
		"tick": analyser.getGameTick(),
	}).Info("A match has been decoded")

	if analyser.matchEncoded {
		dec := gob.NewDecoder(&analyser.b)
		var decodedMatchVars MatchVars
		if err := dec.Decode(&decodedMatchVars); err != nil {
			utils.CheckError(err)
		}

		analyser.MatchVars = decodedMatchVars
		analyser.matchEncoded = false
		analyser.b.Reset()

	} else {
		log.Error("There is no match to decode...")
	}

}

// ##############################################
