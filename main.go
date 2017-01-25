package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gin-contrib/sessions"
	"gopkg.in/gin-gonic/gin.v1"

	"Go-sse/googleauth"

	"github.com/gorilla/securecookie"
)

var redirectURL, credFile string

var hashKey = []byte("very-secret")
var blockKey = []byte("a-lot-secret-123")
var cookie = securecookie.New(hashKey, blockKey)
var appName = "Go-sse"

// var store = sessions.NewCookieStore([]byte("secret"))

func init() {
	bin := path.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Usage of %s
================
`, bin)
		flag.PrintDefaults()
	}
	flag.StringVar(&redirectURL, "redirect", "http://127.0.0.1:4000/auth/", "URL to be redirected to after authorization.")
	flag.StringVar(&credFile, "cred-file", "./sse_secret.json", "Credential JSON file")
}

func main() {
	flag.Parse()

	scopes := []string{
		"https://www.googleapis.com/auth/userinfo.email",
		// You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
	}
	// secret := []byte("secret")
	sessionName := "Go-sse-x"

	router := gin.Default()
	// init settings for google auth
	googleauth.Setup(redirectURL, credFile, scopes)
	fmt.Println("Credfile:", credFile)

	// router.Use(googleauth.Session(sessionName))
	store := sessions.NewCookieStore([]byte("secret"))
	router.Use(sessions.Sessions(sessionName, store))

	router.GET("/", rootHandler)
	router.GET("/auth/login", googleauth.LoginHandler)

	// protected url group
	private := router.Group("/auth")
	private.Use(googleauth.CheckAuth())
	private.GET("/", googleauth.DoAuth)
	private.GET("/info", userInfoHandler)
	private.GET("/api", apiHandler)

	router.Run("127.0.0.1:4000")
}

func apiHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	fmt.Println("API SESSION state:", session.Get("state"))
	fmt.Println("API SESSION user:", session.Get("user"))
	fmt.Println("API SESSION username:", session.Get("username"))
	fmt.Println("API SESSION userid:", session.Get("userid"))
	fmt.Println("API SESSION blah:", session)
	ctx.JSON(200, gin.H{"message": "Hello from private for groups"})
}

func userInfoHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	fmt.Println("USERINFO SESSION state:", session.Get("state"))
	fmt.Println("USERINFO SESSION user:", session.Get("user"))
	fmt.Println("USERINFO SESSION userid:", session.Get("userid"))
	fmt.Println("USERINFO SESSION username:", session.Get("username"))
	fmt.Println("USERINFO SESSION blah:", session)
	ctx.JSON(http.StatusOK, gin.H{"Hello": "from private info", "user": session.Get("username")})
}

// func authDefault(ctx *gin.Context) {
// 	ctx.String(http.StatusOK, "Auth successful.")
// }

// func googleLoginHandler(ctx *gin.Context) {
// 	state := randToken()
// 	session := sessions.Default(ctx)
// 	session.Set("state", state)
// 	session.Save()
// 	ctx.Writer.Write([]byte("<html><title>Golang Google</title> <body> <h3>Login</h3> <a href='" + googleauth.GetLoginURL(state) + "'><button>Login with Google!</button> </a> </body></html>"))
// }

// func randToken() string {
// 	b := make([]byte, 32)
// 	rand.Read(b)
// 	return base64.StdEncoding.EncodeToString(b)
// }

func rootHandler(ctx *gin.Context) {
	// cuser, cusererr := ctx.Cookie("user")
	// fmt.Println("ROOT COOKIE:", cuser)
	// if cusererr == nil {
	// 	ctx.String(http.StatusOK, "Hello %s", cuser)
	// } else {
	// 	ctx.String(http.StatusOK, "Hello unknown person")
	// }

	session := sessions.Default(ctx)
	// user := session.Get("user")
	username := session.Get("username")
	userid := session.Get("userid")
	ctx.String(http.StatusOK, "Hello %s %s", username, userid)

	// fmt.Println("ROOT user", user)
	fmt.Println("ROOTHANDLER ===")
	fmt.Println("ROOT state", session.Get("state"))
	fmt.Println("ROOT username", username)
	fmt.Println("ROOT userid", userid)
	fmt.Println("ROOT blah:", session.Get("blah"))
}
