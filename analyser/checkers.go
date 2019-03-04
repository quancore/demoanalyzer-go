package analyser

import (
	"github.com/markus-wa/demoinfocs-golang/common"
	"github.com/markus-wa/demoinfocs-golang/events"
	utils "github.com/quancore/demoanalyzer-go/common"
	"github.com/quancore/demoanalyzer-go/player"
	logging "github.com/sirupsen/logrus"
)

// ######## Internal checkers###
// checker are methods to check validity of rounds and matches

// checkPlayerValidity check whether a player is connected during
// event handling
func (analyser *Analyser) checkPlayerValidity(e interface{}) bool {
	switch e.(type) {
	case events.Kill:
		killEvent := e.(events.Kill)
		// get ids
		_, _, victimOK, killerOK := analyser.getPlayerID(killEvent.Victim, killEvent.Killer, "Kill")

		// check if victim and attacker exist in the event
		if !victimOK || !killerOK {
			return false
		}
		// get player pointers
		_, _, victimOK, killerOK = analyser.checkEventValidity(killEvent.Victim, killEvent.Killer, "Kill", false)

		// check if victim and attacker exist
		if !victimOK || !killerOK {
			return false
		}

		if killEvent.Assister != nil {
			if _, assisterOK := analyser.getPlayerByID(killEvent.Assister.SteamID, false); !assisterOK {
				return false
			}
		}

	case events.PlayerHurt:
		playerHurtEvent := e.(events.PlayerHurt)

		// get ids
		_, _, victimOK, killerOK := analyser.getPlayerID(playerHurtEvent.Player, playerHurtEvent.Attacker, "playerHurt")

		// check if victim and attacker exist in the event
		if !victimOK || !killerOK {
			return false
		}
		// get player pointers
		_, _, victimOK, attackerOK := analyser.checkEventValidity(playerHurtEvent.Player, playerHurtEvent.Attacker, "playerHurt", false)

		if !victimOK || !attackerOK {
			return false
		}

	case events.WeaponFire:
		weaponFireEvent := e.(events.WeaponFire)
		if _, ok := analyser.getPlayerByID(weaponFireEvent.Shooter.SteamID, false); !ok {
			return false
		}
	case events.BombDefuseStart:
		bombDefuseStartEvent := e.(events.BombDefuseStart)

		if _, ok := analyser.getPlayerByID(bombDefuseStartEvent.Player.SteamID, false); !ok {
			return false
		}

	case events.BombDefused:
		bombDefusedEvent := e.(events.BombDefused)

		if _, ok := analyser.getPlayerByID(bombDefusedEvent.Player.SteamID, false); !ok {
			return false
		}

	case events.BombPlanted:
		bombPlantedEvent := e.(events.BombPlanted)
		if _, ok := analyser.getPlayerByID(bombPlantedEvent.Player.SteamID, false); !ok {
			return false
		}

	case events.PlayerFlashed:

		playerFlashedEvent := e.(events.PlayerFlashed)

		// get ids
		_, _, victimOK, killerOK := analyser.getPlayerID(playerFlashedEvent.Player, playerFlashedEvent.Attacker, "Flash")

		// check if victim and attacker exist in the event
		if !victimOK || !killerOK || playerFlashedEvent.FlashDuration() <= 0 {
			return false
		}
		// get player pointers
		_, _, victimOK, killerOK = analyser.checkEventValidity(playerFlashedEvent.Player, playerFlashedEvent.Attacker, "Flash", false)

		// check if victim and attacker exist
		if !victimOK || !killerOK {
			return false
		}
	}

	return true
}

// checkEventValidity check validity of ids of given attacker and victim IDs
// return player pointers
// suitable for use of kill and player hurt events
func (analyser *Analyser) checkEventValidity(victim, killer *common.Player, eventName string, allPlayer bool) (*player.PPlayer, *player.PPlayer, bool, bool) {
	var victimOK, killerOK bool
	var victimP, killerP *player.PPlayer

	// get player pointers
	victimID := victim.SteamID
	killerID := killer.SteamID
	victimP, victimOK = analyser.getPlayerByID(victimID, allPlayer)
	killerP, killerOK = analyser.getPlayerByID(killerID, allPlayer)

	// check if victim and attacker exist
	if !victimOK || !killerOK {
		analyser.log.WithFields(logging.Fields{
			"event name":  eventName,
			"tick":        analyser.getGameTick(),
			"victim ok":   victimOK,
			"attacker ok": killerOK,
			"victim name": victim.Name,
			"killer name": killer.Name,
		}).Error("Victim or attacker is undefined in map for an event: ")

		return nil, nil, victimOK, killerOK
	}

	return victimP, killerP, victimOK, killerOK

}

