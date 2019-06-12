// Package common includes common definitions and methods
package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/geo/r3"
	player "github.com/markus-wa/demoinfocs-golang/common"
	event "github.com/markus-wa/demoinfocs-golang/events"
	"github.com/quancore/demoanalyzer-go/utils"
	log "github.com/sirupsen/logrus"
	viper "github.com/spf13/viper"
)

const (
	specifier = ","
)

// PPlayer is an abstraction struct on top of parser player struct.
type PPlayer struct {
	// base struct added in composition
	*player.Player
	// logger pointer
	logger *log.Logger
	// old team
	oldTeam player.Team
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
	// The struct keep lastly hurted opponent player
	lastHurt map[int64]*HurtTuples
	// The map keeps spotted opponent players by the player
	spottedPlayers map[int64]*SpottedPlayer
	// The steam id of the player that last flashed the current one
	lastFlashedBy int64
	// The tick value the flash event will be valid
	lastValidTick int64
	// The tick number of the last weapon fire event, 0 if it hasn't happened
	lastWfTick int
	// The tick that player killed
	lastKilledTick int
	// The number of last ended round
	lastEndedRound int
	// The tick value of player last footstep. Used for whether player moved or not.
	lastFootstepTick int
	// The round variable that number of killed team members
	numKilledMembers int
	// The number of rounds played by this player
	numRoundsPlayed uint
	// Number of damage firstly given. Used as normalization on POV to damage
	numFirstDamage int
	// The number of shots hit
	shotsHit uint
	// The number of trades done
	numTrader uint
	// The number of time being the tradee
	numTradee uint
	// The number of mvp
	numMVP uint
	// The number of kast for the player
	kast uint
	// Total saved money
	totalSavedMoney int

	// The amount of time the player blinded other players
	timeFlashingOpponents time.Duration
	// Blindness duration from the flashbang currently affecting the player (seconds)
	flashDuration time.Duration
	// Total amount of time to kill a player after first damage
	timeHurtToKill time.Duration
	// Total amount of time first appear on POV and hit the enemy
	timePOVtoDamage time.Duration
	// Total round time
	totalRoundWinTime time.Duration
	// The number of headshots done
	hs uint
	// The number of headshot kills done
	hsKill uint
	// The number of kill done
	kill uint
	// The number of kills while ducking
	duckKill int
	// The number of kills the player did while victim POV not including the player
	lurkerKill int
	// The total distance to enemy killed by this player
	totalKillDistance float32
	// The cost of unit damage
	damageCost float32
	// The number of rounds first kill done
	firstKill uint
	// The number of focused sniper killed
	sniperKilled int
	// Number of teammates saved from death
	savedFriends int
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

	// The total distance of the teams to dead member
	totalMemberKilledDistance float32
	// The total distance to killed members on this round
	roundMemberKilledDistance float32
	// The number of bomb defused
	bombsDefused uint
	// The number of bomb defused
	bombsPlanted uint
	// The total health of the player after won rounds
	totalHealthWon int
	// The total health of the player after lost rounds
	totalHealthLost int
	// The number of rounds alive as last member survived
	lastMemberSurvived int
	// The map area occupied by the team
	teamOccupiedArea float32

	// Total spent money to weapon dropped and picked up by other member of team
	droppedItemVal int
	// Total weapon value picked up
	pickedItemVal int
	// ******* player consts. *****

	// The amount of money when round start
	roundStartMoney int
	// minimum seconds of killing an attacker will count as saving
	beforeSaveSeconds int
	// max health of hurted player count as savedlastMemberSurvived
	maxHealthSaved int

	// round win percentage
	roundWinPercentage float32

	// the last yaw and pitch value of the player
	lastViewDirectionX float32
	lastViewDirectionY float32

	// ******* weapon stats ******
	numKillMelee       int
	numKillPistol      int
	numKillShotgun     int
	numKillSMG         int
	numKillAssultRifle int
	numKillSniperRifle int
	numKillMachineGun  int
	// spray of weapons **********
	sprayPistol      float32
	sprayShotgun     float32
	spraySMG         float32
	sprayAssultRifle float32
	spraySniperRifle float32
	sprayMachineGun  float32
	// ***** body hit group ******
	numHitHead    int
	numHitChest   int
	numHitArms    int
	numHitLegs    int
	numHitStomach int
}

// NewPPlayer creates a *PPlayer with an *Player struct.
func NewPPlayer(player *player.Player, logger *log.Logger) *PPlayer {
	pplayer := &PPlayer{Player: player,
		logger:            logger,
		roundStartMoney:   viper.GetInt("algorithm.roundStartMoney"),
		beforeSaveSeconds: viper.GetInt("algorithm.before_save_seconds"),
		maxHealthSaved:    viper.GetInt("algorithm.max_health_saved"),
	}
	// map initilization
	pplayer.lastHurt = make(map[int64]*HurtTuples)
	pplayer.spottedPlayers = make(map[int64]*SpottedPlayer)

	return pplayer
}

// *** Getters **********

