// Package analyser package to handle events and update
// player statistics
package analyser

import (
	"bytes"
	"io"

	"github.com/gogo/protobuf/proto"
	dem "github.com/markus-wa/demoinfocs-golang"
	"github.com/markus-wa/demoinfocs-golang/msg"
	utils "github.com/quancore/demoanalyzer-go/common"
	player "github.com/quancore/demoanalyzer-go/player"
	logging "github.com/sirupsen/logrus"
)

// ########### Constants #######################
// csgo match constants
const (
	maxRounds           = 30
	normalTimeWinRounds = maxRounds/2 + 1
)

// ############################################
// ######## Structs ###########################

// tuple struct to store kill event with tick
type killTuples struct {
	Tick   int
	Player *player.PPlayer
}

// tuple struct to store kill event with tick
type disconnectedTuple struct {
	DisconnectedTick int
	ReconnectedTick  int //reconnection of disconnected playre if exist
	Player           *player.PPlayer
	// when player reconnected, its player struct in parser is not
	// updated.This is problematical because if the teams has been swictched
	// after half end, the disconnected player struct will not be update and
	// it leads to incorrect side and player name.So we are keep a player team name
	// to identify the side of the player when it is reconnected
	// teamName string
}

// tuple to store valid rounds after first parse
type roundTuples struct {
	startTick       int
	endTick         int
	officialEndTick int
	// scores after this rounds end
	tScore  int
	ctScore int
}

// MatchVars mainly seperated for serialization purposes
type MatchVars struct {
	// **********match related variables *********
	// map container to store currently connected Players
	Players map[int64]*player.PPlayer
	// map container to store currently disconnected players
	// it is usefull for reconnection of a player
	DisconnectedPlayers map[int64]*disconnectedTuple
	// store reconnected players who has a unassigned team or spectator
	pendingPlayers map[int64]*disconnectedTuple
	// scores
	Tscore      int
	CTscore     int
	RoundPlayed int
	// number of rounds will be played as overtime
	NumOvertime int
	// flags
	// match started flag
	MatchStarted bool
	// match started flag
	MatchEnded bool
	// flag indicating overtime is playing
	IsOvertime bool
	// flag indicating score is swapped
	ScoreSwapped bool
	// new t score after swap.temp value for swap rollback situation.
	swapTscore int
	// new ct score after swap.temp value for swap rollback situation.
	swapCTscore int
	// min number of rounds played not to reset a match
	minPlayedRound int
	// round number lastly score swappped
	lastScoreSwapped int
	// sometimes match start event called twice so
	// we need to avoid duble call on same tick
	lastMatchStartedCalled int
	// round number last round end called
	// used for double call on score update and round end for handling kast and clutches
	lastRoundEndCalled int

	// **********round related variables*******
	// kill map to store pointers to killed players for each player in this round
	// killer id : {killed tick, pointer to victim}
	// // TODO: check consistency
	KilledPlayers map[int64][]*killTuples
	// map container to store alive ct players for each round
	CtAlive map[int64]*player.PPlayer
	// map container to store alive t players for each round
	TAlive map[int64]*player.PPlayer
	// keep round based player did kast
	KastPlayers map[int64]bool
	//store player get disconnected during a round
	// usefull to dispatch events belongs to this players
	// disconnectedLastRound map[int64]*disconnectedTuple

	// player pointer possible to make a clutch
	ClutchPLayer *player.PPlayer
	// player currently defusing the bomb
	Defuser *player.PPlayer
	// current starting money
	currentSMoney float64
	// flag indicate money has been set at least one
	isMoneySet bool
	// flag for bomb defused
	IsBombDefused bool
	// flag for bomb planted
	IsBombPlanted bool
	// flag for bomb planted
	IsBombDefusing bool
	// clutch situation flag
	IsPossibleCLutch bool

	// variables to check validity of a round
	//Truth value for whether we're currently in a round
	InRound bool
	// Truth value whether this round is valid
	IsCancelled bool
	// Truth value whether a player has been hurt
	IsPlayerHurt bool
	// Truth value to check an event occured during a round
	IsEventHappened bool
}

// Analyser main struct to analyse events
type Analyser struct {
	// header related information
	mapName string
	// composition of serialized fields
	MatchVars
	// main parser to parse demofile
	parser *dem.Parser
	// logger (converted struct var for concurent logging)
	log *logging.Logger

	// demo file stream
	demostream io.Reader
	// output path for stat writing
	outPath string
	// buffer
	buf *bytes.Buffer
	// config of parser
	cfg dem.ParserConfig

	// ******* parsing related vars *******
	// first parser flag
	isFirstParse bool
	// store round start tick
	roundStart int
	// store round end tick
	roundEnd int
	// store round end tick
	roundOffEnd int
	// store valid rounds
	// round number: (start tick , end tick, scores)
	validRounds map[int]*roundTuples
	// current valid round tuple
	curValidRound *roundTuples
	// current start and end of the rounds
	// ************************************

	// ***********Serialization**************
	// flag indicating whether there is a serialized match
	matchEncoded bool
	// store old played round number
	oldPlayedROund int
	// store t score
	oldTScore int
	// store ct score
	oldCTScore int
	// store old valid rounds
	oldValidRounds map[int]*roundTuples
	// flags
	// store old score swapped
	oldIsScoreSwapped bool
}

