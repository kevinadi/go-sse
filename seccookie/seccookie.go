package seccookie

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/securecookie"
	"gopkg.in/gin-gonic/gin.v1"
)

var HashKey = []byte("very-secret")
var BlockKey = []byte("a-lot-secret-123")
var appName = "Go-sse"

func StoreSecureCookie(ctx *gin.Context, vals map[string]string, scookie *securecookie.SecureCookie) {
	appName := "Go-sse-secure"

	cookieEncoded, encErr := scookie.Encode(appName, vals)
	if encErr != nil {
		fmt.Println("Cookie encoding error:", encErr)
	}

	cookieStruct := &http.Cookie{
		Name:     appName,
		Value:    cookieEncoded,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(ctx.Writer, cookieStruct)
}

func ReadSecureCookie(ctx *gin.Context, scookie *securecookie.SecureCookie) map[string]string {
	appName := "Go-sse-secure"
	value := make(map[string]string)

	cookie, err := ctx.Request.Cookie(appName)
	if err != nil {
		glog.Errorln("Error fetching cookie:", err)
	}

	err = scookie.Decode(appName, cookie.Value, &value)
	if err != nil {
		glog.Errorln("Error decoding cookie:", err)
	}

	return value
}

func DeleteSecureCookie(ctx *gin.Context, scookie *securecookie.SecureCookie) {
	appName := "Go-sse-secure"
	cookieStruct := &http.Cookie{
		Name:   appName,
		Value:  "",
		Path:   "/",
		MaxAge: 0,
	}
	http.SetCookie(ctx.Writer, cookieStruct)
}
