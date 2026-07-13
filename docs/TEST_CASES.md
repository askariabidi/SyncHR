# SyncHR — Test Cases

This document lists the test cases covering SyncHR's main workflows. Wherever a case is backed by an automated Go test, the **Automated** column names the exact test function — run `go test ./...` from `backend/` to execute all of them for real against a live Postgres database (see `README.md`). Cases without an automated entry are manual/UI checks (frontend behavior that isn't meaningfully unit-testable without a browser automation harness) and are marked **Manual**.

All automated cases below were run at the time of writing: **27/27 passing**, including under Go's `-race` detector (`go test -race ./...`), which specifically exercises the WebSocket hub's concurrent goroutine/channel design under contention.

## Authentication Workflow

| Test Case ID | Workflow | Scenario | Precondition | Steps | Expected Result | Automated |
|---|---|---|---|---|---|---|
| AUTH-01 | Registration | Register a new employee account | None | `POST /api/auth/register` with valid email/password/role | 201 Created, account persisted | `TestAuth_RegisterAndLogin_Success` |
| AUTH-02 | Login | Log in with correct credentials right after registering | Account exists (AUTH-01) | `POST /api/auth/login` with the same credentials | 200 OK, response includes a valid JWT and the correct role | `TestAuth_RegisterAndLogin_Success` |
| AUTH-03 | Login | Log in with the wrong password | Account exists | `POST /api/auth/login` with a bad password | 401 Unauthorized, no token issued | `TestAuth_Login_WrongPassword_Rejected` |
| AUTH-04 | Registration | Register with an email that's already in use | Account exists | `POST /api/auth/register` with a duplicate email | Registration rejected (not 201) | `TestAuth_Register_DuplicateEmail_Rejected` |
| AUTH-05 | Session | Access a protected route with no token | None | `GET /api/leave/balance` with no `Authorization` header | 401 Unauthorized | `TestRBAC_UnauthenticatedRequest_Rejected` |
| AUTH-06 | Session | Access a protected route with an expired token | Have a token whose `exp` claim is in the past | Call any protected route | 401 Unauthorized | `TestParseJWT_ExpiredToken` |
| AUTH-07 | Session | Access a protected route with a tampered/wrong-secret token | Have a token signed with a different secret | Call any protected route | 401 Unauthorized | `TestParseJWT_WrongSecret` |
| AUTH-08 | Session | Access a protected route with a token using the `none` algorithm | Craft a `none`-alg token (classic JWT bypass attempt) | Call any protected route | 401 Unauthorized - signing method is validated, not just presence of a token | `TestParseJWT_RejectsNoneAlgorithm` |
| AUTH-09 | Password reset | HR resets an employee's password | Employee account exists, HR logged in | `PUT /api/users/{id}/reset-password` | 200 OK with a new temporary password returned; the employee's old password stops working and the new one logs in successfully | `TestAuth_HRCanResetEmployeePassword` |
| AUTH-10 | Login UX | Submitting the login form with wrong credentials | On the login page | Enter a bad password, submit | An inline error message is shown; the page does **not** hard-navigate/reload (regression test - the axios 401 interceptor used to force a full page redirect on any 401, including a failed login attempt itself, which wiped the error message and looked like an unexplained refresh) | Manual (frontend, `api.js` interceptor fix) |
| AUTH-11 | Onboarding / recovery | Forgot password or need first-time credentials | On the login page | Click "Forgot password? Contact HR" or "New employee? Ask HR" | Opens a pre-filled `mailto:` to HR - matches the HR-mediated reset model (no self-service reset exists by design) | Manual (frontend) |

## Authorization (RBAC) Workflow

