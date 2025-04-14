package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
}

type UserStorage struct {
	db *pgxpool.Pool
}

func NewUserStorage(db *pgxpool.Pool) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) CreateUser(ctx context.Context, email, passwordHash, role string) (User, error) {
	var user User
	err := s.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id, email, role`,
		email, passwordHash, role,
	).Scan(&user.ID, &user.Email, &user.Role)
	return user, err
}

func (s *UserStorage) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := s.db.QueryRow(ctx,
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
