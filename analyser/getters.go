package analyser

import (
	p_common "github.com/markus-wa/demoinfocs-golang/common"
	common "github.com/quancore/demoanalyzer-go/common"
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
func (analyser *Analyser) getPlayerByID(uid int64, allPlayer bool) (*common.PPlayer, bool) {
	var player *common.PPlayer
	var ok bool

	if player, ok = analyser.players[uid]; !ok {
		if playerDisconnected, discok := analyser.disconnectedPlayers[uid]; discok {
			if allPlayer {
				return playerDisconnected.Player, discok
			}
		}
	}

	return player, ok
}

// getPlayerByID get the pointer to disconnected player by player id
func (analyser *Analyser) getDisconnectedPlayerByID(uid int64) (*common.DisconnectedTuple, bool) {

	player, ok := analyser.disconnectedPlayers[uid]

	return player, ok
}

// getAllPlayers return all connected and disconnected valid players
func (analyser *Analyser) getAllPlayers() []*common.PPlayer {

	var allPLayers []*common.PPlayer

	for _, pplayer := range analyser.players {
		allPLayers = append(allPLayers, pplayer)
	}

	for uid, discpplayer := range analyser.disconnectedPlayers {
		if _, ok := analyser.players[uid]; !ok {
			allPLayers = append(allPLayers, discpplayer.Player)
		}
	}

	return allPLayers
}

// getTeams get t team and ct team players
func (analyser *Analyser) getTeams(isdisconnected bool) ([]*common.PPlayer, []*common.PPlayer) {
	var tTeam []*common.PPlayer
	var ctTeam []*common.PPlayer

	for _, pplayer := range analyser.players {
		if pplayer.Team == p_common.TeamTerrorists {
			tTeam = append(tTeam, pplayer)
		} else if pplayer.Team == p_common.TeamCounterTerrorists {
			ctTeam = append(ctTeam, pplayer)
		}
	}

	if isdisconnected { //if we need disconnected players as well
		for uid, discpplayer := range analyser.disconnectedPlayers {
			if _, ok := analyser.players[uid]; !ok {
				if discpplayer.Player.Team == p_common.TeamTerrorists {
					tTeam = append(tTeam, discpplayer.Player)
				} else if discpplayer.Player.Team == p_common.TeamCounterTerrorists {
					ctTeam = append(ctTeam, discpplayer.Player)
				}
			}
		}
	}

	return tTeam, ctTeam
}

// getPlayerId get player id from players sent in an event
// used for kill event
func (analyser *Analyser) getPlayerID(victim, killer *p_common.Player, eventName string) (int64, int64, bool, bool) {
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

func (analyser *Analyser) getWinnerTeam() p_common.Team {
	// get which team won
	teamWon := p_common.TeamUnassigned
	if analyser.tScore > analyser.ctScore {
		teamWon = p_common.TeamTerrorists
	} else if analyser.tScore < analyser.ctScore {
		teamWon = p_common.TeamCounterTerrorists
	}

	return teamWon
}

// #############################
