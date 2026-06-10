package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"todoListAPI/internal/data"
	h "todoListAPI/internal/helpers"
)

// type Todo struct {
// 	ID          uuid.UUID `json:"id"`
// 	Title       string    `json:"title"`
// 	Description string    `json:"description"`
// 	Completed   bool      `json:"completed"`
// }

type UpdateTodoInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}

func authenticatedUserID(r *http.Request) (uuid.UUID, bool) {
	userIDString, ok := r.Context().Value(userIDKey).(string)
	if !ok || userIDString == "" {
		return uuid.Nil, false
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, false
	}

	return userID, true
}

func (app *application) createTodoHandler(w http.ResponseWriter, r *http.Request) {
	var todo data.Todo

	err := h.ReadJSON(w, r, &todo)
	if err != nil {
		app.badRequestResponse(w, "invalid request body")
		return
	}

	v := h.New()

	v.Check(todo.Title != "", "title", "must be provided")
	v.Check(todo.Description != "", "description", "must be provided")
	v.MinLength(todo.Title, "title", "should be atleast 3 characters long", 3)
	v.MinLength(todo.Description, "description", "should be atleast 10 characters long", 10)

	if !v.Valid() {
		app.failedValidationResponse(w, v.Errors())
		return
	}

	userID, ok := authenticatedUserID(r)
	if !ok {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	todo.ID = uuid.New()
	todo.UserID = userID
	todo.Completed = false
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = todo.CreatedAt

	err = app.models.Todos.InsertTodo(todo)
	if err != nil {
		app.logger.Error("failed to insert todo", "err", err)
		return
	}

}

func (app *application) listTodosHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	AllTodos, err := app.models.Todos.GetAllTodosForUser(userID)
	if err != nil {
		app.logger.Error("failed to return all Todos", "err", err)
		app.serverErrorResponse(w, r, err)
		return
	}

	err = h.WriteJSON(w, 200, envelope{"Todos": AllTodos})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getTodoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := h.CheckIDParam(w, r)
	if err != nil {
		app.invalidParameterResponse(w, "id")
		return
	}

	userID, ok := authenticatedUserID(r)
	if !ok {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	todo, err := app.models.Todos.GetTodo(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, h.ErrRecordNotFound):
			app.notFoundResponse(w)
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
		return
	}

	err = h.WriteJSON(w, 200, envelope{"success": todo})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) updateTodoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := h.CheckIDParam(w, r)
	if err != nil {
		app.invalidParameterResponse(w, "id")
		return
	}

	userID, ok := authenticatedUserID(r)
	if !ok {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	existingTodo, err := app.models.Todos.GetTodo(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, h.ErrRecordNotFound):
			app.notFoundResponse(w)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input UpdateTodoInput

	err = h.ReadJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, "invalid request body")
		return
	}

	if input.Title != nil {
		existingTodo.Title = *input.Title
	}

	if input.Description != nil {
		existingTodo.Description = *input.Description
	}

	if input.Completed != nil {
		existingTodo.Completed = *input.Completed
	}

	existingTodo.UpdatedAt = time.Now()

	err = app.models.Todos.UpdateTodo(existingTodo)
	if err != nil {
		app.logger.Error("failed to update todo", "err", err)
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) deleteTodoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := h.CheckIDParam(w, r)
	if err != nil {
		app.invalidParameterResponse(w, "id")
		return
	}

	userID, ok := authenticatedUserID(r)
	if !ok {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	err = app.models.Todos.DeleteTodo(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, h.ErrRecordNotFound):
			app.notFoundResponse(w)
		default:
			app.logger.Error("failed to delete todo", "err", err)
			app.serverErrorResponse(w, r, err)
		}
		return
	}

}

func (app *application) completeTodoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := h.CheckIDParam(w, r)
	if err != nil {
		app.invalidParameterResponse(w, "id")
		return
	}

	userID, ok := authenticatedUserID(r)
	if !ok {
		app.invalidAuthenticationTokenResponse(w, r)
		return
	}

	err = app.models.Todos.MarkTodoComplete(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, h.ErrRecordNotFound):
			app.notFoundResponse(w)
		default:
			app.logger.Error("failed to mark todo as complete", "err", err)
			app.serverErrorResponse(w, r, err)
		}
		return
	}
}
