package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"

	// "github.com/golang/geo/r3"
	common "github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v2/pkg/demoinfocs/events"
	log "github.com/sirupsen/logrus"
)

// InfoStruct holds the data of one demo file in it's parsed form
type InfoStruct struct {
	MatchID    string            `json:"match_id"     db:"match_id"`
	MatchValid bool              `json:"match_valid"  db:"match_valid"`
	General    ScoreboardGeneral `json:"general"      db:"general"`
	Players    ScoreboardPlayers `json:"players"      db:"players"`
	RdDamages  PlayerRoundDamage `json:"rd_damages"   db:"rd_damages"`
	Rounds     []ScoreboardRound `json:"rounds"       db:"rounds"`

	// Duels             [][]int
	// HeatmapsImageURLs []string
	// Megacoins         []MegacoinPlayer
}

// Value : Make the InfoStruct struct implement the driver.Valuer interface. This method
// simply returns the JSON-encoded representation of the struct.
func (is InfoStruct) Value() (driver.Value, error) {
	return json.Marshal(is)
}

// Scan : Make the InfoStruct struct implement the sql.Scanner interface. This method
// simply decodes a JSON-encoded value into the struct fields.
func (is *InfoStruct) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &is)
}

// ScoreboardGeneral holds general information about the match
type ScoreboardGeneral struct {
	Winner        int           `json:"winner"         db:"winner"`
	ScoreA        int           `json:"score_a"        db:"score_a"`
	ScoreB        int           `json:"score_b"        db:"score_b"`
	MapName       string        `json:"map_name"       db:"map_name"`
	MapIconURL    string        `json:"map_icon_url"   db:"map_icon_url"`
	MatchTime     time.Time     `json:"match_time"    db:"match_time"`
	MatchDuration time.Duration `json:"match_duration" db:"match_duration"`
	DemoLinkURL   string        `json:"demo_link_url"  db:"demo_link_url"`
}

// RoundKill holds information about a kill that happenend during the match
type RoundKill struct {
	Time               time.Duration        `json:"time"                 db:"time"`
	KillerTeamString   string               `json:"killer_team_string"   db:"killer_team_string"`
	VictimTeamString   string               `json:"victim_team_string"   db:"victim_team_string"`
	AssisterTeamString string               `json:"assister_team_string" db:"assister_team_string"`
	IsHeadshot         bool                 `json:"is_headshot"          db:"is_headshot"`
	Victim             *ScoreboardPlayer    `json:"victim"               db:"victim"`
	Killer             *ScoreboardPlayer    `json:"killer"               db:"killer"`
	Assister           *ScoreboardPlayer    `json:"assister"             db:"assister"`
	KillerWeapon       common.EquipmentType `json:"weapon"               db:"weapon"`
}

func allWeapons() []common.EquipmentType {

	return []common.EquipmentType{
		common.EqUnknown,
		common.EqP2000,
		common.EqGlock,
		common.EqP250,
		common.EqDeagle,
		common.EqFiveSeven,
		common.EqDualBerettas,
		common.EqTec9,
		common.EqCZ,
		common.EqUSP,
		common.EqRevolver,
		common.EqMP7,
		common.EqMP9,
		common.EqBizon,
		common.EqMac10,
		common.EqUMP,
		common.EqP90,
		common.EqMP5,
		common.EqSawedOff,
		common.EqNova,
		common.EqMag7,
		common.EqSwag7,
		common.EqXM1014,
		common.EqM249,
		common.EqNegev,
		common.EqGalil,
		common.EqFamas,
		common.EqAK47,
		common.EqM4A4,
		common.EqM4A1,
		common.EqScout,
		common.EqSSG08,
		common.EqSG556,
		common.EqSG553,
		common.EqAUG,
		common.EqAWP,
		common.EqScar20,
		common.EqG3SG1,
		common.EqZeus,
		common.EqKevlar,
		common.EqHelmet,
		common.EqBomb,
		common.EqKnife,
		common.EqDefuseKit,
		common.EqWorld,
		common.EqDecoy,
		common.EqMolotov,
		common.EqIncendiary,
		common.EqFlash,
		common.EqSmoke,
		common.EqHE,
	}
}

// NewWeaponstats constructor weaponstats
func NewWeaponstats() WeaponStats {

	return WeaponStats{
		Kills:     make(map[common.EquipmentType]int),
		Headshots: make(map[common.EquipmentType]int),
		Accuracy:  make(map[common.EquipmentType]int),
		Damage:    make(map[common.EquipmentType]int),
		Shots:     make(map[common.EquipmentType]int),
		Hits:      make(map[common.EquipmentType]int),
	}
}

