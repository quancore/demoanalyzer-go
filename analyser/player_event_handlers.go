package analyser

import (
	"github.com/markus-wa/demoinfocs-golang/common"
	events "github.com/markus-wa/demoinfocs-golang/events"
	logging "github.com/sirupsen/logrus"
)

// ############# Seperate event handlers ########
// all event handlers for player triggered events

// dispatchPlayerEvents common function to handle a player event
// usefull for common checks for all events
// used for second time parsing
func (analyser *Analyser) dispatchPlayerEvents(e interface{}) {
	// set event happened to true for this round
	analyser.IsEventHappened = true

	tick := analyser.getGameTick()
	// it is nearly impossible to a player evetn can happen at the same
	// time with an round start or end.It is much common to be a server event
	// like killing all players at the same time for a game reset
	if !analyser.setRoundStart(tick) || (tick == analyser.roundStart || tick == analyser.roundOffEnd) {
		analyser.log.WithFields(logging.Fields{
			"tick":       tick,
			"start tick": analyser.roundStart,
			"end tick":   analyser.roundEnd,
		}).Info("Invalid event: ")
		return
	}

	switch e.(type) {
	case events.Kill:
		analyser.handleKill(e.(events.Kill))
	case events.PlayerHurt:
		analyser.handleHurt(e.(events.PlayerHurt))
	case events.WeaponFire:
		analyser.handleWeaponFire(e.(events.WeaponFire))
	case events.BombDefuseStart:
		analyser.handleDefuseStart(e.(events.BombDefuseStart))
	case events.BombDefused:
		analyser.handleBombDefused(e.(events.BombDefused))
	case events.BombPlanted:
		analyser.handleBombPlanted(e.(events.BombPlanted))
	case events.PlayerFlashed:
		analyser.handlePlayerFlashed(e.(events.PlayerFlashed))
	}
}

