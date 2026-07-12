// package main

// import (
// 	"database/sql"
// 	"encoding/json"
// 	"net/http"
// 	"time"
// )

// // AttendanceHandler handles attendance-related requests
// type AttendanceHandler struct {
// 	db *sql.DB
// }

// // NewAttendanceHandler creates a new attendance handler
// func NewAttendanceHandler(db *sql.DB) *AttendanceHandler {
// 	return &AttendanceHandler{db: db}
// }

// // CheckIn handles employee check-in
// func (h *AttendanceHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	// Extract user ID from context
// 	userID, ok := r.Context().Value("user_id").(int)
// 	if !ok {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 			Code:    401,
// 		})
// 		return
// 	}

// 	var req CheckInRequest
// 	err := json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Failed to parse request body",
// 			Code:    400,
// 		})
// 		return
// 	}

// 	// Get today's date
// 	today := time.Now().Format("2006-01-02")

// 	// Check if already checked in today
// 	var existingID int
// 	err = h.db.QueryRow(
// 		"SELECT id FROM attendance WHERE user_id = $1 AND date = $2",
// 		userID, today,
// 	).Scan(&existingID)

// 	if err == nil {
// 		// Already checked in today
// 		w.WriteHeader(http.StatusBadRequest)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "already_checked_in",
// 			Message: "You have already checked in today",
// 			Code:    400,
// 		})
// 		return
// 	}

// 	if err != sql.ErrNoRows {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "database_error",
// 			Message: "Failed to query database",
// 			Code:    500,
// 		})
// 		return
// 	}

// 	// Insert check-in record
// 	var attendanceID int
// 	err = h.db.QueryRow(
// 		"INSERT INTO attendance (user_id, check_in_time, date, status) VALUES ($1, $2, $3, $4) RETURNING id",
// 		userID, req.Timestamp, today, "checked_in",
// 	).Scan(&attendanceID)

// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "checkin_error",
// 			Message: "Failed to record check-in",
// 			Code:    500,
// 		})
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(SuccessResponse{
// 		Message: "Check-in successful",
// 		Data: map[string]interface{}{
// 			"attendance_id": attendanceID,
// 			"check_in_time": req.Timestamp,
// 			"date":          today,
// 		},
// 	})
// }

// // CheckOut handles employee check-out
// func (h *AttendanceHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	// Extract user ID from context
// 	userID, ok := r.Context().Value("user_id").(int)
// 	if !ok {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 			Code:    401,
// 		})
// 		return
// 	}

// 	var req CheckOutRequest
// 	err := json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Failed to parse request body",
// 			Code:    400,
// 		})
// 		return
// 	}

// 	// Get today's date
// 	today := time.Now().Format("2006-01-02")

// 	// Find today's check-in record
// 	var attendanceID int
// 	var checkInTime time.Time
// 	err = h.db.QueryRow(
// 		"SELECT id, check_in_time FROM attendance WHERE user_id = $1 AND date = $2 AND status = $3",
// 		userID, today, "checked_in",
// 	).Scan(&attendanceID, &checkInTime)

// 	if err == sql.ErrNoRows {
// 		w.WriteHeader(http.StatusBadRequest)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "no_checkin",
// 			Message: "You have not checked in today",
// 			Code:    400,
// 		})
// 		return
// 	}

// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "database_error",
// 			Message: "Failed to query database",
// 			Code:    500,
// 		})
// 		return
// 	}

// 	// Update check-out record
// 	_, err = h.db.Exec(
// 		"UPDATE attendance SET check_out_time = $1, status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3",
// 		req.Timestamp, "checked_out", attendanceID,
// 	)

// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "checkout_error",
// 			Message: "Failed to record check-out",
// 			Code:    500,
// 		})
// 		return
// 	}

// 	// Calculate duration
// 	duration := req.Timestamp.Sub(checkInTime).Hours()

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(SuccessResponse{
// 		Message: "Check-out successful",
// 		Data: map[string]interface{}{
// 			"attendance_id": attendanceID,
// 			"check_in_time": checkInTime,
// 			"check_out_time": req.Timestamp,
// 			"duration_hours": duration,
// 			"date":           today,
// 		},
// 	})
// }

// // GetAttendanceHistory retrieves user's attendance history
// func (h *AttendanceHandler) GetAttendanceHistory(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	// Extract user ID from context
// 	userID, ok := r.Context().Value("user_id").(int)
// 	if !ok {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "unauthorized",
// 			Message: "User not authenticated",
// 			Code:    401,
// 		})
// 		return
// 	}

// 	// Get month from query parameter (default to current month)
// 	month := r.URL.Query().Get("month")
// 	year := r.URL.Query().Get("year")

// 	if month == "" {
// 		month = time.Now().Format("01")
// 	}
// 	if year == "" {
// 		year = time.Now().Format("2006")
// 	}

// 	// Query attendance records for the month
// 	rows, err := h.db.Query(
// 		`SELECT id, user_id, check_in_time, check_out_time, date, status, created_at, updated_at
// 		 FROM attendance
// 		 WHERE user_id = $1 AND TO_CHAR(date, 'YYYY-MM') = $2
// 		 ORDER BY date DESC`,
// 		userID, year+"-"+month,
// 	)

// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "database_error",
// 			Message: "Failed to retrieve attendance history",
// 			Code:    500,
// 		})
// 		return
// 	}
// 	defer rows.Close()

// 	var attendance []Attendance
// 	for rows.Next() {
// 		var a Attendance
// 		err := rows.Scan(&a.ID, &a.UserID, &a.CheckInTime, &a.CheckOutTime, &a.Date, &a.Status, &a.CreatedAt, &a.UpdatedAt)
// 		if err != nil {
// 			continue
// 		}
// 		attendance = append(attendance, a)
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(SuccessResponse{
// 		Message: "Attendance history retrieved successfully",
// 		Data: map[string]interface{}{
// 			"month":      month,
// 			"year":       year,
// 			"attendance": attendance,
// 		},
// 	})
// }

package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// AttendanceHandler handles attendance-related requests
type AttendanceHandler struct {
	db *sql.DB
}

// NewAttendanceHandler creates a new attendance handler
func NewAttendanceHandler(db *sql.DB) *AttendanceHandler {
	return &AttendanceHandler{db: db}
}

// CheckIn handles employee check-in
func (h *AttendanceHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID from context
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    401,
		})
		return
	}

	log.Printf("✅ CheckIn - UserID: %d", userID)

	// Get today's date
	today := time.Now().Format("2006-01-02")
	currentTime := time.Now()

	// Check if already checked in today
	var existingID int
	err := h.db.QueryRow(
		"SELECT id FROM attendance WHERE user_id = $1 AND date = $2",
		userID, today,
	).Scan(&existingID)

	if err == nil {
		// Already checked in today
		log.Printf("⚠️ Already checked in today: UserID %d", userID)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "already_checked_in",
			Message: "You have already checked in today",
			Code:    400,
		})
		return
	}

	if err != sql.ErrNoRows {
		log.Printf("❌ Database error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to query database",
			Code:    500,
		})
		return
	}

	// Insert check-in record with CURRENT_TIMESTAMP
	var attendanceID int
	err = h.db.QueryRow(
		"INSERT INTO attendance (user_id, check_in_time, date, status) VALUES ($1, $2, $3, $4) RETURNING id",
		userID, currentTime, today, "checked_in",
	).Scan(&attendanceID)

	if err != nil {
		log.Printf("❌ CheckIn error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "checkin_error",
			Message: "Failed to record check-in",
			Code:    500,
		})
		return
	}

	log.Printf("✅ CheckIn successful - AttendanceID: %d", attendanceID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Check-in successful",
		Data: map[string]interface{}{
			"attendance_id": attendanceID,
			"check_in_time": currentTime,
			"date":          today,
		},
	})
}

// CheckOut handles employee check-out
func (h *AttendanceHandler) CheckOut(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID from context
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    401,
		})
		return
	}

	log.Printf("✅ CheckOut - UserID: %d", userID)

	// Get today's date
	today := time.Now().Format("2006-01-02")
	currentTime := time.Now()

	// Find today's check-in record
	var attendanceID int
	var checkInTime time.Time
	err := h.db.QueryRow(
		"SELECT id, check_in_time FROM attendance WHERE user_id = $1 AND date = $2 AND status = $3",
		userID, today, "checked_in",
	).Scan(&attendanceID, &checkInTime)

	if err == sql.ErrNoRows {
		log.Printf("⚠️ No check-in found for today: UserID %d", userID)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "no_checkin",
			Message: "You have not checked in today",
			Code:    400,
		})
		return
	}

	if err != nil {
		log.Printf("❌ Database error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to query database",
			Code:    500,
		})
		return
	}

	// Update check-out record
	_, err = h.db.Exec(
		"UPDATE attendance SET check_out_time = $1, status = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3",
		currentTime, "checked_out", attendanceID,
	)

	if err != nil {
		log.Printf("❌ CheckOut error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "checkout_error",
			Message: "Failed to record check-out",
			Code:    500,
		})
		return
	}

	// Calculate duration
	duration := currentTime.Sub(checkInTime).Hours()

	log.Printf("✅ CheckOut successful - AttendanceID: %d, Duration: %.2f hours", attendanceID, duration)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Check-out successful",
		Data: map[string]interface{}{
			"attendance_id":  attendanceID,
			"check_in_time":  checkInTime,
			"check_out_time": currentTime,
			"duration_hours": duration,
			"date":           today,
		},
	})
}

