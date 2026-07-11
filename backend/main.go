package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database connection
	db, err := ConnectDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connected successfully")

	// Create router
	router := mux.NewRouter()

	// Initialize handlers with database connection
	authHandler := NewAuthHandler(db)
	attendanceHandler := NewAttendanceHandler(db)
	leaveHandler := NewLeaveHandler(db)
	payslipHandler := NewPayslipHandler(db)

	// ============================================================================
	// PUBLIC ROUTES (No Authentication Required)
	// ============================================================================
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")

	// ============================================================================
	// PROTECTED ROUTES (Requires JWT Token)
	// ============================================================================
	// Apply JWT middleware to protected routes
	router.Use(JWTMiddleware)

	// User routes
	router.HandleFunc("/api/users/profile", authHandler.GetProfile).Methods("GET")
	router.HandleFunc("/api/users/profile", authHandler.UpdateProfile).Methods("PUT")

	// Attendance routes
	router.HandleFunc("/api/attendance/checkin", attendanceHandler.CheckIn).Methods("POST")
	router.HandleFunc("/api/attendance/checkout", attendanceHandler.CheckOut).Methods("POST")
	router.HandleFunc("/api/attendance/history", attendanceHandler.GetAttendanceHistory).Methods("GET")

	// Leave routes
	router.HandleFunc("/api/leave/apply", leaveHandler.ApplyLeave).Methods("POST")
	router.HandleFunc("/api/leave/balance", leaveHandler.GetLeaveBalance).Methods("GET")
	router.HandleFunc("/api/leave/requests", leaveHandler.GetLeaveRequests).Methods("GET")
	router.HandleFunc("/api/leave/approve/{id}", leaveHandler.ApproveLeave).Methods("PUT")
	router.HandleFunc("/api/leave/reject/{id}", leaveHandler.RejectLeave).Methods("PUT")

	// Payslip routes
	router.HandleFunc("/api/payslip/list", payslipHandler.GetPayslips).Methods("GET")
	router.HandleFunc("/api/payslip/{id}", payslipHandler.GetPayslipDetails).Methods("GET")

	// Holiday routes
	router.HandleFunc("/api/holidays", NewHolidayHandler(db).GetHolidays).Methods("GET")

	// WebSocket for real-time notifications
	router.HandleFunc("/ws/notifications", HandleWebSocket)

	// Health check
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","message":"SyncHR API is running"}`)
	}).Methods("GET")

	// Enable CORS for frontend communication
	router.Use(CORSMiddleware)

	// Start server
	log.Printf("🚀 SyncHR Server starting on port %s", port)
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// CORSMiddleware enables CORS for cross-origin requests from React frontend
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}