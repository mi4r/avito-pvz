package storage_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/mi4r/avito-pvz/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestAddProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)

	t.Run("success add product", func(t *testing.T) {
		receptionID := uuid.New()
		productID := uuid.New()
		productType := "electronics"
		now := time.Now()

		mock.ExpectQuery(`INSERT INTO products \(type, reception_id\) VALUES \(\$1, \$2\) RETURNING id, created_at, type, reception_id`).
			WithArgs(productType, receptionID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "type", "reception_id"}).
				AddRow(productID, now, productType, receptionID))

		product, err := store.AddProduct(context.Background(), receptionID, productType)

		assert.NoError(t, err)
		assert.Equal(t, productID, product.ID)
		assert.Equal(t, productType, product.Type)
		assert.Equal(t, receptionID, product.ReceptionID)
		assert.WithinDuration(t, now, product.CreatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		receptionID := uuid.New()
		productType := "electronics"

		mock.ExpectQuery(`INSERT INTO products \(type, reception_id\) VALUES \(\$1, \$2\) RETURNING id, created_at, type, reception_id`).
			WithArgs(productType, receptionID).
			WillReturnError(sql.ErrConnDone)

		product, err := store.AddProduct(context.Background(), receptionID, productType)

		assert.Error(t, err)
		assert.Empty(t, product)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetLastProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)

	t.Run("success get last product", func(t *testing.T) {
		receptionID := uuid.New()
		productID := uuid.New()
		productType := "electronics"
		now := time.Now()

		mock.ExpectQuery(`SELECT id, created_at, type, reception_id FROM products WHERE reception_id = \$1 ORDER BY created_at DESC LIMIT 1`).
			WithArgs(receptionID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "type", "reception_id"}).
				AddRow(productID, now, productType, receptionID))

		product, err := store.GetLastProduct(context.Background(), receptionID)

		assert.NoError(t, err)
		assert.Equal(t, productID, product.ID)
		assert.Equal(t, productType, product.Type)
		assert.Equal(t, receptionID, product.ReceptionID)
		assert.WithinDuration(t, now, product.CreatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		receptionID := uuid.New()

		mock.ExpectQuery(`SELECT id, created_at, type, reception_id FROM products WHERE reception_id = \$1 ORDER BY created_at DESC LIMIT 1`).
			WithArgs(receptionID).
			WillReturnError(pgx.ErrNoRows)

		product, err := store.GetLastProduct(context.Background(), receptionID)

		assert.Error(t, err)
		assert.Equal(t, storage.ErrNotFound, err)
		assert.Empty(t, product)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		receptionID := uuid.New()

		mock.ExpectQuery(`SELECT id, created_at, type, reception_id FROM products WHERE reception_id = \$1 ORDER BY created_at DESC LIMIT 1`).
			WithArgs(receptionID).
			WillReturnError(sql.ErrConnDone)

		product, err := store.GetLastProduct(context.Background(), receptionID)

		assert.Error(t, err)
		assert.Empty(t, product)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)

	t.Run("success delete product", func(t *testing.T) {
		productID := uuid.New()

		mock.ExpectExec(`DELETE FROM products WHERE id = \$1`).
			WithArgs(productID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.DeleteProduct(context.Background(), productID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		productID := uuid.New()

		mock.ExpectExec(`DELETE FROM products WHERE id = \$1`).
			WithArgs(productID).
			WillReturnError(sql.ErrConnDone)

		err := store.DeleteProduct(context.Background(), productID)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
