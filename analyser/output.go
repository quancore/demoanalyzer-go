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
			"name":                            currPlayer.Name,
			"team":                            currPlayer.TeamState.ClanName,
			"team number":                     currPlayer.Team,
			"kill":                            currPlayer.GetNumKills(),
			"first kill":                      currPlayer.GetNumFirstKills(),
			"parser kill":                     currPlayer.Player.AdditionalPlayerInformation.Kills,
			"blind kill":                      currPlayer.GetBlindKills(),
			"blinded player killed":           currPlayer.GetPlayerBlindedKills(),
			"hs kll":                          currPlayer.GetNumHSKills(),
			"assist":                          currPlayer.GetNumAssists(),
			"parser assist":                   currPlayer.Player.AdditionalPlayerInformation.Assists,
			"flash assist":                    currPlayer.GetFlashAssist(),
			"death":                           currPlayer.GetNumDeaths(),
			"parser death":                    currPlayer.Player.AdditionalPlayerInformation.Deaths,
			"clutch won":                      currPlayer.GetClutchWon(),
			"pistol won":                      currPlayer.GetPistolRoundWon(),
			"eco won":                         currPlayer.GetEcoRoundWon(),
			"force buy won":                   currPlayer.GetForceBuyRoundWon(),
			"granade damage":                  currPlayer.GetGranadeDamage(),
			"fire damage":                     currPlayer.GetFireDamage(),
			"time flashing":                   currPlayer.GetTimeFlashing(),
			"kast":                            currPlayer.GetKAST(),
			"num trader":                      currPlayer.GetNumTrader(),
			"num tradee":                      currPlayer.GetNumTradee(),
			"bomb defused":                    currPlayer.GetNumBombDefused(),
			"bomb planted":                    currPlayer.GetNumBombPlanted(),
			"num MVP":                         currPlayer.GetMVP(),
			"total money saved":               currPlayer.GetSavedMoney(),
			"kill sniper":                     currPlayer.GetSRifleKills(),
			"kill melee":                      currPlayer.GetMeleeKills(),
			"kill shotgun":                    currPlayer.GetShotgunKills(),
			"kill assult":                     currPlayer.GetARifleeKills(),
			"kill pistol":                     currPlayer.GetPistolKills(),
			"kill machine gun":                currPlayer.GetMachineGunKills(),
			"kill smg":                        currPlayer.GetSmgKills(),
			"hit head":                        currPlayer.GetNumHeadHit(),
			"hit stomach":                     currPlayer.GetNumStomachHit(),
			"hit chest":                       currPlayer.GetNumChestHit(),
			"hit legs":                        currPlayer.GetNumLegsHit(),
			"hit arms":                        currPlayer.GetNumArmsHit(),
			"unit damage cost":                (currPlayer.GetTotalDamageCost() / float32(currPlayer.GetTotalDamage())),
			"avarage kill distance":           currPlayer.GetTotalKillDistance() / float32(currPlayer.GetNumKills()),
			"player saved":                    currPlayer.GetSavedNum(),
			"player won health:":              currPlayer.GetTotalHealthWon(),
			"player lost health:":             currPlayer.GetTotalHealthLost(),
			"number alive rounds":             currPlayer.GetLastMemberSurvived(),
			"time hurt to kill":               currPlayer.GetTimeHurtToKill().Seconds(),
			"spray sniper":                    currPlayer.GetSRifleSpray(),
			"spray shotgun":                   currPlayer.GetShotgunSpray(),
			"spray assult":                    currPlayer.GetARifleSpray(),
			"spray pistol":                    currPlayer.GetPistolSpray(),
			"spray machine gun":               currPlayer.GetMachineGunSpray(),
			"spray smg":                       currPlayer.GetSmgSpray(),
			"round win percentage":            currPlayer.GetRoundWinPercentage(),
			"total round duration":            currPlayer.GetTotalRoundWinTime(),
			"duck kill":                       currPlayer.GetDuckKill(),
			"total distance to killed member": currPlayer.GetTotalKillDistance(),
			"killed snipers":                  currPlayer.GetNumSniperKill(),
			"dropped item val":                currPlayer.GetDroppedItemVal(),
			"picked item val":                 currPlayer.GetPickedItemVal(),
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
	mapnameAlias := viper.GetStringMapString("mapnameAlias")
	roundPlayed := analyser.roundPlayed

	// if test needed for output
	istestrequired := viper.GetBool("checkanalyzer")
	mapname := analyser.mapName
	analyser.log.WithFields(logging.Fields{
		"name": mapname,
	}).Info("Mapname: ")

	if newMapname, ok := mapnameAlias[mapname]; ok {
		analyser.log.WithFields(logging.Fields{
			"new name": newMapname,
			"old name": mapname,
		}).Info("Mapname changed: ")
		mapname = newMapname
	}
	if istestrequired {
		analyser.testGameState()
		analyser.testParticipant()
	}
	w.WriteString(fmt.Sprintf("analyzer_version=%s, mapname=%s, round_played=%d", analyzerVersion, mapname, roundPlayed))
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

		winLabel := 0
		// if there is equality or player team won set to 1
		if teamWon == common.TeamUnassigned || currPlayer.Team == teamWon {
			winLabel = 1
		}
		sb = currPlayer.OutputPlayerState(sb, analyser.roundPlayed, winLabel)

	}

	w.WriteString(sb.String())
	w.Flush()
	sb.Reset()

}
