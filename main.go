package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/bigspaceships/circlejerk/auth"

	// "github.com/gin-contrib/static"
	// "github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed static
var server embed.FS

func main() {
	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}

	// csh := csh_auth.CSHAuth{}
	// csh.Init(
	// 	os.Getenv("OIDC_CLIENT_ID"),
	// 	os.Getenv("OIDC_CLIENT_SECRET"),
	// 	os.Getenv("JWT_SECRET"),
	// 	os.Getenv("STATE"),
	// 	os.Getenv("HOST"),
	// 	os.Getenv("HOST")+"/auth/callback",
	// 	os.Getenv("HOST")+"/auth/login",
	// 	[]string{"profile", "groups"},
	// )

	myAuth := auth.Config{
		ClientId: os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		State: os.Getenv("STATE"),
		RedirectURI: os.Getenv("HOST")+"/auth/callback",
		Issuer: os.Getenv("ISSUER"),
	}

	myAuth.SetupAuth()

	fs := http.FileServer(http.Dir("./static"))

	http.HandleFunc("/auth/login", myAuth.LoginRequest)
	http.HandleFunc("/auth/callback", myAuth.LoginCallback)

	http.HandleFunc("/test", auth.Status)

	http.Handle("/", fs)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
