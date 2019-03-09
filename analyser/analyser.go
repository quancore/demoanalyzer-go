// Package analyser package to handle events and update player statistics
package analyser

import (
	"bytes"
	"io"

	"github.com/gogo/protobuf/proto"
	dem "github.com/markus-wa/demoinfocs-golang"
	p_common "github.com/markus-wa/demoinfocs-golang/common"
	"github.com/markus-wa/demoinfocs-golang/msg"
	common "github.com/quancore/demoanalyzer-go/common"
	utils "github.com/quancore/demoanalyzer-go/utils"
	logging "github.com/sirupsen/logrus"
)

// ########### Constants #######################
// csgo match constants
const (
	maxRounds           = 30
	normalTimeWinRounds = maxRounds/2 + 1
)

// ############################################

// ############# structs ######################

// Analyser main struct to analyse events
type Analyser struct {
	// **********match related variables *********
	// map container to store currently connected players
	players map[int64]*common.PPlayer
	// map container to store currently disconnected players
	// it is usefull for reconnection of a player
	disconnectedPlayers map[int64]*common.DisconnectedTuple
	// scores
	tScore  int
	ctScore int
	// played round number
	roundPlayed int
	// number of rounds will be played as overtime
	NumOvertime int
	// flags
	// match started flag
	matchStarted bool
	// match started flag
	matchEnded bool
	// flag indicating overtime is playing
	isOvertime bool
	// flag indicating score is swapped
	scoreSwapped bool
	// new t score after swap.temp value for swap rollback situation.
	swapTscore int
	// new ct score after swap.temp value for swap rollback situation.
	swapCTscore int
	// min number of rounds played not to reset a match
	minPlayedRound int
	// current round type for t team
	currentTRoundType common.RoundType
	// current round type for ct team
	currentCTRoundType common.RoundType
	// round number lastly score swappped
	lastScoreSwapped int
	// sometimes match start event called twice so
	// we need to avoid duble call on same tick
	lastMatchStartedCalled int
	// round number last round end called
	// used for double call on score update and round end for handling kast and clutches
	lastRoundEndCalled int
	// *************************

	// **********round related variables*******
	// kill map to store pointers to killed players for each player in this round
	// killer id : {killed tick, pointer to victim}
	killedPlayers map[int64][]*common.KillTuples
	// map container to store alive ct players for each round
	ctAlive map[int64]*common.PPlayer
	// map container to store alive t players for each round
	tAlive map[int64]*common.PPlayer
	// keep round based player did kast
	kastPlayers map[int64]bool

	// t player pointer possible to make a clutch
	tClutchPlayer *common.PPlayer
	// ct clutch player
	ctClutchPlayer *common.PPlayer
	// winner team for the last round
	winnerTeam p_common.Team
	// player currently defusing the bomb
	defuser *common.PPlayer
	// current starting money
	currentSMoney float64
	// flag indicate money has been set at least one
	isMoneySet bool
	// flag for bomb defused
	isBombDefused bool
	// flag for bomb planted
	isBombPlanted bool
	// flag for bomb defusing
	isBombDefusing bool
	// clutch situation flag for t
	isTPossibleClutch bool
	// clutch situation flag for ct
	isCTPossibleClutch bool

	// variables to check validity of a round
	//Truth value for whether we're currently in a round
	inRound bool
	// Truth value whether this round is valid
	isCancelled bool
	// Truth value whether a player has been hurt
	isPlayerHurt bool
	// Truth value for a weapon fired for this round
	isWeaponFired bool
	// Truth value to check an event occured during a round
	isEventHappened bool
	// Sometimes, several player has missed the match start but
	// has been connected after match start and before first player hurt event
	// so we will se the flag to wait several player join and trigger start
	// according to status of this flag
	isPlayerWaiting bool
	// ************************

	// ******** parser related vars ***
	// header related information
	mapName string
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
	// ************************************

	// ******* parsing related vars *******
	// first parser flag
	isFirstParse bool
	// flag indicating whole analyze finished and
	// written to disk as a test file
	isSuccesfulAnalyzed bool
	// store round start tick
	roundStart int
	// store round end tick
	roundEnd int
	// store round end tick
	roundOffEnd int
	// store valid rounds
	// round number: (start tick , end tick, scores)
	validRounds map[int]*common.RoundTuples
	// current valid round tuple
	curValidRound *common.RoundTuples
	// ************************************
}

// ######## public interface #######################

// NewAnalyser constructer for getting an analyser
func NewAnalyser(demostream io.Reader, logPath, outPath string, ismethodname, multiplewriter bool) *Analyser {
	// Configure parsing of ConVar net-message (id=6)
	cfg := dem.DefaultParserConfig
	cfg.AdditionalNetMessageCreators = map[int]dem.NetMessageCreator{
		6: func() proto.Message {
			return new(msg.CNETMsg_SetConVar)
		},
	}
	var buf bytes.Buffer
	// buffer demo file for second parsing
	demostream = io.TeeReader(demostream, &buf)

	parser := dem.NewParserWithConfig(demostream, cfg)

	analyser := &Analyser{parser: parser}
	analyser.buf = &buf
	analyser.cfg = cfg
	analyser.outPath = outPath
	analyser.log = utils.InitLogger(logPath, ismethodname, multiplewriter)
	analyser.log.Info("Analyser has been created")

	// create map to store valid rounds
	analyser.validRounds = make(map[int]*common.RoundTuples)

	analyser.resetAnalyserVars()

	return analyser

}

// handleHeader handle header information an initilize related variables
func (analyser *Analyser) handleHeader() {
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
	analyser.handleHeader()
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
	analyser.registerMatchEventHandlers()
	analyser.registerPlayerEventHandlers()

	// beacuse there is no event counting as match start on
	// second parsing, we are calling reset match vars externally
	// on the beginning of second parser
	analyser.resetMatchVars()

	// because net messages handler is not registered, we can use
	// parsetoend, so that we are not protecting sync betwenn net messages
	// and event dispatch
	err := analyser.parser.ParseToEnd()

	// sometimes demo files enden unexpectedly however, it is not important
	// if we already finished the analyze
	if err == dem.ErrUnexpectedEndOfDemo && analyser.isSuccesfulAnalyzed {
		analyser.log.Info("Demo file ended unexpectedly however, analze has been finished")
	} else {
		utils.CheckError(err)
	}
}