// getSpottedPlayer get spotted player by id if exist
func (p *PPlayer) getSpottedPlayer(playerID int64) *SpottedPlayer {
	// for _, spottedplayer := range p.spottedPlayers {
	// 	p.logger.WithFields(log.Fields{
	// 		"id":          playerID,
	// 		"player name": spottedplayer.Player.Name,
	// 	}).Debug("getting spotted player")
	// }

	if spottedplayer, ok := p.spottedPlayers[playerID]; ok {
		return spottedplayer
	}
	return nil
}

// GetOldTeam get old team
func (p *PPlayer) GetOldTeam() player.Team { return p.oldTeam }

// GetUserID get user id
func (p *PPlayer) GetUserID() int { return p.UserID }

// GetUserName get user name
func (p *PPlayer) GetUserName() string { return p.Name }

// GetSteamID get steam id
func (p *PPlayer) GetSteamID() int64 { return p.SteamID }

// GettLastYawPitch get lastly recorded yaw and pitch value
func (p *PPlayer) GettLastYawPitch() (float32, float32) {
	return p.lastViewDirectionX, p.lastViewDirectionY
}

// GetNumKills get number of kill
func (p *PPlayer) GetNumKills() uint { return p.kill }

// GetNumSniperKill get number of sniper killed
func (p *PPlayer) GetNumSniperKill() int { return p.sniperKilled }

// GetTotalKillDistance get total kill distance recorded for the player
func (p *PPlayer) GetTotalKillDistance() float32 { return p.totalKillDistance }

// GetSavedNum get how many player saved in a match
func (p *PPlayer) GetSavedNum() int { return p.savedFriends }

// *** weapon type ***

// GetMeleeKills get number of kill by melee
func (p *PPlayer) GetMeleeKills() int { return p.numKillMelee }

// GetPistolKills get number of kill by pisdamageCosttol
func (p *PPlayer) GetPistolKills() int { return p.numKillPistol }

// GetShotgunKills get number of kill by shotgun
func (p *PPlayer) GetShotgunKills() int { return p.numKillShotgun }

// GetSmgKills get number of kill by smg
func (p *PPlayer) GetSmgKills() int { return p.numKillSMG }

// GetARifleeKills get number of kill by assult rifle
func (p *PPlayer) GetARifleeKills() int { return p.numKillAssultRifle }

// GetSRifleKills get number of kill by sniper rifle
func (p *PPlayer) GetSRifleKills() int { return p.numKillSniperRifle }

// GetMachineGunKills get number of kill by machine gun
func (p *PPlayer) GetMachineGunKills() int { return p.numKillMachineGun }

// **** Spray ********

// GetPistolSpray get pistol spray
func (p *PPlayer) GetPistolSpray() float32 { return p.sprayPistol }

// GetShotgunSpray get shotgun spray
func (p *PPlayer) GetShotgunSpray() float32 { return p.sprayShotgun }

// GetSmgSpray get smg spray
func (p *PPlayer) GetSmgSpray() float32 { return p.spraySMG }

// GetARifleSpray get assult rifle spray
func (p *PPlayer) GetARifleSpray() float32 { return p.sprayAssultRifle }

// GetSRifleSpray get sniper rifle spray
func (p *PPlayer) GetSRifleSpray() float32 { return p.spraySniperRifle }

// GetMachineGunSpray get machine gun spray
func (p *PPlayer) GetMachineGunSpray() float32 { return p.sprayMachineGun }

// **** Hit group ****

// GetNumHeadHit get number of hits done by head
func (p *PPlayer) GetNumHeadHit() int { return p.numHitHead }

// GetNumChestHit get number of hits done by chest
func (p *PPlayer) GetNumChestHit() int { return p.numHitChest }

// GetNumStomachHit get number of hits done by stomach
func (p *PPlayer) GetNumStomachHit() int { return p.numHitStomach }

// GetNumArmsHit get number of hits done by arms
func (p *PPlayer) GetNumArmsHit() int { return p.numHitArms }

// GetNumLegsHit get number of hits done by legs
func (p *PPlayer) GetNumLegsHit() int { return p.numHitLegs }

// ***********

// GetTimeHurtToKill get duration of time hurt to kill
func (p *PPlayer) GetTimeHurtToKill() time.Duration { return p.timeHurtToKill }

// GetLastMemberSurvived get number of last member survived
func (p *PPlayer) GetLastMemberSurvived() int { return p.lastMemberSurvived }

// GetNumFirstKills get number of first kill
func (p *PPlayer) GetNumFirstKills() uint { return p.firstKill }

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

// GetMVP get number of mvps
func (p *PPlayer) GetMVP() uint { return p.numMVP }

// GetFlashAssist get number of flash assist
func (p *PPlayer) GetFlashAssist() uint { return p.flashAssists }

// GetTotalDamage get total damage given
func (p *PPlayer) GetTotalDamage() uint { return p.totalDmg }

// GetTotalDamageCost get total damage cost given
func (p *PPlayer) GetTotalDamageCost() float32 { return p.damageCost }

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

// GetNumFirstDamage get the number of first damage given
func (p *PPlayer) GetNumFirstDamage() int { return p.numFirstDamage }

