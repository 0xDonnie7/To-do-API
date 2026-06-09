package main

import (
	"errors"
	"fmt"
	"net/http"
	h "todoListAPI/internal/helpers"
)

type envelope map[string]any

func (app *application) errorResponse(w http.ResponseWriter, status int, message any) {
	resp := envelope{"error": message}

	err := h.WriteJSON(w, status, resp)
	if err != nil {
		app.logger.Error("failed to write JSON response", "error", err)
	}
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	if r != nil {
		app.logger.Error("server error",
			"method", r.Method,
			"path", r.URL.Path,
			"error", err,
		)
	} else {
		app.logger.Error("server error", "error", err)
	}

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, http.StatusInternalServerError, message)
}

func (app *application) notFoundResponse(w http.ResponseWriter) {
	message := "the requested resource could not be found"
	app.errorResponse(w, http.StatusNotFound, message)
}

func (app *application) failedValidationResponse(w http.ResponseWriter, errors map[string]string) {
	app.errorResponse(w, http.StatusUnprocessableEntity, errors)
}

func (app *application) badRequestResponse(w http.ResponseWriter, message string) {
	app.errorResponse(w, http.StatusBadRequest, message)
}

func (app *application) invalidParameterResponse(w http.ResponseWriter, parameterName string) {
	message := fmt.Sprintf("the %s parameter is invalid", parameterName)
	app.errorResponse(w, http.StatusBadRequest, message)
}

func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, http.StatusUnauthorized, message)
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid or missing authentication token"
	app.errorResponse(w, http.StatusUnauthorized, message)
}

func (app *application) invalidJSONResponse(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, h.ErrEmptyBody):
		app.badRequestResponse(w, "request body must not be empty")
	case errors.Is(err, h.ErrBodyTooLarge):
		app.errorResponse(w, http.StatusRequestEntityTooLarge, "request body too large")
	case errors.Is(err, h.ErrUnknownField):
		app.badRequestResponse(w, err.Error()) // contains the field name
	case errors.Is(err, h.ErrContentType):
		app.errorResponse(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
	case errors.Is(err, h.ErrInvalidJSON):
		app.badRequestResponse(w, "invalid JSON")
	default:
		app.badRequestResponse(w, "invalid JSON")
	}
}
