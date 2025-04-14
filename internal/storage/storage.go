package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreatePVZ(ctx context.Context, city string) (PVZ, error)
	GetPVZsWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]PVZWithReceptions, error)
	CreateReception(ctx context.Context, pvzID uuid.UUID) (Reception, error)
	GetOpenReception(ctx context.Context, pvzID uuid.UUID) (Reception, error)
	CloseReception(ctx context.Context, receptionID uuid.UUID) error
	AddProduct(ctx context.Context, receptionID uuid.UUID, productType string) (Product, error)
	GetLastProduct(ctx context.Context, receptionID uuid.UUID) (Product, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
	CreateUser(ctx context.Context, email, passwordHash, role string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (d *PostgresStorage) Migrate(dsn string) {
	// Try auto-migration
	if err := d.autoDefaultMigrate(dsn); err != nil {
		log.Fatal(err)
	}
}

func (d *PostgresStorage) autoDefaultMigrate(dsn string) error {
	mpath, err := filepath.Abs(
		filepath.Join("internal", "storage", "migrations"))
	if err != nil {
		return err
	}

	migr, err := migrate.New(
		fmt.Sprintf("file://%s", mpath),
		dsn,
	)
	if err != nil {
		return err
	}
	return migr.Up()
}
