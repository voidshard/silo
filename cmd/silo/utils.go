package main

import (
	"strings"
	"encoding/base64"
	"net/http"
	"fmt"
)

// Given some http request, retrieve the username / password, if any.
// At the moment we only support basic auth .. but we could support more things here.
//
func (a *App) getAuth(req *http.Request) (string, string, error) {
	auth, ok := req.Header["Authorization"]
	if !ok {
		return "", "", fmt.Errorf("require Authorization header")
	}

	if len(auth) != 1 {
		return "", "", fmt.Errorf("multiple Authorization headers found")

	}

	token := strings.Replace(auth[0], "Basic ", "", 1)
	authdata, err := base64.StdEncoding.DecodeString(token)

	if err != nil {
		return "", "", fmt.Errorf("invalid Authorization header: unable to decode bas64")
	}

	basicAuthdata := strings.SplitN(string(authdata), ":", 2)
	if len(basicAuthdata) != 2 {
		return "", "", fmt.Errorf("invalid Authorization header: unable to split on ';'")
	}

	return basicAuthdata[0], basicAuthdata[1], nil
}
