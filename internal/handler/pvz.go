package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/mi4r/avito-pvz/internal/storage"
)

func CreatePVZ(pvzStorage storage.PVZRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			City string `json:"city"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		// Check user role
		role := r.Context().Value("role").(string)
		if role != "moderator" {
			respondError(w, http.StatusForbidden, "only moderators can create PVZ")
			return
		}

		pvz, err := pvzStorage.CreatePVZ(r.Context(), req.City)
		if err != nil {
			switch err {
			case storage.ErrInvalidCity:
				respondError(w, http.StatusBadRequest, err.Error())
			default:
				respondError(w, http.StatusInternalServerError, "failed to create PVZ")
			}
			return
		}

		respondJSON(w, http.StatusCreated, pvz)
	}
}

func GetPVZs(pvzStorage storage.PVZRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse pagination
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 30 {
			limit = 10
		}

		// Parse date filters
		var startDate time.Time
		var endDate time.Time = time.Now()
		if sd := r.URL.Query().Get("startDate"); sd != "" {
			startDate, _ = time.Parse(time.RFC3339, sd)
		}
		if ed := r.URL.Query().Get("endDate"); ed != "" {
			endDate, _ = time.Parse(time.RFC3339, ed)
		}

		// Check user role
		role := r.Context().Value("role").(string)
		if role != "moderator" && role != "employee" {
			respondError(w, http.StatusForbidden, "access denied")
			return
		}

		result, err := pvzStorage.GetPVZsWithReceptions(r.Context(), startDate, endDate, page, limit)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "failed to get PVZ data")
			return
		}

		// Transform to response format
		response := make([]storage.PVZWithReceptions, 0, len(result))
		for _, pvzWithRec := range result {
			receptions := make([]storage.ReceptionWithProducts, 0, len(pvzWithRec.Receptions))
			for _, rec := range pvzWithRec.Receptions {
				products := make([]storage.Product, 0, len(rec.Products))
				for _, p := range rec.Products {
					products = append(products, storage.Product{
						ID:          p.ID,
						CreatedAt:   p.CreatedAt,
						Type:        p.Type,
						ReceptionID: p.ReceptionID,
					})
				}

				receptions = append(receptions, storage.ReceptionWithProducts{
					Reception: storage.Reception{
						ID:        rec.Reception.ID,
						CreatedAt: rec.Reception.CreatedAt,
						PVZID:     rec.Reception.PVZID,
						Status:    rec.Reception.Status,
					},
					Products: products,
				})
			}

			response = append(response, storage.PVZWithReceptions{
				PVZ: storage.PVZ{
					ID:               pvzWithRec.PVZ.ID,
					RegistrationDate: pvzWithRec.PVZ.RegistrationDate,
					City:             pvzWithRec.PVZ.City,
				},
				Receptions: receptions,
			})
		}

		respondJSON(w, http.StatusOK, response)
	}
}