| Test Case ID | Workflow | Scenario | Precondition | Steps | Expected Result | Automated |
|---|---|---|---|---|---|---|
| RBAC-01 | Authorization | Employee calls an HR-only endpoint | Employee logged in | `GET /api/attendance/records` with an employee token | 403 Forbidden | `TestRBAC_EmployeeCannotAccessHROnlyAttendance` |
| RBAC-02 | Authorization | Employee tries to approve a leave request | Employee logged in | `PUT /api/leave/approve/{id}` with an employee token | 403 Forbidden | `TestRBAC_EmployeeCannotApproveLeave` |
| RBAC-03 | Authorization | Employee tries to broadcast a company-wide notification | Employee logged in | `POST /api/notifications/broadcast` with an employee token | 403 Forbidden | `TestRBAC_EmployeeCannotBroadcastNotification` |
| RBAC-04 | Authorization | HR views the employee directory | HR logged in | `GET /api/users/employees` | 200 OK, full employee list returned | Manual (exercised via HR Dashboard "All Employees" section) |
| RBAC-05 | Frontend routing | Employee navigates directly to the HR Dashboard URL | Logged in as employee | Visit `/dashboard/hr` | Redirected to `/dashboard/employee` instead of rendering the HR page | Manual (`ProtectedRoute` `allowedRoles`) |
| RBAC-06 | Authorization | Employee tries to reset a password (including their own) | Employee logged in | `PUT /api/users/{id}/reset-password` with an employee token | 403 Forbidden | `TestRBAC_EmployeeCannotResetPassword` |
| RBAC-07 | Multi-tab session | A tab left open on one account stays in sync when a different account logs in on another tab | Two tabs open, same browser | Log in as HR in tab A; in tab B, log in as a different employee account; return to tab A and trigger an HR-only action | Tab A's auth state updates to match the most recently logged-in account instead of silently sending a stale/mismatched token (regression test - previously tab A kept rendering as the old account while actually sending the new account's token, causing confusing 403s) | Manual (frontend, `AuthContext` `storage` event listener) |

## Attendance Workflow

| Test Case ID | Workflow | Scenario | Precondition | Steps | Expected Result | Automated |
|---|---|---|---|---|---|---|
| ATT-01 | Check-in | Employee checks in for the day | Employee logged in, not yet checked in today | `POST /api/attendance/checkin` | 200 OK, attendance row created with status `checked_in` | `TestAttendance_CheckInCheckOutCycle` |
| ATT-02 | Check-in | Employee tries to check in twice the same day | Already checked in today | `POST /api/attendance/checkin` again | 400 Bad Request, `already_checked_in` | `TestAttendance_CheckInCheckOutCycle` |
| ATT-03 | Check-out | Employee checks out after checking in | Checked in today | `POST /api/attendance/checkout` | 200 OK, status becomes `checked_out` | `TestAttendance_CheckInCheckOutCycle` |
| ATT-04 | Check-out | Employee tries to check out without checking in | Not checked in today | `POST /api/attendance/checkout` | 400 Bad Request, `no_checkin` | `TestAttendance_CheckOutWithoutCheckIn_Rejected` |
| ATT-05 | History | Employee views their own attendance history | At least one attendance record exists | `GET /api/attendance/history` | 200 OK, only that employee's records returned | `TestAttendance_CheckInCheckOutCycle` |
| ATT-06 | HR view | HR views a specific day's attendance across all employees | HR logged in | `GET /api/attendance/records?date=YYYY-MM-DD` | 200 OK, all employees' records for that date only | Manual (HR Dashboard day-tabs / date picker) |
| ATT-07 | Live timer | Employee dashboard shows a running session timer while checked in | Checked in | Watch the "Today's Attendance" card | A live `HH:MM:SS` counter increments every second | Manual (frontend) |
| ATT-08 | Confirmation | Employee attempts to check out | Checked in | Click "Check Out" | A confirmation dialog appears before the check-out is submitted | Manual (frontend) |

## Leave Management Workflow

