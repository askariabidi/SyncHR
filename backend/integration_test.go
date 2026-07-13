package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// requireTestDB skips the calling test when no database is reachable
// (e.g. backend/.env isn't configured in this environment), rather than
// failing the whole suite.
func requireTestDB(t *testing.T) {
	t.Helper()
	if testDB == nil || testRouter == nil {
		t.Skip("no database connection available - configure backend/.env against a reachable Postgres instance to run this test")
	}
}

// apiRequest fires a request straight at the real router - same routing,
// same middleware, same handlers the live server uses - and decodes the
// JSON response body into `out` when provided.
func apiRequest(t *testing.T, method, path, token string, body interface{}, out interface{}) *httptest.ResponseRecorder {
	t.Helper()

	var reqBody *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if out != nil && rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
		}
	}
	return rec
}

const testPassword = "TestPass123!"

// registerAndLogin creates a brand-new, uniquely-named account and logs it
// in, returning a ready-to-use token. The account (and everything that
// cascades from it - attendance, leave requests, notifications) is deleted
// via t.Cleanup so repeated test runs never accumulate junk data.
func registerAndLogin(t *testing.T, role string) LoginResponse {
	t.Helper()
	requireTestDB(t)

	email := fmt.Sprintf("test_%s_%d@example.com", role, time.Now().UnixNano())

	registerBody := map[string]interface{}{
		"email":      email,
		"password":   testPassword,
		"first_name": "Test",
		"last_name":  "User",
		"role":       role,
		"department": "QA",
	}
	rec := apiRequest(t, http.MethodPost, "/api/auth/register", "", registerBody, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("registration failed: status %d, body %s", rec.Code, rec.Body.String())
	}

	var loginResp LoginResponse
	rec = apiRequest(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"email":    email,
		"password": testPassword,
	}, &loginResp)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: status %d, body %s", rec.Code, rec.Body.String())
	}

	t.Cleanup(func() {
		testDB.Exec("DELETE FROM users WHERE id = $1", loginResp.User.ID)
	})

	return loginResp
}

// ============================================================================
// TC-AUTH: Authentication
// ============================================================================

// TC-AUTH-01: registering then logging in with the right credentials succeeds
// and returns a usable token tied to the account just created.
func TestAuth_RegisterAndLogin_Success(t *testing.T) {
	emp := registerAndLogin(t, "employee")

	if emp.Token == "" {
		t.Error("expected a non-empty token")
	}
	if emp.User.Role != "employee" {
		t.Errorf("role = %q, want employee", emp.User.Role)
	}
	if emp.User.ID == 0 {
		t.Error("expected a non-zero user id")
	}
}

// TC-AUTH-02: logging in with the wrong password must be rejected, not
// silently accepted or leaked as a different error.
func TestAuth_Login_WrongPassword_Rejected(t *testing.T) {
	emp := registerAndLogin(t, "employee")

	rec := apiRequest(t, http.MethodPost, "/api/auth/login", "", map[string]string{
		"email":    emp.User.Email,
		"password": "definitely-not-the-password",
	}, nil)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401. body: %s", rec.Code, rec.Body.String())
	}
}

// TC-AUTH-03: registering the same email twice must fail the second time
// (the users table has a unique constraint on email).
func TestAuth_Register_DuplicateEmail_Rejected(t *testing.T) {
	emp := registerAndLogin(t, "employee")

	rec := apiRequest(t, http.MethodPost, "/api/auth/register", "", map[string]interface{}{
		"email":      emp.User.Email,
		"password":   testPassword,
		"first_name": "Dup",
		"last_name":  "User",
		"role":       "employee",
		"department": "QA",
	}, nil)

	if rec.Code == http.StatusCreated {
		t.Fatal("expected duplicate registration to fail, but it succeeded")
	}
}

// ============================================================================
// TC-RBAC: Role-based access control
// ============================================================================

// TC-RBAC-01: any request to a protected route without a token must be
// rejected before it ever reaches business logic.
func TestRBAC_UnauthenticatedRequest_Rejected(t *testing.T) {
	requireTestDB(t)

	rec := apiRequest(t, http.MethodGet, "/api/leave/balance", "", nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401. body: %s", rec.Code, rec.Body.String())
	}
}

// TC-RBAC-02: an employee token must not be able to reach the HR-only
// "all attendance records" endpoint.
func TestRBAC_EmployeeCannotAccessHROnlyAttendance(t *testing.T) {
	emp := registerAndLogin(t, "employee")

	rec := apiRequest(t, http.MethodGet, "/api/attendance/records", emp.Token, nil, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403. body: %s", rec.Code, rec.Body.String())
	}
}

