package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Emeditweb/go-auth-api/internal/data"
)

func (app *Application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		// 1. If no header, set as Anonymous and move on
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// 2. Parse the Bearer token
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.writeJSON(w, http.StatusUnauthorized, "Invalid Authorization header format")
			return
		}

		tokenString := headerParts[1]

		// 3. Validate the token
		user, err := app.Models.Users.GetForToken(data.ScopeAuthentication, tokenString)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.writeJSON(w, http.StatusUnauthorized, "Invalid or expired token")
			default:
				app.Logger.Println(err)
				app.writeJSON(w, http.StatusInternalServerError, "Server error")
			}
			return
		}

		// 4. Update the request with the REAL user we found 🆕
		r = app.contextSetUser(r, user)

		app.Logger.Printf("Authenticated request from %s (%s)", user.Email, user.Username)

		// 5. Call the next handler ONCE with the full user data
		next.ServeHTTP(w, r)

	})
}