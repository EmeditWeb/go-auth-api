package validator

import "regexp"

// EmailRX is a regular expression for validating email addresses
var EmailRX = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type Validator struct {
	Errors map[string]string
}

// New creates a new Validator instance
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid returns true if there are no errors
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error to the validator
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error only if the validation check fails
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}

}