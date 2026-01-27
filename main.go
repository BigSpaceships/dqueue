package main

import (
	// "embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bigspaceships/circlejerk/auth"
	"github.com/bigspaceships/circlejerk/queue"
	dq_websocket "github.com/bigspaceships/circlejerk/websocket"

	"github.com/joho/godotenv"
)

// TODO: figure this bit out
//ree go:embed static
// var server embed.FS

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n", auth.GetUserClaims(r))
}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No env file detected, make sure all secrets are loaded into the environment")
		// panic("Error loading .env file")
	}

	cshAuth := auth.Config{
		ClientId: os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"),
		State: os.Getenv("STATE"),
		RedirectURI: os.Getenv("HOST")+"/auth/callback",
		AuthURI: os.Getenv("HOST")+"/auth/login",
		Issuer: os.Getenv("ISSUER"),
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	ws_server := dq_websocket.CreateWSServer()
	queue := queue.SetupQueue(ws_server)

	cshAuth.SetupAuth()

	fs := http.FileServer(http.Dir("./static"))

	http.HandleFunc("/auth/login", cshAuth.LoginRequest)
	http.HandleFunc("/auth/callback", cshAuth.LoginCallback)

	apiMux := http.NewServeMux()

	apiMux.HandleFunc("GET /ping", ping)
	apiMux.HandleFunc("POST /point", queue.NewPoint)
	apiMux.HandleFunc("POST /clarifier", queue.NewClarifier)
	apiMux.HandleFunc("DELETE /point/{id}", queue.DeletePoint)
	apiMux.HandleFunc("DELETE /clarifier/{id}", queue.DeleteClarifier)
	apiMux.HandleFunc("GET /queue", queue.GetQueue)
	apiMux.HandleFunc("/join_ws", ws_server.WebsocketConnect)

	http.Handle("/api/", http.StripPrefix("/api", cshAuth.Handler(apiMux)))

	http.Handle("/", cshAuth.Handler(fs))

	log.Printf("Dairy Queue started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
