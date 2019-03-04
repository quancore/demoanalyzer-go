package analyser

import (
	"github.com/markus-wa/demoinfocs-golang/common"
	logging "github.com/sirupsen/logrus"
)

func (analyser *Analyser) testParticipant() {
	allplayers := analyser.getAllPlayers()
	tTeam, cTeam := analyser.getTeams(true)
	gs := analyser.parser.GameState()

	if nTerrorists := len(tTeam); nTerrorists < 5 {
		analyser.log.WithFields(logging.Fields{
			"terrorist number": nTerrorists,
			"team name":        gs.TeamTerrorists().ClanName,
		}).Fatal("Terrorist team has not enough participant")
	}

	if nCTs := len(cTeam); nCTs < 5 {
		analyser.log.WithFields(logging.Fields{
			"ct number": nCTs,
			"team name": gs.TeamCounterTerrorists().ClanName,
		}).Fatal("CTerrorist team has not enough participant")
	}

	for _, player := range allplayers {
		if (player.Team == common.TeamTerrorists || player.Team == common.TeamCounterTerrorists) &&
			player.GetNumKills() <= 0 && player.GetNumDeaths() <= 0 {
			analyser.log.WithFields(logging.Fields{
				"name":  player.Name,
				"kill":  player.GetNumKills(),
				"death": player.GetNumDeaths(),
			}).Fatal("Player has wrong stats")
		}
	}

	analyser.log.Info("All participant test has succesfully passed")
}

func (analyser *Analyser) testGameState() {

	// for a valid match finish, at least 16 round has to played
	if analyser.RoundPlayed < 16 {
		analyser.log.WithFields(logging.Fields{
			"terrorist score":  analyser.Tscore,
			"cterrorist score": analyser.CTscore,
			"round played":     analyser.RoundPlayed,
		}).Fatal("Played round has wrong")
	}

	if !((analyser.Tscore + analyser.CTscore) == analyser.RoundPlayed) {
		analyser.log.WithFields(logging.Fields{
			"terrorist score":  analyser.Tscore,
			"cterrorist score": analyser.CTscore,
			"round played":     analyser.RoundPlayed,
		}).Fatal("Played round number is not equal to sum of team scores")
	}

	if analyser.Tscore < 0 || analyser.CTscore < 0 {
		analyser.log.WithFields(logging.Fields{
			"terrorist score":  analyser.Tscore,
			"cterrorist score": analyser.CTscore,
		}).Fatal("Scores has wrong")
	}

	// if there is a win it is needed to be at least one team
	// has reach at least 16
	if analyser.Tscore < 16 && analyser.CTscore < 16 {
		analyser.log.WithFields(logging.Fields{
			"terrorist score":  analyser.Tscore,
			"cterrorist score": analyser.CTscore,
		}).Fatal("Match result has wrong")
	}

	if matchEnded, _ := analyser.checkMatchEnd(analyser.Tscore, analyser.CTscore); !matchEnded {
		analyser.log.WithFields(logging.Fields{
			"terrorist score":  analyser.Tscore,
			"cterrorist score": analyser.CTscore,
		}).Fatal("Match is not ended")
	}

	analyser.log.Info("All game state test has succesfully passed")

}
