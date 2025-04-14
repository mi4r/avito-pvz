package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCity = errors.New("city must be Москва, Санкт-Петербург or Казань")
)

type PVZ struct {
	ID               uuid.UUID `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

type PVZWithReceptions struct {
	PVZ        PVZ
	Receptions []ReceptionWithProducts
}

type ReceptionWithProducts struct {
	Reception Reception
	Products  []Product
}

func (s *PostgresStorage) CreatePVZ(ctx context.Context, city string) (PVZ, error) {
	validCities := map[string]bool{
		"Москва":          true,
		"Санкт-Петербург": true,
		"Казань":          true,
	}
	if !validCities[city] {
		return PVZ{}, ErrInvalidCity
	}

	var pvz PVZ
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO pvz (city) 
		VALUES ($1)
		RETURNING id, registration_date, city`,
		city,
	).Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City)

	return pvz, err
}

func (s *PostgresStorage) GetPVZsWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]PVZWithReceptions, error) {
	// Get PVZs with pagination
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, registration_date, city 
		FROM pvz 
		ORDER BY registration_date DESC
		LIMIT $1 OFFSET $2`,
		limit, (page-1)*limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get pvzs: %w", err)
	}
	defer rows.Close()

	var pvzs []PVZ
	for rows.Next() {
		var pvz PVZ
		if err := rows.Scan(&pvz.ID, &pvz.RegistrationDate, &pvz.City); err != nil {
			return nil, err
		}
		pvzs = append(pvzs, pvz)
	}

	// Get receptions and products for each PVZ
	result := make([]PVZWithReceptions, 0, len(pvzs))
	for _, pvz := range pvzs {
		// Get receptions with date filter
		receptions, err := s.getReceptionsForPVZ(ctx, pvz.ID, startDate, endDate)
		if err != nil {
			return nil, err
		}
		result = append(result, PVZWithReceptions{
			PVZ:        pvz,
			Receptions: receptions,
		})
	}

	return result, nil
}

func (s *PostgresStorage) getReceptionsForPVZ(ctx context.Context, pvzID uuid.UUID, startDate, endDate time.Time) ([]ReceptionWithProducts, error) {
	// Get receptions with date filter
	rows, err := s.db.QueryContext(ctx,
		`SELECT r.id, r.created_at, r.pvz_id, r.status 
		FROM receptions r
		WHERE r.pvz_id = $1 
		AND ($2::timestamp IS NULL OR r.created_at >= $2)
		AND ($3::timestamp IS NULL OR r.created_at <= $3)
		ORDER BY r.created_at DESC`,
		pvzID, startDate, endDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get receptions: %w", err)
	}
	defer rows.Close()

	var receptions []ReceptionWithProducts
	for rows.Next() {
		var r Reception
		if err := rows.Scan(&r.ID, &r.CreatedAt, &r.PVZID, &r.Status); err != nil {
			return nil, err
		}

		// Get products for reception
		products, err := s.getProductsForReception(ctx, r.ID)
		if err != nil {
			return nil, err
		}

		receptions = append(receptions, ReceptionWithProducts{
			Reception: r,
			Products:  products,
		})
	}

	return receptions, nil
}

func (s *PostgresStorage) getProductsForReception(ctx context.Context, receptionID uuid.UUID) ([]Product, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, created_at, type, reception_id 
		FROM products 
		WHERE reception_id = $1
		ORDER BY created_at DESC`,
		receptionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.CreatedAt, &p.Type, &p.ReceptionID); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}
