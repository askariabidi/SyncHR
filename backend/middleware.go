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

// parseJWT validates a raw JWT string and returns its claims. Shared by
// JWTMiddleware (Authorization header) and HandleWebSocket (query param,
// since browsers can't set custom headers during a WebSocket handshake).
func parseJWT(tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		jwtSecret := "default_secret_key_change_this"
		if os.Getenv("JWT_SECRET") != "" {
			jwtSecret = os.Getenv("JWT_SECRET")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// JWTMiddleware validates JWT tokens in request headers
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers first
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")

		// Handle OPTIONS requests (preflight)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

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

		claims, err := parseJWT(parts[1])
		if err != nil {
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

		// Debug logging
		log.Printf("🔐 JWT Middleware - Extracted UserID: %d, Email: %s, Role: %s", userID, email, role)

		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "email", email)
		ctx = context.WithValue(ctx, "role", role)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ============================================================================
// WEBSOCKET HUB
// ============================================================================
//
// Real-time notification delivery using Go's "actor" pattern: a single
// goroutine (Hub.run) owns the connected-client registry and is the ONLY
// thing that ever reads or mutates it. Every other goroutine (one per
// connected client, plus every HTTP handler that wants to push a
// notification) talks to it exclusively through channels instead of a
// mutex - this is what makes registration/broadcast/direct-send safe to
// call concurrently from many goroutines at once.

type wsClient struct {
	conn   *websocket.Conn
	userID int
	role   string
}

type wsMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// wsTarget describes who a message should be delivered to: a specific
// user (userID > 0) or everyone currently connected with a given role.
type wsTarget struct {
	userID  int
	role    string
	message wsMessage
}

type wsHub struct {
	clients    map[*wsClient]bool
	register   chan *wsClient
	unregister chan *wsClient
	direct     chan wsTarget
}

var hub = &wsHub{
	clients:    make(map[*wsClient]bool),
	register:   make(chan *wsClient),
	unregister: make(chan *wsClient),
	direct:     make(chan wsTarget),
}

// run is the hub's single goroutine - started once from main().
func (h *wsHub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
			log.Printf("📡 WS client connected - UserID: %d, Role: %s (total: %d)", c.userID, c.role, len(h.clients))

		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				c.conn.Close()
				log.Printf("📡 WS client disconnected - UserID: %d (total: %d)", c.userID, len(h.clients))
			}

		case t := <-h.direct:
			for c := range h.clients {
				if (t.userID > 0 && c.userID == t.userID) || (t.role != "" && c.role == t.role) {
					if err := c.conn.WriteJSON(t.message); err != nil {
						log.Printf("⚠️ WS write failed - UserID: %d: %v", c.userID, err)
					}
				}
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // development only - restrict in production
	},
}

// HandleWebSocket upgrades to a WebSocket connection for real-time
// notifications. Browsers can't set custom headers on a WebSocket
// handshake, so the JWT travels as a query parameter (?token=...) and is
// validated with the same logic as JWTMiddleware. This route is
// registered on the public router (not behind JWTMiddleware) precisely
// because that middleware requires an Authorization header the browser's
// native WebSocket API cannot send.
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims, err := parseJWT(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID := int(claims["user_id"].(float64))
	role := claims["role"].(string)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &wsClient{conn: conn, userID: userID, role: role}
	hub.register <- client

	go readPump(client)
}

// readPump keeps reading from the connection purely to detect
// disconnects/errors; the app doesn't expect the client to send anything.
func readPump(c *wsClient) {
	defer func() {
		hub.unregister <- c
	}()
	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			return
		}
	}
}

// SendNotificationToUser pushes a notification to one connected user (a no-op if they're offline)
func SendNotificationToUser(userID int, notification Notification) {
	hub.direct <- wsTarget{userID: userID, message: wsMessage{Type: "notification", Payload: notification}}
}

// SendNotificationToRole pushes a notification to every connected client with a given role
func SendNotificationToRole(role string, notification Notification) {
	hub.direct <- wsTarget{role: role, message: wsMessage{Type: "notification", Payload: notification}}
}
