package main

import (
	"embed"
	"net/http"
	"os"

	"github.com/computersciencehouse/csh-auth"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed static
var server embed.FS

func main() {
	r := gin.Default()

	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}

	csh := csh_auth.CSHAuth{}
	csh.Init(
		os.Getenv("OIDC_CLIENT_ID"),
		os.Getenv("OIDC_CLIENT_SECRET"),
		os.Getenv("JWT_SECRET"),
		os.Getenv("STATE"),
		os.Getenv("HOST"),
		os.Getenv("HOST")+"/auth/callback",
		os.Getenv("HOST")+"/auth/login",
		[]string{"profile", "groups"},
	)

	fs, err := static.EmbedFolder(server, "static")
	if err != nil {
		panic(err)
	}

	r.GET("/auth/login", csh.AuthRequest)
	r.GET("/auth/callback", csh.AuthCallback)
	r.GET("/auth/logout", csh.AuthLogout)

	r.Use(csh.AuthWrapper(static.Serve("/", fs)))

	r.GET("/api/ping", csh.AuthWrapper(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	}))

	r.Run()
}
