package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"todoListAPI/internal/helpers"

	"github.com/google/uuid"
)

type TodosModel struct {
	DB *sql.DB
}

type Todo struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (t *TodosModel) InsertTodo(todo Todo) error {
	query := `
	INSERT INTO todos (id,  user_id, title, description, completed, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	args := []any{todo.ID, todo.UserID, todo.Title, todo.Description, todo.Completed, todo.CreatedAt, todo.UpdatedAt}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (t *TodosModel) GetAllTodosForUser(userID uuid.UUID) ([]*Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := t.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// var todos []*Todo

	todos := make([]*Todo, 0)

	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, &todo)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (t *TodosModel) GetTodo(id uuid.UUID, userID uuid.UUID) (*Todo, error) {
	query := `
		SELECT id, user_id, title, description, completed, created_at, updated_at
		FROM todos 
		WHERE id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var todo Todo
	err := t.DB.QueryRowContext(ctx, query, id, userID).Scan(&todo.ID, &todo.UserID, &todo.Title, &todo.Description, &todo.Completed, &todo.CreatedAt, &todo.UpdatedAt)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, helpers.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &todo, nil

}

func (t *TodosModel) UpdateTodo(todo *Todo) error {
	query := `
		UPDATE todos 
		SET title = $1, description = $2, completed = $3, updated_at = $4
		WHERE id = $5 AND user_id = $6
	`

	args := []any{todo.Title, todo.Description, todo.Completed, todo.UpdatedAt, todo.ID, todo.UserID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := t.DB.ExecContext(ctx, query, args...)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return helpers.ErrRecordNotFound
	}

	return nil

}

func (t *TodosModel) DeleteTodo(id uuid.UUID, userID uuid.UUID) error {
	query := `
		DELETE FROM todos 
		WHERE id = $1 AND user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := t.DB.ExecContext(ctx, query, id, userID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return helpers.ErrRecordNotFound
	}

	return nil

}

func (t *TodosModel) MarkTodoComplete(id uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE todos
		SET completed = true, updated_at = $3
		WHERE id = $1 AND user_id = $2
	`

	args := []any{id, userID, time.Now()}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := t.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return helpers.ErrRecordNotFound
	}

	return nil
}