// NewPlayerDamages constructor for player damages struct
func NewPlayerDamages() PlayerDamages {

	return PlayerDamages{
		Damages: make(map[uint64]int),
	}
}

// NewRdDamages constructor for player damages struct
func NewRdDamages() RdDamages {

	return RdDamages{
		Damages: make(map[uint64]int),
	}
}

// Damages returns all damages that where dealt during a match
func (is *InfoStruct) Damages() interface{} {

	type pdamage struct {
		Victim string `json:"victim" db:"victim"`
		Amount int    `json:"amount" db:"amount"`
	}

	type pdamages struct {
		PlayerName string    `json:"player"  db:"player"`
		Damages    []pdamage `json:"damages" db:"damages"`
	}

	ret := struct {
		A map[string]pdamages `json:"a"  db:"a"`
		B map[string]pdamages `json:"b" db:"b"`
	}{
		A: make(map[string]pdamages),
		B: make(map[string]pdamages),
	}

	for _, player := range is.Players.Players {

		if player.IsBot {
			continue
		}

		// Prefill with zero damage for all players except BOTs
		dams := make(map[string]int)

		for _, p2 := range is.Players.Players {
			if p2.IsBot {
				continue
			}
			dams[p2.Name] = 0
		}

		for k2, v := range player.PlayerDamages.Damages {
			vicNum, err := is.Players.PlayerNumByID(k2)
			if err != nil {
				panic(err)
			}

			if is.Players.Players[vicNum].IsBot {
				continue
			}

			name := is.Players.Players[vicNum].Name

			dams[name] = v
		}

		tmp := pdamages{
			PlayerName: player.Name,
		}

		for k, v := range dams {
			tmp.Damages = append(tmp.Damages, pdamage{
				Victim: k,
				Amount: v,
			})
		}
		if player.IsAMember {
			ret.A[player.Name] = tmp
		} else {
			ret.B[player.Name] = tmp
		}
	}

	return ret
}

// Weapons lists all weapons used in the match with it's stats
func (is *InfoStruct) Weapons() interface{} {

	type wlist struct {
		//playername to amount
		A map[string]int `json:"a"  db:"a"`
		B map[string]int `json:"b" db:"b"`
	}

	type weapon struct {
		Name           string `json:"name"            db:"name"`
		TotalKills     int    `json:"total_kills"     db:"total_kills"`
		TotalShots     int    `json:"total_shots"     db:"total_shots"`
		TotalHeadshots int    `json:"total_headshots" db:"total_headshots"`
		TotalAccuracy  int    `json:"total_accuracy"  db:"total_accuracy"`
		TotalDamage    int    `json:"total_damage"    db:"total_damage"`
		TotalHits      int    `json:"total_hits"      db:"total_hits"`
		Kills          wlist  `json:"kills"           db:"kills"`
		Shots          wlist  `json:"shots"           db:"shots"`
		Headshots      wlist  `json:"headshots"       db:"headshots"`
		Accuracy       wlist  `json:"accuracy"        db:"accuracy"`
		Damage         wlist  `json:"damage"          db:"damage"`
		Hits           wlist  `json:"hits"            db:"hits"`
	}

	// Weapons           map[common.EquipmentType]map[*ScoreboardPlayer]WeaponStat
	ret := struct {
		// weaponname to stats
		Weapons map[string]*weapon `json:"weapons" db:"weapons"`
	}{Weapons: make(map[string]*weapon)}

	for _, v := range is.Players.AllWeaponsUsed() {

		// Skip non-weapon classes
		if v.Class() == common.EqClassUnknown || v.Class() == common.EqClassEquipment {
			continue
		}

		ret.Weapons[v.String()] = &weapon{
			Name:      v.String(),
			Kills:     wlist{A: make(map[string]int), B: make(map[string]int)},
			Headshots: wlist{A: make(map[string]int), B: make(map[string]int)},
			Accuracy:  wlist{A: make(map[string]int), B: make(map[string]int)},
			Damage:    wlist{A: make(map[string]int), B: make(map[string]int)},
			Shots:     wlist{A: make(map[string]int), B: make(map[string]int)},
			Hits:      wlist{A: make(map[string]int), B: make(map[string]int)},
		}

		for _, player := range is.Players.Players {

			ret.Weapons[v.String()].TotalKills += player.WeaponStats.getKills(v)
			ret.Weapons[v.String()].TotalHeadshots += player.WeaponStats.getHeadshots(v)
			ret.Weapons[v.String()].TotalDamage += player.WeaponStats.getDamage(v)
			ret.Weapons[v.String()].TotalShots += player.WeaponStats.getShots(v)
			ret.Weapons[v.String()].TotalHits += player.WeaponStats.getHits(v)
			// ret.Weapons[v.String()].TotalDamage+= player.WeaponStats.Damage(v)

			if player.IsAMember {
				ret.Weapons[v.String()].Kills.A[player.Name] = player.WeaponStats.getKills(v)
				ret.Weapons[v.String()].Headshots.A[player.Name] = player.WeaponStats.getHeadshots(v)
				ret.Weapons[v.String()].Accuracy.A[player.Name] = player.WeaponStats.getAccuracy(v)
				ret.Weapons[v.String()].Damage.A[player.Name] = player.WeaponStats.getDamage(v)
				ret.Weapons[v.String()].Shots.A[player.Name] = player.WeaponStats.getShots(v)
				ret.Weapons[v.String()].Hits.A[player.Name] = player.WeaponStats.getHits(v)
			} else {
				ret.Weapons[v.String()].Kills.B[player.Name] = player.WeaponStats.getKills(v)
				ret.Weapons[v.String()].Headshots.B[player.Name] = player.WeaponStats.getHeadshots(v)
				ret.Weapons[v.String()].Accuracy.B[player.Name] = player.WeaponStats.getAccuracy(v)
				ret.Weapons[v.String()].Damage.B[player.Name] = player.WeaponStats.getDamage(v)
				ret.Weapons[v.String()].Shots.B[player.Name] = player.WeaponStats.getShots(v)
				ret.Weapons[v.String()].Hits.B[player.Name] = player.WeaponStats.getHits(v)
			}

		}
	}

	return ret
}

