package main

import "net/http"

func (app *Application) routes() http.Handler {
    mux := http.NewServeMux()

    // Public Routes
    mux.HandleFunc("/v1/healthcheck", app.healthCheckHandler)
    mux.HandleFunc("POST /v1/users", app.registerUserHandler)
    mux.HandleFunc("POST /v1/tokens/authentication", app.createAuthenticationTokenHandler)
    
    // User Routes (Protected by app.authenticate at the bottom)
    mux.HandleFunc("GET /v1/users/me", app.showUserProfileHandler)
    mux.HandleFunc("DELETE /v1/tokens/logout", app.logoutHandler)

    // Admin Routes
    // Note: requireRole returns an http.HandlerFunc, which fits mux.HandleFunc
    mux.HandleFunc("GET /v1/admin/metrics", app.requireRole("admin", app.adminMetricsHandler))
    
    // The app.authenticate(mux) wraps the ENTIRE router.
    // This ensures every request first tries to find a user token.

	//wrap the rateLimit and authenticate
	return app.rateLimit(app.authenticate(mux))
}