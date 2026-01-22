package auth

import (
	"context"
	"errors"
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
	AuthURI      string
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

func (auth *Config) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Auth")

		if err != nil || cookie.Value == "" {
			// log.Println("cookie not found")
			http.Redirect(w, r, auth.AuthURI, http.StatusFound)
			return
		}

		token, err := jwt.ParseWithClaims(cookie.Value, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(auth.JwtSecret), nil
		})

		if err != nil {
			log.Println("token failure")
			return
		}

		if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
			log.Printf("%s\n", claims.UserInfo)

			newCtx := context.WithValue(r.Context(), "UserClaims", claims.UserInfo)

			next.ServeHTTP(w, r.WithContext(newCtx))
		}
	})
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
		Name:   "Auth",
		Value:  signedToken,
		MaxAge: expireCookie,
		Path:   "/",
	}

	http.SetCookie(w, cookie)

	// TODO: enable this to redirect to whatever route they tried to access
	http.Redirect(w, r, "/", http.StatusFound)
}