// checkMatchValidity return true if match started and not yet finished
func (analyser *Analyser) checkMatchValidity() bool {
	tick := analyser.getGameTick()

	if analyser.MatchEnded || !analyser.MatchStarted || tick < 0 {
		return false
	}
	return true
}

// checkIsMatchValid return true if there are certain number of round played
// during this match.
// usefull for ignoring match start events after a valid match has already started
func (analyser *Analyser) checkIsMatchValid() bool {
	if analyser.RoundPlayed > analyser.minPlayedRound {
		return true
	}

	return false
}

// checkFinishedMatchValidity check round end reason to validate it is a valid one
func (analyser *Analyser) checkFinishedRoundValidity(e events.RoundEnd) bool {
	reason := e.Reason
	if reason == events.RoundEndReasonCTSurrender || reason == events.RoundEndReasonDraw {
		return false
	}
	return true
}

// checkMatchContinuity check whether match is continuing with overtime
func (analyser *Analyser) checkMatchContinuity() bool {
	tick := analyser.getGameTick()
	ctScore := analyser.CTscore
	tScore := analyser.Tscore
	var isMatchEnded bool

	if analyser.isFirstParse {
		if isMatchEnded, analyser.IsOvertime = analyser.checkMatchEnd(tScore, ctScore); isMatchEnded {
			if !analyser.MatchEnded {
				analyser.log.WithFields(logging.Fields{
					"total round played": analyser.RoundPlayed,
					"tick":               tick,
					"team terrorist":     tScore,
					"team ct terrorist":  ctScore,
				}).Info("Match is over. ")
				// analyser.printPlayers()
			}
			analyser.MatchEnded = isMatchEnded
		}
	} else {
		// check match is ended
		if !analyser.MatchEnded && analyser.RoundPlayed == len(analyser.validRounds) {
			analyser.log.WithFields(logging.Fields{
				"total round played": analyser.RoundPlayed,
				"tick":               tick,
				"team terrorist":     tScore,
				"team ct terrorist":  ctScore,
			}).Info("Match is over. ")
			analyser.printPlayers()
			analyser.writeToFile(analyser.outPath)
			analyser.IsOvertime = false
			analyser.MatchEnded = true
		} else {
			analyser.IsOvertime = true

		}
	}

	return analyser.IsOvertime
}

// checkMatchEnd check whether match should end for given scores
func (analyser *Analyser) checkMatchEnd(tScore, ctScore int) (bool, bool) {
	mpOvertimeMaxrounds := analyser.NumOvertime
	nOvertimeRounds := ctScore + tScore - maxRounds
	var matchOver, isovertime bool

	if ((ctScore == normalTimeWinRounds) != (tScore == normalTimeWinRounds)) || nOvertimeRounds >= 0 {
		// a team won in normal time or at least 30 rounds have been played
		absDiff := utils.Abs(ctScore - tScore)
		x := nOvertimeRounds % mpOvertimeMaxrounds
		nRoundsOfHalf := mpOvertimeMaxrounds / 2
		if nOvertimeRounds < 0 || ((x == 0 && absDiff == 2) || (x > nRoundsOfHalf && absDiff >= nRoundsOfHalf)) {
			matchOver = true
		}
		isovertime = true
	}

	return matchOver, isovertime
}

// checkClutchSituation check alive players for a clutch situation
func (analyser *Analyser) checkClutchSituation() {
	countALiveT := len(analyser.TAlive)
	countALiveCT := len(analyser.CtAlive)

	if !analyser.IsPossibleCLutch {
		// possible clutch for ct
		if countALiveT > 1 && countALiveCT == 1 {
			analyser.IsPossibleCLutch = true
			for _, playerPtr := range analyser.CtAlive {
				analyser.ClutchPLayer = playerPtr
				analyser.log.WithFields(logging.Fields{
					"name": playerPtr.Name,
				}).Info("Possible clutch player ")
			}

		} else if countALiveCT > 1 && countALiveT == 1 { // possible clutch for t
			analyser.IsPossibleCLutch = true
			for _, playerPtr := range analyser.TAlive {
				analyser.ClutchPLayer = playerPtr
				analyser.log.WithFields(logging.Fields{
					"name": playerPtr.Name,
				}).Error("Possible clutch player ")
			}
		}
	}
}