// GetKAST get kast have done
func (p *PPlayer) GetKAST() uint { return p.kast }

// GetNumTrader get number of trader
func (p *PPlayer) GetNumTrader() uint { return p.numTrader }

// GetNumTradee get number of tradee
func (p *PPlayer) GetNumTradee() uint { return p.numTradee }

//GetCurrentEqValue get current eq. value
func (p *PPlayer) GetCurrentEqValue() int { return p.Player.CurrentEquipmentValue }

//GetFreezetEqValue get freeze time end eq. value
func (p *PPlayer) GetFreezetEqValue() int { return p.Player.FreezetimeEndEquipmentValue }

//GetStartEqValue get round start time eq. value
func (p *PPlayer) GetStartEqValue() int { return p.Player.RoundStartEquipmentValue }

//GetMoney get current money
func (p *PPlayer) GetMoney() int { return p.Player.Money }

//GetStartMoney get round start money
func (p *PPlayer) GetStartMoney() int { return p.roundStartMoney }

// GetSavedMoney get total saved money during the match
func (p *PPlayer) GetSavedMoney() int { return p.totalSavedMoney }

// GetTotalHealthWon get total health of a player for all won rounds
func (p *PPlayer) GetTotalHealthWon() int { return p.totalHealthWon }

// GetTotalHealthLost get total health of a player for all lost rounds
func (p *PPlayer) GetTotalHealthLost() int { return p.totalHealthLost }

// GetKilledMemberDistance get total kill distance to killed team members
func (p *PPlayer) GetKilledMemberDistance() float32 { return p.totalMemberKilledDistance }

// GetDroppedItemVal get value of dropped items
func (p *PPlayer) GetDroppedItemVal() int { return p.droppedItemVal }

// GetPickedItemVal get value of picked items
func (p *PPlayer) GetPickedItemVal() int { return p.pickedItemVal }

// GetSide get the side of this player
func (p *PPlayer) GetSide() (player.Team, bool) {
	teamOk := true

	if p.Team == player.TeamUnassigned || p.Team == player.TeamSpectators {
		teamOk = false
	}

	return p.Team, teamOk
}

// GetRoundWinPercentage get round win percentage
func (p *PPlayer) GetRoundWinPercentage() float32 { return p.roundWinPercentage }

// GetTotalRoundWinTime get total round time
func (p *PPlayer) GetTotalRoundWinTime() time.Duration { return p.totalRoundWinTime }

// GetDuckKill get duck kill count
func (p *PPlayer) GetDuckKill() int { return p.duckKill }

// GetLurkerKill get lurker kill count
func (p *PPlayer) GetLurkerKill() int { return p.lurkerKill }

// GetlastFootstepTick get last footstep tick
func (p *PPlayer) GetlastFootstepTick() int { return p.lastFootstepTick }

// GetTeamOccupiedArea get the square meter area occupied by the team
func (p *PPlayer) GetTeamOccupiedArea() float32 { return p.teamOccupiedArea }

// ****** setters ********

// SetOldTeam set old team
func (p *PPlayer) SetOldTeam(oldTeam player.Team) { p.oldTeam = oldTeam }

// SetUserID get user id
func (p *PPlayer) SetUserID(newuserid int) { p.UserID = newuserid }

// SetUserName set user name
func (p *PPlayer) SetUserName(newusername string) { p.Name = newusername }

// SetLastYawPitch record player yaw and pitch value
func (p *PPlayer) SetLastYawPitch() (float32, float32) {
	p.lastViewDirectionX = p.ViewDirectionX
	p.lastViewDirectionY = p.ViewDirectionY
	return p.lastViewDirectionX, p.lastViewDirectionY
}

// SetLastFlashedBy set the id of the player who flashed this player
func (p *PPlayer) SetLastFlashedBy(attackerID int64, lastvalidTick int64) {
	p.lastFlashedBy = attackerID
	p.lastValidTick = lastvalidTick
}

// SetSavedMoney add money saved on a round to total saved
func (p *PPlayer) SetSavedMoney(savedMoney int) {
	p.totalSavedMoney += savedMoney
}

// SetKillDistance add distance to enemy newly killed to tat distance
func (p *PPlayer) SetKillDistance(victimLastAlivePos r3.Vector) {
	distance := victimLastAlivePos.Distance(p.Position)
	// lastposX, lastPosY := victimLastAlivePos.X, victimLastAlivePos.Y
	// distance := FindEuclidianDistance(lastposX, lastPosY, p.Position.X, p.Position.Y)
	p.totalKillDistance += float32(distance)
}

// setTotalHealthWon add total health of a player for all won rounds
func (p *PPlayer) setTotalHealthWon(remHealth int) {
	if remHealth > 0 {
		p.totalHealthWon += remHealth

	}
}

// setTotalHealthLost add total health of a player for all lost rounds
func (p *PPlayer) setTotalHealthLost(remHealth int) {
	if remHealth > 0 {
		p.totalHealthLost += remHealth
	}
}

// ***************event notification *******************

