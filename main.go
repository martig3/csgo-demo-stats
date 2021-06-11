package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"strings"
)

func main() {
	r := gin.Default()
	authUser, _ := os.LookupEnv("DEMO_STATS_USER")
	authPass, _ := os.LookupEnv("DEMO_STATS_PASSWORD")
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	api := r.Group("/api", gin.BasicAuth(gin.Accounts{
		authUser: authPass,
	}))
	api.POST("/parse-stats", func(c *gin.Context) {
		//var bodyBytes []byte
		if c.Request.Body == nil {
			c.JSON(400, "empty request body")
			return
		}
		var matchInfo, err = GetMatchInfo(c)
		if err != nil {
			if strings.Contains(err.Error(), "ErrInvalidFileType") {
				c.JSON(400, err.Error())
				return
			}
			c.JSON(500, err.Error())
			return
		}
		scoreboard := matchInfo.GetScoreboard()
		c.JSON(200, scoreboard)
	})
	api.GET("/parse-stats-disk", func(c *gin.Context) {
		path := c.Query("path")
		log.Println(path)
		if path == "" {
			c.JSON(400, "no path specified")
			return
		}
		var matchInfo, err = GetMatchInfoFromDisk(path)
		if err != nil {
			if strings.Contains(err.Error(), "ErrInvalidFileType") {
				c.JSON(400, err.Error())
				return
			}
			c.JSON(500, err.Error())
			return
		}
		scoreboard := matchInfo.GetScoreboard()
		c.JSON(200, scoreboard)
	})
	api.POST("/parse-stats-disk", func(c *gin.Context) {
		path := c.Query("path")
		deleteAfterParsing := c.Query("delete")
		log.Println(path)
		if path == "" {
			c.JSON(400, "no path specified")
		}
		if c.GetHeader("Content-Length") == "0" {
			c.JSON(400, "empty body")
		}
		saveFile(path, c.Request.Body)
		var matchInfo, err2 = GetMatchInfoFromDisk(path)
		if err2 != nil {
			if strings.Contains(err2.Error(), "ErrInvalidFileType") {
				c.JSON(400, err2.Error())
				return
			}
			c.JSON(500, err2.Error())
			return
		}
		scoreboard := matchInfo.GetScoreboard()
		if deleteAfterParsing == "true" || deleteAfterParsing == "" {
			err := deleteFile(path)
			if err != nil {
				return
			}
		}
		c.JSON(200, scoreboard)
	})
	err := r.Run()
	if err != nil {
		println(err)
		return
	} // listen and serve on 0.0.0.0:8080
}
