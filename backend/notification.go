package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// NotificationHandler handles notification-related requests
type NotificationHandler struct {
	db *sql.DB
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(db *sql.DB) *NotificationHandler {
	return &NotificationHandler{db: db}
}

// createNotification inserts a notification row and returns it fully populated
func (h *NotificationHandler) createNotification(userID int, title, message, notifType string, relatedEntityID *int) (Notification, error) {
	var n Notification
	err := h.db.QueryRow(
		`INSERT INTO notifications (user_id, title, message, type, related_entity_id, is_read)
		 VALUES ($1, $2, $3, $4, $5, false)
		 RETURNING id, user_id, title, message, type, related_entity_id, is_read, created_at`,
		userID, title, message, notifType, relatedEntityID,
	).Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.RelatedEntityID, &n.IsRead, &n.CreatedAt)
	return n, err
}

// notifyUser persists a notification for one user and pushes it live if they're connected.
// Failures are logged, not returned - a notification failing to send should never fail the
// request (leave approval, etc.) that triggered it.
func (h *NotificationHandler) notifyUser(userID int, title, message, notifType string, relatedEntityID *int) {
	n, err := h.createNotification(userID, title, message, notifType, relatedEntityID)
	if err != nil {
		log.Printf("❌ Failed to persist notification for user %d: %v", userID, err)
		return
	}
	SendNotificationToUser(userID, n)
}

// notifyRole persists a notification for every user with the given role and pushes it live
func (h *NotificationHandler) notifyRole(role, title, message, notifType string, relatedEntityID *int) {
	rows, err := h.db.Query("SELECT id FROM users WHERE role = $1", role)
	if err != nil {
		log.Printf("❌ Failed to look up users with role %s: %v", role, err)
		return
	}
	var userIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err == nil {
			userIDs = append(userIDs, id)
		}
	}
	rows.Close()

	for _, id := range userIDs {
		h.notifyUser(id, title, message, notifType, relatedEntityID)
	}
}

// GetNotifications returns the authenticated user's notifications, most recent first
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	rows, err := h.db.Query(
		`SELECT id, user_id, title, message, type, related_entity_id, is_read, created_at
		 FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`,
		userID,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to retrieve notifications",
			Code:    500,
		})
		return
	}
	defer rows.Close()

	notifications := []Notification{}
	unreadCount := 0
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Message, &n.Type, &n.RelatedEntityID, &n.IsRead, &n.CreatedAt); err != nil {
			continue
		}
		if !n.IsRead {
			unreadCount++
		}
		notifications = append(notifications, n)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Notifications retrieved successfully",
		Data: map[string]interface{}{
			"notifications": notifications,
			"unread_count":  unreadCount,
		},
	})
}

// MarkNotificationRead marks a single notification (owned by the caller) as read
func (h *NotificationHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	vars := mux.Vars(r)
	notificationID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid notification ID",
			Code:    400,
		})
		return
	}

	result, err := h.db.Exec(
		"UPDATE notifications SET is_read = true WHERE id = $1 AND user_id = $2",
		notificationID, userID,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update notification",
			Code:    500,
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "not_found",
			Message: "Notification not found",
			Code:    404,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Notification marked as read"})
}

// MarkAllNotificationsRead marks every notification for the caller as read
func (h *NotificationHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	_, err := h.db.Exec("UPDATE notifications SET is_read = true WHERE user_id = $1 AND is_read = false", userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "database_error",
			Message: "Failed to update notifications",
			Code:    500,
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "All notifications marked as read"})
}

// SendBroadcastNotification lets an HR manager compose a notification that's delivered
// to every employee instantly (if online) and persisted for everyone else to see on login
func (h *NotificationHandler) SendBroadcastNotification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	role, _ := r.Context().Value("role").(string)
	if role != "hr_manager" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "forbidden",
			Message: "Only HR managers can send notifications",
			Code:    403,
		})
		return
	}

	var req struct {
		Title   string `json:"title"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" || req.Message == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "invalid_request",
			Message: "Title and message are required",
			Code:    400,
		})
		return
	}

	h.notifyRole("employee", req.Title, req.Message, "hr_announcement", nil)

	log.Printf("📢 HR broadcast sent - Title: %s", req.Title)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: "Notification sent to all employees"})
}
