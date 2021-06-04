package main

// https://github.com/markus-wa/demoinfocs-golang/blob/master/examples/print-events/print_events.go

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"time"

	"github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/events"
)

// DemoParser holds all methods to parse a demo file into a infostruct
type DemoParser struct {
	parser demoinfocs.Parser
	Match  *InfoStruct
	state  parsingState
}

// NewDemoParser constructor for a new demoparser
func NewDemoParser() DemoParser {
	return DemoParser{
		state: parsingState{
			Round:        0,
			RoundOngoing: false,
		},

	}
}

// Used while parsing to hold values while going through the ticks
type parsingState struct {
	Round        int // Current round
	RoundOngoing bool
	WarmupKills  []events.Kill
	TeamA        common.Team
}

// Parse starts the parsing process and fills the infostruct with values
// gathered from the demo file
func (p *DemoParser) Parse(demoStream []byte, m *InfoStruct) error {

	matchID := ""
	m.MatchID = matchID
	// Register handlers for events we care about
	p.Match = m
	var err error
	var reader = bytes.NewReader(demoStream)
	p.parser = demoinfocs.NewParser(reader)
	defer p.parser.Close()

	p.parser.RegisterEventHandler(p.handlerKill)
	p.parser.RegisterEventHandler(p.handlerMatchStart)
	p.parser.RegisterEventHandler(p.handlerRoundEnd)
	p.parser.RegisterEventHandler(p.handlerRoundStart)
	p.parser.RegisterEventHandler(p.handlerRankUpdate)
	p.parser.RegisterEventHandler(p.handlerPlayerHurt)
	p.parser.RegisterEventHandler(p.handlerBombPlanted)
	p.parser.RegisterEventHandler(p.handlerBombDefused)
	p.parser.RegisterEventHandler(p.handlerBombExplode)
	p.parser.RegisterEventHandler(p.handlerScoreUpdated)
	p.parser.RegisterEventHandler(p.handlerWeaponFire)
	log.Debug("registered event handlers")
	// p.RegisterEventHandler(handlerChatMessage)
	err = p.setGeneral()
	// Parse the demo returning errors
	err = p.parser.ParseToEnd()
	// Parse header and set general values

	if err != nil {
		return err
	}
	p.calculate()

	return err

}

