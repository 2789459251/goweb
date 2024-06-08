package zygo

import (
	"encoding/base64"
	"net/http"
)

type Accounts struct {
	UnAuthHandler func(ctx *Context)
	Users         map[string]string
}

func (a *Accounts) BasicAuth(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		username, password, ok := ctx.R.BasicAuth()
		if !ok {
			a.UnAuthHandler(ctx)
			return
		}
		pwd, exists := a.Users[username]
		if !exists {
			a.UnAuthHandler(ctx)
			return
		}
		if password != pwd {
			a.UnAuthHandler(ctx)
			return
		}
		ctx.Set("username", username)
		ctx.Set("password", pwd)
		next(ctx)
	}
}

func (a *Accounts) UnAuthHandlers(ctx *Context) {
	if a.UnAuthHandler != nil {
		a.UnAuthHandler(ctx)
	} else {
		ctx.W.WriteHeader(http.StatusUnauthorized)
	}
}

func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
