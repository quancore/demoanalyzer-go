package analyser

import (
	p_common "github.com/markus-wa/demoinfocs-golang/common"
	events "github.com/markus-wa/demoinfocs-golang/events"
	common "github.com/quancore/demoanalyzer-go/common"
	logging "github.com/sirupsen/logrus"
)

// ############# Seperate event handlers ########
// all event handlers for player triggered events

// dispatchPlayerEvents common function to handle a player event
// usefull for common checks for all events
// used for second time parsing
func (analyser *Analyser) dispatchPlayerEvents(e interface{}) {
	// set event happened to true for this round
	analyser.isEventHappened = true

	tick := analyser.getGameTick()
	// it is nearly impossible to a player event can happen at the same
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

	// dispatch event to its handler
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

// handleKill handler for kill event
func (analyser *Analyser) handleKill(e events.Kill) {
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
		analyser.kastPlayers[killer.SteamID] = true

		// update kill matrix
		newVictim := &common.KillTuples{Tick: tick, Player: victim}
		analyser.killedPlayers[killerID] = append(analyser.killedPlayers[killerID], newVictim)

		// check trader - tradee relationship
		// if the victim killed someone not long ago we consider it's a trade, can be refined with
		// position as well
		if killedByVictim, ok := analyser.killedPlayers[victimID]; ok {
			for _, victimKilled := range killedByVictim {
				if (tick - victimKilled.Tick) <= 640 { // 5s
					victimKilledPlayer := victimKilled.Player
					killer.NotifyTrader()
					victimKilledPlayer.NotifyTradee()
					// analyser.kastPlayers[killer.SteamID] = true
					analyser.kastPlayers[victimKilledPlayer.SteamID] = true
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
			if lastFlashedID != killerID &&
				lastFlashedPlayerSide != victimSide {
				if isVictimBlinded {
					lastFlashedPlayer.NotifyFlashAssist()
					// mark player did a kast
					analyser.kastPlayers[lastFlashedPlayer.SteamID] = true
					// didFlashAssist = true
					analyser.log.WithFields(logging.Fields{
						"tick":    tick,
						"name":    lastFlashedPlayer.Name,
						"user id": lastFlashedID,
						"victim":  victim.Name,
					}).Info("Player did a flash assist for killing: ")
				} else {
					lastValidFlashTick := victim.GetLastValidTick()
					if int(lastValidFlashTick) > tick {
						analyser.log.WithFields(logging.Fields{
							"name":          lastFlashedPlayer.Name,
							"tick":          tick,
							"lastvalidtick": int(lastValidFlashTick),
							"user id":       lastFlashedID,
							"victim":        victim.Name,
						}).Info("Possible flash assist: ")
					}

				}
			}
		}
	}

	// check normal assister
	if e.Assister != nil {
		assisterID := e.Assister.SteamID
		if assister, assisterOK := analyser.getPlayerByID(assisterID, true); assisterOK {
			// we are only checking side of players
			// not whether flash assister also did normal assist
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
				"victim":  victim.Name,
				"user id": assisterID,
			}).Info("Player did an assist for killing: ")

			assister.NotifyAssist()
			// mark player did a kast
			analyser.kastPlayers[assister.SteamID] = true
		}
	}
}