// NotifyOccupiedArea notify occupied area of the team
func (p *PPlayer) NotifyOccupiedArea(area float32) { p.teamOccupiedArea += area }

// NotifyPlayerFootstep handle player footstep
func (p *PPlayer) NotifyPlayerFootstep(tick int) { p.lastFootstepTick = tick }

// NotifyDroppedItem notify a item value dropped by this player and picked up by a team member
func (p *PPlayer) NotifyDroppedItem(value int) { p.droppedItemVal += value }

// NotifyPickedItem notify a item value picked up
func (p *PPlayer) NotifyPickedItem(value int) { p.pickedItemVal += value }

// notifyPOVtoDamage calculate amount of time to victim between first seen in POV and damge given
func (p *PPlayer) notifyPOVtoDamage(victimID int64, tick int, tickrate float64) {

	spottedVictim := p.getSpottedPlayer(victimID)
	if spottedVictim != nil {
		spottedTick := spottedVictim.Tick
		tickDifference := tick - spottedTick
		secondsDifference := TickToSeconds(tickDifference, tickrate)
		// if the time is not an outlier
		if secondsDifference.Seconds() < 3 {
			p.timePOVtoDamage += secondsDifference
			p.numFirstDamage++
			p.logger.WithFields(log.Fields{
				"tick":                 tick,
				"spotted tick":         spottedTick,
				"attacker":             p.Name,
				"spotted victim":       spottedVictim.Player.Name,
				"POVtoDamage duration": secondsDifference.Seconds(),
				"number first damage":  p.numFirstDamage,
			}).Info("Pov to damage event")
		}

		// after recording duration, remove POV record
		delete(p.spottedPlayers, victimID)
	}
}

// notifySniperKilled notify sniper elimination
func (p *PPlayer) notifySniperKilled(victim *PPlayer, tick int) {
	activeWeapon := victim.ActiveWeapon()
	if activeWeapon.Weapon == player.EqAWP || activeWeapon.Weapon == player.EqScar20 ||
		activeWeapon.Weapon == player.EqG3SG1 || activeWeapon.Weapon == player.EqSSG08 {
		if activeWeapon.ZoomLevel == 2 {
			p.logger.WithFields(log.Fields{
				"tick":       tick,
				"victim":     victim.Name,
				"zoom level": activeWeapon.ZoomLevel,
				"weapon":     activeWeapon.Weapon.String(),
			}).Debug("A focused sniper has been killed")
			p.sniperKilled++
		}
	}
}

// NotifyMatchEnd notify match has ended to player
func (p *PPlayer) NotifyMatchEnd(tScore, ctScore int) {
	if p.Team == player.TeamTerrorists {
		p.roundWinPercentage = float32(tScore) / float32(tScore+ctScore)
	} else if p.Team == player.TeamCounterTerrorists {
		p.roundWinPercentage = float32(ctScore) / float32(tScore+ctScore)
	}
}

// NotifyRoundEnd notify player that round is ended
func (p *PPlayer) NotifyRoundEnd(endedRound int, winnerTeam player.Team, duration time.Duration) {
	if endedRound > p.lastEndedRound {
		// add remaining health to total
		if winnerTeam == p.Team {
			// add total round time to won round time
			p.totalRoundWinTime += duration
			p.setTotalHealthWon(p.Hp)
		} else {
			p.setTotalHealthLost(p.Hp)
		}
		p.lastEndedRound = endedRound

		// add avarage distance to killed members to total
		if p.numKilledMembers != 0 {
			avarageDistance := (p.roundMemberKilledDistance / float32(p.numKilledMembers))
			p.totalMemberKilledDistance += float32(avarageDistance)
		}
	}

}

// NotifySavedFriend notify additional tea mate has been saved
func (p *PPlayer) NotifySavedFriend() {
	p.savedFriends++
}

// NotifyFirstKill handle event of first killing
func (p *PPlayer) NotifyFirstKill() {
	p.firstKill++
}

// NotifyKill handle event of killing
func (p *PPlayer) NotifyKill(IsHeadshot bool, victim *PPlayer, weapon *player.Equipment, tick int, tickrate float64) {
	isVictimBlinded := victim.IsBlinded()

	p.kill++

	// hs kill
	if IsHeadshot {
		p.hsKill++
	}
	// killed a player while this player
	// has been blinded
	if p.IsBlinded() {
		p.blindKills++
	}
	// killed a blinded player
	if isVictimBlinded {
		p.blindPlayersKilled++
	}

	// event notifications
	p.notifyTeamMemberSave(victim, tick, tickrate)
	p.notifyWeaponType(weapon)
	p.notifyTimeToKill(victim, tick, tickrate)
	p.notifyCrosshairReplecament(weapon)
	if p.IsDucking {
		p.notifyDuckKill()
	}

	if killerSpottedBy := p.IsSpottedBy(victim.Player); !killerSpottedBy {
		p.logger.WithFields(log.Fields{
			"killer": p.Name,
			"victim": victim.Name,
		}).Info("Lurker kill has been made")
		p.notifyLurkerKill()
	}

	// delete victim from spotted player map of killer
	// if _, ok := p.spottedPlayers[victim.SteamID]; ok {
	// 	delete(p.spottedPlayers, victim.SteamID)
	// }

	p.notifySniperKilled(victim, tick)

}

