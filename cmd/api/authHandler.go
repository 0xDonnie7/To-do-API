package main

import (
	"errors"
	"net/http"
	"time"
	"todoListAPI/internal/data"
	h "todoListAPI/internal/helpers"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func (app *application) generateJWT(userID uuid.UUID, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"email":   email,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(app.cfg.jwt.secret))
}

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

	err = newUser.Password.Set(input.Password)
	if err != nil {
		app.logger.Error("failed to hash password", "err", err)
		app.serverErrorResponse(w, r, err)
		return
	}

	v := h.New()
	data.ValidateUser(v, newUser, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, v.Errors())
		return
	}

	err = app.models.Users.InsertUser(newUser)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(w, v.Errors())
		default:
			app.logger.Error("failed to insert new user", "err", err)
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = h.WriteJSON(w, http.StatusCreated, envelope{"user": newUser})
	if err != nil {
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

	user, err := app.models.Users.GetUserByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, h.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.logger.Error("failed to fetch user details", "err", err)
			app.serverErrorResponse(w, r, err)

		}
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.generateJWT(user.ID, user.Email)
	if err != nil {
		app.logger.Error("failed to generate JWT token", "err", err)
		app.serverErrorResponse(w, r, err)
		return
	}

}
