// Package googleauth provides you access to Google's OAuth2
// infrastructure. The implementation is based on this blog post:
// http://skarlso.github.io/2016/06/12/google-signin-with-go/
package googleauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/golang/glog"
	"github.com/gorilla/securecookie"
	"gopkg.in/gin-gonic/gin.v1"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"Go-sse/seccookie"
)

// Credentials stores google client-ids.
type Credentials struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// User is a retrieved and authenticated user.
type User struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Gender        string `json:"gender"`
	Hd            string `json:"hd"`
}

var cred Credentials
var conf *oauth2.Config
var state string

var store sessions.CookieStore

var scookie = securecookie.New(seccookie.CookieKey.HashKey, seccookie.CookieKey.BlockKey)

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// Setup sets up the OAuth2 parameters
func Setup(redirectURL, credFile string, scopes []string) {
	var c Credentials
	file, err := ioutil.ReadFile(credFile)
	if err != nil {
		glog.Fatalf("[Gin-OAuth] File error: %v\n", err)
	}
	json.Unmarshal(file, &c)

	conf = &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
}

// LoginHandler displays the login page & redirects to google
func LoginHandler(ctx *gin.Context) {
	state = randToken()
	session := sessions.Default(ctx)
	session.Set("state", state)
	session.Save()
	fmt.Println("LOGIN SESSION:", session.Get("userid"))
	ctx.Writer.Write([]byte("<html><title>Golang Google</title> <body> <h3>Hello!</h3> <a href='" + GetLoginURL(state) + "'><button>Login with Google!</button> </a> </body></html>"))
}

// GetLoginURL gets the google login page
func GetLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

// CheckAuth Gin handler function to check if a user is logged in
func CheckAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, err := seccookie.ReadSecureCookie(ctx, scookie)
		if err != nil {
			glog.Errorln("CHECK AUTH: not logged in")
		}
		ctx.Next()
	}
}

// DoAuth does the actual OAuth2 login procedure
func DoAuth(ctx *gin.Context) {
	// Handle the exchange code to initiate a transport.
	session := sessions.Default(ctx)
	retrievedState := session.Get("state")

	if session.Get("userid") != nil {
		return
	}

	if retrievedState != ctx.Query("state") {
		ctx.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Invalid session state: %s", retrievedState))
		return
	}

	tok, err := conf.Exchange(oauth2.NoContext, ctx.Query("code"))
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := conf.Client(oauth2.NoContext, tok)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer email.Body.Close()
	data, err := ioutil.ReadAll(email.Body)
	if err != nil {
		glog.Errorf("[Gin-OAuth] Could not read Body: %s", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var user User
	err = json.Unmarshal(data, &user)
	if err != nil {
		glog.Errorf("[Gin-OAuth] Unmarshal userinfo failed: %s", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	// save userinfo, which could be used in Handlers
	ctx.Set("user", user)

	vals := map[string]string{
		"Name":          user.Name,
		"Email":         user.Email,
		"Picture":       user.Picture,
		"GivenName":     user.GivenName,
		"FamilyName":    user.FamilyName,
		"EmailVerified": fmt.Sprintf("%v", user.EmailVerified),
		"Gender":        user.Gender,
		"Sub":           user.Sub,
		"Profile":       user.Profile,
	}
	seccookie.StoreSecureCookie(ctx, vals, scookie)

	ctx.String(http.StatusOK, "Hello %s %s", user.Name, user.Email)
}
