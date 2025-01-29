package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/mf751/greenlight/internal/validator"
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

var ErrDuplicateEmail = errors.New("duplicate email")

type UserModel struct {
	DB *sql.DB
}

// a pointer to distinguish from preset to ""
type password struct {
	plaintext *string
	hash      []byte
}

func (psd *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	psd.plaintext = &plaintextPassword
	psd.hash = hash

	return nil
}

func (psd *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(psd.hash, []byte(plaintextPassword))
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

func ValidateEmail(vdtr *validator.Validator, email string) {
	vdtr.Check(email != "", "email", "must be a valid email address")
	vdtr.Check(
		validator.Matches(email, validator.EmailRX),
		"email",
		"must be a valid email address",
	)
}

func ValidatePasswordPlaintext(vdtr *validator.Validator, password string) {
	vdtr.Check(password != "", "password", "must be provided")
	vdtr.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	vdtr.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(vdtr *validator.Validator, user *User) {
	vdtr.Check(user.Name != "", "name", "must be provided")
	vdtr.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(vdtr, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(vdtr, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func (model UserModel) Insert(user *User) error {
	sqlQuery := `
INSERT INTO users (name, email, password_hash, activated)
VALUES($1, $2, $3, $4)
RETURNING id, created_at, version
  `
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := model.DB.QueryRowContext(ctx, sqlQuery, args...).
		Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (model UserModel) GetByEmail(email string) (*User, error) {
	sqlQuery := `
SELECT id, created_at, name, email, password_hash, activated, version
FROM users
WHERE email = $2
  `
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := model.DB.QueryRowContext(ctx, sqlQuery, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (model UserModel) Update(user *User) error {
	sqlQuery := `
UPDATE users
SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
WHERE id = $5 AND version = $6
RETURNING version
  `
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := model.DB.QueryRowContext(ctx, sqlQuery, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
