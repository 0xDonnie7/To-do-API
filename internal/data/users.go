package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"todoListAPI/internal/helpers"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type UsersModel struct {
	DB *sql.DB
}

var ErrDuplicateEmail = errors.New("duplicate email")

type User struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	Activated bool      `json:"activated"`
}

type password struct {
	plaintext string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = ""
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *helpers.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(helpers.Matches(email, helpers.EmailRegex), "email", "invalid email address")
}

func ValidatePasswordPlaintext(v *helpers.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be atleast 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *helpers.Validator, user *User, plaintextPassword string) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) >= 3, "name", "must be atleast 3 characters long")
	ValidateEmail(v, user.Email)

	ValidatePasswordPlaintext(v, plaintextPassword)
}

func (u *UsersModel) InsertUser(user *User) error {
	query := ` INSERT INTO users (id, name, email, password_hash, created_at, activated)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	args := []any{user.ID, user.Name, user.Email, user.Password.hash, user.CreatedAt, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := u.DB.ExecContext(ctx, query, args...)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrDuplicateEmail
		}

		return err
	}

	return nil
}

func (u *UsersModel) GetUserByEmail(email string) (*User, error) {
	query := `SELECT id, name, email, password_hash, created_at, activated FROM users WHERE email = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User

	err := u.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Activated,
	)

	if err != nil {
		switch {
		case (errors.Is(err, sql.ErrNoRows)):
			return nil, helpers.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
