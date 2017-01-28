package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/static"
	"github.com/gorilla/securecookie"
	"gopkg.in/gin-gonic/gin.v1"

	"Go-sse/googleauth"
	"Go-sse/seccookie"
)

var redirectURL, credFile string

var scookie = securecookie.New(seccookie.CookieKey.HashKey, seccookie.CookieKey.BlockKey)

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
	sessionName := "Go-sse"

	router := gin.Default()
	// init settings for google auth
	googleauth.Setup(redirectURL, credFile, scopes)
	fmt.Println("Credfile:", credFile)

	store := sessions.NewCookieStore([]byte("secret"))
	router.Use(sessions.Sessions(sessionName, store))

	// static as root
	router.Use(static.Serve("/", static.LocalFile("static", true)))

	// login page
	router.GET("/auth/login", googleauth.LoginHandler)

	// protected url group
	private := router.Group("/auth")
	private.Use(googleauth.CheckAuth())
	private.GET("/", googleauth.DoAuth)
	private.GET("/info", userInfoHandler)
	private.GET("/api", apiHandler)
	private.GET("/logout", logoutHandler)

	router.Run("127.0.0.1:4000")
}

func logoutHandler(ctx *gin.Context) {
	seccookie.DeleteSecureCookie(ctx, scookie)
	ctx.JSON(200, gin.H{"logout": true})
}

func apiHandler(ctx *gin.Context) {
	vals, err := seccookie.ReadSecureCookie(ctx, scookie)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
	} else {
		ctx.JSON(200, gin.H{"message": "Hello from private for groups", "user": vals["Email"], "name": vals["Name"]})
	}
}

func userInfoHandler(ctx *gin.Context) {
	vals, err := seccookie.ReadSecureCookie(ctx, scookie)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
	} else {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"Name":          vals["Name"],
				"GivenName":     vals["GivenName"],
				"FamilyName":    vals["FamilyName"],
				"Email":         vals["Email"],
				"Picture":       vals["Picture"],
				"Gender":        vals["Gender"],
				"EmailVerified": vals["EmailVerified"],
			})
	}
}

func rootHandler(ctx *gin.Context) {
	output := "Hello unknown person"
	vals, err := seccookie.ReadSecureCookie(ctx, scookie)
	if err == nil {
		output = fmt.Sprintf("Hello %s %s", vals["Name"], vals["Email"])
	}
	ctx.String(http.StatusOK, output)
}
