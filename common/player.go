// Package common includes common definitions and methods
package common

import (
	"time"

	player "github.com/markus-wa/demoinfocs-golang/common"
	log "github.com/sirupsen/logrus"
)

// PPlayer is an abstraction struct on top of parser player struct.
type PPlayer struct {
	// base struct added in composition
	*player.Player
	// The number of pistol rounds won by the player
	pistolRoundWon uint
	// The number of pistol rounds lost by the player
	pistolRoundslost uint
	// The number of eco rounds won by the player
	ecoRoundsWon uint
	// The number of eco rounds lost by the player
	ecoRoundsLost uint
	// The number of force rounds won by the player
	forceRoundsWon uint
	// The number of force rounds lost by the player
	forceRoundslost uint
	// The number of clutches won by the player
	clutchesWon uint
	// The number of players killed while they were blinded
	blindPlayersKilled uint
	// The number of players killed while blinded
	blindKills uint
	// The steam id of the player that last flashed the current one
	lastFlashedBy int64
	// The tick value the flash event will be valid
	lastValidTick int64
	// The number of rounds played by this player
	numRoundsPlayed uint
	// The tick number of the last weapon fire event, 0 if it hasn't happened
	lastWfTick int
	// The tick that player killed
	lastKilledTick int
	// The number of shots hit
	shotsHit uint
	// The number of trades done
	numTrader uint
	// The number of time being the tradee
	numTradee uint
	// The number of kast for the player
	kast uint

	// The amount of time the player blinded other players
	timeFlashingOpponents time.Duration
	// Blindness duration from the flashbang currently affecting the player (seconds)
	flashDuration time.Duration
	// The number of headshots done
	hs uint
	// The number of headshot kills done
	hsKill uint
	// The number of kill done
	kill uint
	// The number of death done
	death uint
	// The number of assist done
	assist uint
	// The number of shots fired
	shots uint
	// The amount of grenade damage given
	heDmg uint
	// The amount of fire damage given
	fireDmg uint
	// Total damage given
	totalDmg uint
	// Total damage taken
	totalDmgTaken uint
	// The number of flash assists, i.e. the number of kills done by teamates that this player blinded
	flashAssists uint

	// The number of bomb defused
	bombsDefused uint
	// The number of bomb defused
	bombsPlanted uint
}

// NewPPlayer creates a *PPlayer with an *Player struct.
func NewPPlayer(player *player.Player) *PPlayer {
	return &PPlayer{Player: player}
}

// GetUserID get user id
func (p *PPlayer) GetUserID() int { return p.UserID }

// GetUserName get user name
func (p *PPlayer) GetUserName() string { return p.Name }

// GetSteamID get steam id
func (p *PPlayer) GetSteamID() int64 { return p.SteamID }

// SetUserID get user id
func (p *PPlayer) SetUserID(newuserid int) { p.UserID = newuserid }

// SetUserName set user name
func (p *PPlayer) SetUserName(newusername string) { p.Name = newusername }

// SetLastFlashedBy set the id of the player who flashed this player
func (p *PPlayer) SetLastFlashedBy(attackerID int64, lastvalidTick int64) {
	p.lastFlashedBy = attackerID
	p.lastValidTick = lastvalidTick
}

// SwapTeam swap teams
func (p *PPlayer) SwapTeam() {
	// TODO: implement later
}

// GetNumKills get number of kill
func (p *PPlayer) GetNumKills() uint { return p.kill }

// GetNumHSKills get number of hs kill
func (p *PPlayer) GetNumHSKills() uint { return p.hsKill }

// GetNumDeaths get number of death
func (p *PPlayer) GetNumDeaths() uint { return p.death }

// GetNumAssists get number of assist
func (p *PPlayer) GetNumAssists() uint { return p.assist }

// GetNumBombDefused get number of bomb has been defused
func (p *PPlayer) GetNumBombDefused() uint { return p.bombsDefused }

// GetNumBombPlanted get number of bomb has been planted
func (p *PPlayer) GetNumBombPlanted() uint { return p.bombsPlanted }

// GetPistolRoundWon get number of pistol round won
func (p *PPlayer) GetPistolRoundWon() uint { return p.pistolRoundWon }