// MarshalJSON marshals a RoundKill struct to json, e.g. for the api
func (rk *RoundKill) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Time               time.Duration        `json:"time"                 db:"time"`
		KillerTeamString   string               `json:"killer_team_string"   db:"killer_team_string"`
		VictimTeamString   string               `json:"victim_team_string"   db:"victim_team_string"`
		AssisterTeamString string               `json:"assister_team_string" db:"assister_team_string"`
		IsHeadshot         bool                 `json:"is_headshot"          db:"is_headshot"`
		Victim             *ScoreboardPlayer    `json:"victim"               db:"victim"`
		Killer             *ScoreboardPlayer    `json:"killer"               db:"killer"`
		Assister           *ScoreboardPlayer    `json:"assister"             db:"assister"`
		KillerWeapon       common.EquipmentType `json:"weapon"               db:"weapon"`
		KillerWeaponName   string               `json:"weapon_name"          db:"weapon_name"`
	}{

		Time:               rk.Time,
		KillerTeamString:   rk.KillerTeamString,
		VictimTeamString:   rk.VictimTeamString,
		AssisterTeamString: rk.AssisterTeamString,
		IsHeadshot:         rk.IsHeadshot,
		Victim:             rk.Victim,
		Killer:             rk.Killer,
		Assister:           rk.Assister,
		KillerWeapon:       rk.KillerWeapon,
		KillerWeaponName:   rk.KillerWeapon.String(),
	})
}

//GetMatchInfo parses a demo file and returns a infostruct containing it's data
func GetMatchInfo(c *gin.Context) (*InfoStruct, error) {
	p := NewDemoParser()
	var info InfoStruct
	err := p.Parse(c, &info)
	return p.Match, err
}

func GetMatchInfoFromDisk(path string) (InfoStruct, error) {
	p := NewDemoParser()
	var info InfoStruct
	err := p.ParseFromDisk(path, &info)
	return info, err
}

// GetScoreboard returns the scoreboard of match
func (is InfoStruct) GetScoreboard() ScoreboardPlayers {
	return is.Players
}

// PlayerNumByID returns the a players's position in the Players[] list of a
// match
func (sp ScoreboardPlayers) PlayerNumByID(steamID uint64) (int, error) {
	for k, v := range sp.Players {
		if v.Steamid64 == steamID {
			return k, nil
		}
	}
	return 0, errors.New("Player Number not found" + strconv.FormatUint(steamID, 10))
}

// ScoreboardPlayers thin wrapper around the list of players to implment methods
type ScoreboardPlayers struct {
	Players []ScoreboardPlayer `json:"players" db:"players"`
}

