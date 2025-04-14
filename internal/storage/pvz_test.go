package storage_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/mi4r/avito-pvz/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestCreatePVZ(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)

	t.Run("success create pvz", func(t *testing.T) {
		pvzID := uuid.New()
		city := "Москва"
		registrationDate := time.Now()

		mock.ExpectQuery(`INSERT INTO pvz \(city\) VALUES \(\$1\) RETURNING id, registration_date, city`).
			WithArgs(city).
			WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
				AddRow(pvzID, registrationDate, city))

		pvz, err := store.CreatePVZ(context.Background(), city)

		assert.NoError(t, err)
		assert.Equal(t, pvzID, pvz.ID)
		assert.Equal(t, city, pvz.City)
		assert.WithinDuration(t, registrationDate, pvz.RegistrationDate, time.Second)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("invalid city", func(t *testing.T) {
		city := "Invalid City"

		pvz, err := store.CreatePVZ(context.Background(), city)

		assert.Error(t, err)
		assert.Equal(t, storage.ErrInvalidCity, err)
		assert.Empty(t, pvz)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		city := "Москва"

		mock.ExpectQuery(`INSERT INTO pvz \(city\) VALUES \(\$1\) RETURNING id, registration_date, city`).
			WithArgs(city).
			WillReturnError(sql.ErrConnDone)

		pvz, err := store.CreatePVZ(context.Background(), city)

		assert.Error(t, err)
		assert.Empty(t, pvz)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetPVZsWithReceptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	store := storage.NewPostgresStorage(db)

	t.Run("success get pvzs with receptions", func(t *testing.T) {
		pvzID := uuid.New()
		receptionID := uuid.New()
		productID := uuid.New()
		now := time.Now()
		startDate := now.Add(-24 * time.Hour)
		endDate := now
		page := 1
		limit := 10

		// Mock PVZ query
		mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
				AddRow(pvzID, now, "Москва"))

		// Mock receptions query
		mock.ExpectQuery(`SELECT r.id, r.created_at, r.pvz_id, r.status FROM receptions`).
			WithArgs(pvzID, startDate, endDate).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "pvz_id", "status"}).
				AddRow(receptionID, now, pvzID, "open"))

		// Mock products query
		mock.ExpectQuery(`SELECT id, created_at, type, reception_id FROM products`).
			WithArgs(receptionID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "type", "reception_id"}).
				AddRow(productID, now, "electronics", receptionID))

		pvzs, err := store.GetPVZsWithReceptions(context.Background(), startDate, endDate, page, limit)

		assert.NoError(t, err)
		assert.Len(t, pvzs, 1)
		assert.Equal(t, pvzID, pvzs[0].PVZ.ID)
		assert.Equal(t, "Москва", pvzs[0].PVZ.City)
		assert.Len(t, pvzs[0].Receptions, 1)
		assert.Equal(t, receptionID, pvzs[0].Receptions[0].Reception.ID)
		assert.Len(t, pvzs[0].Receptions[0].Products, 1)
		assert.Equal(t, productID, pvzs[0].Receptions[0].Products[0].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()
		page := 1
		limit := 10

		mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}))

		pvzs, err := store.GetPVZsWithReceptions(context.Background(), startDate, endDate, page, limit)

		assert.NoError(t, err)
		assert.Empty(t, pvzs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on pvz query", func(t *testing.T) {
		startDate := time.Now().Add(-24 * time.Hour)
		endDate := time.Now()
		page := 1
		limit := 10

		mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz`).
			WillReturnError(sql.ErrConnDone)

		pvzs, err := store.GetPVZsWithReceptions(context.Background(), startDate, endDate, page, limit)

		assert.Error(t, err)
		assert.Nil(t, pvzs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error on receptions query", func(t *testing.T) {
		pvzID := uuid.New()
		now := time.Now()
		startDate := now.Add(-24 * time.Hour)
		endDate := now
		page := 1
		limit := 10

		mock.ExpectQuery(`SELECT id, registration_date, city FROM pvz`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "registration_date", "city"}).
				AddRow(pvzID, now, "Москва"))

		mock.ExpectQuery(`SELECT r.id, r.created_at, r.pvz_id, r.status FROM receptions`).
			WithArgs(pvzID, startDate, endDate).
			WillReturnError(sql.ErrConnDone)

		pvzs, err := store.GetPVZsWithReceptions(context.Background(), startDate, endDate, page, limit)

		assert.Error(t, err)
		assert.Nil(t, pvzs)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
