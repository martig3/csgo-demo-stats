package main

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.POST("/parse-stats", func(c *gin.Context) {
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
		} else {
			c.AbortWithStatus(400)
		}
		var demoInfoFromDemo, err = GetMatchInfo(bodyBytes)
		if err != nil {
			c.JSON(500, err)
			return
		}
		scoreboard := demoInfoFromDemo.GetScoreboard()
		c.JSON(200, scoreboard)
	})
	err := r.Run()
	if err != nil {
		println(err)
		return
	} // listen and serve on 0.0.0.0:8080
}
