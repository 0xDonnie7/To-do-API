package helpers

import "regexp"

type Validator struct {
	errors map[string]string
}

var EmailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9-]+\\.[a-zA-Z]{2,}$")

func New() *Validator {
	return &Validator{
		errors: make(map[string]string),
	}
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.errors[key]; !exists {
		v.errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func (v *Validator) MinLength(value, key, message string, n int) {
	v.Check(len(value) >= n, key, message)
}

func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

func (v *Validator) Errors() map[string]string {
	return v.errors
}

// func (v *Validator) IsEmailValid(email string) bool {
// 	return emailRegex.MatchString(email)
// }

func Matches(values string, rx *regexp.Regexp) bool {
	return rx.MatchString(values)
}