// notifyTeamMemberSave handle whether a team member has been saved
func (p *PPlayer) notifyTeamMemberSave(victim *PPlayer, tick int, tickrate float64) {
	// notify the saved team mates to killer
	hurtByVictim := victim.lastHurt
	beforeSaveTick := int(SecondsToTick(float64(p.beforeSaveSeconds), tickrate))

	for steamID, hurtTuples := range hurtByVictim {
		if ((tick - hurtTuples.LastHurtTick) < beforeSaveTick) && (0 < hurtTuples.RemaningHealth) &&
			(hurtTuples.RemaningHealth < p.maxHealthSaved) && (p.SteamID != steamID) {
			p.NotifySavedFriend()
			p.logger.WithFields(log.Fields{
				"tick":            tick,
				"saved name":      hurtTuples.playerName,
				"killed attacker": victim.Name,
				"savior":          p.Name,
			}).Debug("Player has been saved from kill")
		}
	}

	// victim.lastHurt = nil
}

// NotifyTeamMemberDistance handle the distance to recently killed team member
func (p *PPlayer) NotifyTeamMemberDistance(deathPosition r3.Vector) {
	// deathposX, deathPosY := deathPosition.X, deathPosition.Y
	// distance := FindEuclidianDistance(deathposX, deathPosY, p.Position.X, p.Position.Y)
	distance := deathPosition.Distance(p.Position)
	p.numKilledMembers++
	p.roundMemberKilledDistance += float32(distance)
}

// notifyDuckKill handle kill event while player is ducking
func (p *PPlayer) notifyDuckKill() { p.duckKill++ }

// notifyDuckKill handle kill event for lurker kill
func (p *PPlayer) notifyLurkerKill() { p.lurkerKill++ }

// NotifyLastMemberSurvived handle event of last member survived
func (p *PPlayer) NotifyLastMemberSurvived() { p.lastMemberSurvived++ }

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
func (p *PPlayer) NotifyDamageGiven(e event.PlayerHurt, eqRatio float32, tick int, tickrate float64) {
	HealthDamage := e.HealthDamage
	victimID := e.Player.SteamID
	// can be add different type of damage
	p.totalDmg = p.totalDmg + uint(HealthDamage)

	// if this is true we hit someone
	if p.lastWfTick == tick {
		p.shotsHit++
		// notify hot group damage
		p.notifyHitGroup(e.HitGroup)
		// notify time between POV to damage given
		p.notifyPOVtoDamage(victimID, tick, tickrate)
		// add total damage cost
		p.damageCost += (float32(HealthDamage) * eqRatio)
		// add victim to attacker lasthurt pointer
		if hurtedTuple, ok := p.lastHurt[victimID]; ok {
			hurtedTuple.RemaningHealth = e.Health
			hurtedTuple.LastHurtTick = tick

		} else {
			p.lastHurt[victimID] = &HurtTuples{LastHurtTick: tick, FirstHurtTick: tick, playerName: e.Player.Name, RemaningHealth: e.Health}
		}
	}
}

