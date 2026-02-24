package main

import (
	"net"
	"sync"
	"time"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/time/rate"

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

// RBAC mechnism
func (app *Application) requireRole(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user := app.contextGetUser(r)

        if user.IsAnonymous() {
            app.authenticationRequiredResponse(w, r)
            return
        }

        if user.UserRole != "admin" && user.UserRole != requiredRole {
            app.notPermittedResponse(w, r)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func (app *Application) rateLimit(next http.Handler) http.Handler {
    // Define a client struct to hold the limiter and last seen time for each IP
    type client struct {
        limiter  *rate.Limiter
        lastSeen time.Time
    }

    var (
        mu      sync.Mutex
        clients = make(map[string]*client)
    )

    // Launch a background goroutine to remove old entries from the map every minute
    go func() {
        for {
            time.Sleep(time.Minute)
            mu.Lock()
            for ip, client := range clients {
                if time.Since(client.lastSeen) > 3*time.Minute {
                    delete(clients, ip)
                }
            }
            mu.Unlock()
        }
    }()

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract the IP address from the request
        ip, _, err := net.SplitHostPort(r.RemoteAddr)
        if err != nil {
            app.errorResponse(w, r, http.StatusInternalServerError, err.Error())
            return
        }

        mu.Lock()

        // If IP isn't in map, initialize a new limiter (2 requests per second, burst of 4)
        if _, found := clients[ip]; !found {
            clients[ip] = &client{
                limiter: rate.NewLimiter(2, 4), 
            }
        }

        clients[ip].lastSeen = time.Now()

        if !clients[ip].limiter.Allow() {
            mu.Unlock()
            app.rateLimitExceededResponse(w, r)
            return
        }

        mu.Unlock()
        next.ServeHTTP(w, r)
    })
}