package main 

import (
    "encoding/json"
    "net/http"
)

// writeJSON converts Go data to JSON and sends it to the client
func (app *Application) writeJSON(w http.ResponseWriter, status int, data interface{}) error {
    js, err := json.MarshalIndent(data, "", "\t") // Added indentation for better debugging
    if err != nil {
        return err
    }

    js = append(js, '\n')

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(js)
    
    return nil
}

// readJSON decodes the request body into a destination struct safely
func (app *Application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
    // Limit the size of the request body to 1MB to prevent DoS attacks
    maxBytes := 1_048_576
    r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

    dec := json.NewDecoder(r.Body)
    // Disallow unknown fields so the client can't send extra "junk" data
    dec.DisallowUnknownFields()

    err := dec.Decode(dst)
    if err != nil {
        return err
    }

    return nil
}


// errorResponse sends a JSON-formatted error message to the client
func (app *Application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := map[string]interface{}{"error": message}

	err := app.writeJSON(w, status, env)
	if err != nil {
		app.Logger.Println(err)
		w.WriteHeader(500)
	}
}

// Fixes: app.authenticationRequiredResponse undefined
func (app *Application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
    message := "you must be authenticated to access this resource"
    app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// Fixes: app.notPermittedResponse undefined
func (app *Application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
    message := "your user account doesn't have the necessary permissions to access this resource"
    app.errorResponse(w, r, http.StatusForbidden, message)
}

// define the ratelimitExceededResponse
func (app *Application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "ssshhh, rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}