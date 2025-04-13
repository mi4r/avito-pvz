package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Reception struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	PVZID     uuid.UUID `json:"pvzId"`
	Status    string    `json:"status"` // 'in_progress' or 'closed'
}

type ReceptionStorage struct {
	db *pgxpool.Pool
}

func NewReceptionStorage(db *pgxpool.Pool) *ReceptionStorage {
	return &ReceptionStorage{db: db}
}

func (s *ReceptionStorage) CreateReception(ctx context.Context, pvzID uuid.UUID) (Reception, error) {
	var reception Reception
	err := s.db.QueryRow(ctx,
		`INSERT INTO receptions (pvz_id)
		VALUES ($1)
		RETURNING id, created_at, pvz_id, status`,
		pvzID,
	).Scan(&reception.ID, &reception.CreatedAt, &reception.PVZID, &reception.Status)

	if err != nil {
		return Reception{}, fmt.Errorf("failed to create reception: %w", err)
	}
	return reception, nil
}

func (s *ReceptionStorage) GetOpenReception(ctx context.Context, pvzID uuid.UUID) (Reception, error) {
	var reception Reception
	err := s.db.QueryRow(ctx,
		`SELECT id, created_at, pvz_id, status 
		FROM receptions 
		WHERE pvz_id = $1 AND status = 'in_progress'`,
		pvzID,
	).Scan(&reception.ID, &reception.CreatedAt, &reception.PVZID, &reception.Status)

	if errors.Is(err, pgx.ErrNoRows) {
		return Reception{}, ErrNotFound
	}
	return reception, err
}

func (s *ReceptionStorage) CloseReception(ctx context.Context, receptionID uuid.UUID) error {
	_, err := s.db.Exec(ctx,
		`UPDATE receptions 
		SET status = 'closed' 
		WHERE id = $1`,
		receptionID,
	)
	return err
}
