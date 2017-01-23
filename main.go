package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gin-contrib/sessions"
	"gopkg.in/gin-gonic/gin.v1"

	"Go-sse/googleauth"
)

var redirectURL, credFile string

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
	secret := []byte("secret")
	sessionName := "Go-sse"

	router := gin.Default()
	// init settings for google auth
	googleauth.Setup(redirectURL, credFile, scopes, secret)
	fmt.Println("Credfile:", credFile)
	router.Use(googleauth.Session(sessionName))

	router.GET("/", rootHandler)
	router.GET("/login", googleauth.LoginHandler)

	// protected url group
	private := router.Group("/auth")
	private.Use(googleauth.Auth())
	private.GET("/", userInfoHandler)
	private.GET("/info", userInfoHandler)
	private.GET("/api", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		fmt.Println("API SESSION:", session.Get("user"))
		user, found := ctx.Get("user")
		fmt.Println("User:", user, "Found:", found)
		ctx.JSON(200, gin.H{"message": "Hello from private for groups"})
	})

	router.Run("127.0.0.1:4000")
}

func userInfoHandler(ctx *gin.Context) {
	session := sessions.Default(ctx)
	fmt.Println("Keys:", ctx.Keys)
	fmt.Println("USERINFO SESSION:", session.Get("user"))

	user, _ := ctx.GetCookie("user")
	fmt.Println("USERINFO COOKIE:", user)
	// user := ctx.MustGet("user")
	// if user != nil {
	ctx.JSON(http.StatusOK, gin.H{"Hello": "from private", "user": ctx.MustGet("user").(googleauth.User), "Ctx Keys:": ctx.Keys})
	// } else {
	// 	ctx.String(http.StatusOK, "Please login")
	// }
}

func authDefault(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Auth successful.")
}

func googleLoginHandler(ctx *gin.Context) {
	state := randToken()
	session := sessions.Default(ctx)
	session.Set("state", state)
	session.Save()
	ctx.Writer.Write([]byte("<html><title>Golang Google</title> <body> <h3>Login</h3> <a href='" + googleauth.GetLoginURL(state) + "'><button>Login with Google!</button> </a> </body></html>"))
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func rootHandler(ctx *gin.Context) {
	cuser, cusererr := ctx.Cookie("user")
	fmt.Println("ROOT COOKIE:", cuser)
	if cusererr == nil {
		ctx.String(http.StatusOK, "Hello %s", cuser)
	} else {
		ctx.String(http.StatusOK, "Hello unknown person")
	}
}
