package common

import (
	"math"
	"time"

	"github.com/golang/geo/r2"
	"github.com/markus-wa/demoinfocs-golang/common"
)

// ########### type definitions ######

// RoundType base type for round types
type RoundType byte

// ##################################

// ######### constants ##############
// different round types
const (
	NormalRound   RoundType = 1
	PistolRound   RoundType = 2
	EcoRound      RoundType = 3
	ForceBuyRound RoundType = 4
)

// #################################

// ######## Common structs #########

// ItemDrop tuple to store a dropped item
type ItemDrop struct {
	Tick      int
	ItemName  string
	DropperID int64
}

// KillPosition tuple to store a kill position
type KillPosition struct {
	Tick        int
	RoundNumber int
	KillPoint   r2.Point
	VictimID    int64
	KillerID    int64
}

// KillTuples tuple struct to store kill event with tick
type KillTuples struct {
	Tick   int
	Player *PPlayer
}

// HurtTuples tuple struct to player hurt event to find out saviors
type HurtTuples struct {
	FirstHurtTick  int
	LastHurtTick   int
	playerName     string
	RemaningHealth int
}

// DisconnectedTuple tuple struct to store disconnected players
type DisconnectedTuple struct {
	DisconnectedTick int
	Player           *PPlayer
}

// RoundTuples tuple to store valid rounds after first parse
type RoundTuples struct {
	StartTick       int
	EndTick         int
	OfficialEndTick int
	// scores after this rounds end
	TScore  int
	CTScore int
}

// ##################################

// ########## common methods ########

// GetSideString get side string like T or CT using team pointer
func GetSideString(playerSide common.Team) string {
	switch playerSide {
	case common.TeamTerrorists:
		return "T"
	case common.TeamCounterTerrorists:
		return "CT"
	}

	return ""
}

// TickToSeconds convert tick duration to seconds
func TickToSeconds(Tick int, TickRate float64) time.Duration {
	// convert sec to nanoseconds
	sec := time.Duration((float64(Tick) / TickRate) * 1000000000)
	return sec
}

// SecondsToTick convert seconds duration to ticks
func SecondsToTick(Sec, TickRate float64) float64 {
	tick := Sec * TickRate
	return tick
}

// FindEuclidianDistance find distance between two 2D points
func FindEuclidianDistance(x1, y1, x2, y2 float64) float64 {
	res := math.Sqrt(math.Pow((x1-x2), 2) + math.Pow((y1-y2), 2))
	return res
}

// ################################
