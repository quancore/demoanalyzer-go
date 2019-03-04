package testParser

import (
	"fmt"
	"os"
	"testing"

	dem "github.com/markus-wa/demoinfocs-golang"
	"github.com/markus-wa/demoinfocs-golang/common"
	events "github.com/markus-wa/demoinfocs-golang/events"
)

//TestParticipant test non exist flash assist situation
func TestParticipant(t *testing.T) {
	// change here for correct path
	f, err := os.Open("/home/baran/Desktop/demo_files/working_demos/spirit-vs-movistar-riders-nuke.dem")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p := dem.NewParser(f)

	gs := p.GameState()

	p.RegisterEventHandler(func(e events.PlayerTeamChange) {
		tick := gs.IngameTick()

		changedPlayer := e.Player
		oldTeam := e.OldTeam
		newTeam := e.NewTeam
		// uid := changedPlayer.SteamID

		if (newTeam == common.TeamSpectators || newTeam == common.TeamUnassigned) &&
			(oldTeam == common.TeamTerrorists || oldTeam == common.TeamCounterTerrorists) {
			fmt.Printf("Playing player become unactive.Name: %s, new team: %d, old team:%d, tick:%d\n",
				changedPlayer.Name, newTeam, oldTeam, tick)
		}
	})

	p.RegisterEventHandler(func(e events.MatchStart) {
		tick := gs.IngameTick()

		fmt.Printf("Match is starting.tick:%d\n", tick)
		printParticipant(p)
	})

	// Parse to end
	err = p.ParseToEnd()

	if err != nil {
		panic(err)
	}
}

func printParticipant(p *dem.Parser) {
	gs := p.GameState()
	participants := gs.Participants()
	teamTerrorist := participants.TeamMembers(common.TeamTerrorists)
	teamCT := participants.TeamMembers(common.TeamCounterTerrorists)
	fmt.Printf("Terrorist team number :%d, member names: \n", len(teamTerrorist))
	for _, tplayer := range teamTerrorist {
		fmt.Printf("Name: %s \n", tplayer.Name)
	}

	fmt.Printf("CT team number :%d, member names: \n", len(teamCT))
	for _, Ctplayer := range teamCT {
		fmt.Printf("Name: %s \n", Ctplayer.Name)
	}
}
