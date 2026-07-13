package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var hubOnce sync.Once

// startHub launches the WebSocket hub's single goroutine exactly once, no
// matter how many times it's called - SetupRouter is called once from
// main() and once per test file, and the hub must not be started twice.
func startHub() {
	hubOnce.Do(func() {
		go hub.run()
	})
}

// SetupRouter wires every handler and route onto a fresh mux.Router. Pulled
// out of main() so tests can build the exact same routing table against a
// real *sql.DB without going through os.Exit/ListenAndServe.
func SetupRouter(db *sql.DB) *mux.Router {
	startHub()

	// Initialize handlers
	notificationHandler := NewNotificationHandler(db)
	authHandler := NewAuthHandler(db, notificationHandler)
	attendanceHandler := NewAttendanceHandler(db)
	leaveHandler := NewLeaveHandler(db, notificationHandler)
	payslipHandler := NewPayslipHandler(db)
	holidayHandler := NewHolidayHandler(db)

	router := mux.NewRouter()

	// CORS Middleware (must be first!)
	router.Use(corsMiddleware)

	// ============================================================================
	// PUBLIC ROUTES (No authentication required)
	// ============================================================================

	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	}).Methods("GET", "OPTIONS")

	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST", "OPTIONS")

	// WebSocket upgrade - registered on the PUBLIC router (not behind JWTMiddleware)
	// because a browser's native WebSocket API can't set an Authorization header;
	// the JWT is instead passed as a query param and validated inside HandleWebSocket.
	router.HandleFunc("/api/ws/notifications", HandleWebSocket).Methods("GET")

	// ============================================================================
	// PROTECTED ROUTES (JWT authentication required)
	// ============================================================================

	subrouter := router.PathPrefix("/api").Subrouter()
	subrouter.Use(JWTMiddleware)

	// User Profile Routes
	subrouter.HandleFunc("/users/profile", authHandler.GetProfile).Methods("GET", "OPTIONS")
	subrouter.HandleFunc("/users/profile", authHandler.UpdateProfile).Methods("PUT", "OPTIONS")
	subrouter.HandleFunc("/users/employees", authHandler.GetAllEmployees).Methods("GET", "OPTIONS")
	subrouter.HandleFunc("/users/{id}/reset-password", authHandler.ResetEmployeePassword).Methods("PUT", "OPTIONS")

	// Attendance Routes
	subrouter.HandleFunc("/attendance/checkin", attendanceHandler.CheckIn).Methods("POST", "OPTIONS")
	subrouter.HandleFunc("/attendance/checkout", attendanceHandler.CheckOut).Methods("POST", "OPTIONS")
	subrouter.HandleFunc("/attendance/history", attendanceHandler.GetAttendanceHistory).Methods("GET", "OPTIONS")
	subrouter.HandleFunc("/attendance/records", attendanceHandler.GetAllAttendanceRecords).Methods("GET", "OPTIONS")

	// Leave Routes
	subrouter.HandleFunc("/leave/types", leaveHandler.GetLeaveTypes).Methods("GET", "OPTIONS")
	subrouter.HandleFunc("/leave/apply", leaveHandler.ApplyLeave).Methods("POST", "OPTIONS")
	subrouter.HandleFunc("/leave/balance", leaveHandler.GetLeaveBalance).Methods("GET", "OPTIONS")
	subrouter.HandleFunc("/leave/requests", leaveHandler.GetLeaveRequests).Methods("GET", "OPTIONS")
	subrouter.HandleFunc("/leave/approve/{id}", leaveHandler.ApproveLeave).Methods("PUT", "OPTIONS")
	subrouter.HandleFunc("/leave/reject/{id}", leaveHandler.RejectLeave).Methods("PUT", "OPTIONS")

	// Payslip Routes
	subrouter.HandleFunc("/payslip/list", payslipHandler.GetPayslips).Methods("GET", "OPTIONS")
	subrouter.HandleFunc("/payslip/{id}", payslipHandler.GetPayslipDetails).Methods("GET", "OPTIONS")

	// Holiday Routes
	subrouter.HandleFunc("/holidays", holidayHandler.GetHolidays).Methods("GET", "OPTIONS")

	// Notification Routes
	subrouter.HandleFunc("/notifications", notificationHandler.GetNotifications).Methods("GET", "OPTIONS")
	subrouter.HandleFunc("/notifications/read-all", notificationHandler.MarkAllNotificationsRead).Methods("PUT", "OPTIONS")
	subrouter.HandleFunc("/notifications/{id}/read", notificationHandler.MarkNotificationRead).Methods("PUT", "OPTIONS")
	subrouter.HandleFunc("/notifications/broadcast", notificationHandler.SendBroadcastNotification).Methods("POST", "OPTIONS")

	return router
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No .env file found, using default values")
	}

	// Initialize database
	db, err := ConnectDatabase()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connected successfully")

	router := SetupRouter(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 SyncHR Server starting on port %s", port)
	log.Printf("📍 API Base URL: http://localhost:%s/api", port)

	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}

// corsMiddleware handles CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for all requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
