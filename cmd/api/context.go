package main

import (
	"context"
	"net/http"

	"github.com/Emeditweb/go-auth-api/internal/data"
)


// define the context type for our context key 
type contextKey string

const (
	CtxKeyUser contextKey = "user" // define the context key for the user
)
// this returns a new copy of the request with the user added to the context
func (app *Application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), CtxKeyUser, user)
	return r.WithContext(ctx)
}

// retrieve the user from the context
func (app *Application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(CtxKeyUser).(*data.User)
	if !ok {
		return data.AnonymousUser
	}
	// return the user context
	return user
}