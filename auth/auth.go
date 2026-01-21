package auth

import (
	"context"
	"fmt"
	"time"

	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
)

type Config struct {
	ClientId     string
	ClientSecret string
	JwtSecret    string
	State        string
	RedirectURI  string
	Issuer       string
}

type AuthClaims struct {
	Token    string   `json:"token"`
	UserInfo UserInfo `"json:user_info"`
	jwt.StandardClaims
}

type UserInfo struct {
	Name     string `json:"name"`
	Username string `json:"preferred_username"`
	IsEboard bool
}

var oauthConfig oauth2.Config
var ctx = context.Background()
var provider *oidc.Provider

func (auth *Config) SetupAuth() {
	var err error
	provider, err = oidc.NewProvider(ctx, auth.Issuer)

	if err != nil {
		log.Fatal(err)
	}

	oauthConfig = oauth2.Config{
		ClientID:     auth.ClientId,
		ClientSecret: auth.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  auth.RedirectURI,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
}

func Status(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("Auth")

	fmt.Printf("%s, %s\n", cookie, err)

	fmt.Fprintf(w, "hii")
}

func (auth *Config) LoginRequest(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, oauthConfig.AuthCodeURL(auth.State), http.StatusFound)
}

func (auth *Config) LoginCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")

	if state != auth.State {
		http.Error(w, "Bad state", http.StatusBadRequest)
	}

	oauthToken, err := oauthConfig.Exchange(ctx, r.URL.Query().Get("code"))

	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	oidcUserInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauthToken))

	userInfo := &UserInfo{}
	oidcUserInfo.Claims(userInfo)

	expireToken := time.Now().Add(time.Hour * 1).Unix()
	expireCookie := 3600
	claims := AuthClaims{
		oauthToken.AccessToken,
		*userInfo,
		jwt.StandardClaims{
			ExpiresAt: expireToken,
			Issuer:    auth.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(auth.JwtSecret))

	cookie := &http.Cookie{
		Name: "Auth",
		Value: signedToken,
		MaxAge: expireCookie,
		Path: "/",
	}

	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)

	// json_data, err := json.MarshalIndent(jsonRaw, "", "    ")

	// log.Println(string(jsonRaw))
}