| Test Case ID | Workflow | Scenario | Precondition | Steps | Expected Result | Automated |
|---|---|---|---|---|---|---|
| LEAVE-00 | Leave types | Employee fetches the full list of leave types to apply for | Employee logged in | `GET /api/leave/types` | 200 OK, all 5 seeded types returned (Sick, Casual, Earned, Maternity, Paternity) - regression test for a bug where the frontend hardcoded only 3 of them by ID | `TestLeave_GetLeaveTypes` |
| LEAVE-01 | Apply | Employee applies for leave within their balance | Employee has a positive balance for some leave type | `POST /api/leave/apply` | 201 Created, request stored with status `pending` | `TestLeave_FullApprovalFlow` |
| LEAVE-02 | Apply | Employee applies for more days than their balance allows | Balance is known | `POST /api/leave/apply` with `number_of_days` far beyond the balance | 400 Bad Request, `insufficient_balance` | `TestLeave_ApplyInsufficientBalance_Rejected` |
| LEAVE-03 | Approve | HR approves a pending request | Request exists, HR logged in | `PUT /api/leave/approve/{id}` | 200 OK with a success message; employee's balance decremented by the approved days | `TestLeave_FullApprovalFlow` |
| LEAVE-04 | Approve | Employee receives a notification when their leave is approved | Leave just approved | `GET /api/notifications` as the employee | A `leave_approved` notification is present | `TestLeave_FullApprovalFlow` |
| LEAVE-05 | Reject | HR rejects a pending request | Request exists, HR logged in | `PUT /api/leave/reject/{id}` | 200 OK, request status becomes `rejected` | `TestLeave_RejectFlow` |
| LEAVE-06 | Reject | Employee receives a notification when their leave is rejected | Leave just rejected | `GET /api/notifications` as the employee | A `leave_rejected` notification is present | `TestLeave_RejectFlow` |
| LEAVE-07 | HR view | HR sees a new request appear live without refreshing | HR dashboard open in a browser tab, WebSocket connected | An employee submits a leave request | HR's browser receives a live push and the notification bell updates immediately | Manual — verified via automated E2E script during development (WebSocket push confirmed working) |
| LEAVE-08 | List/filter | HR filters requests by status | Multiple requests with different statuses exist | Click Pending/Approved/Rejected/All tabs on HR Dashboard | Only matching requests are shown | Manual (frontend) |
| LEAVE-09 | Display | A leave request's type is shown by name, not a raw ID | A leave request exists | `GET /api/leave/requests` | The response includes `leave_type_name` resolved via a join, matching what `MyLeaveRequests.jsx` and `HRDashboard.jsx` render directly - regression test for the same hardcoded-ID bug as LEAVE-00 | `TestLeave_RequestIncludesLeaveTypeName` |

## Notifications & Real-Time Workflow

| Test Case ID | Workflow | Scenario | Precondition | Steps | Expected Result | Automated |
|---|---|---|---|---|---|---|
| NOTIF-01 | WebSocket auth | Connect to the notification socket with no token | None | Open `ws://.../ws/notifications` with no `?token=` | Handshake rejected with 401 | `TestHandleWebSocket_RejectsMissingToken` |
| NOTIF-02 | WebSocket auth | Connect with a garbage/invalid token | None | Open the socket with `?token=not-a-real-jwt` | Handshake rejected with 401 | `TestHandleWebSocket_RejectsInvalidToken` |
| NOTIF-03 | Targeted delivery | A notification sent to one user reaches only that user | Two users connected via WebSocket | Send a notification targeted at user A's ID | User A's socket receives it; user B's socket receives nothing | `TestHub_SendNotificationToUser_OnlyReachesTargetUser` |
| NOTIF-04 | Role broadcast | A notification sent to a role reaches every connected client with that role | One HR and one employee connected | Send a notification targeted at role `hr_manager` | HR's socket receives it; the employee's socket receives nothing | `TestHub_SendNotificationToRole_OnlyReachesThatRole` |
| NOTIF-05 | Resilience | One client disconnecting doesn't affect delivery to others | Two clients connected | Close client A's connection, then message client B | Client B still receives its message; the hub goroutine keeps running | `TestHub_SurvivesClientDisconnect` |
| NOTIF-06 | Concurrency | Many clients connecting and being messaged at once doesn't race or deadlock | 20 clients connect concurrently | Send each of them a targeted notification concurrently | All 20 receive their message; `go test -race` reports no data race | `TestHub_ConcurrentClientsNoRace` |
| NOTIF-07 | HR broadcast | HR sends an announcement to all employees | HR logged in | `POST /api/notifications/broadcast` with title/message | Notification persisted for every employee and pushed live to anyone currently connected | Manual — verified via automated E2E script during development (broadcast → live push → persisted → correct unread count, all confirmed working) |
| NOTIF-08 | Read state | Employee opens the notification bell and reads a notification | Unread notification exists | Click the notification | `is_read` becomes true, unread badge count decreases | Manual (frontend) + covered server-side by `MarkNotificationRead` logic exercised in NOTIF-04/05 setups |
| NOTIF-09 | Reconnect | Browser reconnects automatically after a dropped connection | WebSocket connected, then network interrupted | Simulate a dropped connection | Client automatically retries and re-establishes the socket within a few seconds | Manual (frontend `NotificationContext` reconnect logic) |

