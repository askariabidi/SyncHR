package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// HolidayHandler handles holiday-related requests
type HolidayHandler struct {
	db *sql.DB
}

// NewHolidayHandler creates a new holiday handler
func NewHolidayHandler(db *sql.DB) *HolidayHandler {
	return &HolidayHandler{db: db}
}

// GetHolidays retrieves all public holidays
func (h *HolidayHandler) GetHolidays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Query all holidays
	rows, err := h.db.Query(
		"SELECT id, name, holiday_date, description, country, created_at, updated_at FROM public_holidays ORDER BY holiday_date ASC",
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve holidays",
			Code:    500,
		})
		return
	}
	defer rows.Close()

	var holidays []PublicHoliday
	for rows.Next() {
		var h PublicHoliday
		err := rows.Scan(&h.ID, &h.Name, &h.HolidayDate, &h.Description, &h.Country, &h.CreatedAt, &h.UpdatedAt)
		if err != nil {
			continue
		}
		holidays = append(holidays, h)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Holidays retrieved successfully",
		Data: map[string]interface{}{
			"holidays": holidays,
			"count":    len(holidays),
		},
	})
}