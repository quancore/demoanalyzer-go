package analyser

import (
	"github.com/markus-wa/demoinfocs-golang/common"
	"github.com/quancore/demoanalyzer-go/player"
	logging "github.com/sirupsen/logrus"
)

// ########## Internal getters ##############
// all getter methods
// getGameTick get current tick of game
func (analyser *Analyser) getGameTick() int {
	tick := analyser.parser.GameState().IngameTick()
	if tick < 0 {
		analyser.log.WithFields(logging.Fields{
			"tick": tick,
		}).Error("Negative tick number")
	}
	return tick
}

// getPlayerByID get the pointer to pplayer by player id
func (analyser *Analyser) getPlayerByID(uid int64, allPlayer bool) (*player.PPlayer, bool) {
	var player *player.PPlayer
	var ok bool

	if player, ok = analyser.Players[uid]; !ok {
		if playerDisconnected, discok := analyser.DisconnectedPlayers[uid]; discok {
			if allPlayer {
				return playerDisconnected.Player, discok
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

// getAllPlayers return all connected and disconnected valid players
func (analyser *Analyser) getAllPlayers() []*player.PPlayer {

	var allPLayers []*player.PPlayer

	for _, pplayer := range analyser.Players {
		allPLayers = append(allPLayers, pplayer)
	}

	for uid, discpplayer := range analyser.DisconnectedPlayers {
		if _, ok := analyser.Players[uid]; !ok {
			allPLayers = append(allPLayers, discpplayer.Player)
		}
	}

	return allPLayers
}

// getTeams get t team and ct team players
func (analyser *Analyser) getTeams(isdisconnected bool) ([]*player.PPlayer, []*player.PPlayer) {
	var tTeam []*player.PPlayer
	var ctTeam []*player.PPlayer

	for _, pplayer := range analyser.Players {
		if pplayer.Team == common.TeamTerrorists {
			tTeam = append(tTeam, pplayer)
		} else if pplayer.Team == common.TeamCounterTerrorists {
			ctTeam = append(ctTeam, pplayer)
		}
	}

	if isdisconnected { //if we need disconnected players as well
		for uid, discpplayer := range analyser.DisconnectedPlayers {
			if _, ok := analyser.Players[uid]; !ok {
				if discpplayer.Player.Team == common.TeamTerrorists {
					tTeam = append(tTeam, discpplayer.Player)
				} else if discpplayer.Player.Team == common.TeamCounterTerrorists {
					ctTeam = append(ctTeam, discpplayer.Player)
				}
			}
		}
	}

	return tTeam, ctTeam
}

// getPlayerId get player id from players sent in an event
// used for kill event
func (analyser *Analyser) getPlayerID(victim, killer *common.Player, eventName string) (int64, int64, bool, bool) {
	var victimID, killerID int64
	var victimName, killerName string
	var victimOK, killerOK bool

	if victim != nil {
		victimID = victim.SteamID
		victimName = victim.Name
		victimOK = true
	}
	if killer != nil {
		killerID = killer.SteamID
		killerName = killer.Name
		killerOK = true
	}

	if killer == nil || victim == nil {
		analyser.log.WithFields(logging.Fields{
			"event":       eventName,
			"tick":        analyser.getGameTick(),
			"killer name": killerName,
			"victim name": victimName,
		}).Error("Victim or killer is nill for event: ")
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
