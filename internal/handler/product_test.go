package handler_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mi4r/avito-pvz/internal/handler"
	"github.com/mi4r/avito-pvz/internal/storage"
	"github.com/mi4r/avito-pvz/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddProduct(t *testing.T) {
	mockStorage := mocks.NewStorage(t)
	handler := handler.AddProduct(mockStorage)

	t.Run("success add product", func(t *testing.T) {
		pvzID := uuid.New()
		receptionID := uuid.New()
		productType := "электроника"

		reqBody := []byte(`{
			"type": "электроника",
			"pvzId": "` + pvzID.String() + `"
		}`)

		// Mock expectations
		mockStorage.On("GetOpenReception", mock.Anything, pvzID).
			Return(storage.Reception{ID: receptionID}, nil)

		mockStorage.On("AddProduct", mock.Anything, receptionID, productType).
			Return(storage.Product{ID: uuid.New(), Type: productType}, nil)

		req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/products", bytes.NewBuffer([]byte("{")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("no open reception", func(t *testing.T) {
		pvzID := uuid.New()
		reqBody := []byte(`{
			"type": "электроника",
			"pvzId": "` + pvzID.String() + `"
		}`)

		mockStorage.On("GetOpenReception", mock.Anything, pvzID).
			Return(storage.Reception{}, storage.ErrNotFound)

		req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage error on add product", func(t *testing.T) {
		pvzID := uuid.New()
		receptionID := uuid.New()
		reqBody := []byte(`{
			"type": "электроника",
			"pvzId": "` + pvzID.String() + `"
		}`)

		mockStorage.On("GetOpenReception", mock.Anything, pvzID).
			Return(storage.Reception{ID: receptionID}, nil)

		mockStorage.On("AddProduct", mock.Anything, receptionID, "электроника").
			Return(storage.Product{}, errors.New("database error"))

		req := httptest.NewRequest("POST", "/products", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockStorage.AssertExpectations(t)
	})
}

func TestDeleteLastProduct(t *testing.T) {
	mockStorage := new(mocks.Storage)
	handler := handler.DeleteLastProduct(mockStorage)

	t.Run("success delete product", func(t *testing.T) {
		pvzID := uuid.New()
		receptionID := uuid.New()
		productID := uuid.New()

		// Setup router with URL param
		r := chi.NewRouter()
		r.Post("/pvz/{pvzId}/delete_last_product", handler)

		// Mock expectations
		mockStorage.On("GetOpenReception", mock.Anything, pvzID).
			Return(storage.Reception{ID: receptionID}, nil)

		mockStorage.On("GetLastProduct", mock.Anything, receptionID).
			Return(storage.Product{ID: productID}, nil)

		mockStorage.On("DeleteProduct", mock.Anything, productID).
			Return(nil)

		req := httptest.NewRequest("POST", "/pvz/"+pvzID.String()+"/delete_last_product", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("invalid pvz id", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/pvz/invalid/delete_last_product", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("no open reception", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/pvz/{pvzId}/delete_last_product", handler)
		pvzID := uuid.New()

		mockStorage.On("GetOpenReception", mock.Anything, pvzID).
			Return(storage.Reception{}, storage.ErrNotFound)

		req := httptest.NewRequest("POST", "/pvz/"+pvzID.String()+"/delete_last_product", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("no products to delete", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/pvz/{pvzId}/delete_last_product", handler)

		pvzID := uuid.New()
		receptionID := uuid.New()

		mockStorage.On("GetOpenReception", mock.Anything, pvzID).
			Return(storage.Reception{ID: receptionID}, nil)

		mockStorage.On("GetLastProduct", mock.Anything, receptionID).
			Return(storage.Product{}, storage.ErrNotFound)

		req := httptest.NewRequest("POST", "/pvz/"+pvzID.String()+"/delete_last_product", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("delete product error", func(t *testing.T) {
		pvzID := uuid.New()
		receptionID := uuid.New()
		productID := uuid.New()
		r := chi.NewRouter()
		r.Post("/pvz/{pvzId}/delete_last_product", handler)

		mockStorage.On("GetOpenReception", mock.Anything, pvzID).
			Return(storage.Reception{ID: receptionID}, nil)

		mockStorage.On("GetLastProduct", mock.Anything, receptionID).
			Return(storage.Product{ID: productID}, nil)

		mockStorage.On("DeleteProduct", mock.Anything, productID).
			Return(errors.New("database error"))

		req := httptest.NewRequest("POST", "/pvz/"+pvzID.String()+"/delete_last_product", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