func (analyser *Analyser) handleKill(e events.Kill) {
	// defer utils.RecoverPanic()

	// declare variables
	var isVictimBlinded bool

	tick := analyser.getGameTick()

	// get ids
	victimID, killerID, victimOK, killerOK := analyser.getPlayerID(e.Victim, e.Killer, "Kill")

	// check if victim and attacker exist in the event
	if !victimOK || !killerOK {
		return
	}
	// get player pointers
	// we are checking players disconnected as well
	// since event validity already confirmed
	victim, killer, victimOK, killerOK := analyser.checkEventValidity(e.Victim, e.Killer, "Kill", true)

	// check if victim and attacker exist
	if !victimOK || !killerOK {
		return
	}

	// get side of players
	victimSide, killerSide, sideOK := analyser.checkTeamSideValidity(victim, killer, "kill")

	// if !sideOK{
	// 	return
	// }

	// handle victim
	analyser.log.WithFields(logging.Fields{
		"tick":        tick,
		"victim":      victim.Name,
		"killer":      killer.Name,
		"killer side": killerSide,
		"victim side": victimSide,
		// "user id": victimID,
	}).Info("Player has been killed: ")

	victim.NotifyDeath(tick)
	analyser.deleteAlivePlayer(victimSide, victimID)
	isVictimBlinded = victim.IsBlinded()

	// if two players are in the same side, there is no frag for killer and
	// there is no trader - tradee relationship and KAST for killing
	// However, death count for victim and assist count for assister
	if sideOK {
		// handle killer
		IsHeadshot := e.IsHeadshot
		killer.NotifyKill(IsHeadshot, isVictimBlinded)
		analyser.KastPlayers[killer.SteamID] = true

		// update kill matrix
		newVictim := &killTuples{tick, victim}
		analyser.KilledPlayers[killerID] = append(analyser.KilledPlayers[killerID], newVictim)

		// check trader - tradee relationship
		// if the victim killed someone not long ago we consider it's a trade, can be refined with
		// position as well
		if killedByVictim, ok := analyser.KilledPlayers[victimID]; ok {
			for _, victimKilled := range killedByVictim {
				if (tick - victimKilled.Tick) <= 640 { // 5s
					victimKilledPlayer := victimKilled.Player
					killer.NotifyTrader()
					victimKilledPlayer.NotifyTradee()
					// analyser.kastPlayers[killer.SteamID] = true
					analyser.KastPlayers[victimKilledPlayer.SteamID] = true
					break
				}
			}
		}
	}

	// we will check is there a flash assist
	// first get the player that lastly flashed victim
	lastFlashedID := victim.GetLastFlashedBy()
	// flag to indicate whether there is a flash assist
	// var didFlashAssist bool
	// if there is a player flashed victim
	if lastFlashedID > 0 {
		if lastFlashedPlayer, ok := analyser.getPlayerByID(lastFlashedID, true); ok {
			// if victim is blinded while killed and flash bomb has not thrown by killer
			// or players in the same team, then there is a flash assist

			// Right now, underflowed flash counts are caused by is blinded method
			// sometimes it falsely finds a player not blind so there is not flash assist
			lastFlashedPlayerSide, _ := lastFlashedPlayer.GetSide()
			if isVictimBlinded && lastFlashedID != killerID &&
				lastFlashedPlayerSide != victimSide {
				lastFlashedPlayer.NotifyFlashAssist()
				// mark player did a kast
				analyser.KastPlayers[lastFlashedPlayer.SteamID] = true
				// didFlashAssist = true
				analyser.log.WithFields(logging.Fields{
					"tick":       tick,
					"name":       lastFlashedPlayer.Name,
					"user id":    lastFlashedID,
					"is blinded": isVictimBlinded,
					"victim":     victim.Name,
				}).Info("Player did a flash assist for killing: ")
			} //else if lastFlashedID != killerID &&
			// 	lastFlashedPlayerSide != victimSide {
			// 	analyser.log.WithFields(logging.Fields{
			// 		"tick":           tick,
			// 		"flasher name":   lastFlashedPlayer.Name,
			// 		"victim":         victim.Name,
			// 		"victim blinded": isVictimBlinded,
			// 		"killer":         killer.Name,
			// 		"user id":        lastFlashedID,
			// 	}).Info("Player not a flash assist for killing: ")
			// }
		}
	}
	// check normal assister
	if e.Assister != nil {
		assisterID := e.Assister.SteamID
		if assister, assisterOK := analyser.getPlayerByID(assisterID, true); assisterOK {
			// we are only checking side of players
			// not whenther flash assister also did normal assist
			// because at the same time a player can flash assist and normal assist as well
			if assister.Team == victim.Team {
				return
			}

			// if (didFlashAssist && (lastFlashedID == assister.SteamID)) || (assister.Team == victim.Team) {
			// 	analyser.log.WithFields(logging.Fields{
			// 		"tick":             tick,
			// 		"name":             assister.Name,
			// 		"assiter steam id": assisterID,
			// 		"last flasher id":  lastFlashedID,
			// 		"did flash":        didFlashAssist,
			// 		"user id":          assisterID,
			// 		"victim":           victim.Name,
			// 	}).Info("Player did invalid an assist for killing: ")
			// 	return
			// }
			analyser.log.WithFields(logging.Fields{
				"tick":    tick,
				"name":    e.Assister.Name,
				"user id": assisterID,
			}).Info("Player did an assist for killing: ")

			assister.NotifyAssist()
			// mark player did a kast
			analyser.KastPlayers[assister.SteamID] = true
		}
	}
}

