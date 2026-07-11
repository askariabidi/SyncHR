package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// JWTMiddleware validates JWT tokens in request headers
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip middleware for health check endpoint
		if r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "Missing authorization header",
				Code:    401,
			})
			return
		}

		// Extract Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
				Code:    401,
			})
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			// Get JWT secret from environment
			jwtSecret := "default_secret_key_change_this"
			if os.Getenv("JWT_SECRET") != "" {
				jwtSecret = os.Getenv("JWT_SECRET")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
				Code:    401,
			})
			return
		}

		// Extract claims and add to context
		userID := int(claims["user_id"].(float64))
		email := claims["email"].(string)
		role := claims["role"].(string)

		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "email", email)
		ctx = context.WithValue(ctx, "role", role)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WebSocket clients and hub
var (
	clients = make(map[*websocket.Conn]int) // Map of WebSocket connections to user IDs
	broadcast = make(chan interface{})
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for development
		},
	}
)

// HandleWebSocket handles WebSocket connections for real-time notifications
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from query parameter or header
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "missing_user_id",
			Message: "user_id query parameter is required",
			Code:    400,
		})
		return
	}

	// Parse user ID
	var userID int
	_, err := json.Marshal(userIDStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	// Add client to map
	clients[conn] = userID
	log.Printf("📱 WebSocket client connected: User ID %d", userID)

	// Handle messages from client
	go handleClientMessages(conn, userID)
}

// handleClientMessages reads messages from client and broadcasts them
func handleClientMessages(conn *websocket.Conn, userID int) {
	defer func() {
		delete(clients, conn)
		conn.Close()
		log.Printf("📱 WebSocket client disconnected: User ID %d", userID)
	}()

	for {
		// Read message from client
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			return
		}

		// Process message (e.g., leave approval notification)
		log.Printf("📩 Message from User %d: %v", userID, msg)

		// Broadcast to all connected clients
		broadcast <- map[string]interface{}{
			"from_user": userID,
			"data":      msg,
		}
	}
}

// BroadcastNotification sends a notification to all connected WebSocket clients
func BroadcastNotification(notification Notification) {
	broadcast <- map[string]interface{}{
		"type":         "notification",
		"title":        notification.Title,
		"message":      notification.Message,
		"notification": notification,
	}
}

// SendNotificationToUser sends a notification to a specific user via WebSocket
func SendNotificationToUser(userID int, notification Notification) {
	for conn, cID := range clients {
		if cID == userID {
			err := conn.WriteJSON(map[string]interface{}{
				"type":         "notification",
				"title":        notification.Title,
				"message":      notification.Message,
				"notification": notification,
			})
			if err != nil {
				log.Printf("Error sending notification: %v", err)
				conn.Close()
				delete(clients, conn)
			}
		}
	}
}