// GetPistolRoundLost get number of pistol round lost
func (p *PPlayer) GetPistolRoundLost() uint { return p.pistolRoundslost }

// GetEcoRoundWon get number of eco round won
func (p *PPlayer) GetEcoRoundWon() uint { return p.ecoRoundsWon }

// GetEcoRoundLost get number of eco round lost
func (p *PPlayer) GetEcoRoundLost() uint { return p.ecoRoundsLost }

// GetForceBuyRoundWon get number of force buy round won
func (p *PPlayer) GetForceBuyRoundWon() uint { return p.forceRoundsWon }

// GetForceBuyRoundLost get number of force buy round lost
func (p *PPlayer) GetForceBuyRoundLost() uint { return p.forceRoundslost }

// GetBlindKills get number of kills while kiiler was blinded
func (p *PPlayer) GetBlindKills() uint { return p.blindKills }

// GetPlayerBlindedKills get number of player killed while blinded
func (p *PPlayer) GetPlayerBlindedKills() uint { return p.blindPlayersKilled }

// GetClutchWon get number of bomb has been planted
func (p *PPlayer) GetClutchWon() uint { return p.clutchesWon }

// GetFlashAssist get number of flash assist
func (p *PPlayer) GetFlashAssist() uint { return p.flashAssists }

// GetTotalDamage get total damage given
func (p *PPlayer) GetTotalDamage() uint { return p.totalDmg }

// GetGranadeDamage get granade damage given
func (p *PPlayer) GetGranadeDamage() uint { return p.heDmg }

// GetFireDamage get fire damage given
func (p *PPlayer) GetFireDamage() uint { return p.fireDmg }

// GetLastFlashedBy get the id of the player who flashed this player
func (p *PPlayer) GetLastFlashedBy() int64 { return p.lastFlashedBy }

// GetLastValidTick get the last valid tick of last flash event
func (p *PPlayer) GetLastValidTick() int64 { return p.lastValidTick }

// GetTimeFlashing get time flashing opponent
func (p *PPlayer) GetTimeFlashing() time.Duration { return p.timeFlashingOpponents }

// GetShots get shots have done
func (p *PPlayer) GetShots() uint { return p.shots }

// GetShotsHit get shots hit have done
func (p *PPlayer) GetShotsHit() uint { return p.shotsHit }

// GetKAST get kast have done
func (p *PPlayer) GetKAST() uint { return p.kast }

// GetNumTrader get number of trader
func (p *PPlayer) GetNumTrader() uint { return p.numTrader }

// GetNumTradee get number of tradee
func (p *PPlayer) GetNumTradee() uint { return p.numTradee }

//GetCurrentEqValue get current eq. value
func (p *PPlayer) GetCurrentEqValue() int { return p.Player.CurrentEquipmentValue }

// GetSide get the side of this player
func (p *PPlayer) GetSide() (player.Team, bool) {
	teamOk := true

	if p.Team == player.TeamUnassigned || p.Team == player.TeamSpectators {
		teamOk = false
	}

	return p.Team, teamOk
}

// ***************event notification *******************

// NotifyKill handle event of killing
func (p *PPlayer) NotifyKill(IsHeadshot, VictemBlinded bool) {
	p.kill++

	// hs kill
	if IsHeadshot {
		p.hsKill++
		// this kill is hs as well
		p.hs++
	}
	// killed a player while this player
	// has been blinded
	if p.IsBlinded() {
		p.blindKills++
	}
	// killed a blinded player
	if VictemBlinded {
		p.blindPlayersKilled++
	}
}

// NotifyDeath handle event of death
func (p *PPlayer) NotifyDeath(tick int) {
	p.death++

	// remember last death event of player
	if tick > p.lastKilledTick {
		p.lastKilledTick = tick
	}
}

// NotifyAssist handle event of assist
func (p *PPlayer) NotifyAssist() { p.assist++ }

// NotifyWeaponFire handle event of weapon fire
func (p *PPlayer) NotifyWeaponFire(tick int) {
	p.shots++
	p.lastWfTick = tick
}

// NotifyDamageTaken handle event of damage taken
func (p *PPlayer) NotifyDamageTaken(HealthDamage int) {
	p.totalDmgTaken = p.totalDmgTaken + uint(HealthDamage)
}

