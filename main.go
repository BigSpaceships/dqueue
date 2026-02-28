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
		ClientId:     os.Getenv("CSH_OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("CSH_OIDC_CLIENT_SECRET"),
		Issuer:       os.Getenv("CSH_ISSUER"),
		IssuerShort:  "csh",
		State:        os.Getenv("STATE"),
		RedirectURI:  os.Getenv("HOST") + "/auth/callback-csh",
		AuthURI:      os.Getenv("HOST") + "/auth/login-csh",
	}

	googleAuth := auth.Config{
		ClientId:     os.Getenv("GOOGLE_OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OIDC_CLIENT_SECRET"),
		Issuer:       os.Getenv("GOOGLE_ISSUER"),
		IssuerShort:  "google",
		State:        os.Getenv("STATE"),
		RedirectURI:  os.Getenv("HOST") + "/auth/callback-google",
		AuthURI:      os.Getenv("HOST") + "/auth/login-google",
	}

	cshAuth.SetupAuth()
	googleAuth.SetupAuth()

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	ws_server := dq_websocket.CreateWSServer()
	discussion := queue.SetupDiscussion(ws_server)

	fs := http.FileServer(http.Dir("./static"))

	http.HandleFunc("/auth/login-csh", cshAuth.LoginRequest)
	http.HandleFunc("/auth/callback-csh", cshAuth.LoginCallback)

	http.HandleFunc("/auth/login-google", googleAuth.LoginRequest)
	http.HandleFunc("/auth/callback-google", googleAuth.LoginCallback)

	apiMux := http.NewServeMux()

	apiMux.HandleFunc("POST /queue/{queue}/point", discussion.NewPoint)
	apiMux.HandleFunc("POST /queue/{queue}/clarifier", discussion.NewClarifier)
	apiMux.HandleFunc("DELETE /queue/{queue}/point/{id}", discussion.DeletePoint)
	apiMux.HandleFunc("DELETE /queue/{queue}/clarifier/{id}", discussion.DeleteClarifier)
	apiMux.HandleFunc("PATCH /queue/{queue}", discussion.ChangeTopic)
	apiMux.HandleFunc("GET /queue/{queue}", discussion.GetQueue)
	apiMux.HandleFunc("POST /queue/{queue}/new-child", discussion.NewQueue)
	apiMux.HandleFunc("DELETE /queue/{queue}", discussion.DeleteQueue)
	apiMux.HandleFunc("GET /queue/{queue}/path", discussion.GetQueuePath)
	apiMux.HandleFunc("GET /discussion", discussion.GetDiscussion)
	apiMux.HandleFunc("/joinws", ws_server.WebsocketConnect)

	http.Handle("/api/", http.StripPrefix("/api", auth.Handler(apiMux)))

	http.Handle("/auth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/auth.html")
	}))

	http.Handle("/auth/logout", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := &http.Cookie{
			Name:   "Auth",
			Value:  "",
			MaxAge: -1,
			Path:   "/",
		}

		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/auth", http.StatusFound)
	}))

	http.Handle("/", fs)

	log.Printf("Dairy Queue started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