// handleHurt handler for hurt event
func (analyser *Analyser) handleHurt(e events.PlayerHurt) {
	tick := analyser.getGameTick()

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
	victim, attacker, victimOK, attackerOK := analyser.checkEventValidity(e.Player, e.Attacker, "playerHurt", true)

	if !victimOK || !attackerOK {
		return
	}

	// check is there any player waiting for round start
	// if we are in first parse and several players are waiting to join and
	// this is the first hurt event for this round, check missed players
	// has been join the game
	if analyser.isFirstParse && analyser.isPlayerWaiting && !analyser.isPlayerHurt {
		if _, _, ok := analyser.checkParticipantValidity(); ok {
			analyser.log.WithFields(logging.Fields{
				"tick": analyser.getGameTick(),
			}).Info("Late match start has been triggered with player hurt event")
			analyser.isCancelled = false
			analyser.isPlayerWaiting = false
			// call match start again
			analyser.handleMatchStart("late_match_start")
		}
	}

	// handle victim
	analyser.log.WithFields(logging.Fields{
		"tick":     tick,
		"victim":   e.Player.Name,
		"attacker": e.Attacker.Name,
		"health":   e.Health,
		"damage":   damage,
	}).Info("Player has been hurt: ")

	// debug purpose
	// if e.Health <= 0 {
	// 	// handle victim
	// 	analyser.log.WithFields(logging.Fields{
	// 		"tick":          tick,
	// 		"victim":        e.Player.Name,
	// 		"attacker":      e.Attacker.Name,
	// 		"damage":        damage,
	// 		"victim health": e.Health,
	// 	}).Info("Player has been hurt with zero health: ")
	// }

	// set a player has been hurt in this round
	analyser.isPlayerHurt = true

	// handle attacker
	victimHealth := e.Health

	// did hurt with grenade class
	if weaponType == p_common.EqClassGrenade {
		if e.Weapon.Weapon == p_common.EqHE { // he damage
			attacker.NotifyGranadeDamage(uint(damage))
		} else if e.Weapon.Weapon == p_common.EqIncendiary || e.Weapon.Weapon == p_common.EqMolotov { // inferno damage
			attacker.NotifyFireDamage(uint(damage))
		}
	}

	// check if this is an headshot
	isHeadshot := false
	if e.HitGroup == events.HitGroupHead {
		isHeadshot = true
	}

	// if damage given by a weapon
	if weaponType != p_common.EqClassUnknown {
		attacker.NotifyDamageGiven(damage, victimHealth, isHeadshot, tick)
		victim.NotifyDamageTaken(damage)
	}
}