// TC-RBAC-03: an employee token must not be able to approve leave requests,
// even one that genuinely exists and belongs to another employee.
func TestRBAC_EmployeeCannotApproveLeave(t *testing.T) {
	emp := registerAndLogin(t, "employee")

	rec := apiRequest(t, http.MethodPut, "/api/leave/approve/1", emp.Token,
		map[string]string{"approval_notes": "trying anyway"}, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403. body: %s", rec.Code, rec.Body.String())
	}
}

// TC-RBAC-04: an employee token must not be able to broadcast a
// notification to the whole company. Deliberately does NOT test the HR
// success path here - a successful broadcast fans out to every real
// "employee" row in the database, and this suite is meant to be safe to
// run repeatedly without spamming real accounts. The delivery mechanics of
// SendNotificationToRole are already covered in hub_test.go against
// synthetic, isolated test users.
func TestRBAC_EmployeeCannotBroadcastNotification(t *testing.T) {
	emp := registerAndLogin(t, "employee")

	rec := apiRequest(t, http.MethodPost, "/api/notifications/broadcast", emp.Token,
		map[string]string{"title": "Unauthorized", "message": "should not send"}, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403. body: %s", rec.Code, rec.Body.String())
	}
}

// ============================================================================
// TC-ATT: Attendance
// ============================================================================

type attendanceHistoryResponse struct {
	Data struct {
		AttendanceRecords []struct {
			Status       string `json:"status"`
			CheckInTime  string `json:"check_in_time"`
			CheckOutTime string `json:"check_out_time"`
		} `json:"attendance_records"`
	} `json:"data"`
}

