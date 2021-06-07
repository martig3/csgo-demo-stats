package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	//"os"
	"testing"
)

func TestParseStats(t *testing.T) {
	f := "G:\\replays\\_scrimbot_backup\\2021-6-6_pug_de_overpass_60bc5197ffc453a0de48212d.dem"
	demoInfoFromDem, err := GetMatchInfoFromDisk(f)
	assert.NoError(t, err)
	log.Println("Parsed Match with ID: ", demoInfoFromDem.MatchID)
}

func createKeyValuePairs(players []ScoreboardPlayer) {
	b := new(bytes.Buffer)
	for value := range players {
		_, err := fmt.Fprintf(b, "%d\"\n", value)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println(b.String())
}