// #########################################
// ######## serialization ##########################
// // saveState save valid rounds and count of played rounds
// func (analyser *Analyser) saveState() {
// 	analyser.log.WithFields(logging.Fields{
// 		"tick": analyser.getGameTick(),
// 	}).Info("A state has been saved")
// 	// Create the target map
// 	analyser.oldValidRounds = make(map[int]*roundTuples)
//
// 	// Copy from the original map to the target map
// 	for key, value := range analyser.validRounds {
// 		analyser.oldValidRounds[key] = value
// 	}
// 	analyser.oldPlayedROund = analyser.RoundPlayed
// 	analyser.oldTScore = analyser.Tscore
// 	analyser.oldCTScore = analyser.CTscore
// 	analyser.oldIsScoreSwapped = analyser.ScoreSwapped
// 	analyser.matchEncoded = true
// }
//
// // loadState load a match state
// func (analyser *Analyser) loadState() {
// 	analyser.log.WithFields(logging.Fields{
// 		"tick": analyser.getGameTick(),
// 	}).Info("A state has been loaded")
//
// 	if analyser.matchEncoded {
// 		// Create the target map
// 		analyser.validRounds = make(map[int]*roundTuples)
//
// 		// Copy from the original map to the target map
// 		for key, value := range analyser.oldValidRounds {
// 			analyser.validRounds[key] = value
// 		}
// 		analyser.RoundPlayed = analyser.oldPlayedROund
// 		analyser.Tscore = analyser.oldTScore
// 		analyser.CTscore = analyser.oldCTScore
// 		analyser.ScoreSwapped = analyser.oldIsScoreSwapped
// 		analyser.matchEncoded = false
// 	} else {
// 		analyser.log.Error("There is no match to decode...")
// 	}
//
// }

// #################################################
// ######## public interface #######################

// NewAnalyser constructer for getting an analyser
func NewAnalyser(demostream io.Reader, logPath, outPath string, ismethodname, multiplewriter bool) *Analyser {

	// parser := dem.NewParser(demostream)
	// Configure parsing of ConVar net-message (id=6)
	cfg := dem.DefaultParserConfig
	cfg.AdditionalNetMessageCreators = map[int]dem.NetMessageCreator{
		6: func() proto.Message {
			return new(msg.CNETMsg_SetConVar)
		},
	}
	var buf bytes.Buffer
	// analyser.demostream = io.TeeReader(demostream, &analyser.buf)
	demostream = io.TeeReader(demostream, &buf)

	parser := dem.NewParserWithConfig(demostream, cfg)

	analyser := &Analyser{MatchVars: MatchVars{}, parser: parser}
	analyser.buf = &buf
	analyser.cfg = cfg
	analyser.outPath = outPath
	analyser.log = utils.InitLogger(logPath, ismethodname, multiplewriter)
	analyser.log.Info("Analyser has been created")

	analyser.validRounds = make(map[int]*roundTuples)

	analyser.resetAnalyserVars()

	// // register net message handlers
	// analyser.registerNetMessageHandlers()

	return analyser

}

// HandleHeader handle header information an initilize related variables
func (analyser *Analyser) HandleHeader() {
	analyser.log.Info("Parsing header of demo file")
	// Parse header
	header, err := analyser.parser.ParseHeader()
	utils.CheckError(err)
	analyser.mapName = header.MapName

	analyser.log.WithFields(logging.Fields{
		"server name": header.ServerName,
		"client name": header.ClientName,
		"map name":    header.MapName,
		"tick rate":   header.TickRate(),
	}).Info("Several fields of header: ")
}

// Analyze parse demofile
func (analyser *Analyser) Analyze() {
	// first handle parser header
	analyser.HandleHeader()
	analyser.log.Info("Analyzing first time")
	analyser.isFirstParse = true
	analyser.registerNetMessageHandlers()
	analyser.registerMatchEventHandlers()
	for hasMoreFrames, err := true, error(nil); hasMoreFrames; hasMoreFrames, err = analyser.parser.ParseNextFrame() {
		utils.CheckError(err)
	}

	analyser.log.Info("Analyzing second time")
	analyser.isFirstParse = false
	analyser.resetAnalyser()
	// analyser.registerNetMessageHandlers()
	analyser.registerMatchEventHandlers()
	analyser.registerEventHandlers()
	// beacuse there is no event sounting as match start on
	// second parsing, we are calling reset match vars externally
	// on the beginning of second parser
	analyser.resetMatchVars()

	// because net messages handler is not registered, we can use
	// parsetoend
	err := analyser.parser.ParseToEnd()
	utils.CheckError(err)
}
