package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db       *sql.DB
	notifier *NotificationHandler
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *sql.DB, notifier *NotificationHandler) *AuthHandler {
	return &AuthHandler{db: db, notifier: notifier}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest
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
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.Role == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "validation_error",
			Message: "Email, password, first name, last name, and role are required",
			Code:    400,
		})
		return
	}

	// Validate role
	if req.Role != "employee" && req.Role != "hr_manager" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "validation_error",
			Message: "Role must be 'employee' or 'hr_manager'",
			Code:    400,
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "password_hash_error",
			Message: "Failed to hash password",
			Code:    500,
		})
		return
	}

	// Insert user into database
	var userID int
	err = h.db.QueryRow(
		"INSERT INTO users (email, password_hash, first_name, last_name, role, department, phone_number, date_of_joining) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id",
		req.Email, hashedPassword, req.FirstName, req.LastName, req.Role, req.Department, req.PhoneNumber, time.Now(),
	).Scan(&userID)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "registration_error",
			Message: "Email already exists or database error",
			Code:    400,
		})
		return
	}

	// Initialize leave balances for new employee
	if req.Role == "employee" {
		h.initializeLeaveBalance(userID)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "User registered successfully",
		Data: map[string]interface{}{
			"user_id": userID,
			"email":   req.Email,
		},
	})
}

// Login handles user login and generates JWT token
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
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
	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "validation_error",
			Message: "Email and password are required",
			Code:    400,
		})
		return
	}

	// Fetch user from database
	var user User
	var hashedPassword string
	err = h.db.QueryRow(
		"SELECT id, email, password_hash, first_name, last_name, role, department, phone_number, date_of_joining, created_at, updated_at FROM users WHERE email = $1",
		req.Email,
	).Scan(&user.ID, &user.Email, &hashedPassword, &user.FirstName, &user.LastName, &user.Role, &user.Department, &user.PhoneNumber, &user.DateOfJoining, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Email or password is incorrect",
				Code:    401,
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to query database",
			Code:    500,
		})
		return
	}

	// Verify password against the bcrypt hash created at registration
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Email or password is incorrect",
			Code:    401,
		})
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "token_generation_error",
			Message: "Failed to generate authentication token",
			Code:    500,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Message: "Login successful",
		Token:   token,
		User:    user,
	})
}

// GetProfile retrieves the authenticated user's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract user ID from context (set by JWT middleware)
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

	// Fetch user from database
	var user User
	err := h.db.QueryRow(
		"SELECT id, email, first_name, last_name, role, department, phone_number, date_of_joining, created_at, updated_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Role, &user.Department, &user.PhoneNumber, &user.DateOfJoining, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "user_not_found",
			Message: "User profile not found",
			Code:    404,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Profile retrieved successfully",
		Data:    user,
	})
}

// UpdateProfile updates the authenticated user's profile
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
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

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
			Code:    400,
		})
		return
	}

	// Update user in database
	_, err = h.db.Exec(
		"UPDATE users SET first_name = $1, last_name = $2, department = $3, phone_number = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5",
		user.FirstName, user.LastName, user.Department, user.PhoneNumber, userID,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "update_error",
			Message: "Failed to update profile",
			Code:    500,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Profile updated successfully",
	})
}

// GenerateJWT generates a JWT token for the user
func GenerateJWT(userID int, email, role string) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_secret_key_change_this"
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// initializeLeaveBalance sets up default leave balance for new employee
func (h *AuthHandler) initializeLeaveBalance(userID int) error {
	// Get all leave types
	rows, err := h.db.Query("SELECT id FROM leave_types")
	if err != nil {
		return err
	}
	defer rows.Close()

	currentYear := time.Now().Year()

	// Create leave balance for each leave type
	for rows.Next() {
		var leaveTypeID int
		if err := rows.Scan(&leaveTypeID); err != nil {
			return err
		}

		// Get max days for this leave type
		var maxDays int
		h.db.QueryRow("SELECT max_days_per_year FROM leave_types WHERE id = $1", leaveTypeID).Scan(&maxDays)

		// Insert leave balance
		_, err := h.db.Exec(
			"INSERT INTO leave_balance (user_id, leave_type_id, balance, year) VALUES ($1, $2, $3, $4)",
			userID, leaveTypeID, maxDays, currentYear,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAllEmployees retrieves all employees (HR only)
func (h *AuthHandler) GetAllEmployees(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract role from context
	role, _ := r.Context().Value("role").(string)

	// Only HR managers can access this
	if role != "hr_manager" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "forbidden",
			Message: "Only HR managers can access the employee list",
			Code:    403,
		})
		return
	}

	rows, err := h.db.Query(
		"SELECT id, email, first_name, last_name, role, department, phone_number, date_of_joining, created_at, updated_at FROM users ORDER BY first_name ASC",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve employees",
			Code:    500,
		})
		return
	}
	defer rows.Close()

	var employees []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.Role, &u.Department, &u.PhoneNumber, &u.DateOfJoining, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			continue
		}
		employees = append(employees, u)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Employees retrieved successfully",
		Data: map[string]interface{}{
			"employees": employees,
		},
	})
}

const tempPasswordChars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"

// generateTempPassword returns a random 12-character password drawn from a
// crypto/rand source (not math/rand - this value gets handed to HR to relay,
// so it needs to be unguessable, not just look random).
func generateTempPassword() (string, error) {
	const length = 12
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(tempPasswordChars))))
		if err != nil {
			return "", err
		}
		result[i] = tempPasswordChars[n.Int64()]
	}
	return string(result), nil
}

// ResetEmployeePassword lets an HR manager generate a new temporary password
// for an employee (HR-mediated reset - there is no self-service flow; the
// temporary password is returned once here for HR to relay to the employee
// out of band, e.g. by chat or phone, matching how "ask HR for your
// credentials" already works for new hires).
func (h *AuthHandler) ResetEmployeePassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	role, _ := r.Context().Value("role").(string)
	if role != "hr_manager" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "forbidden",
			Message: "Only HR managers can reset passwords",
			Code:    403,
		})
		return
	}

	vars := mux.Vars(r)
	targetUserID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid user ID",
			Code:    400,
		})
		return
	}

	var targetEmail string
	err = h.db.QueryRow("SELECT email FROM users WHERE id = $1", targetUserID).Scan(&targetEmail)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "not_found",
			Message: "Employee not found",
			Code:    404,
		})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to look up employee",
			Code:    500,
		})
		return
	}

	tempPassword, err := generateTempPassword()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "generation_error",
			Message: "Failed to generate a new password",
			Code:    500,
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "password_hash_error",
			Message: "Failed to hash the new password",
			Code:    500,
		})
		return
	}

	_, err = h.db.Exec("UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", hashedPassword, targetUserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to reset password",
			Code:    500,
		})
		return
	}

	// Let the employee know it happened (not the password itself - HR relays that out of band)
	h.notifier.notifyUser(targetUserID, "Password Reset", "HR has reset your password. Please contact HR to receive your new temporary password.", "password_reset", nil)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Password reset successfully",
		Data: map[string]interface{}{
			"email":               targetEmail,
			"temporary_password":  tempPassword,
		},
	})
}
