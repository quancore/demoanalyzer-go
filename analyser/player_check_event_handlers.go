package analyser

import (
	p_common "github.com/markus-wa/demoinfocs-golang/common"
	events "github.com/markus-wa/demoinfocs-golang/events"
	logging "github.com/sirupsen/logrus"
)

func (analyser *Analyser) handleCheckKill(e events.Kill) {
	tick := analyser.getGameTick()
	// get player pointers
	// we are checking players disconnected as well
	victim, killer, victimOK, killerOK := analyser.checkEventValidity(e.Victim, e.Killer, "Kill", true)

	// check if victim and attacker exist
	if !victimOK || !killerOK {
		return
	}

	// get side of players
	victimSide, killerSide, sideOK := analyser.checkTeamSideValidity(victim, killer, "kill")

	if sideOK {
		// check crosshair replecament
		ev := preCroshairReplecament{eventCommon: eventCommon{Tick: tick, analyser: analyser},
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

		// check first kill
		var IsFirstKill bool
		if killerSide == p_common.TeamTerrorists && analyser.isTFirstKill == false {
			IsFirstKill = true
			analyser.isTFirstKill = true

		} else if killerSide == p_common.TeamCounterTerrorists && analyser.isCTFirstKill == false {
			IsFirstKill = true
			analyser.isCTFirstKill = true
		}
		if IsFirstKill {
			ev := postFirstKillChecker{eventCommon: eventCommon{Tick: tick, analyser: analyser},
				killerID: killer.GetSteamID()}

			executionTick := analyser.customScheduler.addEvent(tick, analyser.afterFirstKill, ev)
			analyser.log.WithFields(logging.Fields{
				"tick":         tick,
				"victim":       victim.Name,
				"killer":       killer.Name,
				"killer side":  killerSide,
				"victim side":  victimSide,
				"will execute": executionTick,
				// "user id": victimID,
			}).Info("First kill has been scheduled: ")
		}
	}
}
