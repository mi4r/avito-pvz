package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mi4r/avito-pvz/internal/storage"
)

func AddProduct(
	productStorage storage.ProductRepository,
	receptionStorage storage.ReceptionRepository,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Type  string    `json:"type"`
			PVZID uuid.UUID `json:"pvzId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request")
			return
		}

		// Get open reception
		reception, err := receptionStorage.GetOpenReception(r.Context(), req.PVZID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "no open reception")
			return
		}

		product, err := productStorage.AddProduct(r.Context(), reception.ID, req.Type)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to add product")
			return
		}

		respondJSON(w, http.StatusCreated, product)
	}
}

func DeleteLastProduct(
	productStorage storage.ProductRepository,
	receptionStorage storage.ReceptionRepository,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pvzID, err := uuid.Parse(chi.URLParam(r, "pvzId"))
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid pvz id")
			return
		}

		// Get open reception
		reception, err := receptionStorage.GetOpenReception(r.Context(), pvzID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "no open reception")
			return
		}

		product, err := productStorage.GetLastProduct(r.Context(), reception.ID)
		if err != nil {
			respondError(w, http.StatusNotFound, "no products to delete")
			return
		}

		if err := productStorage.DeleteProduct(r.Context(), product.ID); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to delete product")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
