package main

import (
	"net/http"
	"time"
	"todoListAPI/internal/data"
	h "todoListAPI/internal/helpers"

	"github.com/google/uuid"
)

func (app *application) signupHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := h.ReadJSON(w, r, &input)
	if err != nil {
		app.logger.Error("failed to read signUp JSON", "err", err)
		app.invalidJSONResponse(w, err)
		return
	}

	newUser := &data.User{
		ID:        uuid.New(),
		Name:      input.Name,
		Email:     input.Email,
		CreatedAt: time.Now(),
		Activated: false,
	}

	v := h.New()

	data.ValidateUser(v, newUser, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, v.Errors())
		return
	}

	err = newUser.Password.Set(input.Password)
	if err != nil {
		app.logger.Error("failed to hash password", "err", err)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Users.InsertUser(newUser)
	if err != nil {
		app.logger.Error("failed to insert new user", "err", err)
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := h.ReadJSON(w, r, &input)
	if err != nil {
		app.logger.Error("failed to read Login JSON details", "err", err)
		app.invalidJSONResponse(w, err)
		return
	}

}
