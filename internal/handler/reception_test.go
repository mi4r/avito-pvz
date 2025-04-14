package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mi4r/avito-pvz/internal/handler"
	"github.com/mi4r/avito-pvz/internal/storage"
	"github.com/mi4r/avito-pvz/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateReception(t *testing.T) {
	mockReceptionRepo := new(mocks.ReceptionRepository)
	handler := handler.CreateReception(mockReceptionRepo)

	t.Run("success create reception", func(t *testing.T) {
		pvzID := uuid.New()
		expectedReception := storage.Reception{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			PVZID:     pvzID,
			Status:    "in_progress",
		}

		// Mock setup
		mockReceptionRepo.On("GetOpenReception", mock.Anything, pvzID).
			Return(storage.Reception{}, storage.ErrNotFound)

		mockReceptionRepo.On("CreateReception", mock.Anything, pvzID).
			Return(expectedReception, nil)

		reqBody := []byte(`{"pvzId":"` + pvzID.String() + `"}`)
		req := httptest.NewRequest("POST", "/receptions", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var reception storage.Reception
		json.NewDecoder(w.Body).Decode(&reception)
		assert.Equal(t, expectedReception.ID, reception.ID)
		assert.Equal(t, expectedReception.PVZID, reception.PVZID)
		assert.Equal(t, expectedReception.Status, reception.Status)
		mockReceptionRepo.AssertExpectations(t)
	})

	t.Run("open reception already exists", func(t *testing.T) {
		pvzID := uuid.New()
		existingReception := storage.Reception{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			PVZID:     pvzID,
			Status:    "in_progress",
		}

		// Mock setup
		mockReceptionRepo.On("GetOpenReception", mock.Anything, pvzID).
			Return(existingReception, nil)

		reqBody := []byte(`{"pvzId":"` + pvzID.String() + `"}`)
		req := httptest.NewRequest("POST", "/receptions", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockReceptionRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/receptions", bytes.NewBuffer([]byte("{")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("storage error on create", func(t *testing.T) {
		pvzID := uuid.New()

		// Mock setup
		mockReceptionRepo.On("GetOpenReception", mock.Anything, pvzID).
			Return(storage.Reception{}, storage.ErrNotFound)

		mockReceptionRepo.On("CreateReception", mock.Anything, pvzID).
			Return(storage.Reception{}, errors.New("db error"))

		reqBody := []byte(`{"pvzId":"` + pvzID.String() + `"}`)
		req := httptest.NewRequest("POST", "/receptions", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockReceptionRepo.AssertExpectations(t)
	})
}

func TestCloseLastReception(t *testing.T) {
	mockReceptionRepo := new(mocks.ReceptionRepository)
	handler := handler.CloseLastReception(mockReceptionRepo)

	t.Run("success close reception", func(t *testing.T) {
		pvzID := uuid.New()
		receptionID := uuid.New()
		reception := storage.Reception{
			ID:        receptionID,
			CreatedAt: time.Now(),
			PVZID:     pvzID,
			Status:    "in_progress",
		}

		// Setup router with URL param
		r := chi.NewRouter()
		r.Post("/pvz/{pvzId}/close_last_reception", handler)

		// Mock setup
		mockReceptionRepo.On("GetOpenReception", mock.Anything, pvzID).
			Return(reception, nil)

		mockReceptionRepo.On("CloseReception", mock.Anything, receptionID).
			Return(nil)

		req := httptest.NewRequest("POST", "/pvz/"+pvzID.String()+"/close_last_reception", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response storage.Reception
		json.NewDecoder(w.Body).Decode(&response)
		assert.Equal(t, receptionID, response.ID)
		assert.Equal(t, "closed", response.Status)
		mockReceptionRepo.AssertExpectations(t)
	})

	t.Run("invalid pvz id", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/pvz/invalid/close_last_reception", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("no open reception found", func(t *testing.T) {
		pvzID := uuid.New()
		r := chi.NewRouter()
		r.Post("/pvz/{pvzId}/close_last_reception", handler)

		mockReceptionRepo.On("GetOpenReception", mock.Anything, pvzID).
			Return(storage.Reception{}, storage.ErrNotFound)

		req := httptest.NewRequest("POST", "/pvz/"+pvzID.String()+"/close_last_reception", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockReceptionRepo.AssertExpectations(t)
	})

	t.Run("storage error on close", func(t *testing.T) {
		pvzID := uuid.New()
		receptionID := uuid.New()
		reception := storage.Reception{
			ID:        receptionID,
			CreatedAt: time.Now(),
			PVZID:     pvzID,
			Status:    "in_progress",
		}
		r := chi.NewRouter()
		r.Post("/pvz/{pvzId}/close_last_reception", handler)

		mockReceptionRepo.On("GetOpenReception", mock.Anything, pvzID).
			Return(reception, nil)

		mockReceptionRepo.On("CloseReception", mock.Anything, receptionID).
			Return(errors.New("db error"))

		req := httptest.NewRequest("POST", "/pvz/"+pvzID.String()+"/close_last_reception", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockReceptionRepo.AssertExpectations(t)
	})
}
