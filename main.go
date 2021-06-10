package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"strings"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/parse-stats", func(c *gin.Context) {
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
	r.GET("/parse-stats-disk", func(c *gin.Context) {
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
	err := r.Run()
	if err != nil {
		println(err)
		return
	} // listen and serve on 0.0.0.0:8080
}