func (p *DemoParser) ParseFromDisk(path string, m *InfoStruct) error {

	matchID := strings.Split(filepath.Base(path), "_")[0]
	m.MatchID = matchID
	// Register handlers for events we care about
	p.Match = m
	var f *os.File
	var err error

	if f, err = os.Open(path); err != nil {
		return err
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	p.parser = demoinfocs.NewParser(f)
	defer p.parser.Close()

	p.parser.RegisterEventHandler(p.handlerKill)
	p.parser.RegisterEventHandler(p.handlerMatchStart)
	p.parser.RegisterEventHandler(p.handlerRoundEnd)
	p.parser.RegisterEventHandler(p.handlerRoundStart)
	p.parser.RegisterEventHandler(p.handlerRankUpdate)
	p.parser.RegisterEventHandler(p.handlerPlayerHurt)
	p.parser.RegisterEventHandler(p.handlerBombPlanted)
	p.parser.RegisterEventHandler(p.handlerBombDefused)
	p.parser.RegisterEventHandler(p.handlerBombExplode)
	p.parser.RegisterEventHandler(p.handlerScoreUpdated)
	p.parser.RegisterEventHandler(p.handlerWeaponFire)
	// p.RegisterEventHandler(handlerChatMessage)

	// Parse header and set general values
	err = p.setGeneral()

	if err != nil {
		return err
	}

	// Parse the demo returning errors
	err = p.parser.ParseToEnd()
	p.calculate()

	return err

}

func (p *DemoParser) playersBySteamID(steamID uint64) *common.Player {
	for _, v := range p.parser.GameState().Participants().All() {
		if v.SteamID64 == steamID {
			return v
		}
	}
	panic("player not found")
}

func (p *DemoParser) calculate() {

	for k, player := range p.Match.Players.Players {

		// Set Kills, Deaths, Assists, MVPs
		p.Match.Players.Players[k].Kills = p.playersBySteamID(player.Steamid64).Kills()
		p.Match.Players.Players[k].Deaths = p.playersBySteamID(player.Steamid64).Deaths()
		p.Match.Players.Players[k].Assists = p.playersBySteamID(player.Steamid64).Assists()
		p.Match.Players.Players[k].MVPs = p.playersBySteamID(player.Steamid64).MVPs()

		var playeradr int = 0

		for _, v := range allWeapons() {
			playeradr += p.Match.Players.Players[k].WeaponStats.getDamage(v)
		}

		p.Match.Players.Players[k].Adr = playeradr / len(p.Match.Rounds)

		// Calculate player's K/D
		if p.Match.Players.Players[k].Deaths != 0 {
			p.Match.Players.Players[k].Kd = float64(p.Match.Players.Players[k].Kills) / float64(p.Match.Players.Players[k].Deaths)
		}

		for _, round := range p.Match.Rounds {

			// Find player's kills and hs
			roundKills := 0
			for _, kill := range append(round.BKills, round.AKills...) {
				if kill.Killer.Steamid64 == player.Steamid64 {
					if kill.IsHeadshot {
						p.Match.Players.Players[k].Headshots++
					}
					roundKills++
					// p.Match.Players.Players[k].WeaponStats.AddKill(kill.KillerWeapon, kill)
					// p.Match.Players.Players[k].WeaponStats.Kills[kill.KillerWeapon]++
				}
			}

			// Calculate player's he percentage
			if p.Match.Players.Players[k].Kills != 0 {
				p.Match.Players.Players[k].Hsprecent = float64(p.Match.Players.Players[k].Headshots) / float64(p.Match.Players.Players[k].Kills) * 100
			}

			// Set player's 3k, 4k, 5k rounds
			if roundKills == 5 {
				p.Match.Players.Players[k].Rounds5K++
			}
			if roundKills == 4 {
				p.Match.Players.Players[k].Rounds4K++
			}
			if roundKills == 3 {
				p.Match.Players.Players[k].Rounds3K++
			}
			if roundKills == 2 {
				p.Match.Players.Players[k].Rounds2K++
			}
			if roundKills == 1 {
				p.Match.Players.Players[k].Rounds1K++
			}
		}
		var killRating float64
		var survivalRating float64
		var multiRating float64

		const AVERAGE_KPR float64 = 0.679 // average kills per round
		const AVERAGE_SPR float64 = 0.317 // average survived rounds per round
		const AVERAGE_RMK = 1.277         // average value calculated from rounds with multiple kills
		// Calculate HLTV rating
		var r int = len(p.Match.Rounds)

		// Kills/Rounds/AverageKPR
		killRating = float64(p.Match.Players.Players[k].Kills/r) / AVERAGE_KPR
		// (Rounds-Deaths)/Rounds/AverageSPR
		survivalRating = float64(r-p.Match.Players.Players[k].Deaths) / (float64(r)) / AVERAGE_SPR
		// (1K + 4*2K + 9*3K + 16*4K + 25*5K)/Rounds/AverageRMK
		multiRating = float64(p.Match.Players.Players[k].Rounds1K+
			4*p.Match.Players.Players[k].Rounds2K+
			9*p.Match.Players.Players[k].Rounds3K+
			16*p.Match.Players.Players[k].Rounds4K+
			25*p.Match.Players.Players[k].Rounds5K) / float64(r) / AVERAGE_RMK
		var rating float64 = math.Round((killRating + 0.7*survivalRating + multiRating) / 2.7)

		p.Match.Players.Players[k].Rating = rating
	}
}

func (p *DemoParser) setGeneral() error {

	var header common.DemoHeader
	var err error

	if header, err = p.parser.ParseHeader(); err != nil {
		return err
	}

	p.Match.General.MapName = header.MapName
	p.Match.General.MapIconURL = header.MapName
	p.Match.General.MatchTime = time.Now()
	//p.Match.General.DemoLinkURL = "https:TODO/"
	p.Match.General.ScoreA = 0
	p.Match.General.ScoreB = 0

	return nil
}

// NewScoreBoardPlayer constructor for a ScoreboardPlayer. Initializes some
// values with defaults
func (p *DemoParser) NewScoreBoardPlayer(player *common.Player) ScoreboardPlayer {

	name := "BOT"

	if !player.IsBot {
		name = player.Name
	}

	return ScoreboardPlayer{
		IsBot:            player.IsBot,
		IsAMember:        player.Team == p.state.TeamA,
		Name:             name,
		Rank:             0,
		Atag:             player.ClanTag(),
		Steamid64:        player.SteamID64,
		Kills:            0,
		Deaths:           0,
		Assists:          0,
		Kd:               0,
		Adr:              0,
		Rws:              0,
		Rating:           0,
		Hsprecent:        0,
		Firstkills:       0,
		Firstdeaths:      0,
		Tradekills:       0,
		Tradedeaths:      0,
		Tradefirstkills:  0,
		Tradefirstdeaths: 0,
		Roundswonv5:      0,
		Roundswonv4:      0,
		Roundswonv3:      0,
		Rounds5K:         0,
		Rounds4K:         0,
		Rounds3K:         0,
		Rounds2K:         0,
		Rounds1K:         0,
		WeaponStats:      NewWeaponstats(),
		PlayerDamages:    NewPlayerDamages(),
	}
}

func (p *DemoParser) handlerWeaponFire(e events.WeaponFire) {

	if p.parser.GameState().IsWarmupPeriod() {
		return
	}

	p.playerByID(e.Shooter)
	shooter, err := p.Match.Players.PlayerNumByID(e.Shooter.SteamID64)

	if err != nil {
		panic(err)
	}
	p.Match.Players.Players[shooter].WeaponStats.addShot(e)

}

func (p *DemoParser) handlerKill(e events.Kill) {

	if e.Killer == nil || e.Victim == nil {
		return
	}

	// Skip all calculations for kills during warmup
	if p.parser.GameState().IsWarmupPeriod() {
		p.state.WarmupKills = append(p.state.WarmupKills, e)
		return
	}

	// Find killer
	killer := p.playerByID(e.Killer)
	killerNum, err := p.Match.Players.PlayerNumByID(e.Killer.SteamID64)
	if err != nil {
		panic(err)
	}

	p.Match.Players.Players[killerNum].WeaponStats.addKill(e)

	if e.IsHeadshot {
		p.Match.Players.Players[killerNum].WeaponStats.addHeadshot(e)
	}

	// Find victim
	victim := p.playerByID(e.Victim)
	victimNum, err := p.Match.Players.PlayerNumByID(e.Victim.SteamID64)
	if err != nil {
		panic(err)
	}

	kill := RoundKill{
		Time:         p.parser.CurrentTime(),
		IsHeadshot:   e.IsHeadshot,
		KillerWeapon: e.Weapon.Type,
		Killer:       killer,
		Victim:       victim,
	}

	if e.Assister != nil {
		assister := p.playerByID(e.Assister)
		p.Match.Players.addAssist(e.Assister.SteamID64)
		kill.Assister = assister
	}

	// Find fistkills and firstdeaths
	if e.Killer.Team == p.state.TeamA {

		// Check if it's the first kill of the round
		if len(p.Match.Rounds[p.state.Round-1].AKills) == 0 {
			p.Match.Players.Players[killerNum].Firstkills++
			p.Match.Players.Players[victimNum].Firstdeaths++
		}

		for _, v := range p.Match.Rounds[p.state.Round-1].AKills {
			if v.Killer.Steamid64 == e.Victim.SteamID64 && ((p.parser.CurrentTime() - v.Time) < (5 * time.Second)) {
				p.Match.Players.Players[killerNum].Tradekills++
				p.Match.Players.Players[victimNum].Tradedeaths++

				if len(p.Match.Rounds[p.state.Round-1].AKills) == 0 {
					p.Match.Players.Players[killerNum].Tradefirstkills++
					p.Match.Players.Players[victimNum].Tradefirstdeaths++
				}
			}
		}

		// Append to akills
		p.Match.Rounds[p.state.Round-1].AKills = append(p.Match.Rounds[p.state.Round-1].AKills, kill)
	} else {

		// Check if it's the first kill of the round
		if len(p.Match.Rounds[p.state.Round-1].BKills) == 0 {
			p.Match.Players.Players[killerNum].Firstkills++
			p.Match.Players.Players[victimNum].Firstdeaths++
		}

		for _, v := range p.Match.Rounds[p.state.Round-1].BKills {
			if v.Killer.Steamid64 == e.Victim.SteamID64 && ((p.parser.CurrentTime() - v.Time) < (5 * time.Second)) {
				p.Match.Players.Players[killerNum].Tradekills++
				p.Match.Players.Players[victimNum].Tradedeaths++
			}

			if len(p.Match.Rounds[p.state.Round-1].BKills) == 0 {
				p.Match.Players.Players[killerNum].Tradefirstkills++
				p.Match.Players.Players[victimNum].Tradefirstdeaths++
			}
		}

		// Append to bkills
		p.Match.Rounds[p.state.Round-1].BKills = append(p.Match.Rounds[p.state.Round-1].BKills, kill)
	}

	// Find 1v5, 1v4, 1v3
	if p.matesAlive(e.Killer) == 1 {
		switch p.matesAlive(e.Victim) {
		case 5:
			p.Match.Players.Players[killerNum].Roundswonv5++
		case 4:
			p.Match.Players.Players[killerNum].Roundswonv4++
		case 3:
			p.Match.Players.Players[killerNum].Roundswonv3++
		}
	}
}

func (p DemoParser) matesAlive(player *common.Player) int {
	alive := 0
	for _, v := range p.parser.GameState().Participants().Playing() {
		if v.IsAlive() && v.Team == player.Team {
			alive++
		}
	}
	return alive
}

func (p *DemoParser) handlerPlayerHurt(e events.PlayerHurt) {

	if e.Attacker == nil || e.Player == nil {
		return
	}

	for k, v := range p.Match.Players.Players {
		if v.Steamid64 == e.Attacker.SteamID64 {

			// Add damage stats for weapon
			p.Match.Players.Players[k].WeaponStats.addDamage(e)

			// Add hit stats for weapon
			p.Match.Players.Players[k].WeaponStats.addHit(e)

			// Add damage stats for PvP
			_ = p.playerByID(e.Player)
			victimNum, err := p.Match.Players.PlayerNumByID(e.Player.SteamID64)

			if err != nil {
				panic(err)
			}

			p.Match.Players.Players[k].addDamage(e.HealthDamage, &p.Match.Players.Players[victimNum])

			// Add damage to attackers round share for RWS.
			// Only count attacks on the opposing team.
			if e.Attacker.Team != e.Player.Team {
				var dmg int = e.HealthDamage
				// Cap damage at player health for a max total damage of 500
				if dmg > e.Player.Health() {
					dmg = e.Player.Health()
				}
				p.Match.RdDamages.addDamage(dmg, e.Attacker.SteamID64)
			}

			return
		}
	}
}

// func handlerChatMessage(e events.ChatMessage) {
// 	fmt.Printf("Chat - %s says: %s\n", formatPlayer(e.Sender), e.Text)
// }

// Handlers
func (p *DemoParser) handlerRankUpdate(e events.RankUpdate) {

	for k, v := range p.Match.Players.Players {
		if v.Steamid64 == e.SteamID64() {
			p.Match.Players.Players[k].Rank = e.RankNew
			return
		}
	}
	log.Error("player not found setting rank")
}

func (p *DemoParser) handlerMatchStart(e events.MatchStart) {
	// We will treat the team playing CT first as "Team A"
	p.state.TeamA = common.TeamCounterTerrorists
	p.Match.MatchValid = true

	// Add all players to the match
	for _, ct := range p.parser.GameState().Participants().Playing() {
		player := p.NewScoreBoardPlayer(ct)

		p.Match.Players.Players = append(p.Match.Players.Players, player)
	}
	p.Match.RdDamages.RdDamages = NewRdDamages()
}

func (p *DemoParser) handlerRoundStart(e events.RoundStart) {

	p.state.RoundOngoing = true

	// An new round has started, increase counter and add it to slice of the
	// output. The counter should be increased here and *not* in the RoundEnd
	// handler, sice there might happen things "between" the rounds, i.e in the
	// time when a round has ended but the new one has not yet started
	p.state.Round++

	var a_member_steamID uint64
	/*
	 * Each round, find the first player on team A and use them to determine
	 * whether team A is CT or T.
	 */
	for _, pl := range p.Match.Players.Players {
		if pl.IsAMember {
			a_member_steamID = pl.Steamid64
			break
		}
	}

	for _, ct := range p.parser.GameState().TeamCounterTerrorists().Members() {
		if ct.SteamID64 == a_member_steamID {
			p.state.TeamA = common.TeamCounterTerrorists
			break
		}
	}

	for _, t := range p.parser.GameState().TeamTerrorists().Members() {
		if t.SteamID64 == a_member_steamID {
			p.state.TeamA = common.TeamTerrorists
			break
		}
	}

	round := ScoreboardRound{}
	p.Match.Rounds = append(p.Match.Rounds, round)

}

func (p *DemoParser) handlerBombPlanted(e events.BombPlanted) {
	p.Match.Rounds[p.state.Round-1].BombPlanter = e.Player.SteamID64
}

func (p *DemoParser) handlerBombDefused(e events.BombDefused) {
	p.Match.Rounds[p.state.Round-1].BombDefuser = e.Player.SteamID64
}

func (p *DemoParser) handlerBombExplode(e events.BombExplode) {
}

func (p *DemoParser) handlerScoreUpdated(e events.ScoreUpdated) {

	scoreCT := p.parser.GameState().TeamCounterTerrorists().Score()
	scoreT := p.parser.GameState().TeamTerrorists().Score()

	if p.state.TeamA == common.TeamCounterTerrorists {
		p.Match.Rounds[p.state.Round-1].ScoreA = scoreCT
		p.Match.Rounds[p.state.Round-1].ScoreB = scoreT
		p.Match.General.ScoreA = scoreCT
		p.Match.General.ScoreB = scoreT
		return
	}

	if p.state.TeamA == common.TeamTerrorists {
		p.Match.Rounds[p.state.Round-1].ScoreA = scoreT
		p.Match.Rounds[p.state.Round-1].ScoreB = scoreCT
		p.Match.General.ScoreA = scoreT
		p.Match.General.ScoreB = scoreCT
		return
	}

	// log.Warning("Scoreparsing did something strange", p.state.TeamA)
}

func (p *DemoParser) handlerRoundEnd(e events.RoundEnd) {

	if !p.state.RoundOngoing {
		return
	}

	p.state.RoundOngoing = false
	var rd_idx int = p.state.Round - 1

	// Set the winning team
	p.Match.Rounds[rd_idx].TeamWon = e.Winner

	if e.Winner == p.state.TeamA {
		p.Match.Rounds[rd_idx].AWonRound = true
		p.Match.General.ScoreA++
	} else {
		p.Match.General.ScoreB++
	}

	// Calculate RWS for the winning team
	var winning_team_damage int
	var shares_left float64 = 100.0
	var winners []*common.Player

	if e.Winner == common.TeamCounterTerrorists {
		log.Debug("CounterTerrorists win")
		// If the CTs won due to defuse, give defuser 30 shares.
		if e.Reason == events.RoundEndReasonBombDefused {
			log.Debug("Bomb defusal")
			var defuser uint64 = p.Match.Rounds[rd_idx].BombDefuser
			playerNum, err := p.Match.Players.PlayerNumByID(defuser)
			if err != nil {
				panic(err)
			}

			p.Match.Players.Players[playerNum].Rws += 30.0
			shares_left -= 30.0
		}

		winners = p.parser.GameState().TeamCounterTerrorists().Members()

	} else {
		log.Debug("Terrorists win")
		// If the Ts won due to bomb, give planter 30 shares.
		if e.Reason == events.RoundEndReasonTargetBombed {
			log.Debug("Bomb explosion")
			var planter uint64 = p.Match.Rounds[rd_idx].BombPlanter
			playerNum, err := p.Match.Players.PlayerNumByID(planter)
			if err != nil {
				panic(err)
			}

			p.Match.Players.Players[playerNum].Rws += 30.0
			shares_left -= 30.0
		}
		winners = p.parser.GameState().TeamTerrorists().Members()
	}

	// Find total damage for winning team to evenly split 100 RWS
	for _, pl := range winners {
		winning_team_damage += p.Match.RdDamages.RdDamages.Damages[pl.SteamID64]
		//log.Info("Steamid ", pl.SteamID64)
	}

	log.Debug("Win reason: ", e.Reason, " total damage: ", winning_team_damage)
	p.Match.Rounds[rd_idx].WinReason = e.Reason

	// Split the rest of the shares by damage.
	for _, pl := range winners {
		var d int = p.Match.RdDamages.RdDamages.Damages[pl.SteamID64]

		playerNum, err := p.Match.Players.PlayerNumByID(pl.SteamID64)
		if err != nil {
			panic(err)
		}

		var rws float64 = float64(shares_left * float64(d) / float64(winning_team_damage))
		shares_left -= rws
		p.Match.Players.Players[playerNum].Rws += rws

		log.Debugln(p.Match.Players.Players[playerNum].Name, " has ", rws, " this round and ", p.Match.Players.Players[playerNum].Rws, " RWS this game.")
	}

	// Reset all round damage
	for _, pl := range p.Match.Players.Players {
		p.Match.RdDamages.resetDamage(pl.Steamid64)
	}
}
