package testParser

import (
	"fmt"
	"os"
	"testing"

	dem "github.com/markus-wa/demoinfocs-golang"
	events "github.com/markus-wa/demoinfocs-golang/events"
)

//TestKill test non exist flash assist situation
func TestKill(t *testing.T) {
	// change here for correct path
	f, err := os.Open("/home/baran/Desktop/demo_files/working_demos/mibr-vs-astralis-m1-overpass.dem")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p := dem.NewParser(f)

	gs := p.GameState()

	// name of specific player
	playerName := "Xyp9x"
	numKilledPLayer := 0

	p.RegisterEventHandler(func(e events.Kill) {
		if e.Killer != nil && e.Victim.TeamState == e.Killer.TeamState.Opponent {
			if e.Killer.Name == playerName {
				numKilledPLayer++
				fmt.Printf("Player killed a player: killer:%s victim:%s count: %d tick:%d\n", playerName, e.Victim.Name, numKilledPLayer, gs.IngameTick())
			}
		} else {
			fmt.Printf("Nill killer killed a player:tick:%d\n", gs.IngameTick())

		}

	})

	// Parse to end
	err = p.ParseToEnd()

	if err != nil {
		panic(err)
	}
}
