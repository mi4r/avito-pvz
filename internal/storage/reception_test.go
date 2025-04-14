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

func TestCreateReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)

	t.Run("success create reception", func(t *testing.T) {
		pvzID := uuid.New()
		receptionID := uuid.New()
		now := time.Now()

		mock.ExpectQuery(`INSERT INTO receptions \(pvz_id\) VALUES \(\$1\) RETURNING id, created_at, pvz_id, status`).
			WithArgs(pvzID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "pvz_id", "status"}).
				AddRow(receptionID, now, pvzID, "in_progress"))

		reception, err := store.CreateReception(context.Background(), pvzID)

		assert.NoError(t, err)
		assert.Equal(t, receptionID, reception.ID)
		assert.Equal(t, pvzID, reception.PVZID)
		assert.Equal(t, "in_progress", reception.Status)
		assert.WithinDuration(t, now, reception.CreatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		pvzID := uuid.New()

		mock.ExpectQuery(`INSERT INTO receptions \(pvz_id\) VALUES \(\$1\) RETURNING id, created_at, pvz_id, status`).
			WithArgs(pvzID).
			WillReturnError(sql.ErrConnDone)

		reception, err := store.CreateReception(context.Background(), pvzID)

		assert.Error(t, err)
		assert.Empty(t, reception)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetOpenReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)

	t.Run("success get open reception", func(t *testing.T) {
		pvzID := uuid.New()
		receptionID := uuid.New()
		now := time.Now()

		mock.ExpectQuery(`SELECT id, created_at, pvz_id, status FROM receptions WHERE pvz_id = \$1 AND status = 'in_progress'`).
			WithArgs(pvzID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "pvz_id", "status"}).
				AddRow(receptionID, now, pvzID, "in_progress"))

		reception, err := store.GetOpenReception(context.Background(), pvzID)

		assert.NoError(t, err)
		assert.Equal(t, receptionID, reception.ID)
		assert.Equal(t, pvzID, reception.PVZID)
		assert.Equal(t, "in_progress", reception.Status)
		assert.WithinDuration(t, now, reception.CreatedAt, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		pvzID := uuid.New()

		mock.ExpectQuery(`SELECT id, created_at, pvz_id, status FROM receptions WHERE pvz_id = \$1 AND status = 'in_progress'`).
			WithArgs(pvzID).
			WillReturnError(pgx.ErrNoRows)

		reception, err := store.GetOpenReception(context.Background(), pvzID)

		assert.Error(t, err)
		assert.Equal(t, storage.ErrNotFound, err)
		assert.Empty(t, reception)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		pvzID := uuid.New()

		mock.ExpectQuery(`SELECT id, created_at, pvz_id, status FROM receptions WHERE pvz_id = \$1 AND status = 'in_progress'`).
			WithArgs(pvzID).
			WillReturnError(sql.ErrConnDone)

		reception, err := store.GetOpenReception(context.Background(), pvzID)

		assert.Error(t, err)
		assert.Empty(t, reception)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCloseReception(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)

	t.Run("success close reception", func(t *testing.T) {
		receptionID := uuid.New()

		mock.ExpectExec(`UPDATE receptions SET status = 'closed' WHERE id = \$1`).
			WithArgs(receptionID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := store.CloseReception(context.Background(), receptionID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		receptionID := uuid.New()

		mock.ExpectExec(`UPDATE receptions SET status = 'closed' WHERE id = \$1`).
			WithArgs(receptionID).
			WillReturnError(sql.ErrConnDone)

		err := store.CloseReception(context.Background(), receptionID)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