// PlayerRoundDamage thin wrapper around the list of player round damages to implment methods
type PlayerRoundDamage struct {
	RdDamages RdDamages `json:"rd_damages" db:"rd_damages"`
}

// AllWeaponsUsed returns all weapons shot at least once during the match
func (sp *ScoreboardPlayers) AllWeaponsUsed() []common.EquipmentType {
	list := []common.EquipmentType{}

	for _, w := range allWeapons() {
		for _, p := range sp.Players {
			if p.WeaponStats.getShots(w) > 0 {
				list = append(list, w)
			}
		}
	}

	return list
}

func (sp *ScoreboardPlayers) addAssist(steamID uint64) {
	for k := range sp.Players {
		if sp.Players[k].Steamid64 == steamID {
			sp.Players[k].Assists++
		}
	}
}

// A returns all players of the a team
func (sp ScoreboardPlayers) A() []ScoreboardPlayer {

	out := []ScoreboardPlayer{}

	for _, p := range sp.Players {
		if p.IsAMember {
			out = append(out, p)
		}
	}
	return out
}

// B returns all players of the b team
func (sp ScoreboardPlayers) B() []ScoreboardPlayer {

	out := []ScoreboardPlayer{}

	for _, p := range sp.Players {
		if !p.IsAMember {
			out = append(out, p)
		}
	}

	return out
}

func (p *DemoParser) playerByID(player *common.Player) *ScoreboardPlayer {

	if player == nil {
		return nil
	}

	for _, v := range p.Match.Players.Players {
		if v.Steamid64 == player.SteamID64 {
			return &v
		}
	}

	log.Warning("Created new player for ID:", player.SteamID64)
	newplayer := p.NewScoreBoardPlayer(player)
	p.Match.Players.Players = append(p.Match.Players.Players, newplayer)

	return &newplayer
}

// WeaponStats maps weapons to the player stats
type WeaponStats struct {

	// Number of Kills
	Kills map[common.EquipmentType]int `json:"kills" db:"kills"`

	// Number of Headshots
	Headshots map[common.EquipmentType]int `json:"headshots" db:"headshots"`

	// Percent shots hit of shots fired
	Accuracy map[common.EquipmentType]int `json:"accuracy" db:"accuracy"`

	// Damage caused
	Damage map[common.EquipmentType]int `json:"damage" db:"damage"`

	// Shots fired
	Shots map[common.EquipmentType]int `json:"shots" db:"shots"`

	// Shots hit
	Hits map[common.EquipmentType]int `json:"hits" db:"hits"`
}

func (ws *WeaponStats) addKill(e events.Kill) {
	ws.Kills[e.Weapon.Type]++
}

func (ws *WeaponStats) addHeadshot(e events.Kill) {
	if e.IsHeadshot {
		ws.Headshots[e.Weapon.Type]++
	}
}

func (ws *WeaponStats) addDamage(e events.PlayerHurt) {
	ws.Damage[e.Weapon.Type] += e.HealthDamage
}

func (ws *WeaponStats) addShot(e events.WeaponFire) {
	ws.Shots[e.Weapon.Type]++
	ws.Accuracy[e.Weapon.Type] = (ws.getHits(e.Weapon.Type) * 100) / ws.getShots(e.Weapon.Type)
}

func (ws *WeaponStats) addHit(e events.PlayerHurt) {
	ws.Hits[e.Weapon.Type]++
}

func (ws WeaponStats) getKills(w common.EquipmentType) int {
	return ws.Kills[w]
}

// getAccuracy returns
func (ws WeaponStats) getAccuracy(w common.EquipmentType) int {
	return ws.Accuracy[w]
}

func (ws WeaponStats) getHeadshots(w common.EquipmentType) int {
	return ws.Headshots[w]
}

func (ws WeaponStats) getDamage(w common.EquipmentType) int {
	return ws.Damage[w]
}

func (ws WeaponStats) getShots(w common.EquipmentType) int {
	return ws.Shots[w]
}

func (ws WeaponStats) getHits(w common.EquipmentType) int {
	return ws.Hits[w]
}

func (sp *ScoreboardPlayer) addDamage(damage int, victim *ScoreboardPlayer) {
	sp.PlayerDamages.Damages[victim.Steamid64] += damage
}
func (prd *PlayerRoundDamage) addDamage(damage int, steamID uint64) {
	prd.RdDamages.Damages[steamID] += damage
}

func (prd *PlayerRoundDamage) resetDamage(steamID uint64) {
	prd.RdDamages.Damages[steamID] = 0
}

