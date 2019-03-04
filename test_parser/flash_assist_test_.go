package testParser

import (
	"fmt"
	"os"
	"testing"

	dem "github.com/markus-wa/demoinfocs-golang"
	events "github.com/markus-wa/demoinfocs-golang/events"
)

type flasherTuple struct {
	name    string
	steamID int64
}

//TestFlashAssist test non exist flash assist situation
func TestFlashAssist(t *testing.T) {
	// change here for correct path
	f, err := os.Open("/home/baran/Desktop/demo_files/gunrunners-vs-izako-boars-m1-inferno.dem")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p := dem.NewParser(f)

	gs := p.GameState()

	// store flashed player id: flasher tuple pairs
	var lastFlashedPlayer map[int64]flasherTuple
	// store flasher name: flash assist counts
	var flashAssistCounts map[string]int
	// name of specific player
	playerName := "GunrunnersDEV7L"

	lastFlashedPlayer = make(map[int64]flasherTuple)
	flashAssistCounts = make(map[string]int)

	p.RegisterEventHandler(func(e events.PlayerFlashed) {
		if e.Attacker.Team != e.Player.Team && playerName == e.Attacker.Name {
			fmt.Printf("Player has been flashed: flasher:%s flashed:%s tick:%d\n", e.Attacker.Name, e.Player.Name, gs.IngameTick())
			newFlasherTuple := flasherTuple{name: e.Attacker.Name, steamID: e.Attacker.SteamID}
			lastFlashedPlayer[e.Player.SteamID] = newFlasherTuple
		} else {
			delete(lastFlashedPlayer, e.Player.SteamID)
		}

	})

	p.RegisterEventHandler(func(e events.Kill) {
		if e.Victim != nil && e.Killer != nil && e.Victim.TeamState == e.Killer.TeamState.Opponent {
			if lastFlasherTuple, ok := lastFlashedPlayer[e.Victim.SteamID]; ok {
				if lastFlasherTuple.name == playerName && lastFlasherTuple.steamID != e.Killer.SteamID {
					// if e.Victim.IsBlinded() && lastFlasherTuple.steamID != e.Killer.SteamID {
					fmt.Printf("Player did an flash assist: name:%s victim:%s tick:%d\n", lastFlasherTuple.name, e.Victim.Name, gs.IngameTick())
					fmt.Printf("Remaning flash time: %f\n", e.Victim.FlashDuration)
					flashAssistCounts[lastFlasherTuple.name]++
				}
			}
		}

	})

	// Parse to end
	err = p.ParseToEnd()

	// print players and flash assist counts
	for flasherName, count := range flashAssistCounts {
		fmt.Printf("Player name:%s assist count:%d\n", flasherName, count)
	}

	if err != nil {
		panic(err)
	}
}
