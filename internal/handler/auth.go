package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mi4r/avito-pvz/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}

func DummyLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Role string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}
		if req.Role != "employee" && req.Role != "moderator" {
			respondError(w, http.StatusBadRequest, "invalid role")
			return
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"role": req.Role,
			"exp":  time.Now().Add(72 * time.Hour).Unix(),
		})
		tokenStr, _ := token.SignedString([]byte("secret"))
		respondJSON(w, http.StatusOK, map[string]string{"token": tokenStr})
	}
}

func Register(db storage.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to hash password")
			return
		}

		user, err := db.CreateUser(r.Context(), req.Email, string(hash), req.Role)
		if err != nil {
			respondError(w, http.StatusConflict, "user already exists")
			return
		}

		respondJSON(w, http.StatusCreated, user)
	}
}

func Login(db storage.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}

		user, err := db.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":  user.ID.String(),
			"role": user.Role,
			"exp":  time.Now().Add(72 * time.Hour).Unix(),
		})
		tokenStr, _ := token.SignedString([]byte("secret"))
		respondJSON(w, http.StatusOK, map[string]string{"token": tokenStr})
	}
}
