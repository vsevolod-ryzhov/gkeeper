package storage

import (
	"context"
	"errors"
	"fmt"

	"gkeeper/internal/model"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *PostgresStorage) CreateUser(ctx context.Context, email string, passwordHash string) (*model.UserRecord, error) {
	var user model.UserRecord

	err := (sq.Insert("users").
		Columns("email", "password_hash").
		Values(email, passwordHash).
		Suffix("RETURNING id, email, password_hash, created_at").
		PlaceholderFormat(sq.Dollar).
		RunWith(s.db).
		QueryRowContext(ctx)).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrUserAlreadyExists
		}

		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}
