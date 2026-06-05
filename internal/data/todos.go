package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type TodosModel struct {
	DB *sql.DB
}

type Todo struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
}

func (t *TodosModel) InsertTodo(todo Todo) error {
	query := `
	INSERT INTO todos (id, title, description, completed)
	VALUES ($1, $2, $3, $4)
	`

	args := []any{todo.ID, todo.Title, todo.Description, todo.Completed}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := t.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (t *TodosModel) GetAllTodos() ([]*Todo, error) {
	query := `
		SELECT id, title, description, completed FROM todos
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := t.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// var todos []*Todo

	todos := make([]*Todo, 0)

	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed)
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

func (t *TodosModel) GetTodo(id uuid.UUID) (*Todo, error) {
	query := `
		SELECT id, title, description, completed 
		FROM todos 
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var todo Todo
	err := t.DB.QueryRowContext(ctx, query, id).Scan(&todo.ID, &todo.Title, &todo.Description, &todo.Completed)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, errors.New("todo record not found")
		default:
			return nil, err
		}
	}

	return &todo, nil

}

func (t *TodosModel) UpdateTodo(todo *Todo) error {
	query := `
		UPDATE todos 
		SET title = $1, description = $2, completed = $3
		WHERE id = $4
	`

	args := []any{todo.Title, todo.Description, todo.Completed, todo.ID}

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
		return errors.New("record not found")
	}

	return nil

}

func (t *TodosModel) DeleteTodo(id uuid.UUID) error {
	query := `
		DELETE FROM todos 
		WHERE id = $1 
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := t.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil

}

func (t *TodosModel) MarkTodoComplete(id uuid.UUID) error {
	query := `
		UPDATE todos
		SET completed = true
		WHERE id = $1
	`

	args := []any{id}

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
		return errors.New("record not found")
	}

	return nil
}
