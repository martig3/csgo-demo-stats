package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	r := gin.Default()
	authUser, _ := os.LookupEnv("DEMO_STATS_USER")
	authPass, _ := os.LookupEnv("DEMO_STATS_PASSWORD")
	api := r.Group("/api", gin.BasicAuth(gin.Accounts{
		authUser: authPass,
	}))
	api.POST("/parse", func(c *gin.Context) {
		if c.Request.Body == nil {
			c.JSON(400, "empty request body")
			return
		}
		var matchInfo, err = GetMatchInfo(c.Request.Body)
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
	api.GET("/parse-remote", func(c *gin.Context) {
		url := c.Query("url")
		authStr := c.Query("auth")
		if url == "" {
			c.JSON(400, "no url specified")
			return
		}
		req, httperr := http.NewRequest("GET", url, nil)
		if httperr != nil {
			c.JSON(500, httperr)
			return
		}
		if authStr != "" {
			req.Header.Set("Authorization", authStr)
		}
		client := &http.Client{
			Timeout: time.Minute * 20,
		}
		resp, respErr := client.Do(req)
		if resp != nil && resp.StatusCode != 200 {
			c.JSON(400, "remote url returned: "+resp.Status)
			return
		}
		if respErr != nil {
			c.JSON(400, respErr)
			return
		}
		if c.Request.Body == nil {
			c.JSON(400, "empty request body")
			return
		}
		var matchInfo, err = GetMatchInfo(resp.Body)
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
	}
}