// handleWeaponFire handler for weapon fire event
func (analyser *Analyser) handleWeaponFire(e events.WeaponFire) {
	tick := analyser.getGameTick()
	shooterID := e.Shooter.SteamID

	// determine the type of round in the first
	// player hurt event of second parsing
	if !(analyser.isWeaponFired || analyser.isFirstParse) {
		analyser.setRoundType()
	}

	analyser.isWeaponFired = true
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

// handleBombDefused handler for bomb defuse event
func (analyser *Analyser) handleBombDefused(e events.BombDefused) {
	tick := analyser.getGameTick()

	defuserID := e.Player.SteamID

	if defuser, ok := analyser.getPlayerByID(defuserID, true); ok {
		defuser.NotifyBombDefused()
		analyser.isBombDefused = true
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

// handleBombPlanted handler for bomb planted event
func (analyser *Analyser) handleBombPlanted(e events.BombPlanted) {
	tick := analyser.getGameTick()

	planterID := e.Player.SteamID

	if planter, ok := analyser.getPlayerByID(planterID, true); ok {
		planter.NotifyBombPlanted()
		analyser.isBombPlanted = true
		analyser.log.WithFields(logging.Fields{
			"tick":    tick,
			"planter": planter.Name,
			"user id": planterID,
		}).Info("Bomb has been planted: ")
	} else {
		analyser.log.WithFields(logging.Fields{
			"tick":    tick,
			"user id": planterID,
		}).Error("Bomb has been planted by non exist player: ")
	}
}

// handleDefuseStart handler for bomb defuse start event
func (analyser *Analyser) handleDefuseStart(e events.BombDefuseStart) {
	tick := analyser.getGameTick()

	defuserID := e.Player.SteamID

	if defuser, ok := analyser.getPlayerByID(defuserID, true); ok {
		analyser.isBombDefusing = true
		analyser.defuser = defuser
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
		flashed.SetLastFlashedBy(0, 0)
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
		// calculate last valid tick the flash event will be effective
		tickRate := analyser.parser.Header().TickRate()
		flashLenght := tickRate * duration.Seconds()
		lastValidTick := int64(flashLenght) + int64(tick)

		// set last flasher
		flashed.SetLastFlashedBy(flasher.SteamID, lastValidTick)
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

	for currPlayerID, kastbool := range analyser.kastPlayers {
		if kastbool {
			player, isOK := analyser.getPlayerByID(currPlayerID, false)
			if isOK {
				player.NotifyKAST()
			}
		}
	}
}

// handleClutchSituation notify players who did a clutch for this round
func (analyser *Analyser) handleClutchSituation(winnerTS p_common.Team, tick int) {
	var clutchPlayer *common.PPlayer
	var clutchHappen bool

	// get clutch players of each side
	tClutchPLayer := analyser.tClutchPlayer
	ctClutchPLayer := analyser.ctClutchPlayer

	// check whether player did a clutch
	// any kind of 1 to n winning count as clutch
	// clutch player may be killed but the important thing is that
	// the team of clutch player has to win
	if winnerTS == p_common.TeamTerrorists {
		if analyser.isTPossibleClutch && tClutchPLayer != nil {
			// if the player still alive
			if _, ok := analyser.tAlive[tClutchPLayer.SteamID]; ok {
				clutchHappen = true
				clutchPlayer = tClutchPLayer
			}
		}

	} else if winnerTS == p_common.TeamCounterTerrorists {
		if analyser.isCTPossibleClutch && ctClutchPLayer != nil {
			if _, ok := analyser.ctAlive[ctClutchPLayer.SteamID]; ok {
				clutchHappen = true
				clutchPlayer = ctClutchPLayer
			}
		}
	}

	if clutchHappen {
		clutchPlayer.NotifyClutchWon()
		analyser.log.WithFields(logging.Fields{
			"name":    clutchPlayer.Name,
			"user id": clutchPlayer.SteamID,
			"tick":    tick,
		}).Info("Player did a clutch ")
	}
}

// handleSpecialRound handle special round won and loss
func (analyser *Analyser) handleSpecialRound(winnerT, loserT p_common.Team) {
	// get team members
	gs := analyser.parser.GameState()
	participants := gs.Participants()
	winnerTeamPlayers := participants.TeamMembers(winnerT)
	loserTeamPlayers := participants.TeamMembers(loserT)
	winnerTS := gs.Team(winnerT)
	loserTS := gs.Team(loserT)

	// check team states
	if winnerTS == nil || loserTS == nil {
		return
	}

	var winnerRoundType, loserRoundType common.RoundType

	// find out winner and loser team round type
	if winnerT == p_common.TeamTerrorists {
		winnerRoundType, loserRoundType = analyser.currentTRoundType, analyser.currentCTRoundType
	} else if winnerT == p_common.TeamCounterTerrorists {
		winnerRoundType, loserRoundType = analyser.currentCTRoundType, analyser.currentTRoundType
	} else {
		analyser.log.WithFields(logging.Fields{
			"winner team": winnerT,
			"loser team":  loserT,
			"tick":        analyser.getGameTick(),
		}).Error("Invalid team type for handling round type")
		return
	}

	analyser.log.WithFields(logging.Fields{
		"winner team":   winnerTS.ClanName,
		"loser team":    loserTS.ClanName,
		"T round type":  analyser.currentTRoundType,
		"CT round type": analyser.currentCTRoundType,
		"tick":          analyser.getGameTick(),
	}).Info("Handling type of the round")

	// winner team
	for _, currPlayer := range winnerTeamPlayers {
		if NewPPlayer, ok := analyser.getPlayerByID(currPlayer.SteamID, false); ok {
			NewPPlayer.NotifySpecialRoundWon(winnerRoundType)
		}
	}
	// loser team
	for _, currPlayer := range loserTeamPlayers {
		if NewPPlayer, ok := analyser.getPlayerByID(currPlayer.SteamID, false); ok {
			NewPPlayer.NotifySpecialRoundLost(loserRoundType)
		}
	}

}

// **************************************************************
// ##############################################
