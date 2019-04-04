package analyser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	common "github.com/markus-wa/demoinfocs-golang/common"
	utils "github.com/quancore/demoanalyzer-go/utils"
	logging "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

const (
	specifier = ","
	// if special char is in a string
	// it will replace with space
	replaceWithSpace = true
)

// ############## Printer / writers #############
// printPlayers print player stats after finished the match
func (analyser *Analyser) printPlayers() {
	analyser.log.Info("#########################################")

	teamWon := analyser.getWinnerTeam()
	gs := analyser.parser.GameState()

	analyser.log.WithFields(logging.Fields{
		"t score":      analyser.tScore,
		"ct score":     analyser.ctScore,
		"winner team":  gs.Team(teamWon).ClanName,
		"played round": analyser.roundPlayed,
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
			"team number":           currPlayer.Team,
			"kill":                  currPlayer.GetNumKills(),
			"parser kill":           currPlayer.Player.AdditionalPlayerInformation.Kills,
			"blind kill":            currPlayer.GetBlindKills(),
			"blinded player killed": currPlayer.GetPlayerBlindedKills(),
			"hs kll":                currPlayer.GetNumHSKills(),
			"assist":                currPlayer.GetNumAssists(),
			"parser assist":         currPlayer.Player.AdditionalPlayerInformation.Assists,
			"flash assist":          currPlayer.GetFlashAssist(),
			"death":                 currPlayer.GetNumDeaths(),
			"parser death":          currPlayer.Player.AdditionalPlayerInformation.Deaths,
			"clutch won":            currPlayer.GetClutchWon(),
			"pistol won":            currPlayer.GetPistolRoundWon(),
			"eco won":               currPlayer.GetEcoRoundWon(),
			"force buy won":         currPlayer.GetForceBuyRoundWon(),
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
}

// printPlayers print player stats after finished the match
func (analyser *Analyser) writeToFile(path string) {
	file, err := os.Create(path)
	utils.CheckError(err)
	w := bufio.NewWriter(file)
	var sb strings.Builder
	defer file.Close()
	features := viper.GetString("output.features")
	analyzerVersion := viper.GetString("output.analyzer_version")

	// if test needed for output
	istestrequired := viper.GetBool("checkanalyzer")
	mapname := analyser.mapName
	if istestrequired {
		analyser.testGameState()
		analyser.testParticipant()
	}
	w.WriteString(fmt.Sprintf("analyzer_version=%s, mapname=%s", analyzerVersion, mapname))
	w.WriteByte('\n')
	w.Flush()
	w.WriteString(features)
	w.WriteByte('\n')
	w.Flush()
	gs := analyser.parser.GameState()

	teamWon := analyser.getWinnerTeam()

	analyser.log.WithFields(logging.Fields{
		"t score":      analyser.tScore,
		"ct score":     analyser.ctScore,
		"played round": analyser.roundPlayed,
		"winner team":  gs.Team(teamWon).ClanName,
		"writing path": path,
	}).Info("Writing to file: ")

	for _, currPlayer := range analyser.getAllPlayers() {
		// if currPlayer.TeamState == nil {
		// 	analyser.log.WithFields(logging.Fields{
		// 		"name": currPlayer.Name,
		// 		"team": currPlayer.Team,
		// 	}).Info("Team state is null for player ")
		// 	continue
		if !(analyser.checkTeamValidity(currPlayer.Team)) {
			analyser.log.WithFields(logging.Fields{
				"name":     currPlayer.Name,
				"team":     currPlayer.Team,
				"old team": currPlayer.GetOldTeam(),
			}).Info("Player team and old team is wrong ")
			continue
		} else if currPlayer.GetNumKills() <= 0 && currPlayer.GetNumDeaths() <= 0 {
			analyser.log.WithFields(logging.Fields{
				"name":      currPlayer.Name,
				"team":      currPlayer.Team,
				"num kill":  currPlayer.GetNumKills(),
				"num death": currPlayer.GetNumDeaths(),
			}).Info("Player has wrong stat ")
			continue
		}
		// teamState := gs.Team(currPlayer.Team)
		roundPlayed := float32(analyser.roundPlayed)
		playerName := currPlayer.Name
		if replaceWithSpace {
			playerName = strings.Replace(playerName, specifier, " ", -1)
		}
		sb.WriteString(fmt.Sprintf("%s%s", playerName, specifier))

		var pistolRoundWonPercentage float32
		pistolROundsWon := float32(currPlayer.GetPistolRoundWon())
		pistolROundsLost := float32(currPlayer.GetPistolRoundLost())
		if (pistolROundsWon + pistolROundsLost) > 0 {
			pistolRoundWonPercentage = pistolROundsWon / (pistolROundsWon + pistolROundsLost)
		}
		sb.WriteString(fmt.Sprintf("%s%s", fmt.Sprintf("%.2f", pistolRoundWonPercentage), specifier))

		var hsPercentage float32
		hsKills := float32(currPlayer.GetNumHSKills())
		totalKIlls := float32(currPlayer.GetNumKills())
		if totalKIlls > 0 {
			hsPercentage = hsKills / totalKIlls
		}
		sb.WriteString(fmt.Sprintf("%s%s", fmt.Sprintf("%.2f", hsPercentage), specifier))

		clutchesWon := fmt.Sprintf("%.2f", float32(currPlayer.GetClutchWon())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", clutchesWon, specifier))

		adr := fmt.Sprintf("%.2f", float32(currPlayer.GetTotalDamage())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", adr, specifier))

		fpr := fmt.Sprintf("%.2f", float32(currPlayer.GetNumKills())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", fpr, specifier))

		apr := fmt.Sprintf("%.2f", float32(currPlayer.GetNumAssists())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", apr, specifier))

		kdDiff := fmt.Sprintf("%.2f", (float32(currPlayer.GetNumKills())-float32(currPlayer.GetNumDeaths()))/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", kdDiff, specifier))

		flashAssist := fmt.Sprintf("%.2f", float32(currPlayer.GetFlashAssist())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", flashAssist, specifier))

		blindPlayerKilled := fmt.Sprintf("%.2f", float32(currPlayer.GetPlayerBlindedKills())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", blindPlayerKilled, specifier))

		blindKills := fmt.Sprintf("%.2f", float32(currPlayer.GetBlindKills())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", blindKills, specifier))

		granedaDamage := fmt.Sprintf("%.2f", float32(currPlayer.GetGranadeDamage())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", granedaDamage, specifier))

		fireDamage := fmt.Sprintf("%.2f", float32(currPlayer.GetFireDamage())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", fireDamage, specifier))

		timeFlashingOpponent := fmt.Sprintf("%.2f", float32(currPlayer.GetTimeFlashing().Seconds())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", timeFlashingOpponent, specifier))

		accuracy := float32(currPlayer.GetShotsHit()) / float32(currPlayer.GetShots())
		accuracyStr := fmt.Sprintf("%.2f", accuracy)
		sb.WriteString(fmt.Sprintf("%s%s", accuracyStr, specifier))

		numTrader := fmt.Sprintf("%.2f", float32(currPlayer.GetNumTrader())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", numTrader, specifier))

		numTradee := fmt.Sprintf("%.2f", float32(currPlayer.GetNumTradee())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", numTradee, specifier))

		kast := fmt.Sprintf("%.2f", float32(currPlayer.GetKAST())/roundPlayed)
		sb.WriteString(fmt.Sprintf("%s%s", kast, specifier))
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