// GetAttendanceHistory retrieves user's attendance history (returns ALL records, frontend filters)
func (h *AttendanceHandler) GetAttendanceHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID from context
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
			Code:    401,
		})
		return
	}

	log.Printf("📍 GetAttendanceHistory - UserID: %d", userID)

	// Query ALL attendance records for this user (no month filter)
	rows, err := h.db.Query(
		`SELECT id, user_id, check_in_time, check_out_time, date, status, created_at, updated_at 
		 FROM attendance 
		 WHERE user_id = $1
		 ORDER BY date DESC`,
		userID,
	)

	if err != nil {
		log.Printf("❌ Database error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve attendance history",
			Code:    500,
		})
		return
	}
	defer rows.Close()

	var attendanceRecords []map[string]interface{}
	for rows.Next() {
		var id, userID int
		var checkInTime, checkOutTime sql.NullTime
		var date string
		var status string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &userID, &checkInTime, &checkOutTime, &date, &status, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("❌ Scan error: %v", err)
			continue
		}

		record := map[string]interface{}{
			"id":             id,
			"user_id":        userID,
			"check_in_time":  checkInTime.Time,
			"check_out_time": checkOutTime.Time,
			"date":           date,
			"status":         status,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		}
		attendanceRecords = append(attendanceRecords, record)
	}

	log.Printf("✅ Found %d attendance records", len(attendanceRecords))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Attendance history retrieved successfully",
		Data: map[string]interface{}{
			"attendance_records": attendanceRecords,
		},
	})
}

// GetAllAttendanceRecords retrieves all employees' attendance for current month (HR only)
func (h *AttendanceHandler) GetAllAttendanceRecords(w http.ResponseWriter, r *http.Request) {
	log.Printf("🚨 GetAllAttendanceRecords CALLED!")
	w.Header().Set("Content-Type", "application/json")

	// Extract role from context
	role, _ := r.Context().Value("role").(string)

	// Only HR managers can access this
	if role != "hr_manager" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "forbidden",
			Message: "Only HR managers can access attendance records",
			Code:    403,
		})
		return
	}

	log.Printf("📊 GetAllAttendanceRecords - Fetching for HR manager")

	// Get current month and year
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)

	log.Printf("📊 Date range: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Query all attendance records for current month
	rows, err := h.db.Query(
		`SELECT id, user_id, check_in_time, check_out_time, date, status, created_at, updated_at 
		 FROM attendance 
		 WHERE date >= $1 AND date <= $2
		 ORDER BY date DESC, user_id ASC`,
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"),
	)

	if err != nil {
		log.Printf("❌ Database error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve attendance records",
			Code:    500,
		})
		return
	}
	defer rows.Close()

	var attendanceRecords []map[string]interface{}
	for rows.Next() {
		var id, userID int
		var checkInTime, checkOutTime sql.NullTime
		var date string
		var status string
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &userID, &checkInTime, &checkOutTime, &date, &status, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("❌ Scan error: %v", err)
			continue
		}

		record := map[string]interface{}{
			"id":             id,
			"user_id":        userID,
			"check_in_time":  checkInTime.Time,
			"check_out_time": checkOutTime.Time,
			"date":           date,
			"status":         status,
			"created_at":     createdAt,
			"updated_at":     updatedAt,
		}
		attendanceRecords = append(attendanceRecords, record)
	}

	log.Printf("✅ Found %d attendance records for current month", len(attendanceRecords))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Attendance records retrieved successfully",
		Data: map[string]interface{}{
			"attendance_records": attendanceRecords,
		},
	})
}