func (p *PPlayer) notifyHitGroup(hitGroup event.HitGroup) {
	switch hitGroup {
	case event.HitGroupHead:
		p.numHitHead++
	case event.HitGroupChest:
		p.numHitChest++
	case event.HitGroupStomach:
		p.numHitStomach++
	case event.HitGroupLeftArm, event.HitGroupRightArm:
		p.numHitArms++
	case event.HitGroupLeftLeg, event.HitGroupRightLeg:
		p.numHitLegs++
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

// NotifyMVP handle event of updating mvp
func (p *PPlayer) NotifyMVP() { p.numMVP++ }

// notifyWeaponType handle type of the weapon
func (p *PPlayer) notifyWeaponType(weapon *player.Equipment) {
	switch weapon.Class() {
	case player.EqClassEquipment:
		if weapon.Weapon == player.EqKnife {
			p.numKillMelee++
		}
	case player.EqClassHeavy:
		// shotguns
		if weapon.Weapon == player.EqM249 || weapon.Weapon == player.EqNegev {
			p.numKillMachineGun++
		} else {
			p.numKillShotgun++
		}
	case player.EqClassSMG:
		p.numKillSMG++
	case player.EqClassRifle:
		// sniper rifles
		if weapon.Weapon == player.EqAWP || weapon.Weapon == player.EqScar20 ||
			weapon.Weapon == player.EqG3SG1 || weapon.Weapon == player.EqSSG08 {
			p.numKillSniperRifle++
		} else {
			p.numKillAssultRifle++
		}
	case player.EqClassPistols:
		p.numKillPistol++
	}
}

// NotifyRoundStart handle event of round start
func (p *PPlayer) NotifyRoundStart() {
	p.lastFlashedBy = 0
	p.lastValidTick = 0
	p.lastHurt = make(map[int64]*HurtTuples)
	p.spottedPlayers = make(map[int64]*SpottedPlayer)
	p.lastWfTick = 0
	p.roundStartMoney = p.GetMoney()
	p.numKilledMembers = 0
	p.roundMemberKilledDistance = 0

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

// NotifyPlayerSpotted notify a player has been spotted by the player
func (p *PPlayer) NotifyPlayerSpotted(player *PPlayer, tick int) {
	isAlive := player.IsAlive()
	p.logger.WithFields(log.Fields{
		"spotter":               p.Name,
		"tick":                  tick,
		"spotted player":        player.Name,
		"spotted payer isalive": isAlive,
	}).Info("Player spotted by a player")
	if isAlive {
		playerID := player.GetSteamID()
		spottedplayer := p.getSpottedPlayer(playerID)
		if spottedplayer != nil {
			spottedplayer.Tick = tick
		} else {
			newSpottedPlayer := SpottedPlayer{Tick: tick, Player: player}
			p.spottedPlayers[playerID] = &newSpottedPlayer
		}
	}

}

// NotifyBlindDuration handle event of updating blind duration
func (p *PPlayer) NotifyBlindDuration(Duration time.Duration) {
	p.timeFlashingOpponents += Duration
	p.flashDuration = Duration
}

// notifyTimeToKill handle event of time of damage to kill
func (p *PPlayer) notifyTimeToKill(victim *PPlayer, tick int, tickrate float64) {
	// notify the saved team mates to killer
	hurtByKiller := p.lastHurt
	if hurtTuple, ok := hurtByKiller[victim.GetSteamID()]; ok {
		firstHurtTick := hurtTuple.FirstHurtTick
		timeToKillTick := tick - firstHurtTick
		timeToKillSec := TickToSeconds(timeToKillTick, tickrate)
		if timeToKillSec.Seconds() < 10 {
			p.logger.WithFields(log.Fields{
				"sec":  timeToKillSec.Seconds(),
				"tick": timeToKillTick,
			}).Debug("Amount of time to kill")
			p.timeHurtToKill += timeToKillSec
		}
	}
}

// notifyTimeToKill handle event of time of damage to kill
func (p *PPlayer) notifyCrosshairReplecament(weapon *player.Equipment) {
	var weaponType string
	distance := float32(FindEuclidianDistance(float64(p.lastViewDirectionX), float64(p.lastViewDirectionY),
		float64(p.ViewDirectionX), float64(p.ViewDirectionY)))

	switch weapon.Class() {
	case player.EqClassHeavy:
		// machine gun
		if weapon.Weapon == player.EqM249 || weapon.Weapon == player.EqNegev {
			p.sprayMachineGun += distance
			weaponType = "machine gun"
		} else {
			p.sprayShotgun += distance
			weaponType = "shotgun"

		}
	case player.EqClassSMG:
		p.spraySMG += distance
		weaponType = "smg"

	case player.EqClassRifle:
		if weapon.Weapon == player.EqAWP || weapon.Weapon == player.EqScar20 ||
			weapon.Weapon == player.EqG3SG1 || weapon.Weapon == player.EqSSG08 {
			p.spraySniperRifle += distance
			weaponType = "sniper rifle"

		} else {
			p.sprayAssultRifle += distance
			weaponType = "assult rifle"

		}
	case player.EqClassPistols:
		p.sprayPistol += distance
		weaponType = "pistol"

	}

	p.logger.WithFields(log.Fields{
		"before x": p.lastViewDirectionX,
		"before y": p.lastViewDirectionY,
		"after x":  p.ViewDirectionX,
		"after y":  p.ViewDirectionY,
		"distance": distance,
		"weapon":   weaponType,
	}).Debug("Amount of crosshair replecament")
}

// ResetPlayerState resest all player stats
func (p *PPlayer) ResetPlayerState() {

	p.pistolRoundWon = 0
	p.pistolRoundslost = 0
	p.ecoRoundsWon = 0
	p.ecoRoundsLost = 0
	p.forceRoundsWon = 0
	p.forceRoundslost = 0
	p.roundStartMoney = p.GetMoney()
	p.clutchesWon = 0
	p.blindPlayersKilled = 0
	p.blindKills = 0
	p.sniperKilled = 0
	p.lastEndedRound = 0
	p.lastFlashedBy = 0
	p.lastValidTick = 0
	p.numRoundsPlayed = 0
	p.lastWfTick = 0
	p.lastKilledTick = 0
	p.shotsHit = 0
	p.numFirstDamage = 0
	p.numTrader = 0
	p.numTradee = 0
	p.kast = 0
	p.numMVP = 0
	p.timeFlashingOpponents = 0
	p.timeHurtToKill = 0
	p.hsKill = 0
	p.kill = 0
	p.duckKill = 0
	p.lurkerKill = 0
	p.totalKillDistance = 0
	p.lastMemberSurvived = 0
	p.lastFootstepTick = 0
	p.firstKill = 0
	p.death = 0
	p.assist = 0
	p.shots = 0
	p.heDmg = 0
	p.fireDmg = 0
	p.totalDmg = 0
	p.damageCost = 0
	p.totalDmgTaken = 0
	p.flashAssists = 0
	p.bombsDefused = 0
	p.bombsPlanted = 0
	p.roundStartMoney = 0
	p.beforeSaveSeconds = 0
	p.maxHealthSaved = 0
	p.droppedItemVal = 0
	p.pickedItemVal = 0

	// *** weapon ***
	p.numKillMelee = 0
	p.numKillPistol = 0
	p.numKillShotgun = 0
	p.numKillSMG = 0
	p.numKillAssultRifle = 0
	p.numKillSniperRifle = 0
	p.numKillMachineGun = 0
	// *** spray ******
	p.sprayPistol = 0
	p.sprayShotgun = 0
	p.spraySMG = 0
	p.sprayAssultRifle = 0
	p.spraySniperRifle = 0
	p.sprayMachineGun = 0
	// *** hit group **
	p.numHitHead = 0
	p.numHitChest = 0
	p.numHitArms = 0
	p.numHitLegs = 0
	p.numHitStomach = 0
	// maps
	p.lastHurt = make(map[int64]*HurtTuples)
	p.spottedPlayers = make(map[int64]*SpottedPlayer)

}

// OutputPlayerState output as string form of current player state
func (p *PPlayer) OutputPlayerState(sb strings.Builder, roundPlayed, Won int) strings.Builder {
	roundPlayedf := float32(roundPlayed)

	playerName := p.Name
	playerName = strings.Replace(playerName, specifier, " ", -1)
	sb.WriteString(fmt.Sprintf("%s%s", playerName, specifier))

	var pistolRoundWonPercentage float32
	pistolROundsWon := float32(p.pistolRoundWon)
	pistolROundsLost := float32(p.pistolRoundslost)
	if (pistolROundsWon + pistolROundsLost) > 0 {
		pistolRoundWonPercentage = pistolROundsWon / (pistolROundsWon + pistolROundsLost)
	}
	sb.WriteString(fmt.Sprintf("%s%s", fmt.Sprintf("%.3f", pistolRoundWonPercentage), specifier))

	sb.WriteString(fmt.Sprintf("%s%s", fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.hsKill), float32(p.kill))), specifier))

	clutchesWon := fmt.Sprintf("%.3f", float32(p.clutchesWon)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", clutchesWon, specifier))
	// sb.WriteString(fmt.Sprintf("%s%s", fmt.Sprint(currPlayer.GetClutchWon()), specifier))

	adr := fmt.Sprintf("%.3f", float32(p.totalDmg)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", adr, specifier))

	fpr := fmt.Sprintf("%.3f", float32(p.kill)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", fpr, specifier))

	fkr := fmt.Sprintf("%.3f", float32(p.firstKill)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", fkr, specifier))

	apr := fmt.Sprintf("%.3f", float32(p.assist)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", apr, specifier))

	kdDiff := fmt.Sprintf("%.3f", (float32(p.kill)-float32(p.death))/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", kdDiff, specifier))

	flashAssist := fmt.Sprintf("%.3f", float32(p.flashAssists)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", flashAssist, specifier))

	blindPlayerKilled := fmt.Sprintf("%.3f", float32(p.blindPlayersKilled)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", blindPlayerKilled, specifier))

	blindKills := fmt.Sprintf("%.3f", float32(p.blindKills)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", blindKills, specifier))

	granedaDamage := fmt.Sprintf("%.3f", float32(p.heDmg)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", granedaDamage, specifier))

	fireDamage := fmt.Sprintf("%.3f", float32(p.fireDmg)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", fireDamage, specifier))

	timeFlashingOpponent := fmt.Sprintf("%.3f", float32(p.timeFlashingOpponents.Seconds())/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", timeFlashingOpponent, specifier))

	accuracyStr := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.shotsHit), float32(p.shots)))
	sb.WriteString(fmt.Sprintf("%s%s", accuracyStr, specifier))

	numTrader := fmt.Sprintf("%.3f", float32(p.numTrader)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", numTrader, specifier))

	numTradee := fmt.Sprintf("%.3f", float32(p.numTradee)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", numTradee, specifier))

	kast := fmt.Sprintf("%.3f", float32(p.kast)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", kast, specifier))

	mvp := fmt.Sprintf("%.3f", float32(p.numMVP)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", mvp, specifier))

	moneySaved := fmt.Sprintf("%.3f", float32(p.totalSavedMoney)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", moneySaved, specifier))

	sniperKill := fmt.Sprintf("%.3f", float32(p.numKillSniperRifle)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", sniperKill, specifier))

	meleeKill := fmt.Sprintf("%.3f", float32(p.numKillMelee)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", meleeKill, specifier))

	shoutgunKill := fmt.Sprintf("%.3f", float32(p.numKillShotgun)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", shoutgunKill, specifier))

	assultRKill := fmt.Sprintf("%.3f", float32(p.numKillAssultRifle)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", assultRKill, specifier))

	pistolKill := fmt.Sprintf("%.3f", float32(p.numKillPistol)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", pistolKill, specifier))

	machineGunKill := fmt.Sprintf("%.3f", float32(p.numKillMachineGun)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", machineGunKill, specifier))

	smgKill := fmt.Sprintf("%.3f", float32(p.numKillSMG)/roundPlayedf)
	sb.WriteString(fmt.Sprintf("%s%s", smgKill, specifier))

	totalHit := float32(p.shotsHit)
	headHit := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.numHitHead), totalHit))
	sb.WriteString(fmt.Sprintf("%s%s", headHit, specifier))

	stomachHit := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.numHitStomach), totalHit))
	sb.WriteString(fmt.Sprintf("%s%s", stomachHit, specifier))

	chestHit := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.numHitChest), totalHit))
	sb.WriteString(fmt.Sprintf("%s%s", chestHit, specifier))

	legsHit := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.numHitLegs), totalHit))
	sb.WriteString(fmt.Sprintf("%s%s", legsHit, specifier))

	armsHit := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.numHitArms), totalHit))
	sb.WriteString(fmt.Sprintf("%s%s", armsHit, specifier))

	unitDamageCost := fmt.Sprintf("%.3f", utils.SafeDivision(p.damageCost, float32(p.totalDmg)))
	sb.WriteString(fmt.Sprintf("%s%s", unitDamageCost, specifier))

	avarageKillDistance := fmt.Sprintf("%.3f", utils.SafeDivision(p.totalKillDistance, float32(p.kill)))
	sb.WriteString(fmt.Sprintf("%s%s", avarageKillDistance, specifier))

	avaragePlayerSaved := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.savedFriends), roundPlayedf))
	sb.WriteString(fmt.Sprintf("%s%s", avaragePlayerSaved, specifier))

	playerWonHealth := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.totalHealthWon), roundPlayedf))
	sb.WriteString(fmt.Sprintf("%s%s", playerWonHealth, specifier))

	playerLostHealth := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.totalHealthLost), roundPlayedf))
	sb.WriteString(fmt.Sprintf("%s%s", playerLostHealth, specifier))

	lastMemberSurvived := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.lastMemberSurvived), roundPlayedf))
	sb.WriteString(fmt.Sprintf("%s%s", lastMemberSurvived, specifier))

	timeHurtToKill := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.timeHurtToKill.Seconds()), float32(p.kill)))
	sb.WriteString(fmt.Sprintf("%s%s", timeHurtToKill, specifier))

	spraySniper := fmt.Sprintf("%.3f", utils.SafeDivision(p.spraySniperRifle, float32(p.kill)))
	sb.WriteString(fmt.Sprintf("%s%s", spraySniper, specifier))

	sprayShotgun := fmt.Sprintf("%.3f", utils.SafeDivision(p.sprayShotgun, float32(p.kill)))
	sb.WriteString(fmt.Sprintf("%s%s", sprayShotgun, specifier))

	sprayARifle := fmt.Sprintf("%.3f", utils.SafeDivision(p.sprayAssultRifle, float32(p.kill)))
	sb.WriteString(fmt.Sprintf("%s%s", sprayARifle, specifier))

	sprayPistol := fmt.Sprintf("%.3f", utils.SafeDivision(p.sprayPistol, float32(p.kill)))
	sb.WriteString(fmt.Sprintf("%s%s", sprayPistol, specifier))

	sprayMachineGun := fmt.Sprintf("%.3f", utils.SafeDivision(p.sprayMachineGun, float32(p.kill)))
	sb.WriteString(fmt.Sprintf("%s%s", sprayMachineGun, specifier))

	spraySMG := fmt.Sprintf("%.3f", utils.SafeDivision(p.spraySMG, float32(p.kill)))
	sb.WriteString(fmt.Sprintf("%s%s", spraySMG, specifier))

	roundWinPercentage := fmt.Sprintf("%.3f", p.roundWinPercentage)
	sb.WriteString(fmt.Sprintf("%s%s", roundWinPercentage, specifier))

	roundWinTime := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.totalRoundWinTime.Seconds()), roundPlayedf))
	sb.WriteString(fmt.Sprintf("%s%s", roundWinTime, specifier))

	duckKill := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.duckKill), float32(p.kill)))
	sb.WriteString(fmt.Sprintf("%s%s", duckKill, specifier))

	memberDeathDistance := fmt.Sprintf("%.3f", utils.SafeDivision(p.totalMemberKilledDistance, roundPlayedf))
	sb.WriteString(fmt.Sprintf("%s%s", memberDeathDistance, specifier))

	sniperKilled := fmt.Sprintf("%.3f", utils.SafeDivision(float32(p.sniperKilled), float32(p.kill)))
	sb.WriteString(fmt.Sprintf("%s%s", sniperKilled, specifier))

	occupiedArea := fmt.Sprintf("%.3f", utils.SafeDivision(p.teamOccupiedArea, roundPlayedf))
	sb.WriteString(fmt.Sprintf("%s%s", occupiedArea, specifier))

	sb.WriteString(fmt.Sprintf("%s", fmt.Sprint(Won)))

	sb.WriteByte('\n')

	return sb
}
