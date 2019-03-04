package analyser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	common "github.com/markus-wa/demoinfocs-golang/common"
	utils "github.com/quancore/demoanalyzer-go/common"
	logging "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

// ############## Printer / writers #############
// printPlayers print player stats after finished the match
func (analyser *Analyser) printPlayers() {
	analyser.log.Info("#########################################")

	analyser.log.WithFields(logging.Fields{
		"t score":      analyser.Tscore,
		"ct score":     analyser.CTscore,
		"played round": analyser.RoundPlayed,
	}).Info("Match has been finished: ")
	for _, currPlayer := range analyser.getAllPlayers() {
		if currPlayer.TeamState == nil {
			analyser.log.WithFields(logging.Fields{
				"name": currPlayer.Name,
				"team": currPlayer.Team,
			}).Info("Team state is null for player ")
			continue
		}
		analyser.log.Info("**************************************")
		analyser.log.Info(currPlayer.Name)
		analyser.log.WithFields(logging.Fields{
			"name":                  currPlayer.Name,
			"team":                  currPlayer.TeamState.ClanName,
			"kill":                  currPlayer.GetNumKills(),
			"blind kill":            currPlayer.GetBlindKills(),
			"blinded player killed": currPlayer.GetPlayerBlindedKills(),
			"hs kll":                currPlayer.GetNumHSKills(),
			"assist":                currPlayer.GetNumAssists(),
			"flash assist":          currPlayer.GetFlashAssist(),
			"death":                 currPlayer.GetNumDeaths(),
			"clutch won":            currPlayer.GetClutchWon(),
			"pistol won":            currPlayer.GetPistolRoundWon(),
			"granade damage":        currPlayer.GetGranadeDamage(),
			"fire damage":           currPlayer.GetFireDamage(),
			"time flashing":         currPlayer.GetTimeFlashing(),
			"kast":                  currPlayer.GetKAST(),
			"num trader":            currPlayer.GetNumTrader(),
			"num tradee":            currPlayer.GetNumTradee(),
			"bomb defused":          currPlayer.GetNumBombDefused(),
			"bomb planted":          currPlayer.GetNumBombPlanted(),
		}).Info("Player: ")
	}

	// os.Exit(0)
}

// printPlayers print player stats after finished the match
func (analyser *Analyser) writeToFile(path string) {
	file, err := os.Create(path)
	utils.CheckError(err)
	w := bufio.NewWriter(file)
	var sb strings.Builder
	defer file.Close()
	features := viper.GetString("output.features")

	// if test needed for output
	istestrequired := viper.GetBool("checkanalyzer")
	if istestrequired {
		analyser.testGameState()
		analyser.testParticipant()
	}

	w.WriteString(features)
	w.WriteByte('\n')
	w.Flush()
	gs := analyser.parser.GameState()
	// get which team won
	teamWon := common.TeamUnassigned
	if analyser.Tscore > analyser.CTscore {
		teamWon = common.TeamTerrorists
	} else if analyser.Tscore < analyser.CTscore {
		teamWon = common.TeamCounterTerrorists
	}

	analyser.log.WithFields(logging.Fields{
		"t score":      analyser.Tscore,
		"ct score":     analyser.CTscore,
		"played round": analyser.RoundPlayed,
		"winner team":  gs.Team(teamWon).ClanName,
		"writing path": path,
	}).Info("Writing to file: ")

	for _, currPlayer := range analyser.getAllPlayers() {
		if currPlayer.TeamState == nil {
			analyser.log.WithFields(logging.Fields{
				"name": currPlayer.Name,
				"team": currPlayer.Team,
			}).Info("Team state is null for player ")
			continue
		}
		// teamState := gs.Team(currPlayer.Team)
		roundPlayed := float32(analyser.RoundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", currPlayer.Name))

		var pistolRoundWonPercentage float32
		pistolROundsWon := float32(currPlayer.GetPistolRoundWon())
		pistolROundsLost := float32(currPlayer.GetPistolRoundLost())
		if (pistolROundsWon + pistolROundsLost) > 0 {
			pistolRoundWonPercentage = pistolROundsWon / (pistolROundsWon + pistolROundsLost)
		}
		sb.WriteString(fmt.Sprintf("%s,", fmt.Sprintf("%.2f", pistolRoundWonPercentage)))

		var hsPercentage float32
		hsKills := float32(currPlayer.GetNumHSKills())
		totalKIlls := float32(currPlayer.GetNumKills())
		if totalKIlls > 0 {
			hsPercentage = hsKills / totalKIlls
		}
		sb.WriteString(fmt.Sprintf("%s,", fmt.Sprintf("%.2f", hsPercentage)))

		clutchesWon := fmt.Sprintf("%.2f", float32(currPlayer.GetClutchWon())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", clutchesWon))

		adr := fmt.Sprintf("%.2f", float32(currPlayer.GetTotalDamage())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", adr))

		fpr := fmt.Sprintf("%.2f", float32(currPlayer.GetNumKills())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", fpr))

		apr := fmt.Sprintf("%.2f", float32(currPlayer.GetNumAssists())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", apr))

		kdDiff := fmt.Sprintf("%.2f", (float32(currPlayer.GetNumKills())-float32(currPlayer.GetNumDeaths()))/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", kdDiff))

		flashAssist := fmt.Sprintf("%.2f", float32(currPlayer.GetFlashAssist())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", flashAssist))

		blindPlayerKilled := fmt.Sprintf("%.2f", float32(currPlayer.GetPlayerBlindedKills())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", blindPlayerKilled))

		blindKills := fmt.Sprintf("%.2f", float32(currPlayer.GetBlindKills())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", blindKills))

		granedaDamage := fmt.Sprintf("%.2f", float32(currPlayer.GetGranadeDamage())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", granedaDamage))

		fireDamage := fmt.Sprintf("%.2f", float32(currPlayer.GetFireDamage())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", fireDamage))

		timeFlashingOpponent := fmt.Sprintf("%.2f", float32(currPlayer.GetTimeFlashing().Seconds())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", timeFlashingOpponent))

		accuracy := float32(currPlayer.GetShotsHit()) / float32(currPlayer.GetShots())
		accuracyStr := fmt.Sprintf("%.2f", accuracy)
		sb.WriteString(fmt.Sprintf("%s,", accuracyStr))

		numTrader := fmt.Sprintf("%.2f", float32(currPlayer.GetNumTrader())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", numTrader))

		numTradee := fmt.Sprintf("%.2f", float32(currPlayer.GetNumTradee())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", numTradee))

		kast := fmt.Sprintf("%.2f", float32(currPlayer.GetKAST())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s,", kast))
		winLabel := 0
		// if there is equality or player team won set to 1
		if teamWon == common.TeamUnassigned || currPlayer.Team == teamWon {
			winLabel = 1
		}
		sb.WriteString(fmt.Sprintf("%s", fmt.Sprint(winLabel)))
		w.WriteString(sb.String())
		w.WriteByte('\n')
		w.Flush()
		sb.Reset()

	}

	// os.Exit(0)

}
