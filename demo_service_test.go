package main

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	//"os"
	"path/filepath"
	"testing"
)

func TestParseStats(t *testing.T) {
	log.Info("Looking for demos in: ", ".")
	files, err := filepath.Glob("."+ "/*.dem")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		demoInfoFromDem, err := GetMatchInfoFromDisk(f)
		assert.NoError(t, err)
		log.Println("Parsed Match with ID: ", demoInfoFromDem.MatchID)
		createKeyValuePairs(demoInfoFromDem.Players.Players)
	}
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
