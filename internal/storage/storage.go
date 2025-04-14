package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PVZRepository interface {
	CreatePVZ(ctx context.Context, city string) (PVZ, error)
	GetPVZsWithReceptions(ctx context.Context, startDate, endDate time.Time, page, limit int) ([]PVZWithReceptions, error)
}

type ReceptionRepository interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (Reception, error)
	GetOpenReception(ctx context.Context, pvzID uuid.UUID) (Reception, error)
	CloseReception(ctx context.Context, receptionID uuid.UUID) error
}

type ProductRepository interface {
	AddProduct(ctx context.Context, receptionID uuid.UUID, productType string) (Product, error)
	GetLastProduct(ctx context.Context, receptionID uuid.UUID) (Product, error)
	DeleteProduct(ctx context.Context, productID uuid.UUID) error
}

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash, role string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
}
