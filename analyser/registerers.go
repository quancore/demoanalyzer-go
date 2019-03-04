package analyser

import (
	"strconv"

	"github.com/markus-wa/demoinfocs-golang/events"
	"github.com/markus-wa/demoinfocs-golang/msg"
	logging "github.com/sirupsen/logrus"
)

// registerNetMessageHandlers register net message handlers
func (analyser *Analyser) registerNetMessageHandlers() {
	// Register handler for ConVar updates
	analyser.parser.RegisterNetMessageHandler(func(m *msg.CNETMsg_SetConVar) {
		for _, cvar := range m.Convars.Cvars {
			if cvar.Name == "mp_overtime_maxrounds" {
				analyser.NumOvertime, _ = strconv.Atoi(cvar.Value)
			} else if cvar.Name == "mp_startmoney" {
				analyser.currentSMoney, _ = strconv.ParseFloat(cvar.Value, 64)
				analyser.isMoneySet = true
			}
			analyser.log.WithFields(logging.Fields{
				"cvar name":  cvar.Name,
				"cvar value": cvar.Value,
			}).Info("Cvars ")
		}
	})
}

func (analyser *Analyser) registerMatchEventHandlers() {
	// *********** match events ********************
	// this event is not triggered by the actions of players
	// Register handler on match start
	analyser.parser.RegisterEventHandler(func(e events.MatchStart) { analyser.handleMatchStart("matchStart") })
	// Register handler on match start.Sometimes, match start event is not called
	analyser.parser.RegisterEventHandler(func(e events.MatchStartedChanged) { analyser.handleMatchStart("matchStartChanged") })
	// Register handler on player connected
	analyser.parser.RegisterEventHandler(func(e events.PlayerConnect) { analyser.handlePlayerConnect(e) })
	// Register handler on player disconnected
	analyser.parser.RegisterEventHandler(func(e events.PlayerDisconnected) { analyser.handlePlayerDisconnect(e) })
	// Register handler on game phase changed. Useful for match end
	// analyser.parser.RegisterEventHandler(func(e events.GamePhaseChanged) { analyser.handleGamePhaseChange(e) })
	// Register handler on round start
	analyser.parser.RegisterEventHandler(func(e events.RoundStart) { analyser.handleRoundStart(e) })
	// Register handler on score updated event
	analyser.parser.RegisterEventHandler(func(e events.ScoreUpdated) { analyser.handleRoundEnd(e) })
	// Register handler on round end event
	analyser.parser.RegisterEventHandler(func(e events.RoundEnd) { analyser.handleRoundEnd(e) })
	// Register handler on round end official event
	analyser.parser.RegisterEventHandler(func(e events.RoundEndOfficial) { analyser.handleRoundOfficiallyEnd(e) })
	// Register handler on player team change
	// mainly needed for reconnected players
	analyser.parser.RegisterEventHandler(func(e events.PlayerTeamChange) { analyser.handleTeamChange(e) })

	// // Register handler on round end official event
	// analyser.parser.RegisterEventHandler(func(e events.TeamSideSwitch) {
	// 	// switch scores
	// 	// first swap scores if needed
	// 	mpOvertimeMaxrounds := analyser.NumOvertime
	// 	nOvertimeRounds := analyser.RoundPlayed - maxRounds
	// 	if analyser.RoundPlayed == 15 || analyser.RoundPlayed == maxRounds {
	// 		if !analyser.ScoreSwapped {
	// 			log.Info("Score has been swapped with team switch")
	// 			analyser.Tscore, analyser.CTscore = analyser.CTscore, analyser.Tscore
	// 			analyser.ScoreSwapped = true
	// 		}
	// 	} else if nOvertimeRounds > 0 && nOvertimeRounds%mpOvertimeMaxrounds == 0 {
	// 		if !analyser.ScoreSwapped {
	// 			log.Info("Score has been swapped with team switch")
	// 			analyser.Tscore, analyser.CTscore = analyser.CTscore, analyser.Tscore
	// 			analyser.ScoreSwapped = true
	// 		}
	// 	} else { //once score swapped and after swap situation we reset flag
	// 		analyser.ScoreSwapped = false
	// 	}
	// })

	// Register handler on round end official event
	// registered for testing purpose
	analyser.parser.RegisterEventHandler(func(e events.IsWarmupPeriodChanged) {
		analyser.log.WithFields(logging.Fields{
			"tick":        analyser.getGameTick(),
			"old warm up": e.OldIsWarmupPeriod,
			"new warm up": e.NewIsWarmupPeriod,
		}).Info("Warm up period")
	})

	// it is used for indicating a player has been hurt in a round for the first parse
	// therefore it is registered in the first parsing
	if analyser.isFirstParse {
		analyser.parser.RegisterEventHandler(func(e events.PlayerHurt) { analyser.handleHurt(e) })
	}

	// **************************************************
}

// registerEventHandlers register handlers for each needed events
func (analyser *Analyser) registerEventHandlers() {
	// ************** player events *********************
	// Register handler on match start
	analyser.parser.RegisterEventHandler(func(e events.Kill) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.PlayerHurt) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.WeaponFire) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.BombDefused) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.BombPlanted) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.BombDefuseStart) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.PlayerFlashed) { analyser.dispatchPlayerEvents(e) })
	// **************************************************
}

// #################################################
