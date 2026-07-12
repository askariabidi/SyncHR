package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// LeaveHandler handles leave-related requests
type LeaveHandler struct {
	db *sql.DB
}

// NewLeaveHandler creates a new leave handler
func NewLeaveHandler(db *sql.DB) *LeaveHandler {
	return &LeaveHandler{db: db}
}

// ApplyLeave handles leave request submission
func (h *LeaveHandler) ApplyLeave(w http.ResponseWriter, r *http.Request) {
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

	var req ApplyLeaveRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
			Code:    400,
		})
		return
	}

	// Validate required fields
	if req.LeaveTypeID == 0 || req.StartDate == "" || req.EndDate == "" || req.NumberOfDays == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "validation_error",
			Message: "Leave type, start date, end date, and number of days are required",
			Code:    400,
		})
		return
	}

	// Check leave balance
	var balance int
	currentYear := time.Now().Year()
	err = h.db.QueryRow(
		"SELECT balance FROM leave_balance WHERE user_id = $1 AND leave_type_id = $2 AND year = $3",
		userID, req.LeaveTypeID, currentYear,
	).Scan(&balance)

	if err != nil || balance < req.NumberOfDays {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "insufficient_balance",
			Message: "Insufficient leave balance for this request",
			Code:    400,
		})
		return
	}

	// Insert leave request
	var leaveID int
	err = h.db.QueryRow(
		`INSERT INTO leave_request (user_id, leave_type_id, start_date, end_date, number_of_days, reason, status) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		userID, req.LeaveTypeID, req.StartDate, req.EndDate, req.NumberOfDays, req.Reason, "pending",
	).Scan(&leaveID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "insert_error",
			Message: "Failed to submit leave request",
			Code:    500,
		})
		return
	}

	// // Send notification to HR managers via WebSocket
	// notification := Notification{
	// 	UserID:  userID,
	// 	Title:   "New Leave Request",
	// 	Message: "A new leave request has been submitted",
	// 	Type:    "leave_request",
	// 	RelatedEntityID: &leaveID,
	// 	IsRead:  false,
	// 	CreatedAt: time.Now(),
	// }
	// BroadcastNotification(notification)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Leave request submitted successfully",
		Data: map[string]interface{}{
			"leave_id": leaveID,
		},
	})
}

// GetLeaveBalance retrieves user's leave balance
func (h *LeaveHandler) GetLeaveBalance(w http.ResponseWriter, r *http.Request) {
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

	// Query leave balance for current year
	rows, err := h.db.Query(
		`SELECT lb.id, lb.user_id, lb.leave_type_id, lb.balance, lb.year, lt.name 
		 FROM leave_balance lb 
		 JOIN leave_types lt ON lb.leave_type_id = lt.id 
		 WHERE lb.user_id = $1 AND lb.year = $2`,
		userID, time.Now().Year(),
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve leave balance",
			Code:    500,
		})
		return
	}
	defer rows.Close()

	var balances []map[string]interface{}
	for rows.Next() {
		var id, userID, leaveTypeID, balance, year int
		var leaveTypeName string
		err := rows.Scan(&id, &userID, &leaveTypeID, &balance, &year, &leaveTypeName)
		if err != nil {
			continue
		}
		balances = append(balances, map[string]interface{}{
			"id":            id,
			"user_id":       userID,
			"leave_type_id": leaveTypeID,
			"leave_type":    leaveTypeName,
			"balance":       balance,
			"year":          year,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Leave balance retrieved successfully",
		Data: map[string]interface{}{
			"balances": balances,
			"year":     time.Now().Year(),
		},
	})
}

// GetLeaveRequests retrieves leave requests for authenticated user
// func (h *LeaveHandler) GetLeaveRequests(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	// Extract user ID and role from context
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

// 	role, _ := r.Context().Value("role").(string)

// 	var query string
// 	var queryParam interface{}

// 	// HR managers see all leave requests; employees see their own
// 	if role == "hr_manager" {
// 		query = `SELECT id, user_id, leave_type_id, start_date, end_date, number_of_days, reason, status, approved_by, approval_date, approval_notes, created_at, updated_at
// 		         FROM leave_request ORDER BY created_at DESC`
// 		queryParam = nil
// 	} else {
// 		query = `SELECT id, user_id, leave_type_id, start_date, end_date, number_of_days, reason, status, approved_by, approval_date, approval_notes, created_at, updated_at
// 		         FROM leave_request WHERE user_id = $1 ORDER BY created_at DESC`
// 		queryParam = userID
// 	}

// 	var rows *sql.Rows
// 	var err error

// 	if queryParam == nil {
// 		rows, err = h.db.Query(query)
// 	} else {
// 		rows, err = h.db.Query(query, queryParam)
// 	}

// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "database_error",
// 			Message: "Failed to retrieve leave requests",
// 			Code:    500,
// 		})
// 		return
// 	}
// 	defer rows.Close()

// 	var leaveRequests []LeaveRequest = []LeaveRequest{}
// 	for rows.Next() {
// 		var lr LeaveRequest
// 		err := rows.Scan(&lr.ID, &lr.UserID, &lr.LeaveTypeID, &lr.StartDate, &lr.EndDate, &lr.NumberOfDays, &lr.Reason, &lr.Status, &lr.ApprovedBy, &lr.ApprovalDate, &lr.ApprovalNotes, &lr.CreatedAt, &lr.UpdatedAt)
// 		if err != nil {
// 			continue
// 		}
// 		leaveRequests = append(leaveRequests, lr)
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(SuccessResponse{
// 		Message: "Leave requests retrieved successfully",
// 		Data: map[string]interface{}{
// 			"leave_requests": leaveRequests,
// 		},
// 	})
// }

// GetLeaveRequests retrieves leave requests for authenticated user
// func (h *LeaveHandler) GetLeaveRequests(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	// Extract user ID and role from context
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

// 	role, _ := r.Context().Value("role").(string)

// 	var query string
// 	var rows *sql.Rows
// 	var err error

// 	// HR managers see all leave requests; employees see their own
// 	if role == "hr_manager" {
// 		query = `SELECT id, user_id, leave_type_id, start_date, end_date, number_of_days, reason, status, approved_by, approval_date, approval_notes, created_at, updated_at
// 		         FROM leave_request ORDER BY created_at DESC`
// 		rows, err = h.db.Query(query)
// 	} else {
// 		query = `SELECT id, user_id, leave_type_id, start_date, end_date, number_of_days, reason, status, approved_by, approval_date, approval_notes, created_at, updated_at
// 		         FROM leave_request WHERE user_id = $1 ORDER BY created_at DESC`
// 		rows, err = h.db.Query(query, userID)
// 	}

// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		json.NewEncoder(w).Encode(ErrorResponse{
// 			Error:   "database_error",
// 			Message: "Failed to retrieve leave requests",
// 			Code:    500,
// 		})
// 		return
// 	}
// 	defer rows.Close()

// 	leaveRequests := []LeaveRequest{}
// 	for rows.Next() {
// 		var lr LeaveRequest
// 		err := rows.Scan(&lr.ID, &lr.UserID, &lr.LeaveTypeID, &lr.StartDate, &lr.EndDate, &lr.NumberOfDays, &lr.Reason, &lr.Status, &lr.ApprovedBy, &lr.ApprovalDate, &lr.ApprovalNotes, &lr.CreatedAt, &lr.UpdatedAt)
// 		if err != nil {
// 			continue
// 		}
// 		leaveRequests = append(leaveRequests, lr)
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(SuccessResponse{
// 		Message: "Leave requests retrieved successfully",
// 		Data: map[string]interface{}{
// 			"leave_requests": leaveRequests,
// 		},
// 	})
// }

// GetLeaveRequests retrieves leave requests for authenticated user
func (h *LeaveHandler) GetLeaveRequests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID and role from context
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

	role, _ := r.Context().Value("role").(string)

	// Debug logging
	log.Printf("🔍 GetLeaveRequests - UserID: %d, Role: %s", userID, role)

	var query string
	var rows *sql.Rows
	var err error

	// HR managers see all leave requests; employees see their own
	if role == "hr_manager" {
		query = `SELECT id, user_id, leave_type_id, start_date, end_date, number_of_days, reason, status, approved_by, approval_date, approval_notes, created_at, updated_at 
		         FROM leave_request ORDER BY created_at DESC`
		log.Printf("📊 HR Query executing for all requests")
		rows, err = h.db.Query(query)
	} else {
		query = `SELECT id, user_id, leave_type_id, start_date, end_date, number_of_days, reason, status, approved_by, approval_date, approval_notes, created_at, updated_at 
		         FROM leave_request WHERE user_id = $1 ORDER BY created_at DESC`
		log.Printf("📊 Employee Query executing for user_id: %d", userID)
		rows, err = h.db.Query(query, userID)
	}

	if err != nil {
		log.Printf("❌ Database error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve leave requests",
			Code:    500,
		})
		return
	}
	defer rows.Close()

	leaveRequests := []LeaveRequest{}
	for rows.Next() {
		var lr LeaveRequest
		err := rows.Scan(&lr.ID, &lr.UserID, &lr.LeaveTypeID, &lr.StartDate, &lr.EndDate, &lr.NumberOfDays, &lr.Reason, &lr.Status, &lr.ApprovedBy, &lr.ApprovalDate, &lr.ApprovalNotes, &lr.CreatedAt, &lr.UpdatedAt)
		if err != nil {
			log.Printf("❌ Scan error: %v", err)
			continue
		}
		leaveRequests = append(leaveRequests, lr)
	}

	log.Printf("✅ Found %d leave requests", len(leaveRequests))

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Leave requests retrieved successfully",
		Data: map[string]interface{}{
			"leave_requests": leaveRequests,
		},
	})
}

// ApproveLeave handles leave request approval (HR only)
func (h *LeaveHandler) ApproveLeave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID and role from context
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

	role, _ := r.Context().Value("role").(string)
	if role != "hr_manager" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "forbidden",
			Message: "Only HR managers can approve leaves",
			Code:    403,
		})
		return
	}

	// Get leave ID from URL
	vars := mux.Vars(r)
	leaveID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid leave ID",
			Code:    400,
		})
		return
	}

	var req ApproveLeaveRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
			Code:    400,
		})
		return
	}

	// Get leave request details
	var lr LeaveRequest
	err = h.db.QueryRow("SELECT id, user_id, leave_type_id, number_of_days FROM leave_request WHERE id = $1", leaveID).
		Scan(&lr.ID, &lr.UserID, &lr.LeaveTypeID, &lr.NumberOfDays)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "not_found",
			Message: "Leave request not found",
			Code:    404,
		})
		return
	}

	// Update leave request
	_, err = h.db.Exec(
		`UPDATE leave_request SET status = $1, approved_by = $2, approval_date = CURRENT_TIMESTAMP, approval_notes = $3, updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $4`,
		"approved", userID, req.ApprovalNotes, leaveID,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "update_error",
			Message: "Failed to approve leave",
			Code:    500,
		})
		return
	}

	// Deduct from leave balance
	_, err = h.db.Exec(
		`UPDATE leave_balance SET balance = balance - $1 WHERE user_id = $2 AND leave_type_id = $3 AND year = $4`,
		lr.NumberOfDays, lr.UserID, lr.LeaveTypeID, time.Now().Year(),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "balance_error",
			Message: "Failed to update leave balance",
			Code:    500,
		})
		return
	}

	// Send notification to employee via WebSocket
	// notification := Notification{
	// 	UserID:          lr.UserID,
	// 	Title:           "Leave Approved",
	// 	Message:         "Your leave request has been approved",
	// 	Type:            "leave_approved",
	// 	RelatedEntityID: &leaveID,
	// 	IsRead:          false,
	// 	CreatedAt:       time.Now(),
	// }
	// SendNotificationToUser(lr.UserID, notification)

	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(SuccessResponse{
	// 	Message: "Leave request approved successfully",
	// })
}

// RejectLeave handles leave request rejection (HR only)
func (h *LeaveHandler) RejectLeave(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID and role from context
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

	role, _ := r.Context().Value("role").(string)
	if role != "hr_manager" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "forbidden",
			Message: "Only HR managers can reject leaves",
			Code:    403,
		})
		return
	}

	// Get leave ID from URL
	vars := mux.Vars(r)
	leaveID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid leave ID",
			Code:    400,
		})
		return
	}

	var req ApproveLeaveRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
			Code:    400,
		})
		return
	}

	// Get leave request details (to get employee user ID)
	var employeeUserID int
	err = h.db.QueryRow("SELECT user_id FROM leave_request WHERE id = $1", leaveID).Scan(&employeeUserID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "not_found",
			Message: "Leave request not found",
			Code:    404,
		})
		return
	}

	// Update leave request
	_, err = h.db.Exec(
		`UPDATE leave_request SET status = $1, approved_by = $2, approval_date = CURRENT_TIMESTAMP, approval_notes = $3, updated_at = CURRENT_TIMESTAMP 
		 WHERE id = $4`,
		"rejected", userID, req.ApprovalNotes, leaveID,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "update_error",
			Message: "Failed to reject leave",
			Code:    500,
		})
		return
	}

	// Send notification to employee via WebSocket
	// notification := Notification{
	// 	UserID:          employeeUserID,
	// 	Title:           "Leave Rejected",
	// 	Message:         "Your leave request has been rejected",
	// 	Type:            "leave_rejected",
	// 	RelatedEntityID: &leaveID,
	// 	IsRead:          false,
	// 	CreatedAt:       time.Now(),
	// }
	// SendNotificationToUser(employeeUserID, notification)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Leave request rejected successfully",
	})
}
