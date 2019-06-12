package analyser

import (
	p_common "github.com/markus-wa/demoinfocs-golang/common"
	common "github.com/quancore/demoanalyzer-go/common"
	"github.com/quancore/demoanalyzer-go/utils"
	logging "github.com/sirupsen/logrus"
)

// #### first kill ###################
type postFirstKillChecker struct {
	eventCommon
	killerID int64
}

func (ec postFirstKillChecker) handleEvent(tick int) {

	if ec.analyser.isTFirstKill && ec.analyser.isCTFirstKill {
		return
	}

	ec.analyser.log.WithFields(logging.Fields{
		"tick": tick,
	}).Debug("Handling first kill event")

	// first handle first kill event
	if player, playerOK := ec.analyser.getPlayerByID(ec.killerID, false); playerOK {
		// check first kill
		killerSide := player.Team
		var isFirstKill bool
		if killerSide == p_common.TeamTerrorists && ec.analyser.isTFirstKill == false {
			isFirstKill = true

		} else if killerSide == p_common.TeamCounterTerrorists && ec.analyser.isCTFirstKill == false {
			isFirstKill = true
		}
		if isFirstKill {
			if player.Player.IsAlive() {
				player.NotifyFirstKill()
				if killerSide == p_common.TeamTerrorists {
					ec.analyser.isTFirstKill = true
				} else if killerSide == p_common.TeamCounterTerrorists {
					ec.analyser.isCTFirstKill = true
				}
				ec.analyser.log.WithFields(logging.Fields{
					"tick":               tick,
					"killer":             player.Name,
					"player team":        killerSide,
					"t first kill bool":  ec.analyser.isTFirstKill,
					"ct first kill bool": ec.analyser.isCTFirstKill,
					"round number":       ec.analyser.roundPlayed,
				}).Info("First kill has been done: ")

			} else {
				ec.analyser.log.WithFields(logging.Fields{
					"tick":   tick,
					"killer": player.Name,
				}).Info("Invalid first kill. Player has been killed.")
			}
		}
	}
}
func (ec postFirstKillChecker) postEventHandler() { return }

// ######## croshair replecament ####
type preCroshairReplecament struct {
	eventCommon
	killerID int64
}

func (ec preCroshairReplecament) handleEvent(tick int) {
	// first handle first kill event
	if player, playerOK := ec.analyser.getPlayerByID(ec.killerID, false); playerOK {
		x, y := player.SetLastYawPitch()
		ec.analyser.log.WithFields(logging.Fields{
			"tick":   tick,
			"killer": player.Name,
			"x":      x,
			"y":      y,
		}).Info("Yaw and pitch value have been recorded ")
	}
}

func (ec preCroshairReplecament) postEventHandler() { return }

// #### map control ###################
type mapControl struct {
	eventCommon
}

func (ec mapControl) handleEvent(tick int) {
	ec.analyser.log.WithFields(logging.Fields{
		"tick":       tick,
		"event name": "map control",
	}).Debug("Map place assignment has been called")
	tAlive := ec.analyser.tAlive
	ctAlive := ec.analyser.ctAlive
	ec.handleTeam(tAlive, tick)
	ec.handleTeam(ctAlive, tick)
	// set last check of map occupancy done after all team has been checked
	ec.analyser.navigator.SetLastAssignedTick(tick)

}

func (ec mapControl) handleTeam(team map[int64]*common.PPlayer, tick int) {
	for _, player := range team {
		lastMovedTick := player.GetlastFootstepTick()
		lastMapAssignmentTick := ec.analyser.navigator.GetLastAssignedTick()
		// if the player has not moved since last map assingment check, no point to process
		if lastMovedTick > lastMapAssignmentTick {
			ec.analyser.log.WithFields(logging.Fields{
				"tick":              tick,
				"name":              player.Name,
				"last check done":   lastMapAssignmentTick,
				"player last moved": lastMovedTick,
			}).Debug("Place assignment for moved player has been called")
			ec.analyser.navigator.AssignPlace(player, tick)

		} else {
			ec.analyser.log.WithFields(logging.Fields{
				"tick":              tick,
				"name":              player.Name,
				"last check done":   lastMapAssignmentTick,
				"player last moved": lastMovedTick,
			}).Debug("Player has not moved since last map assignment check")
			// we need to update tick value of assigned places to this stationary player
			// and check the assigned place situation (exitence of opponent team etc.)
			ec.analyser.navigator.UpdateTeamAssignment(player.GetSteamID(), player.TeamState, tick)

		}
	}

}

func (ec mapControl) postEventHandler() {
	tick, err := ec.analyser.getGameTick()
	if err {
		return
	}
	ec.analyser.log.Info("Calculating team map occupancy")
	tArea, ctArea := ec.analyser.navigator.CalculateMapOccupancy()
	totalMapArea := ec.analyser.mapArea
	ec.analyser.notifySquareMeter(utils.SafeDivision(tArea, totalMapArea), utils.SafeDivision(ctArea, totalMapArea), tick)
}
