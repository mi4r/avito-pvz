package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
}

func (s *PostgresStorage) CreateUser(ctx context.Context, email, passwordHash, role string) (User, error) {
	var user User
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, role`,
		email, passwordHash, role,
	).Scan(&user.ID, &user.Email, &user.Role)
	return user, err
}

func (s *PostgresStorage) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, role 
		FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	return user, err
}

var ErrNotFound = errors.New("not found")
