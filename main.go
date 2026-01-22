package main

import (
	// "embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bigspaceships/circlejerk/auth"

	"github.com/joho/godotenv"
)

// TODO: figure this bit out
//ree go:embed static
// var server embed.FS

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n", r.Context().Value("UserClaims"))
}

func main() {
	err := godotenv.Load()

	if err != nil {
		panic("Error loading .env file")
	}

	cshAuth := auth.Config{
		ClientId: os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		State: os.Getenv("STATE"),
		RedirectURI: os.Getenv("HOST")+"/auth/callback",
		AuthURI: os.Getenv("HOST")+"/auth/login",
		Issuer: os.Getenv("ISSUER"),
	}

	cshAuth.SetupAuth()

	fs := http.FileServer(http.Dir("./static"))

	http.HandleFunc("/auth/login", cshAuth.LoginRequest)
	http.HandleFunc("/auth/callback", cshAuth.LoginCallback)

	apiMux := http.NewServeMux()

	apiMux.HandleFunc("/ping", ping)

	http.Handle("/api/", http.StripPrefix("/api", cshAuth.Handler(apiMux)))

	http.Handle("/", cshAuth.Handler(fs))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