// checkTeamSideValidity check validity of players with respect to their teams
// return player pointers
// suitable for use of kill and player hurt events
func (analyser *Analyser) checkTeamSideValidity(victim, killer *player.PPlayer, eventName string) (string, string, bool) {
	tick := analyser.getGameTick()
	// get side of players
	victimSide, vSideOK := victim.GetSide()
	killerSide, KSideOK := killer.GetSide()

	if !vSideOK || !KSideOK {
		analyser.log.WithFields(logging.Fields{
			"victim side":   victimSide,
			"attacker side": killerSide,
			"attacker name": killer.Name,
			"victim name":   victim.Name,
			"event name":    eventName,
			"tick":          tick,
		}).Error("Victim or attacker has no side: ")

		return victimSide, killerSide, false

	} else if victimSide == killerSide {
		analyser.log.WithFields(logging.Fields{
			"victim side":   victimSide,
			"attacker side": killerSide,
			"attacker name": killer.Name,
			"victim name":   victim.Name,
			"event name":    eventName,
			"tick":          tick,
		}).Error("Victim and attacker is the same side: ")
		return victimSide, killerSide, false
	}

	return victimSide, killerSide, true

}

// checkParticipantValidity check number of participant for each team
// return teams for each side
// usefull to check before a match begins
func (analyser *Analyser) checkParticipantValidity() ([]*common.Player, []*common.Player, bool) {
	// sometimes, at tick 0, players are getting connected after matchstarted
	// so there is a special case for tick 0
	if analyser.getGameTick() == 0 {
		return nil, nil, true
	}
	// first get players
	nTerrorists, nCTs := 5, 5
	gs := analyser.parser.GameState()
	participants := gs.Participants()
	teamTerrorist := participants.TeamMembers(common.TeamTerrorists)
	teamCT := participants.TeamMembers(common.TeamCounterTerrorists)
	// all := participants.All()
	// players := participants.Playing()

	// check participants number etc
	if nTerrorists != len(teamTerrorist) || nCTs != len(teamCT) {
		// We know there should be 5 terrorists at match start in the default demo
		return teamTerrorist, teamCT, false
	}

	return teamTerrorist, teamCT, true
}

// checkMoneyValidity check starting money of each round
func (analyser *Analyser) checkMoneyValidity() bool {
	// if the money value is not set, no need to check
	if !analyser.isMoneySet {
		return true
	}
	// normal time half starts
	if analyser.RoundPlayed == 0 || analyser.RoundPlayed == 15 {
		if analyser.currentSMoney != 800 {
			return false
		}
	} else if analyser.RoundPlayed >= maxRounds { //overtime
		ctScore := analyser.CTscore
		tScore := analyser.Tscore
		mpOvertimeMaxrounds := analyser.NumOvertime
		nOvertimeRounds := ctScore + tScore - maxRounds
		nRoundsOfHalf := mpOvertimeMaxrounds / 2
		if nOvertimeRounds%nRoundsOfHalf == 0 {
			if analyser.currentSMoney != 16000 {
				return false
			}
		}
	}

	return true
}

// checkHalfBreak return true for if we are in half break
func (analyser *Analyser) checkHalfBreak(tScore, ctScore int) bool {
	RoundPlayed := ctScore + tScore
	mpOvertimeMaxrounds := analyser.NumOvertime
	nOvertimeRounds := RoundPlayed - maxRounds
	nRoundsOfHalf := mpOvertimeMaxrounds / 2

	// normal time
	if nOvertimeRounds <= 0 {
		if RoundPlayed == 15 || (RoundPlayed == 30 && analyser.MatchEnded == false) {
			return true
		}
	} else { //overtime
		if nOvertimeRounds%nRoundsOfHalf == 0 {
			return true
		}
	}

	return false
}

// checkPlayerTeamValidity check whether a player is assigned a valid team
func (analyser *Analyser) checkPlayerTeamValidity(NewPlayer *common.Player) bool {
	if NewPlayer.IsBot || NewPlayer.Team == common.TeamSpectators || NewPlayer.Team == common.TeamUnassigned {
		return false
	}
	return true
}

// #####################################
