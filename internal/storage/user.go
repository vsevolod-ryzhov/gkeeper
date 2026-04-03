package storage

import (
	"context"
	"errors"
	"fmt"

	"gkeeper/internal/model"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
)

// CreateUser inserts a new user record and returns it. Returns ErrUserAlreadyExists on duplicate email.
func (s *PostgresStorage) CreateUser(ctx context.Context, email string, passwordHash string, salt string) (*model.UserRecord, error) {
	var user model.UserRecord

	err := (sq.Insert("users").
		Columns("email", "password_hash", "salt").
		Values(email, passwordHash, salt).
		Suffix("RETURNING id, email, password_hash, salt, created_at").
		PlaceholderFormat(sq.Dollar).
		RunWith(s.db).
		QueryRowContext(ctx)).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Salt, &user.CreatedAt)

	if err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrUserAlreadyExists
		}

		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

// GetUserByEmail retrieves a user by email address. Returns ErrUserNotFound if not found.
func (s *PostgresStorage) GetUserByEmail(ctx context.Context, email string) (*model.UserRecord, error) {
	var user model.UserRecord

	err := (sq.Select("id", "email", "password_hash", "salt", "created_at").
		From("users").
		Where(sq.Eq{"email": email}).
		RunWith(s.db).
		PlaceholderFormat(sq.Dollar).
		QueryRowContext(ctx)).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Salt, &user.CreatedAt)

	if err != nil {
		return nil, ErrUserNotFound
	}

	return &user, nil
}