// PlayerDamages holds the damage of a player to each other player
type PlayerDamages struct {
	Damages map[uint64]int `json:"damages" db:"damages"`
}

// RdDamages holds the damage of a player to each other player for the round
type RdDamages struct {
	Damages map[uint64]int `json:"rd_damages" db:"rd_damages"`
}

// ScoreboardPlayer holds the information about the player of a match
type ScoreboardPlayer struct {
	WeaponStats      WeaponStats   `json:"weapon_stats" db:"weapon_stats"`
	PlayerDamages    PlayerDamages `json:"player_damages" db:"player_damages"`
	IsBot            bool          `json:"isbot" db:"isbot"`
	IsAMember        bool          `json:"isamember" db:"isamember"`
	TeamChar         string        `json:"team" db:"team"`
	SteamId          string        `json:"steamid" db:"steamid"`
	Steamid64        uint64        `json:"steamid64" db:"steamid64"`
	Name             string        `json:"name" db:"name"`
	Atag             string        `json:"atag" db:"atag"`
	Rank             int           `json:"rank" db:"rank"`
	Kills            int           `json:"kills" db:"kills"`
	MVPs             int           `json:"mvps" db:"mvps"`
	Deaths           int           `json:"deaths" db:"deaths"`
	Assists          int           `json:"assists" db:"assists"`
	Kd               float64       `json:"kd" db:"kd"`
	Adr              float64       `json:"adr" db:"adr"`
	Kast             float64       `json:"kast" db:"kast"`
	KastRounds       int           `json:"kastRounds" db:"kastRounds"`
	Rws              float64       `json:"rws" db:"rws"`
	Rating           float64       `json:"rating" db:"rating"`
	Headshots        int           `json:"headshots" db:"headshots"`
	Hsprecent        float64       `json:"hsprecent" db:"hsprecent"`
	Firstkills       int           `json:"firstkills" db:"firstkills"`
	Firstdeaths      int           `json:"firstdeaths" db:"firstdeaths"`
	Tradekills       int           `json:"tradekills" db:"tradekills"`
	Tradedeaths      int           `json:"tradedeaths" db:"tradedeaths"`
	Tradefirstkills  int           `json:"tradefirstkills" db:"tradefirstkills"`
	Tradefirstdeaths int           `json:"tradefirstdeaths" db:"tradefirstdeaths"`
	Roundswonv5      int           `json:"roundswonv5" db:"roundswonv5"`
	Roundswonv4      int           `json:"roundswonv4" db:"roundswonv4"`
	Roundswonv3      int           `json:"roundswonv3" db:"roundswonv3"`
	Rounds5K         int           `json:"rounds5k" db:"rounds5k"`
	Rounds4K         int           `json:"rounds4k" db:"rounds4k"`
	Rounds3K         int           `json:"rounds3k" db:"rounds3k"`
	Rounds2K         int           `json:"rounds2k" db:"rounds2k"`
	Rounds1K         int           `json:"rounds1k" db:"rounds1k"`
	EffFlashes       int           `json:"effFlashes" db:"effFlashes"`
	FlashDuration    int64         `json:"flashDuration" db:"flashDuration"`
}

// ScoreboardRound holds the information about a round in a match
type ScoreboardRound struct {
	AWonRound        bool                  `json:"a_won_round" db:"a_won_round"`
	Duration         time.Duration         `json:"duration" db:"duration"`
	AKills           []RoundKill           `json:"kills_a" db:"kills_a"`
	BKills           []RoundKill           `json:"kills_b" db:"kills_b"`
	ScoreA           int                   `json:"score_a" db:"score_a"`
	ScoreB           int                   `json:"score_b" db:"score_b"`
	ASurvivors       int                   `json:"survivivors_a" db:"survivivors_a"`
	BSurvivors       int                   `json:"survivors_b" db:"survivors_b"`
	TeamWon          common.Team           `json:"team_won" db:"team_won"`
	TotalDamageGiven int                   `json:"total_damage_given" db:"total_damage_given"`
	TotalDamageTaken int                   `json:"total_damage_taken" db:"total_damage_taken"`
	WinReason        events.RoundEndReason `json:"win_reason" db:"win_reason"`
	WinnerTeam       common.Team           `json:"winner_team" db:"winner_team"`
	BombPlanter      uint64                `json:"bomb_planter" db:"bomb_planter"`
	BombDefuser      uint64                `json:"bomb_defuser" db:"bomb_defuser"`
}
