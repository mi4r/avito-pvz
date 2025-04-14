package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mi4r/avito-pvz/internal/handler"
	"github.com/mi4r/avito-pvz/internal/storage"
	"github.com/mi4r/avito-pvz/internal/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists = errors.New("user already exists")
	ErrNotFound   = errors.New("not found")
)

func TestDummyLogin(t *testing.T) {
	t.Run("success employee", func(t *testing.T) {
		reqBody := []byte(`{"role":"employee"}`)
		req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler.DummyLogin()(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]string
		json.NewDecoder(resp.Body).Decode(&response)
		token := response["token"]
		assert.NotEmpty(t, token)

		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "employee", claims["role"])
	})

	t.Run("invalid role", func(t *testing.T) {
		reqBody := []byte(`{"role":"admin"}`)
		req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler.DummyLogin()(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/dummyLogin", bytes.NewBuffer([]byte("{")))
		w := httptest.NewRecorder()

		handler.DummyLogin()(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestRegister(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	handler := handler.Register(mockRepo)

	t.Run("success registration", func(t *testing.T) {
		reqBody := []byte(`{
			"email": "test@example.com",
			"password": "password123",
			"role": "employee"
		}`)

		expectedUser := storage.User{
			Email: "test@example.com",
			Role:  "employee",
		}

		mockRepo.On("CreateUser", mock.Anything, "test@example.com", mock.Anything, "employee").
			Return(expectedUser, nil)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var user storage.User
		json.NewDecoder(resp.Body).Decode(&user)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.Equal(t, expectedUser.Role, user.Role)
	})

	t.Run("user already exists", func(t *testing.T) {
		reqBody := []byte(`{
            "email": "exists@example.com",
            "password": "password123",
            "role": "moderator"
        }`)

		// Настраиваем мок
		mockRepo.On(
			"CreateUser",
			mock.Anything,
			"exists@example.com",
			mock.Anything,
			"moderator",
		).Return(
			storage.User{},
			ErrUserExists,
		)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer([]byte("invalid")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLogin(t *testing.T) {
	mockRepo := new(mocks.UserRepository)
	handler := handler.Login(mockRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)

	t.Run("success login", func(t *testing.T) {
		reqBody := []byte(`{
			"email": "user@example.com",
			"password": "correct_password"
		}`)

		mockUser := storage.User{
			Email:        "user@example.com",
			PasswordHash: string(hashedPassword),
			Role:         "moderator",
		}

		mockRepo.On("GetUserByEmail", mock.Anything, "user@example.com").
			Return(mockUser, nil)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]string
		json.NewDecoder(resp.Body).Decode(&response)
		assert.NotEmpty(t, response["token"])
	})

	t.Run("invalid credentials", func(t *testing.T) {
		reqBody := []byte(`{
			"email": "user@example.com",
			"password": "wrong_password"
		}`)

		mockUser := storage.User{
			PasswordHash: string(hashedPassword),
		}

		mockRepo.On("GetUserByEmail", mock.Anything, "user@example.com").
			Return(mockUser, nil)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		reqBody := []byte(`{
			"email": "notfound@example.com",
			"password": "password123"
		}`)

		mockRepo.On("GetUserByEmail", mock.Anything, "notfound@example.com").
			Return(storage.User{}, errors.New("not found"))

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
