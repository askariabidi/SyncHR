package main

import "time"

// ============================================================================
// USER MODELS
// ============================================================================

// User represents an employee or HR manager
type User struct {
	ID            int       `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"` // Never send password to client
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	Role          string    `json:"role"` // "employee" or "hr_manager"
	Department    string    `json:"department"`
	PhoneNumber   string    `json:"phone_number"`
	DateOfJoining time.Time `json:"date_of_joining"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// LoginRequest represents the login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
	User    User   `json:"user"`
}

// RegisterRequest represents new user registration
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Role        string `json:"role"` // "employee" or "hr_manager"
	Department  string `json:"department"`
	PhoneNumber string `json:"phone_number"`
}

// ============================================================================
// ATTENDANCE MODELS
// ============================================================================

// Attendance represents daily attendance record
type Attendance struct {
	ID           int        `json:"id"`
	UserID       int        `json:"user_id"`
	CheckInTime  time.Time  `json:"check_in_time"`
	CheckOutTime *time.Time `json:"check_out_time"` // Nullable until checked out
	Date         string     `json:"date"`           // YYYY-MM-DD format
	Status       string     `json:"status"`         // "checked_in", "checked_out", "absent"
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CheckInRequest represents check-in request
type CheckInRequest struct {
	Timestamp time.Time `json:"timestamp"`
}

// CheckOutRequest represents check-out request
type CheckOutRequest struct {
	Timestamp time.Time `json:"timestamp"`
}

// BreakTime represents employee break time
type BreakTime struct {
	ID              int        `json:"id"`
	UserID          int        `json:"user_id"`
	BreakStartTime  time.Time  `json:"break_start_time"`
	BreakEndTime    *time.Time `json:"break_end_time"`
	DurationMinutes *int       `json:"duration_minutes"`
	Date            string     `json:"date"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ============================================================================
// LEAVE MODELS
// ============================================================================

// LeaveType represents types of leave (Sick, Casual, etc.)
type LeaveType struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	MaxDaysPerYear int       `json:"max_days_per_year"`
	CreatedAt      time.Time `json:"created_at"`
}

// LeaveBalance represents remaining leave balance for employee
type LeaveBalance struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	LeaveTypeID int       `json:"leave_type_id"`
	Balance     int       `json:"balance"`
	Year        int       `json:"year"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LeaveRequest represents a leave request with approval workflow
type LeaveRequest struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	LeaveTypeID   int        `json:"leave_type_id"`
	StartDate     string     `json:"start_date"` // YYYY-MM-DD
	EndDate       string     `json:"end_date"`   // YYYY-MM-DD
	NumberOfDays  int        `json:"number_of_days"`
	Reason        string     `json:"reason"`
	Status        string     `json:"status"` // "pending", "approved", "rejected"
	ApprovedBy    *int       `json:"approved_by"`
	ApprovalDate  *time.Time `json:"approval_date"`
	ApprovalNotes *string    `json:"approval_notes"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	// Additional fields from user join
	EmployeeFirstName  *string `json:"employee_first_name"`
	EmployeeLastName   *string `json:"employee_last_name"`
	EmployeeDepartment *string `json:"employee_department"`
}

// ApplyLeaveRequest represents the request to apply for leave
type ApplyLeaveRequest struct {
	LeaveTypeID  int    `json:"leave_type_id"`
	StartDate    string `json:"start_date"`
	EndDate      string `json:"end_date"`
	NumberOfDays int    `json:"number_of_days"`
	Reason       string `json:"reason"`
}

// ApproveLeaveRequest represents approval of leave request
type ApproveLeaveRequest struct {
	ApprovalNotes string `json:"approval_notes"`
}

// ============================================================================
// PAYSLIP MODELS
// ============================================================================

// Payslip represents monthly salary slip
type Payslip struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Month       int       `json:"month"` // 1-12
	Year        int       `json:"year"`
	BasicSalary float64   `json:"basic_salary"`
	Allowances  float64   `json:"allowances"`
	Deductions  float64   `json:"deductions"`
	Tax         float64   `json:"tax"`
	NetSalary   float64   `json:"net_salary"`
	WorkingDays int       `json:"working_days"`
	LeaveTaken  int       `json:"leave_taken"`
	Bonus       float64   `json:"bonus"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ============================================================================
// HOLIDAY MODELS
// ============================================================================

// PublicHoliday represents public/company holidays
type PublicHoliday struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	HolidayDate string    `json:"holiday_date"` // YYYY-MM-DD
	Description string    `json:"description"`
	Country     string    `json:"country"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ============================================================================
// NOTIFICATION MODELS
// ============================================================================

// Notification represents real-time notifications
type Notification struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	Title           string    `json:"title"`
	Message         string    `json:"message"`
	Type            string    `json:"type"` // "leave_approved", "leave_rejected", etc.
	RelatedEntityID *int      `json:"related_entity_id"`
	IsRead          bool      `json:"is_read"`
	CreatedAt       time.Time `json:"created_at"`
}

// ============================================================================
// RESPONSE MODELS
// ============================================================================

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SuccessResponse represents a successful response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ============================================================================
// JWT CLAIMS
// ============================================================================

// Claims represents JWT token claims
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}
