package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/Emeditweb/go-auth-api/internal/data"
	"github.com/Emeditweb/go-auth-api/internal/validator"
)

// healthCheckHandler writes a json response with a message 
func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the user from the request context using the helper we built
	user := app.contextGetUser(r)

	// Create a dynamic response map
	envData := map[string]interface{}{
		"status":      "available",
		"environment": app.Config.Env,
		"location":    "Learn2Earn HQ, Lagos Nigeria",
		"developer":   "EmeditWeb",
	}

	// If the user is authenticated (not anonymous), add their info to the response
	if !user.IsAnonymous() {
		envData["authenticated_user"] = map[string]string{
			"username": user.Username,
			"email":    user.Email,
		}
	}

	err := app.writeJSON(w, http.StatusOK, envData)
	if err != nil {
		app.Logger.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// registerUserHandler handles the registration of new users
func (app *Application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.Logger.Println(err)
		app.writeJSON(w, http.StatusBadRequest, "Invalid Request Payload")
		return
	}

	v := validator.New()

	v.Check(input.Username != "", "username", "must be provided")
	v.Check(len(input.Username) >= 3, "username", "must be at least 3 characters long")
	v.Check(len(input.Username) <= 72, "username", "must not be more than 72 characters")

	v.Check(input.Email != "", "email", "must be provided")
	v.Check(validator.EmailRX.MatchString(input.Email), "email", "must be a valid email address")

	v.Check(input.Password != "", "password", "must be provided")
	v.Check(len(input.Password) >= 8, "password", "must be at least 8 characters long")
	v.Check(len(input.Password) <= 72, "password", "must not be more than 72 characters")

	if !v.Valid() {
		app.writeJSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	user := data.User{
		Username: input.Username,
		Email:    input.Email,
		UserRole:     "user",
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.Logger.Println(err)
		app.writeJSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	err = app.Models.Users.InsertUser(&user)
	if err != nil {
		app.Logger.Println(err)
		app.writeJSON(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	app.writeJSON(w, http.StatusCreated, user)
}

// createAuthenticationTokenHandler verifies credentials and issues a token
func (app *Application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	v := validator.New()
	v.Check(input.Email != "", "email", "must be provided")
	v.Check(input.Password != "", "password", "must be provided")

	if !v.Valid() {
		app.writeJSON(w, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	user, err := app.Models.Users.GetUserByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.writeJSON(w, http.StatusUnauthorized, "Invalid authentication credentials")
		default:
			app.Logger.Println(err)
			app.writeJSON(w, http.StatusInternalServerError, "Server error")
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.Logger.Println(err)
		app.writeJSON(w, http.StatusInternalServerError, "Server error")
		return
	}

	if !match {
		app.writeJSON(w, http.StatusUnauthorized, "Invalid authentication credentials")
		return
	}

	token, err := app.Models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.Logger.Println(err)
		app.writeJSON(w, http.StatusInternalServerError, "Server error")
		return
	}

	app.writeJSON(w, http.StatusCreated, map[string]interface{}{"authentication_token": token})
}

// lets check who the user is
// showUserProfileHandler returns the details of the currently authenticated user
func (app *Application) showUserProfileHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Get the user from the context
    user := app.contextGetUser(r)

    // 2. Check if the user is anonymous (e.g., they didn't provide a token)
    if user.IsAnonymous() {
        app.writeJSON(w, http.StatusUnauthorized, "You must be authenticated to access this resource")
        return
    }

    // 3. Send the user details back as JSON
    // The struct tags in models.go will automatically hide the password
    err := app.writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
    if err != nil {
        app.Logger.Println(err)
        app.writeJSON(w, http.StatusInternalServerError, "Internal Server Error")
    }
}

func (app *Application) logoutHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Get the user from context
    user := app.contextGetUser(r)

    // 2. If they are already anonymous, they are technically already "logged out"
    if user.IsAnonymous() {
        app.writeJSON(w, http.StatusUnauthorized, "You are not logged in")
        return
    }

    // 3. Delete all authentication tokens for this user
    err := app.Models.Tokens.DeleteAllForUser(data.ScopeAuthentication, user.ID)
    if err != nil {
        app.Logger.Println(err)
        app.writeJSON(w, http.StatusInternalServerError, "Server error")
        return
    }

    // 4. Send a success message
    app.writeJSON(w, http.StatusOK, map[string]string{"message": "Successfully logged out"})
}