package analyser

import (
	"github.com/markus-wa/demoinfocs-golang/common"
	logging "github.com/sirupsen/logrus"
)

// testParticipant test participant counts and individual stats
func (analyser *Analyser) testParticipant() {
	allplayers := analyser.getAllPlayers()
	var numActiveT, numActiveCT int

	gs := analyser.parser.GameState()

	for _, player := range allplayers {
		if player.GetNumKills() > 0 && player.GetNumDeaths() > 0 {
			analyser.log.WithFields(logging.Fields{
				"name":     player.Name,
				"team":     player.Team,
				"old team": player.GetOldTeam(),
			}).Info("Checking player")
			if player.Team == common.TeamTerrorists {
				numActiveT++
			} else if player.Team == common.TeamCounterTerrorists {
				numActiveCT++
			}
		} else {
			if player.Team == common.TeamTerrorists || player.Team == common.TeamCounterTerrorists {
				analyser.log.WithFields(logging.Fields{
					"name":  player.Name,
					"kill":  player.GetNumKills(),
					"death": player.GetNumDeaths(),
				}).Error("Player has wrong stats")
			}

		}
	}

	if numActiveT < 5 {
		analyser.log.WithFields(logging.Fields{
			"terrorist number": numActiveT,
			"team name":        gs.TeamTerrorists().ClanName,
		}).Fatal("Terrorist team has not enough participant")
	}

	if numActiveCT < 5 {
		analyser.log.WithFields(logging.Fields{
			"ct number": numActiveCT,
			"team name": gs.TeamCounterTerrorists().ClanName,
		}).Fatal("CTerrorist team has not enough participant")
	}

	analyser.log.Info("All participant test has succesfully passed")
}

// testGameState test game state, played round etc.
func (analyser *Analyser) testGameState() {

	// for a valid match finish, at least 16 round has to played
	if analyser.roundPlayed < 16 {
		analyser.log.WithFields(logging.Fields{
			"terrorist score":  analyser.tScore,
			"cterrorist score": analyser.ctScore,
			"round played":     analyser.roundPlayed,
		}).Fatal("Played round has wrong")
	}

	if !((analyser.tScore + analyser.ctScore) == analyser.roundPlayed) {
		analyser.log.WithFields(logging.Fields{
			"terrorist score":  analyser.tScore,
			"cterrorist score": analyser.ctScore,
			"round played":     analyser.roundPlayed,
		}).Fatal("Played round number is not equal to sum of team scores")
	}

	if analyser.tScore < 0 || analyser.ctScore < 0 {
		analyser.log.WithFields(logging.Fields{
			"terrorist score":  analyser.tScore,
			"cterrorist score": analyser.ctScore,
		}).Fatal("Scores has wrong")
	}

	// if there is a win it is needed to be at least one team
	// has reach at least 16
	if analyser.tScore < 16 && analyser.ctScore < 16 {
		analyser.log.WithFields(logging.Fields{
			"terrorist score":  analyser.tScore,
			"cterrorist score": analyser.ctScore,
		}).Fatal("Match result has wrong")
	}

	if matchEnded, _ := analyser.checkMatchEnd(analyser.tScore, analyser.ctScore); !matchEnded {
		analyser.log.WithFields(logging.Fields{
			"terrorist score":  analyser.tScore,
			"cterrorist score": analyser.ctScore,
		}).Fatal("Match is not ended")
	}

	analyser.log.Info("All game state test has succesfully passed")

}
