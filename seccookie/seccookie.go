package seccookie

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/securecookie"
	"gopkg.in/gin-gonic/gin.v1"
)

// CookieKey defines secure cookie parameters
var CookieKey struct {
	AppName  string `json:"app_name"`
	HashKey  []byte `json:"hash_key"`
	BlockKey []byte `json:"block_key"`
}

var credFile = "cookie_secret.json"

func init() {
	file, err := ioutil.ReadFile(credFile)
	if err != nil {
		glog.Fatalf("[SecureCookie] File error: %v\n", err)
	}
	json.Unmarshal(file, &CookieKey)
}

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

func ReadSecureCookie(ctx *gin.Context, scookie *securecookie.SecureCookie) (map[string]string, error) {
	appName := "Go-sse-secure"
	value := make(map[string]string)

	cookie, err := ctx.Request.Cookie(appName)
	if err != nil {
		glog.Errorln("Error fetching cookie:", err)
		return value, err
	}

	err = scookie.Decode(appName, cookie.Value, &value)
	if err != nil {
		glog.Errorln("Error decoding cookie:", err)
		return value, err
	}

	return value, nil
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
