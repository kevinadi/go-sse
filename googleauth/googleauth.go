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
	"gopkg.in/gin-gonic/gin.v1"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Credentials stores google client-ids.
type Credentials struct {
	ClientID     string `json:"clientid"`
	ClientSecret string `json:"secret"`
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

// var store sessions.CookieStore

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// Setup the authorization path
// func Setup(redirectURL, credFile string, scopes []string, secret []byte) {
// 	store = sessions.NewCookieStore(secret)
// 	var c Credentials
// 	file, err := ioutil.ReadFile(credFile)
// 	if err != nil {
// 		glog.Fatalf("[Gin-OAuth] File error: %v\n", err)
// 	}
// 	json.Unmarshal(file, &c)

// 	conf = &oauth2.Config{
// 		ClientID:     c.ClientID,
// 		ClientSecret: c.ClientSecret,
// 		RedirectURL:  redirectURL,
// 		Scopes:       scopes,
// 		Endpoint:     google.Endpoint,
// 	}
// }
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

// func Session(name string) gin.HandlerFunc {
// 	return sessions.Sessions(name, store)
// }

func LoginHandler(ctx *gin.Context) {
	state = randToken()
	session := sessions.Default(ctx)
	session.Set("state", state)
	session.Save()
	fmt.Println("LOGIN SESSION:", session.Get("userid"))
	ctx.Writer.Write([]byte("<html><title>Golang Google</title> <body> <h3>Hello!</h3> <a href='" + GetLoginURL(state) + "'><button>Login with Google!</button> </a> </body></html>"))
}

func GetLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

// Auth is the google authorization middleware. You can use them to protect a routergroup.
// Example:
//
//        private.Use(google.Auth())
//        private.GET("/", UserInfoHandler)
//        private.GET("/api", func(ctx *gin.Context) {
//            ctx.JSON(200, gin.H{"message": "Hello from private for groups"})
//        })
//    func UserInfoHandler(ctx *gin.Context) {
//        ctx.JSON(http.StatusOK, gin.H{"Hello": "from private", "user": ctx.MustGet("user").(google.User)})
//    }
func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Handle the exchange code to initiate a transport.
		session := sessions.Default(ctx)
		retrievedState := session.Get("state")
		fmt.Println("BEFORE AUTH: RETRIEVED STATE:", retrievedState)
		fmt.Println("BEFORE AUTH: SESSION username:", session.Get("username"))
		fmt.Println("BEFORE AUTH: SESSION userid:", session.Get("userid"))
		fmt.Println("BEFORE AUTH: SESSION blah:", session.Get("blah"))

		// cuser, cusererr := ctx.Cookie("user")
		// fmt.Println("AUTH: COOKIE USER:", cuser)
		// if cusererr == nil {
		// 	return
		// }
		if session.Get("userid") != nil {
			return
		}

		sessionUserID := session.Get("userid")
		fmt.Println("SESSION USER ID:", sessionUserID)
		sessionUser := session.Get("user")
		fmt.Println("SESSION USER:", sessionUser)
		ctxKeys := ctx.Keys
		fmt.Println("CTX KEYS:", ctxKeys)

		if retrievedState != ctx.Query("state") {
			ctx.String(http.StatusUnauthorized, "Not logged in")
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
		session.Set("user", user)
		session.Set("username", user.Name)
		session.Set("userid", user.Email)
		session.Set("blah", "blah1")
		session.Save()

		fmt.Println("AFTER AUTH SESSION state:", session.Get("state"))
		fmt.Println("AFTER AUTH SESSION user:", session.Get("user"))
		fmt.Println("AFTER AUTH SESSION userid:", session.Get("userid"))
		fmt.Println("AFTER AUTH SESSION username:", session.Get("username"))
		fmt.Println("AFTER AUTH SESSION blah:", session.Get("blah"))

		session.Set("blah", "blah2")
		session.Save()
		fmt.Println("AFTER AUTH SESSION blah:", session.Get("blah"))

		ctx.SetCookie("user", user.Email, 300, "/", "127.0.0.1", false, true)

	}
}