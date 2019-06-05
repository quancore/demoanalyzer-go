package analyser

import (
	logging "github.com/sirupsen/logrus"
)

// #### first kill ###################
type postFirstKillChecker struct {
	eventCommon
	killerID int64
}

func (ec postFirstKillChecker) handleEvent() {
	// first handle first kill event
	if player, playerOK := ec.analyser.getPlayerByID(ec.killerID, false); playerOK {
		if player.Player.IsAlive() {
			player.NotifyFirstKill()
			ec.analyser.log.WithFields(logging.Fields{
				"tick":   ec.Tick,
				"killer": player.Name,
				// "user id": victimID,
			}).Info("First kill has been done: ")
		} else {
			ec.analyser.log.WithFields(logging.Fields{
				"tick":   ec.Tick,
				"killer": player.Name,
				// "user id": victimID,
			}).Info("Invalid first kill. Player has been killed.")
		}
	}
}

// ######## croshair replecament ####
type preCroshairReplecament struct {
	eventCommon
	killerID int64
}

func (ec preCroshairReplecament) handleEvent() {
	// first handle first kill event
	if player, playerOK := ec.analyser.getPlayerByID(ec.killerID, false); playerOK {
		x, y := player.SetLastYawPitch()
		ec.analyser.log.WithFields(logging.Fields{
			"tick":   ec.Tick,
			"killer": player.Name,
			"x":      x,
			"y":      y,
		}).Info("Yaw and pitch value have been recorded ")
	}
}