## Payslip Workflow

| Test Case ID | Workflow | Scenario | Precondition | Steps | Expected Result | Automated |
|---|---|---|---|---|---|---|
| PAY-01 | View list | Employee views their own payslips | Payslip records exist for the employee | `GET /api/payslip/list` | 200 OK, only that employee's payslips returned | Manual |
| PAY-02 | View details | Employee views one payslip's full breakdown | A payslip exists | `GET /api/payslip/{id}` | 200 OK, full earnings/deductions breakdown | Manual |
| PAY-03 | Access control | Employee tries to view another employee's payslip by ID | Another employee's payslip ID is known | `GET /api/payslip/{other_id}` | Access denied (not the requester's own payslip, and not HR) | Manual |

## Known Issues Surfaced By Testing

Writing this suite surfaced real, previously-undiscovered bugs, all now fixed:

1. **Login accepted no real passwords for API-registered accounts.** `Login` compared the bcrypt hash directly against the plaintext password (`hashedPassword != req.Password`) instead of calling `bcrypt.CompareHashAndPassword`. Every account created through `POST /api/auth/register` was permanently unable to log in. The only reason the demo accounts (`employee@example.com`, `hr@example.com`) appeared to work throughout earlier development was that their `password_hash` column literally contained the plaintext string `"password123"`, not a real hash — confirmed by direct inspection and fixed by re-hashing them with bcrypt. Fixed in `backend/auth.go`; regression-covered by `TestAuth_RegisterAndLogin_Success`.
2. **Negative duration in `CheckOut`.** During `TestAttendance_CheckInCheckOutCycle`, the backend logged `Duration: -2.00 hours` for a check-in immediately followed by a check-out - both happening within the same test, milliseconds apart. Root cause: `time.Now()` (server's local timezone) was written to a `TIMESTAMP WITHOUT TIME ZONE` Postgres column, then read back and treated as UTC, silently skewing every duration calculation by the server's UTC offset. Fixed by writing `time.Now().UTC()` consistently in both `CheckIn` and `CheckOut` (`backend/attendance.go`); regression-covered by an assertion in `TestAttendance_CheckInCheckOutCycle` that `check_out_time` is never before `check_in_time`.
3. **Leave type names were hardcoded to IDs 1-3 in the frontend.** `ApplyLeave.jsx`'s dropdown and `MyLeaveRequests.jsx`'s table both hardcoded a 3-entry switch (`1 → Sick, 2 → Casual, 3 → Earned`), even though the backend seeds 5 leave types including Maternity and Paternity. Employees could never apply for the last two, and any leave request using them would render with a blank type. Fixed by adding `GET /api/leave/types` and joining `leave_types` into `GetLeaveRequests` so the actual name is always returned from the backend instead of guessed on the frontend. Regression-covered by `TestLeave_GetLeaveTypes` and `TestLeave_RequestIncludesLeaveTypeName`.
