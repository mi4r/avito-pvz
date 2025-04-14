package handler_test

import (
	"bytes"
	"context"
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

func TestCreatePVZ(t *testing.T) {
	mockPVZRepo := new(mocks.PVZRepository)
	handler := handler.CreatePVZ(mockPVZRepo)

	t.Run("success create pvz", func(t *testing.T) {
		expectedCity := "Москва"
		expectedPVZ := storage.PVZ{
			ID:               uuid.New(),
			City:             expectedCity,
			RegistrationDate: time.Now(),
		}

		// Mock setup
		mockPVZRepo.On("CreatePVZ", mock.Anything, expectedCity).
			Return(expectedPVZ, nil)

		reqBody := []byte(`{"city":"Москва"}`)
		req := httptest.NewRequest("POST", "/pvz", bytes.NewBuffer(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), "role", "moderator"))

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var pvz storage.PVZ
		json.NewDecoder(w.Body).Decode(&pvz)
		assert.Equal(t, expectedCity, pvz.City)
		mockPVZRepo.AssertExpectations(t)
	})

	t.Run("forbidden for non-moderator", func(t *testing.T) {
		reqBody := []byte(`{"city":"Москва"}`)
		req := httptest.NewRequest("POST", "/pvz", bytes.NewBuffer(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), "role", "employee"))

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("invalid city", func(t *testing.T) {
		mockPVZRepo.On("CreatePVZ", mock.Anything, "Новосибирск").
			Return(storage.PVZ{}, storage.ErrInvalidCity)

		reqBody := []byte(`{"city":"Новосибирск"}`)
		req := httptest.NewRequest("POST", "/pvz", bytes.NewBuffer(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), "role", "moderator"))

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockPVZRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/pvz", bytes.NewBuffer([]byte("{")))
		req = req.WithContext(context.WithValue(req.Context(), "role", "moderator"))

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("storage error", func(t *testing.T) {
		r := chi.NewRouter()
		r.Post("/pvz", handler)
		mockPVZRepo.On("CreatePVZ", mock.Anything, "Москва").
			Return(storage.PVZ{}, errors.New("db error"))

		reqBody := []byte(`{"city":"Москва"}`)
		req := httptest.NewRequest("POST", "/pvz", bytes.NewBuffer(reqBody))
		req = req.WithContext(context.WithValue(req.Context(), "role", "moderator"))

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		mockPVZRepo.AssertExpectations(t)
	})
}

func TestGetPVZs(t *testing.T) {
	mockPVZRepo := new(mocks.PVZRepository)
	handler := handler.GetPVZs(mockPVZRepo)

	t.Run("success get pvzs", func(t *testing.T) {
		expectedPVZs := []storage.PVZWithReceptions{
			{
				PVZ: storage.PVZ{
					ID:               uuid.New(),
					City:             "Москва",
					RegistrationDate: time.Now(),
				},
				Receptions: []storage.ReceptionWithProducts{
					{
						Reception: storage.Reception{
							ID:        uuid.New(),
							CreatedAt: time.Now(),
							PVZID:     uuid.New(),
							Status:    "in_progress",
						},
						Products: []storage.Product{
							{
								ID:          uuid.New(),
								CreatedAt:   time.Now(),
								Type:        "электроника",
								ReceptionID: uuid.New(),
							},
						},
					},
				},
			},
		}

		// Mock setup
		mockPVZRepo.On("GetPVZsWithReceptions", mock.Anything, mock.Anything, mock.Anything, 1, 10).
			Return(expectedPVZs, nil)

		req := httptest.NewRequest("GET", "/pvz?page=1&limit=10", nil)
		req = req.WithContext(context.WithValue(req.Context(), "role", "moderator"))

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []storage.PVZWithReceptions
		json.NewDecoder(w.Body).Decode(&response)
		assert.Len(t, response, 1)
		assert.Equal(t, expectedPVZs[0].PVZ.City, response[0].PVZ.City)
		assert.Len(t, response[0].Receptions, 1)
		assert.Len(t, response[0].Receptions[0].Products, 1)
		mockPVZRepo.AssertExpectations(t)
	})

	t.Run("access denied for client role", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/pvz", nil)
		req = req.WithContext(context.WithValue(req.Context(), "role", "client"))

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("invalid pagination params", func(t *testing.T) {
		testCases := []struct {
			query string
			page  int
			limit int
		}{
			{"page=0&limit=5", 1, 5},
			{"page=abc&limit=def", 1, 10},
		}

		for _, tc := range testCases {
			t.Run(tc.query, func(t *testing.T) {
				// Reset mock for each test case
				mockPVZRepo.ExpectedCalls = nil

				// Setup mock for this specific test case with correct parameter types
				mockPVZRepo.On("GetPVZsWithReceptions",
					mock.MatchedBy(func(ctx context.Context) bool { return true }), // context
					mock.MatchedBy(func(t time.Time) bool { return true }),         // startDate
					mock.MatchedBy(func(t time.Time) bool { return true }),         // endDate
					tc.page,  // page
					tc.limit, // limit
				).Return([]storage.PVZWithReceptions{}, nil)

				req := httptest.NewRequest("GET", "/pvz?"+tc.query, nil)
				req = req.WithContext(context.WithValue(req.Context(), "role", "employee"))

				w := httptest.NewRecorder()
				handler(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				mockPVZRepo.AssertExpectations(t)
			})
		}
	})

	t.Run("date filters", func(t *testing.T) {
		start := time.Now().AddDate(0, -1, 0)
		end := time.Now()

		req := httptest.NewRequest("GET", "/pvz?startDate="+start.Format(time.RFC3339)+"&endDate="+end.Format(time.RFC3339), nil)
		req = req.WithContext(context.WithValue(req.Context(), "role", "moderator"))

		mockPVZRepo.On("GetPVZsWithReceptions",
			mock.MatchedBy(func(ctx context.Context) bool { return true }), // context
			start, // startDate
			end,   // endDate
			1,     // page
			10,    // limit
		).Return([]storage.PVZWithReceptions{}, nil)

		w := httptest.NewRecorder()
		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("storage error", func(t *testing.T) {
		mockPVZRepo.On("GetPVZsWithReceptions",
			mock.MatchedBy(func(ctx context.Context) bool { return true }), // context
			mock.MatchedBy(func(t time.Time) bool { return true }),         // startDate
			mock.MatchedBy(func(t time.Time) bool { return true }),         // endDate
			1,  // page
			10, // limit
		).Return(nil, errors.New("db error"))

		req := httptest.NewRequest("GET", "/pvz", nil)
		req = req.WithContext(context.WithValue(req.Context(), "role", "moderator"))

		w := httptest.NewRecorder()
		handler(w, req)

	})
}
