package main

import (
	"time"
)

type Match struct {
	team1Score int
	team2Score int
	team1Name  string
	team2Name  string
	players    map[uint64]PlayerRow
	mapName    string
}

type PlayerRow struct {
	name           string
	steamid        uint64
	teamName       string
	kills          int
	assists        int
	death          int
	damageTotal    int
	hsKills        int
	flashAssist    int
	playersFlashed int
	flashTime      time.Duration
	kastRounds     int
	rating         float32
}

//func ParseStats(demoStream io.Reader) (matchResult InfoStruct, error error) {
//	info, err := GetMatchInfo(demoStream)
//	return info, err
//}

