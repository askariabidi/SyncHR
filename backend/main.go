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

	// ============================================================================
	// APPLY CORS MIDDLEWARE FIRST (MUST be before all other middleware)
	// ============================================================================
	router.Use(CORSMiddleware)

	// Initialize handlers with database connection
	authHandler := NewAuthHandler(db)
	attendanceHandler := NewAttendanceHandler(db)
	leaveHandler := NewLeaveHandler(db)
	payslipHandler := NewPayslipHandler(db)

	// ============================================================================
	// PUBLIC ROUTES (No Authentication Required)
	// ============================================================================
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/health", HealthCheck).Methods("GET", "OPTIONS")

	// ============================================================================
	// PROTECTED ROUTES (Requires JWT Token)
	// ============================================================================
	// Create a subrouter for protected routes and apply JWT middleware to it
	protectedRoutes := router.PathPrefix("/api").Subrouter()
	protectedRoutes.Use(JWTMiddleware)

	// User routes
	protectedRoutes.HandleFunc("/users/profile", authHandler.GetProfile).Methods("GET", "OPTIONS")
	protectedRoutes.HandleFunc("/users/profile", authHandler.UpdateProfile).Methods("PUT", "OPTIONS")

	// Attendance routes
	protectedRoutes.HandleFunc("/attendance/checkin", attendanceHandler.CheckIn).Methods("POST", "OPTIONS")
	protectedRoutes.HandleFunc("/attendance/checkout", attendanceHandler.CheckOut).Methods("POST", "OPTIONS")
	protectedRoutes.HandleFunc("/attendance/history", attendanceHandler.GetAttendanceHistory).Methods("GET", "OPTIONS")

	// Leave routes
	protectedRoutes.HandleFunc("/leave/apply", leaveHandler.ApplyLeave).Methods("POST", "OPTIONS")
	protectedRoutes.HandleFunc("/leave/balance", leaveHandler.GetLeaveBalance).Methods("GET", "OPTIONS")
	protectedRoutes.HandleFunc("/leave/requests", leaveHandler.GetLeaveRequests).Methods("GET", "OPTIONS")
	protectedRoutes.HandleFunc("/leave/approve/{id}", leaveHandler.ApproveLeave).Methods("PUT", "OPTIONS")
	protectedRoutes.HandleFunc("/leave/reject/{id}", leaveHandler.RejectLeave).Methods("PUT", "OPTIONS")

	// Payslip routes
	protectedRoutes.HandleFunc("/payslip/list", payslipHandler.GetPayslips).Methods("GET", "OPTIONS")
	protectedRoutes.HandleFunc("/payslip/{id}", payslipHandler.GetPayslipDetails).Methods("GET", "OPTIONS")

	// Holiday routes
	protectedRoutes.HandleFunc("/holidays", NewHolidayHandler(db).GetHolidays).Methods("GET", "OPTIONS")

	// WebSocket for real-time notifications
	router.HandleFunc("/ws/notifications", HandleWebSocket)

	// Start server
	log.Printf("🚀 SyncHR Server starting on port %s", port)
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// ============================================================================
// MIDDLEWARE: CORS
// ============================================================================
// CORSMiddleware enables CORS for cross-origin requests from React frontend
// This MUST be applied first, before any other middleware
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins (for development; restrict in production)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests (OPTIONS method)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ============================================================================
// HEALTH CHECK ENDPOINT
// ============================================================================
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","message":"SyncHR API is running"}`)
}