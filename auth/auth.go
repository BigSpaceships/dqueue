package auth

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
)

type Config struct {
	IssuerShort  string
	ClientId     string
	ClientSecret string
	State        string
	AuthURI      string
	RedirectURI  string
	Issuer       string
	oauthConfig  oauth2.Config
	provider     *oidc.Provider
}

type AuthClaims struct {
	Token    string   `json:"token"`
	UserInfo UserInfo `"json:user_info"`
	jwt.StandardClaims
}

type UserInfo struct {
	Name     string   `json:"name"`
	Username string   `json:"preferred_username"`
	IsEboard bool     `json:"is_eboard"`
	Groups   []string `json:"groups"`
	Issuer   string   `json:"issuer"`
	Picture  string   `json:"picture"`
	Email    string   `json:"email"`
}

var ctx = context.Background()

var authmap = make(map[string]*Config)
var jwtSecret = os.Getenv("JWT_SECRET")

var nonEboardAdmins []string

func (auth *Config) SetupAuth() {
	provider, err := oidc.NewProvider(ctx, auth.Issuer)
	nonEboardAdmins = strings.Split(os.Getenv("NON_EBOARD_ADMINS"), ",")

	if err != nil {
		log.Fatal(err)
	}

	oauthConfig := oauth2.Config{
		ClientID:     auth.ClientId,
		ClientSecret: auth.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  auth.RedirectURI,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	auth.provider = provider
	auth.oauthConfig = oauthConfig

	authmap[auth.IssuerShort] = auth
}

func GetUserClaims(r *http.Request) UserInfo {
	return r.Context().Value("UserInfo").(UserInfo)
}

func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("Auth")

		if err != nil || cookie.Value == "" {
			// log.Println("cookie not found")
			w.WriteHeader(http.StatusForbidden)
			http.Redirect(w, r, "auth", http.StatusFound)
			return
		}

		token, err := jwt.ParseWithClaims(cookie.Value, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			log.Println("token failure")
			return
		}

		if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
			newCtx := context.WithValue(r.Context(), "UserInfo", claims.UserInfo)

			next.ServeHTTP(w, r.WithContext(newCtx))
		}
	})
}

func (auth *Config) LoginRequest(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, auth.oauthConfig.AuthCodeURL(auth.State), http.StatusFound)
}

func (auth *Config) LoginCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")

	if state != auth.State {
		http.Error(w, "Bad state", http.StatusBadRequest)
	}

	issuer := auth.IssuerShort

	oauthToken, err := auth.oauthConfig.Exchange(ctx, r.URL.Query().Get("code"))

	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	oidcUserInfo, err := auth.provider.UserInfo(ctx, oauth2.StaticTokenSource(oauthToken))

	userInfo := &UserInfo{}
	oidcUserInfo.Claims(userInfo)

	switch issuer {
	case "google":
		userInfo.Username = strings.Split(userInfo.Email, "@")[0]
	case "csh":
		userInfo.Picture = fmt.Sprintf("https://profiles.csh.rit.edu/image/%s", userInfo.Username)
	}

	userInfo.IsEboard = slices.Contains(userInfo.Groups, "eboard") || slices.Contains(nonEboardAdmins, userInfo.Username)

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
	signedToken, err := token.SignedString([]byte(jwtSecret))

	cookie := &http.Cookie{
		Name:   "Auth",
		Value:  signedToken,
		MaxAge: expireCookie,
		Path:   "/",
	}

	http.SetCookie(w, cookie)

	// TODO: enable this to redirect to whatever route they tried to access, do I even need to do this?
	http.Redirect(w, r, "/", http.StatusFound)
}
