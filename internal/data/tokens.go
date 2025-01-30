package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/mf751/greenlight/internal/validator"
)

const (
	ScopeActivation = "activation"
)

type Token struct {
	PlainText string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

type TokenModel struct {
	DB *sql.DB
}

func generateToken(userID int64, timeToLive time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(timeToLive),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	// Genrerate random 16 bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlaintext(vdtr *validator.Validator, tokenPlaintext string) {
	vdtr.Check(tokenPlaintext != "", "token", "must be provided")
	vdtr.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

func (model TokenModel) New(userID int64, timeToLive time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, timeToLive, scope)
	if err != nil {
		return nil, err
	}

	err = model.Insert(token)
	return token, nil
}

func (model TokenModel) Insert(token *Token) error {
	sqlQuery := `
INSERT INTO tokens (hash, user_id, expiry, scope)
VALUES ($1, $2, $3, $4)
  `

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := model.DB.ExecContext(ctx, sqlQuery, args...)
	return err
}

func (model TokenModel) DeleteAllForUser(scope string, userID int64) error {
	sqlQuery := `
DELETE FROM tokens
WHERE scope = $1 AND user_id = $2
  `
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := model.DB.ExecContext(ctx, sqlQuery)
	return err
}
