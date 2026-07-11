package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// PayslipHandler handles payslip-related requests
type PayslipHandler struct {
	db *sql.DB
}

// NewPayslipHandler creates a new payslip handler
func NewPayslipHandler(db *sql.DB) *PayslipHandler {
	return &PayslipHandler{db: db}
}

// GetPayslips retrieves user's payslips
func (h *PayslipHandler) GetPayslips(w http.ResponseWriter, r *http.Request) {
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

	role, _ := r.Context().Value("role").(string)

	var query string
	var queryParam interface{}

	// HR managers can view all payslips; employees can only view their own
	if role == "hr_manager" {
		query = `SELECT id, user_id, month, year, basic_salary, allowances, deductions, tax, net_salary, working_days, leave_taken, bonus, created_at, updated_at 
		         FROM payslip ORDER BY year DESC, month DESC`
		queryParam = nil
	} else {
		query = `SELECT id, user_id, month, year, basic_salary, allowances, deductions, tax, net_salary, working_days, leave_taken, bonus, created_at, updated_at 
		         FROM payslip WHERE user_id = $1 ORDER BY year DESC, month DESC`
		queryParam = userID
	}

	var rows *sql.Rows
	var err error

	if queryParam == nil {
		rows, err = h.db.Query(query)
	} else {
		rows, err = h.db.Query(query, queryParam)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve payslips",
			Code:    500,
		})
		return
	}
	defer rows.Close()

	var payslips []Payslip
	for rows.Next() {
		var p Payslip
		err := rows.Scan(&p.ID, &p.UserID, &p.Month, &p.Year, &p.BasicSalary, &p.Allowances, &p.Deductions, &p.Tax, &p.NetSalary, &p.WorkingDays, &p.LeaveTaken, &p.Bonus, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		payslips = append(payslips, p)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Payslips retrieved successfully",
		Data: map[string]interface{}{
			"payslips": payslips,
		},
	})
}

// GetPayslipDetails retrieves a specific payslip
func (h *PayslipHandler) GetPayslipDetails(w http.ResponseWriter, r *http.Request) {
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

	role, _ := r.Context().Value("role").(string)

	// Get payslip ID from URL
	vars := mux.Vars(r)
	payslipID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid payslip ID",
			Code:    400,
		})
		return
	}

	// Fetch payslip
	var p Payslip
	err = h.db.QueryRow(
		"SELECT id, user_id, month, year, basic_salary, allowances, deductions, tax, net_salary, working_days, leave_taken, bonus, created_at, updated_at FROM payslip WHERE id = $1",
		payslipID,
	).Scan(&p.ID, &p.UserID, &p.Month, &p.Year, &p.BasicSalary, &p.Allowances, &p.Deductions, &p.Tax, &p.NetSalary, &p.WorkingDays, &p.LeaveTaken, &p.Bonus, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "not_found",
			Message: "Payslip not found",
			Code:    404,
		})
		return
	}

	// Check authorization: employees can only view their own payslip
	if role != "hr_manager" && p.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have permission to view this payslip",
			Code:    403,
		})
		return
	}

	// Fetch employee details
	var employeeName string
	var employeeEmail string
	h.db.QueryRow("SELECT first_name || ' ' || last_name, email FROM users WHERE id = $1", p.UserID).
		Scan(&employeeName, &employeeEmail)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Payslip details retrieved successfully",
		Data: map[string]interface{}{
			"payslip": p,
			"employee_name": employeeName,
			"employee_email": employeeEmail,
			"month_name": time.Month(p.Month).String(),
		},
	})
}