// NotifyDamageGiven handle event of damage given
func (p *PPlayer) NotifyDamageGiven(HealthDamage int, health int, isHeadshot bool, tick int) {
	// can be add different type of damage
	p.totalDmg = p.totalDmg + uint(HealthDamage)

	// if this is true we hit someone
	if p.lastWfTick == tick {
		p.shotsHit++
	}
	// health need to bigger than zero for avoiding double increment
	// of headshot count (kill event increment this variable as well)
	if health > 0 && isHeadshot {
		p.hs++
	}
}

// NotifyBombDefused handle event of bomb defused
func (p *PPlayer) NotifyBombDefused() { p.bombsDefused++ }

// NotifyBombPlanted handle event of bomb planted
func (p *PPlayer) NotifyBombPlanted() { p.bombsPlanted++ }

// NotifyTrader handle event of being a trader
func (p *PPlayer) NotifyTrader() { p.numTrader++ }

// NotifyTradee handle event of being a tradee
func (p *PPlayer) NotifyTradee() { p.numTradee++ }

// NotifyKAST handle event of updating kast
func (p *PPlayer) NotifyKAST() { p.kast++ }

// NotifyRoundStart handle event of round start
func (p *PPlayer) NotifyRoundStart() {
	p.lastFlashedBy = 0
	p.lastValidTick = 0
}

// NotifySpecialRoundWon handle event of won a special round
func (p *PPlayer) NotifySpecialRoundWon(RoundType RoundType) {
	switch RoundType {
	case PistolRound:
		p.pistolRoundWon++
	case EcoRound:
		p.ecoRoundsWon++
	case ForceBuyRound:
		p.forceRoundsWon++
	case NormalRound:

	default:
		log.Fatal("Unexpected type of round")
	}
}

// NotifySpecialRoundLost handle event of lost a special round
func (p *PPlayer) NotifySpecialRoundLost(RoundType RoundType) {
	switch RoundType {
	case PistolRound:
		p.pistolRoundslost++
	case EcoRound:
		p.ecoRoundsLost++
	case ForceBuyRound:
		p.forceRoundslost++
	case NormalRound:

	default:
		log.Fatal("Unexpected type of round")
	}
}

// NotifyClutchWon handle event of updating clutch
func (p *PPlayer) NotifyClutchWon() { p.clutchesWon++ }

// NotifyFlashAssist handle event of updating flash assist
func (p *PPlayer) NotifyFlashAssist() { p.assist++; p.flashAssists++ }

// NotifyGranadeDamage handle event of updating granade damage given
func (p *PPlayer) NotifyGranadeDamage(Damage uint) { p.heDmg += Damage }

// NotifyFireDamage handle event of updating fire damage given
func (p *PPlayer) NotifyFireDamage(Damage uint) { p.fireDmg += Damage }

// NotifyBlindDuration handle event of updating blind duration
func (p *PPlayer) NotifyBlindDuration(Duration time.Duration) {
	p.timeFlashingOpponents = p.timeFlashingOpponents + Duration
	p.flashDuration = Duration
}

// ResetPlayerState resest all player stats
func (p *PPlayer) ResetPlayerState() {
	p.pistolRoundWon = 0
	p.pistolRoundslost = 0
	p.ecoRoundsWon = 0
	p.ecoRoundsLost = 0
	p.forceRoundsWon = 0
	p.forceRoundslost = 0
	p.clutchesWon = 0
	p.blindPlayersKilled = 0
	p.blindKills = 0
	p.lastFlashedBy = 0
	p.lastValidTick = 0
	p.numRoundsPlayed = 0
	p.lastWfTick = 0
	p.lastKilledTick = 0
	p.shotsHit = 0
	p.numTrader = 0
	p.numTradee = 0
	p.kast = 0
	p.timeFlashingOpponents = 0
	p.hs = 0
	p.hsKill = 0
	p.kill = 0
	p.death = 0
	p.assist = 0
	p.shots = 0
	p.heDmg = 0
	p.fireDmg = 0
	p.totalDmg = 0
	p.totalDmgTaken = 0
	p.flashAssists = 0
	p.bombsDefused = 0
	p.bombsPlanted = 0
}
