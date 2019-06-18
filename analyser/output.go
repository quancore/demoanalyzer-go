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
	teamWonState := gs.Team(teamWon)

	tScore, ctScore := analyser.tScore, analyser.ctScore

	analyser.log.WithFields(logging.Fields{
		"t score":             tScore,
		"ct score":            ctScore,
		"winner team":         teamWonState.ClanName,
		"played round":        analyser.roundPlayed,
		"round winner string": analyser.createRoundString(teamWonState.ClanName),
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
			"name":                               currPlayer.Name,
			"team":                               currPlayer.TeamState.ClanName,
			"team number":                        currPlayer.Team,
			"player team score":                  currPlayer.GetPlayerScore(tScore, ctScore),
			"kill":                               currPlayer.GetNumKills(),
			"first kill":                         currPlayer.GetNumFirstKills(),
			"parser kill":                        currPlayer.Player.AdditionalPlayerInformation.Kills,
			"blind kill":                         currPlayer.GetBlindKills(),
			"blinded player killed":              currPlayer.GetPlayerBlindedKills(),
			"hs kll":                             currPlayer.GetNumHSKills(),
			"assist":                             currPlayer.GetNumAssists(),
			"parser assist":                      currPlayer.Player.AdditionalPlayerInformation.Assists,
			"flash assist":                       currPlayer.GetFlashAssist(),
			"death":                              currPlayer.GetNumDeaths(),
			"parser death":                       currPlayer.Player.AdditionalPlayerInformation.Deaths,
			"clutch won":                         currPlayer.GetClutchWon(),
			"pistol won":                         currPlayer.GetPistolRoundWon(),
			"eco won":                            currPlayer.GetEcoRoundWon(),
			"force buy won":                      currPlayer.GetForceBuyRoundWon(),
			"granade damage":                     currPlayer.GetGranadeDamage(),
			"fire damage":                        currPlayer.GetFireDamage(),
			"time flashing":                      currPlayer.GetTimeFlashing(),
			"kast":                               currPlayer.GetKAST(),
			"num trader":                         currPlayer.GetNumTrader(),
			"num tradee":                         currPlayer.GetNumTradee(),
			"bomb defused":                       currPlayer.GetNumBombDefused(),
			"bomb planted":                       currPlayer.GetNumBombPlanted(),
			"num MVP":                            currPlayer.GetMVP(),
			"total money saved":                  currPlayer.GetSavedMoney(),
			"kill sniper":                        currPlayer.GetSRifleKills(),
			"kill melee":                         currPlayer.GetMeleeKills(),
			"kill shotgun":                       currPlayer.GetShotgunKills(),
			"kill assult":                        currPlayer.GetARifleeKills(),
			"kill pistol":                        currPlayer.GetPistolKills(),
			"kill machine gun":                   currPlayer.GetMachineGunKills(),
			"kill smg":                           currPlayer.GetSmgKills(),
			"hit head":                           currPlayer.GetNumHeadHit(),
			"hit stomach":                        currPlayer.GetNumStomachHit(),
			"hit chest":                          currPlayer.GetNumChestHit(),
			"hit legs":                           currPlayer.GetNumLegsHit(),
			"hit arms":                           currPlayer.GetNumArmsHit(),
			"unit damage cost":                   (currPlayer.GetTotalDamageCost() / float32(currPlayer.GetTotalDamage())),
			"avarage kill distance":              currPlayer.GetTotalKillDistance() / float32(currPlayer.GetNumKills()),
			"player saved":                       currPlayer.GetSavedNum(),
			"player won health:":                 currPlayer.GetTotalHealthWon(),
			"player lost health:":                currPlayer.GetTotalHealthLost(),
			"number alive rounds":                currPlayer.GetLastMemberSurvived(),
			"time hurt to kill":                  currPlayer.GetTimeHurtToKill().Seconds(),
			"spray sniper":                       currPlayer.GetSRifleSpray(),
			"spray shotgun":                      currPlayer.GetShotgunSpray(),
			"spray assult":                       currPlayer.GetARifleSpray(),
			"spray pistol":                       currPlayer.GetPistolSpray(),
			"spray machine gun":                  currPlayer.GetMachineGunSpray(),
			"spray smg":                          currPlayer.GetSmgSpray(),
			"round win percentage":               currPlayer.GetRoundWinPercentage(),
			"total round duration":               currPlayer.GetTotalRoundWinTime(),
			"duck kill":                          currPlayer.GetDuckKill(),
			"lurker kill":                        currPlayer.GetLurkerKill(),
			"total distance to killed opponents": currPlayer.GetTotalKillDistance(),
			"killed members distance":            currPlayer.GetKilledMemberDistance(),
			"killed snipers":                     currPlayer.GetNumSniperKill(),
			"dropped item val":                   currPlayer.GetDroppedItemVal(),
			"picked item val":                    currPlayer.GetPickedItemVal(),
			"total occupied area":                currPlayer.GetTeamOccupiedArea(),
		}).Info("Player: ")
	}
}

// createRoundString create each round winners represented in a string
// zero represent won rounds by the winner of the match, one otherwise.
func (analyser *Analyser) createRoundString(winnerTeamName string) string {
	var sb strings.Builder
	roundPlayed := analyser.roundPlayed
	for roundNum := 1; roundNum <= roundPlayed; roundNum++ {
		if roundWinnerName, ok := analyser.roundWinners[roundNum]; ok {
			var roundWinnerBit int
			if winnerTeamName != roundWinnerName {
				roundWinnerBit = 1
			}

			sb.WriteString(fmt.Sprintf("%s|", fmt.Sprint(roundWinnerBit)))
		}

	}
	return sb.String()
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

	gs := analyser.parser.GameState()

	teamWon := analyser.getWinnerTeam()
	roundString := analyser.createRoundString(gs.Team(teamWon).ClanName)

	w.WriteString(fmt.Sprintf("analyzer_version=%s, mapname=%s, round_played=%d, round_string=%s",
		analyzerVersion, mapname, roundPlayed, roundString))
	w.WriteByte('\n')
	w.Flush()
	w.WriteString(features)
	w.WriteByte('\n')
	w.Flush()

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
			}).Info("Player team or old team is wrong ")
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
		sb = currPlayer.OutputPlayerState(sb, analyser.roundPlayed, winLabel, analyser.tScore, analyser.ctScore)

	}

	w.WriteString(sb.String())
	w.Flush()
	sb.Reset()

}
