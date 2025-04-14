package storage_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mi4r/avito-pvz/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)

	t.Run("success create user", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"
		passwordHash := "hashed_password"
		role := "employee"

		// Mock setup
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(email, passwordHash, role).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "role"}).
				AddRow(userID, email, role))

		user, err := store.CreateUser(context.Background(), email, passwordHash, role)

		assert.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, role, user.Role)
		assert.Empty(t, user.PasswordHash) // Password hash should not be returned
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate email", func(t *testing.T) {
		email := "test@example.com"
		passwordHash := "hashed_password"
		role := "employee"

		// Mock setup
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(email, passwordHash, role).
			WillReturnError(sql.ErrNoRows)

		_, err := store.CreateUser(context.Background(), email, passwordHash, role)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid role", func(t *testing.T) {
		email := "test@example.com"
		passwordHash := "hashed_password"
		role := "invalid_role"

		// Mock setup
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(email, passwordHash, role).
			WillReturnError(sql.ErrNoRows)

		_, err := store.CreateUser(context.Background(), email, passwordHash, role)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)

	t.Run("success get user", func(t *testing.T) {
		userID := uuid.New()
		email := "test@example.com"
		passwordHash := "hashed_password"
		role := "employee"

		// Mock setup
		mock.ExpectQuery(`SELECT id, email, password_hash, role FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
				AddRow(userID, email, passwordHash, role))

		user, err := store.GetUserByEmail(context.Background(), email)

		assert.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, passwordHash, user.PasswordHash)
		assert.Equal(t, role, user.Role)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("user not found", func(t *testing.T) {
		email := "nonexistent@example.com"

		// Mock setup
		mock.ExpectQuery(`SELECT id, email, password_hash, role FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		_, err := store.GetUserByEmail(context.Background(), email)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		email := "test@example.com"

		// Mock setup
		mock.ExpectQuery(`SELECT id, email, password_hash, role FROM users WHERE email = \$1`).
			WithArgs(email).
			WillReturnError(sql.ErrConnDone)

		_, err := store.GetUserByEmail(context.Background(), email)

		assert.Error(t, err)
		assert.NotEqual(t, storage.ErrNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
