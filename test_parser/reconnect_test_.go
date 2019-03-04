package testParser

import (
	"fmt"
	"os"
	"testing"

	dem "github.com/markus-wa/demoinfocs-golang"
	events "github.com/markus-wa/demoinfocs-golang/events"
)

//TestDisconnectConnect test reconnection situation
func TestDisconnectConnect(t *testing.T) {
	f, err := os.Open("/home/baran/Desktop/demo_files/natus-vincere-vs-vitality-mirage.dem")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p := dem.NewParser(f)

	gs := p.GameState()

	var reconnectedPlayerID int64
	// var reconnectedPlayer *common.Player

	p.RegisterEventHandler(func(e events.PlayerDisconnected) {
		if e.Player != nil {
			fmt.Printf("Player has been disconnected: name:%s team:%s tick:%d\n", e.Player.Name, e.Player.TeamState.ClanName, gs.IngameTick())
			reconnectedPlayerID = e.Player.SteamID
		}
	})

	// p.RegisterEventHandler(func(e events.Kill) {
	// 	fmt.Printf("Player killed a player: attacker: %s team:%d tick:%d\n", e.Killer.Name, e.Killer.Team, gs.IngameTick())
	// 	if e.Killer != nil && reconnectedPlayerID == e.Killer.SteamID {
	// 		fmt.Printf("Player killed a player: attacker: %s team:%s tick:%d\n", e.Killer.Name, e.Killer.TeamState.ClanName, gs.IngameTick())
	// 	}
	// })

	p.RegisterEventHandler(func(e events.PlayerConnect) {
		if e.Player != nil && reconnectedPlayerID == e.Player.SteamID {
			fmt.Printf("Player has been reconnected: name:%s team:%d tick:%d\n", e.Player.Name, e.Player.Team, gs.IngameTick())
			// reconnectedPlayer = e.Player
		}
	})

	// Parse to end
	err = p.ParseToEnd()
	if err != nil {
		panic(err)
	}
}