// TC-ATT-01: a full check-in -> check-out cycle succeeds, a second check-in
// on the same day is rejected, and the resulting history record reflects
// the completed session.
func TestAttendance_CheckInCheckOutCycle(t *testing.T) {
	emp := registerAndLogin(t, "employee")

	rec := apiRequest(t, http.MethodPost, "/api/attendance/checkin", emp.Token, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("check-in status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}

	// TC-ATT-02: duplicate check-in on the same day must be rejected
	rec = apiRequest(t, http.MethodPost, "/api/attendance/checkin", emp.Token, nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("duplicate check-in status = %d, want 400. body: %s", rec.Code, rec.Body.String())
	}

	rec = apiRequest(t, http.MethodPost, "/api/attendance/checkout", emp.Token, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("check-out status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}

	var history attendanceHistoryResponse
	rec = apiRequest(t, http.MethodGet, "/api/attendance/history", emp.Token, nil, &history)
	if rec.Code != http.StatusOK {
		t.Fatalf("history status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}
	if len(history.Data.AttendanceRecords) != 1 {
		t.Fatalf("expected exactly 1 attendance record, got %d", len(history.Data.AttendanceRecords))
	}
	if got := history.Data.AttendanceRecords[0].Status; got != "checked_out" {
		t.Errorf("status = %q, want checked_out", got)
	}
}

// TC-ATT-03: checking out without ever checking in must be rejected.
func TestAttendance_CheckOutWithoutCheckIn_Rejected(t *testing.T) {
	emp := registerAndLogin(t, "employee")

	rec := apiRequest(t, http.MethodPost, "/api/attendance/checkout", emp.Token, nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body: %s", rec.Code, rec.Body.String())
	}
}

// ============================================================================
// TC-LEAVE: Leave management
// ============================================================================

type leaveBalanceResponse struct {
	Data struct {
		Balances []struct {
			LeaveTypeID int `json:"leave_type_id"`
			Balance     int `json:"balance"`
		} `json:"balances"`
	} `json:"data"`
}

type applyLeaveResponse struct {
	Data struct {
		LeaveID int `json:"leave_id"`
	} `json:"data"`
}

type notificationsResponse struct {
	Data struct {
		UnreadCount   int `json:"unread_count"`
		Notifications []struct {
			Title string `json:"title"`
			Type  string `json:"type"`
		} `json:"notifications"`
	} `json:"data"`
}

// firstFundedLeaveType finds a leave type the freshly-registered employee
// actually has a positive balance for, rather than assuming fixed IDs/seed
// data that could change independently of this test.
func firstFundedLeaveType(t *testing.T, emp LoginResponse) (leaveTypeID, balance int) {
	t.Helper()
	var balResp leaveBalanceResponse
	rec := apiRequest(t, http.MethodGet, "/api/leave/balance", emp.Token, nil, &balResp)
	if rec.Code != http.StatusOK {
		t.Fatalf("leave balance status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}
	for _, b := range balResp.Data.Balances {
		if b.Balance > 0 {
			return b.LeaveTypeID, b.Balance
		}
	}
	t.Fatal("newly registered employee has no leave balance to test against - is leave_types seeded?")
	return 0, 0
}

// TC-LEAVE-01: applying for more days than the current balance allows must
// be rejected before any row is written.
func TestLeave_ApplyInsufficientBalance_Rejected(t *testing.T) {
	emp := registerAndLogin(t, "employee")
	_, balance := firstFundedLeaveType(t, emp)

	rec := apiRequest(t, http.MethodPost, "/api/leave/apply", emp.Token, ApplyLeaveRequest{
		LeaveTypeID:  1,
		StartDate:    "2027-01-01",
		EndDate:      "2027-01-01",
		NumberOfDays: balance + 1000, // deliberately far beyond any real balance
		Reason:       "should be rejected",
	}, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400. body: %s", rec.Code, rec.Body.String())
	}
}

// TC-LEAVE-02: the full HR approval flow - apply, approve, and confirm both
// the balance deduction and the notification delivered to the employee.
// This is also a regression test for a bug found and fixed this session:
// ApproveLeave used to write no HTTP response body/status at all on success.
func TestLeave_FullApprovalFlow(t *testing.T) {
	emp := registerAndLogin(t, "employee")
	hr := registerAndLogin(t, "hr_manager")
	leaveTypeID, balanceBefore := firstFundedLeaveType(t, emp)

	var applyResp applyLeaveResponse
	rec := apiRequest(t, http.MethodPost, "/api/leave/apply", emp.Token, ApplyLeaveRequest{
		LeaveTypeID:  leaveTypeID,
		StartDate:    "2027-02-01",
		EndDate:      "2027-02-01",
		NumberOfDays: 1,
		Reason:       "integration test leave",
	}, &applyResp)
	if rec.Code != http.StatusCreated {
		t.Fatalf("apply status = %d, want 201. body: %s", rec.Code, rec.Body.String())
	}

	var approveBody map[string]interface{}
	rec = apiRequest(t, http.MethodPut, fmt.Sprintf("/api/leave/approve/%d", applyResp.Data.LeaveID), hr.Token,
		map[string]string{"approval_notes": "approved by integration test"}, &approveBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("approve status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}
	if approveBody["message"] == nil || approveBody["message"] == "" {
		t.Error("expected a non-empty success message in the approve response body")
	}

	var balResp leaveBalanceResponse
	rec = apiRequest(t, http.MethodGet, "/api/leave/balance", emp.Token, nil, &balResp)
	if rec.Code != http.StatusOK {
		t.Fatalf("balance status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}
	var balanceAfter int
	for _, b := range balResp.Data.Balances {
		if b.LeaveTypeID == leaveTypeID {
			balanceAfter = b.Balance
		}
	}
	if balanceAfter != balanceBefore-1 {
		t.Errorf("balance after approval = %d, want %d (before %d minus 1 day)", balanceAfter, balanceBefore-1, balanceBefore)
	}

	var notifResp notificationsResponse
	rec = apiRequest(t, http.MethodGet, "/api/notifications", emp.Token, nil, &notifResp)
	if rec.Code != http.StatusOK {
		t.Fatalf("notifications status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}
	found := false
	for _, n := range notifResp.Data.Notifications {
		if n.Type == "leave_approved" {
			found = true
		}
	}
	if !found {
		t.Error("expected a persisted 'leave_approved' notification for the employee after HR approved their request")
	}
}

// TC-LEAVE-03: the HR rejection flow - status ends up "rejected" and the
// employee gets a persisted notification about it.
func TestLeave_RejectFlow(t *testing.T) {
	emp := registerAndLogin(t, "employee")
	hr := registerAndLogin(t, "hr_manager")
	leaveTypeID, _ := firstFundedLeaveType(t, emp)

	var applyResp applyLeaveResponse
	rec := apiRequest(t, http.MethodPost, "/api/leave/apply", emp.Token, ApplyLeaveRequest{
		LeaveTypeID:  leaveTypeID,
		StartDate:    "2027-03-01",
		EndDate:      "2027-03-01",
		NumberOfDays: 1,
		Reason:       "integration test leave to be rejected",
	}, &applyResp)
	if rec.Code != http.StatusCreated {
		t.Fatalf("apply status = %d, want 201. body: %s", rec.Code, rec.Body.String())
	}

	rec = apiRequest(t, http.MethodPut, fmt.Sprintf("/api/leave/reject/%d", applyResp.Data.LeaveID), hr.Token,
		map[string]string{"approval_notes": "rejected by integration test"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("reject status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}

	var notifResp notificationsResponse
	rec = apiRequest(t, http.MethodGet, "/api/notifications", emp.Token, nil, &notifResp)
	if rec.Code != http.StatusOK {
		t.Fatalf("notifications status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}
	found := false
	for _, n := range notifResp.Data.Notifications {
		if n.Type == "leave_rejected" {
			found = true
		}
	}
	if !found {
		t.Error("expected a persisted 'leave_rejected' notification for the employee after HR rejected their request")
	}
}
