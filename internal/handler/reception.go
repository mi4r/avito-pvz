package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mi4r/avito-pvz/internal/storage"
)

func CreateReception(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			PVZID uuid.UUID `json:"pvzId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}

		// Check for existing open reception
		_, err := db.GetOpenReception(r.Context(), req.PVZID)
		if err == nil {
			respondError(w, http.StatusConflict, "open reception already exists")
			return
		}

		reception, err := db.CreateReception(r.Context(), req.PVZID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to create reception")
			return
		}

		respondJSON(w, http.StatusCreated, reception)
	}
}

func CloseLastReception(db storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pvzID, err := uuid.Parse(chi.URLParam(r, "pvzId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid pvz id")
			return
		}

		reception, err := db.GetOpenReception(r.Context(), pvzID)
		if err != nil {
			respondError(w, http.StatusNotFound, "no open reception found")
			return
		}

		if err := db.CloseReception(r.Context(), reception.ID); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to close reception")
			return
		}
		reception.Status = "closed"
		respondJSON(w, http.StatusOK, reception)
	}
}
