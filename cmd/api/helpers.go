// writing Json by hand as a string isdangerous and error-prone because a comma miss means a break,
// so we create a helper function to convert any go struct to json

package main 

import (
	"encoding/json"
	"net/http"
)

// our helper function to convert any go struct to json and sends to our client

func (app *Application) writeJSON(w http.ResponseWriter, status int, data interface{}) error {

	// takes the Go data (struct or any other) and convert it to json
	js, err := json.Marshal(data)
	if err != nil {
		app.Logger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	// new line for readability
	js = append(js, '\n')

	// set the header to json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	
	return nil
}

// readJSON decodes the request body into a destination struct
func (app *Application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Use json.NewDecoder to read the body
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		// If there is an error, return it so the handler can handle it
		return err
	}

	return nil
}