func (analyser *Analyser) handleHurt(e events.PlayerHurt) {
	tick := analyser.getGameTick()
	// defer utils.RecoverPanic()

	// get entities in the event and game state variables
	damage := e.HealthDamage
	weaponType := e.Weapon.Class()

	// get ids
	_, _, victimOK, killerOK := analyser.getPlayerID(e.Player, e.Attacker, "playerHurt")

	// check if victim and attacker exist in the event
	if !victimOK || !killerOK {
		return
	}
	// get player pointers
	_, attacker, victimOK, attackerOK := analyser.checkEventValidity(e.Player, e.Attacker, "playerHurt", true)

	if !victimOK || !attackerOK {
		return
	}

	// handle victim
	analyser.log.WithFields(logging.Fields{
		"tick":     tick,
		"victim":   e.Player.Name,
		"attacker": e.Attacker.Name,
		"health":   e.Health,
		"damage":   damage,
	}).Info("Player has been hurt: ")

	if e.Health <= 0 {
		// handle victim
		analyser.log.WithFields(logging.Fields{
			"tick":          tick,
			"victim":        e.Player.Name,
			"attacker":      e.Attacker.Name,
			"damage":        damage,
			"victim health": e.Health,
		}).Info("Player has been hurt with zero health: ")
	}

	analyser.IsPlayerHurt = true

	// handle killer
	victimHealth := e.Health

	// did hurt with grenade class
	if weaponType == common.EqClassGrenade {
		if e.Weapon.Weapon == common.EqHE { // he damage
			attacker.NotifyGranadeDamage(uint(damage))
		} else if e.Weapon.Weapon == common.EqIncendiary || e.Weapon.Weapon == common.EqMolotov { // inferno damage
			attacker.NotifyFireDamage(uint(damage))
		}
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

func (analyser *Analyser) handleWeaponFire(e events.WeaponFire) {
	tick := analyser.getGameTick()
	shooterID := e.Shooter.SteamID

	if shooter, ok := analyser.getPlayerByID(shooterID, true); ok {
		shooter.NotifyWeaponFire(tick)
		analyser.log.WithFields(logging.Fields{
			"tick":    tick,
			"name":    e.Shooter.Name,
			"user id": shooterID,
		}).Info("Player fired a weapon: ")
	} else {
		analyser.log.WithFields(logging.Fields{
			"tick":    tick,
			"name":    e.Shooter.Name,
			"user id": shooterID,
		}).Error("Non exist player fired a weapon: ")
	}
}

func (analyser *Analyser) handleBombDefused(e events.BombDefused) {
	tick := analyser.getGameTick()

	defuserID := e.Player.SteamID

	if defuser, ok := analyser.getPlayerByID(defuserID, true); ok {
		defuser.NotifyBombDefused()
		analyser.IsBombDefused = true
		analyser.log.WithFields(logging.Fields{
			"tick":    tick,
			"defuser": defuser.Name,
			"user id": defuserID,
		}).Info("Bomb has been defused: ")
	} else {
		analyser.log.WithFields(logging.Fields{
			"tick":    tick,
			"name":    defuser.Name,
			"user id": defuserID,
		}).Error("Bomb has been defused by non-exist player: ")
	}
}

func (analyser *Analyser) handleBombPlanted(e events.BombPlanted) {
	tick := analyser.getGameTick()

	planterID := e.Player.SteamID

	if planter, ok := analyser.getPlayerByID(planterID, true); ok {
		planter.NotifyBombPlanted()
		analyser.IsBombPlanted = true
		analyser.log.WithFields(logging.Fields{
			"tick":    tick,
			"planter": planter.Name,
			"user id": planterID,
		}).Info("Bomb has been planted: ")
	} else {
		analyser.log.WithFields(logging.Fields{
			"tick":    tick,
			"name":    planter.Name,
			"user id": planterID,
		}).Error("Bomb has been planted by non exist player: ")
	}
}

func (analyser *Analyser) handleDefuseStart(e events.BombDefuseStart) {
	tick := analyser.getGameTick()

	defuserID := e.Player.SteamID

	if defuser, ok := analyser.getPlayerByID(defuserID, true); ok {
		analyser.IsBombDefusing = true
		analyser.Defuser = defuser
		analyser.log.WithFields(logging.Fields{
			"tick":    tick,
			"defuser": defuser.Name,
			"user id": defuserID,
		}).Info("Player started to defuse bomb: ")
	} else {
		analyser.log.WithFields(logging.Fields{
			"tick":    tick,
			"defuser": defuser.Name,
			"user id": defuserID,
		}).Error("Non exist player started to defuse bomb: ")
	}
}

// handlePlayerFlashed handle player flashed event
func (analyser *Analyser) handlePlayerFlashed(e events.PlayerFlashed) {
	tick := analyser.getGameTick()

	// get ids
	_, _, flashedOK, flasherOK := analyser.getPlayerID(e.Player, e.Attacker, "Flash")

	// check if victim and attacker not nill
	if !flashedOK || !flasherOK {
		return
	}

	// get player pointers
	// we are checking players disconnected as well
	flashed, flasher, flashedOK, flasherOK := analyser.checkEventValidity(e.Player, e.Attacker, "Flash", true)

	// check if victim and attacker exist
	if !flashedOK || !flasherOK {
		return
	}

	// get side of players
	_, _, ok := analyser.checkTeamSideValidity(flashed, flasher, "Flash")

	if !ok {
		// if two player in the same side we need to reset last flasher
		flashed.SetLastFlashedBy(0)
		return
	}

	duration := e.FlashDuration()
	if duration <= 0 {
		analyser.log.WithFields(logging.Fields{
			"flashed":        flashed.Name,
			"attacker":       flasher.Name,
			"flash duration": duration,
			"tick":           tick,
		}).Error("Player flashed error: ")
		return
	}
	analyser.log.WithFields(logging.Fields{
		"player name":    flashed.Name,
		"tick":           tick,
		"attacker":       flasher.Name,
		"flash duration": duration,
	}).Info("Player flashed: ")
	if !analyser.isFirstParse {
		flashed.SetLastFlashedBy(flasher.SteamID)
		if flashed.Team != flasher.Team {
			flashed.NotifyBlindDuration(duration)
		}
	}

}

// *********** User defined event handlers ***************

// handleKAST notify players who did kast for this round
func (analyser *Analyser) handleKAST() {
	analyser.log.WithFields(logging.Fields{
		"tick": analyser.getGameTick(),
	}).Info("Handling KAST for players ")
	for currPlayerID, kastbool := range analyser.KastPlayers {
		if kastbool {
			player, isOK := analyser.getPlayerByID(currPlayerID, false)
			if isOK {
				player.NotifyKAST()
			}
		}
	}
}

// handleClutchSituation notify players who did a clutch for this round
func (analyser *Analyser) handleClutchSituation(winnerTS, loserTS common.Team, tick int) {
	clutchPLayer := analyser.ClutchPLayer
	// check whether player did a clutch
	// any kind of 1 to n winning count as clutch
	// clutch player may be killed but the important thing is that
	// the team of clutch player has to win
	if analyser.IsPossibleCLutch && clutchPLayer != nil {
		clutchPlayerSide := clutchPLayer.Team
		if clutchPlayerSide == winnerTS {
			clutchHappen := true
			clutchPLayer.NotifyClutchWon()

			// clutchHappen := false
			// switch clutchPlayerSide {
			// case common.TeamTerrorists:
			// 	opponentAliveNum := len(analyser.CtAlive)
			// 	if opponentAliveNum == 0 {
			// 		clutchPLayer.NotifyClutchWon()
			// 		clutchHappen = true
			// 	}
			// case common.TeamCounterTerrorists:
			// 	opponentAliveNum := len(analyser.TAlive)
			// 	if opponentAliveNum == 0 {
			// 		clutchPLayer.NotifyClutchWon()
			// 		clutchHappen = true
			// 	}
			// default:
			// 	analyser.log.WithFields(logging.Fields{
			// 		"name":    clutchPLayer.Name,
			// 		"user id": clutchPLayer.SteamID,
			// 		"tick":    tick,
			// 	}).Error("Player has no side for clutch situation: ")
			// }

			if clutchHappen {
				analyser.log.WithFields(logging.Fields{
					"name":    clutchPLayer.Name,
					"user id": clutchPLayer.SteamID,
					"tick":    tick,
				}).Info("Player did a clutch ")
			}

		}
	}
}

// handleSpecialRound handle special round won and loss
func (analyser *Analyser) handleSpecialRound(Winner, Loser []*common.Player) {
	// pistol round handling
	// only normal time
	// first round of each halfs
	if analyser.RoundPlayed <= 30 && analyser.RoundPlayed%15 == 1 {
		analyser.log.WithFields(logging.Fields{
			"winner team":        Winner[0].TeamState.ClanName,
			"loser team":         Loser[0].TeamState.ClanName,
			"special round type": "pistol_round",
			"tick":               analyser.getGameTick(),
		}).Info("Team has won a special round")
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

// **************************************************************
// ##############################################
