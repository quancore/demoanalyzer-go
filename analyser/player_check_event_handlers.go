package analyser

import (
	events "github.com/markus-wa/demoinfocs-golang/events"
	logging "github.com/sirupsen/logrus"
)

// Event handlers for checking whether a round is valid or not in the first parsing stage

// handleCheckKill check whether a kill event is valid. If valid schedule related pre/post event checkers.
func (analyser *Analyser) handleCheckKill(e events.Kill, tick int) {

	// get player pointers
	// we are checking players disconnected as well
	victim, killer, victimOK, killerOK := analyser.checkEventValidity(e.Victim, e.Killer, "Kill", true, tick)

	// check if victim and attacker exist
	if !victimOK || !killerOK {
		return
	}

	// get side of players
	victimSide, killerSide, sideOK := analyser.checkTeamSideValidity(victim, killer, "kill", tick)

	if sideOK {
		// check crosshair replecament
		ev := preCroshairReplecament{eventCommon: eventCommon{analyser: analyser, offsetSec: -analyser.beforeCrosshair, isPeriodic: false},
			killerID: killer.GetSteamID()}

		executionTick := analyser.customScheduler.addEvent(tick, -analyser.beforeCrosshair, ev)
		analyser.log.WithFields(logging.Fields{
			"tick":         tick,
			"victim":       victim.Name,
			"killer":       killer.Name,
			"killer side":  killerSide,
			"victim side":  victimSide,
			"will execute": executionTick,
			// "user id": victimID,
		}).Info("Recording for crosshair replecament has been scheduled ")

		firstKillEv := postFirstKillChecker{eventCommon: eventCommon{analyser: analyser, offsetSec: analyser.afterFirstKill, isPeriodic: false},
			killerID: killer.GetSteamID()}

		executionTick = analyser.customScheduler.addEvent(tick, analyser.afterFirstKill, firstKillEv)
		analyser.log.WithFields(logging.Fields{
			"tick":         tick,
			"victim":       victim.Name,
			"killer":       killer.Name,
			"killer side":  killerSide,
			"victim side":  victimSide,
			"will execute": executionTick,
		}).Info("First kill has been scheduled: ")

	}
}
