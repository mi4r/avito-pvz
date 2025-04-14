package storage_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/mi4r/avito-pvz/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompleteWorkflow(t *testing.T) {
	// Connect to the test database
	dbURL := "postgres://mi4r:1234@localhost:5432/pvz_storage?sslmode=disable" // Change with your test DB URL for work
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Initialize storage
	store := storage.NewPostgresStorage(db)

	// Create a new PVZ
	pvz, err := store.CreatePVZ(context.Background(), "Москва")
	require.NoError(t, err)
	assert.NotEmpty(t, pvz.ID)
	assert.Equal(t, "Москва", pvz.City)

	// Create a new reception
	reception, err := store.CreateReception(context.Background(), pvz.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, reception.ID)
	assert.Equal(t, pvz.ID, reception.PVZID)
	assert.Equal(t, "in_progress", reception.Status)

	// Add 50 products to the reception
	productTypes := []string{"электроника", "одежда", "обувь"}
	for i := range 50 {
		productType := productTypes[i%len(productTypes)]
		product, err := store.AddProduct(context.Background(), reception.ID, productType)
		require.NoError(t, err)
		assert.NotEmpty(t, product.ID)
		assert.Equal(t, reception.ID, product.ReceptionID)
		assert.Equal(t, productType, product.Type)
	}

	// Verify the last product
	lastProduct, err := store.GetLastProduct(context.Background(), reception.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, lastProduct.ID)
	assert.Equal(t, reception.ID, lastProduct.ReceptionID)
	assert.Equal(t, productTypes[49%len(productTypes)], lastProduct.Type)

	// Close the reception
	err = store.CloseReception(context.Background(), reception.ID)
	require.NoError(t, err)

	// Verify the reception is closed
	closedReception, err := store.GetOpenReception(context.Background(), pvz.ID)
	assert.Error(t, err)
	assert.Empty(t, closedReception)

	// Get PVZ with receptions to verify the complete data
	pvzs, err := store.GetPVZsWithReceptions(context.Background(), time.Now().Add(-24*time.Hour), time.Now(), 1, 10)
	require.NoError(t, err)
	assert.Equal(t, pvz.ID, pvzs[0].PVZ.ID)
	assert.Len(t, pvzs[0].Receptions, 1)
	assert.Equal(t, reception.ID, pvzs[0].Receptions[0].Reception.ID)
	assert.Equal(t, "closed", pvzs[0].Receptions[0].Reception.Status)
	assert.Len(t, pvzs[0].Receptions[0].Products, 50)
}
