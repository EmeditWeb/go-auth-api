package main

import "net/http"

/*

Auth Lifecycle:

Register (POST /v1/users)

Login (POST /v1/tokens/authentication)

Identify (GET /v1/users/me)

Logout (DELETE /v1/tokens/logout)

*/

func (app *Application) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/healthcheck", app.healthCheckHandler)
	mux.HandleFunc("POST /v1/users", app.registerUserHandler)
	mux.HandleFunc("POST /v1/tokens/authentication", app.createAuthenticationTokenHandler)
	
	// route for knowing who the user is
	mux.HandleFunc("GET /v1/users/me", app.showUserProfileHandler)
	
	// route for logging out
	mux.HandleFunc("DELETE /v1/tokens/logout", app.logoutHandler)
	
	return app.authenticate(mux)
}



