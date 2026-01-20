package main

import (
	"embed"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/static"
)

//go:embed static
var server embed.FS

func main() {
	fs, err := static.EmbedFolder(server, "static")
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	r.Use(static.Serve("/", fs))

	r.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run()
}
