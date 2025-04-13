package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Product struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Type        string    `json:"type"`
	ReceptionID uuid.UUID `json:"receptionId"`
}

type ProductStorage struct {
	db *pgxpool.Pool
}

func NewProductStorage(db *pgxpool.Pool) *ProductStorage {
	return &ProductStorage{db: db}
}

func (s *ProductStorage) AddProduct(ctx context.Context, receptionID uuid.UUID, productType string) (Product, error) {
	var product Product
	err := s.db.QueryRow(ctx,
		`INSERT INTO products (type, reception_id)
		VALUES ($1, $2)
		RETURNING id, created_at, type, reception_id`,
		productType, receptionID,
	).Scan(&product.ID, &product.CreatedAt, &product.Type, &product.ReceptionID)

	return product, err
}

func (s *ProductStorage) GetLastProduct(ctx context.Context, receptionID uuid.UUID) (Product, error) {
	var product Product
	err := s.db.QueryRow(ctx,
		`SELECT id, created_at, type, reception_id 
		FROM products 
		WHERE reception_id = $1 
		ORDER BY created_at DESC 
		LIMIT 1`,
		receptionID,
	).Scan(&product.ID, &product.CreatedAt, &product.Type, &product.ReceptionID)

	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	return product, err
}

func (s *ProductStorage) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`DELETE FROM products 
		WHERE id = $1`,
		productID,
	)
	return err
}
