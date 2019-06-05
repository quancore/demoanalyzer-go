package analyser

import (
	"strconv"

	"github.com/markus-wa/demoinfocs-golang/events"
	"github.com/markus-wa/demoinfocs-golang/msg"
	logging "github.com/sirupsen/logrus"
)

// registerNetMessageHandlers register net message handlers
func (analyser *Analyser) registerNetMessageHandlers() {
	// Register handler for net messages updates
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

	analyser.log.Info("Match event handlers have been registered.")
}

// registerPlayerEventHandlers register handlers for each needed events
func (analyser *Analyser) registerPlayerEventHandlers() {
	// ************** player events *********************
	// Register handler on match start
	analyser.parser.RegisterEventHandler(func(e events.Kill) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.PlayerHurt) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.WeaponFire) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.BombDefused) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.BombPlanted) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.BombDefuseStart) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.PlayerFlashed) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.RoundMVPAnnouncement) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.ItemDrop) { analyser.dispatchPlayerEvents(e) })
	analyser.parser.RegisterEventHandler(func(e events.ItemPickup) { analyser.dispatchPlayerEvents(e) })

	// **************************************************
	// registered for testing purpose
	// analyser.parser.RegisterEventHandler(func(e events.ItemDrop) {
	// 	ownerName := "-"
	// 	if e.Weapon.Owner != nil {
	// 		ownerName = e.Weapon.Owner.Name
	// 	}
	// 	if e.Player == nil {
	// 		return
	// 	}
	// 	analyser.log.WithFields(logging.Fields{
	// 		"tick":    analyser.getGameTick(),
	// 		"weapon":  e.Weapon.Class(),
	// 		"owner":   ownerName,
	// 		"dropper": e.Player.Name,
	// 	}).Info("Item has been dropped")
	// })
	// analyser.parser.RegisterEventHandler(func(e events.ItemEquip) {
	// 	ownerName := "-"
	// 	if e.Weapon.Owner != nil {
	// 		ownerName = e.Weapon.Owner.Name
	// 	}
	// 	if e.Player == nil {
	// 		return
	// 	}
	// 	analyser.log.WithFields(logging.Fields{
	// 		"tick":     analyser.getGameTick(),
	// 		"weapon":   e.Weapon.Class(),
	// 		"owner":    ownerName,
	// 		"equipper": e.Player.Name,
	// 	}).Info("Item has been equipped")
	// })
	// analyser.parser.RegisterEventHandler(func(e events.ItemPickup) {
	// 	ownerName := "-"
	// 	if e.Weapon.Owner != nil {
	// 		ownerName = e.Weapon.Owner.Name
	// 	}
	// 	if e.Player == nil {
	// 		return
	// 	}
	// 	analyser.log.WithFields(logging.Fields{
	// 		"tick":   analyser.getGameTick(),
	// 		"weapon": e.Weapon.Class(),
	// 		"owner":  ownerName,
	// 		"picker": e.Player.Name,
	// 	}).Info("Item has been pickup")
	// })
	analyser.log.Info("Player event handlers have been registered for second parse.")

}

// ########## first parse related  handlers #########################################################
// registerFirstPlayerEventHandlers register handlers for each needed events in the first parse session
func (analyser *Analyser) registerFirstPlayerEventHandlers() {
	// it is used for indicating a player has been hurt in a round for the first parse
	// therefore it is registered in the first parsing
	if analyser.isFirstParse {
		analyser.parser.RegisterEventHandler(func(e events.PlayerHurt) { analyser.handleHurt(e) })
		// it is very rare however, there could be some round that no one get hurts and bomb has been
		// exploded and a team win the round. So we need to record bomb planted event as well in the
		// first parse to understand a round is valid.
		analyser.parser.RegisterEventHandler(func(e events.BombPlanted) { analyser.handleBombPlanted(e) })
		analyser.parser.RegisterEventHandler(func(e events.Kill) { analyser.handleCheckKill(e) })
		analyser.log.Info("Player event handlers have been registered for first parse.")

	}
}

func (analyser *Analyser) registerScheduler() {
	analyser.parser.RegisterEventHandler(func(e events.TickDone) { analyser.customScheduler.checkEvent(analyser.getGameTick()) })
	analyser.log.Info("Scheduler for custom event handlers has been registered.")

}

// #################################################
