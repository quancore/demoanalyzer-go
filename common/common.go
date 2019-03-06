package common

import "github.com/markus-wa/demoinfocs-golang/common"

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

// KillTuples tuple struct to store kill event with tick
type KillTuples struct {
	Tick   int
	Player *PPlayer
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

// ################################
