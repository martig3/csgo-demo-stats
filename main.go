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
		ByteBody, _ := ioutil.ReadAll(c.Request.Body)
		demoInfoFromDem, _ := GetMatchInfo(ByteBody)
		demoInfoFromDemo := demoInfoFromDem
		c.JSON(200, demoInfoFromDemo.GetScoreboard())
	})
	err := r.Run()
	if err != nil {
		println(err)
		return
	} // listen and serve on 0.0.0.0:8080
}
