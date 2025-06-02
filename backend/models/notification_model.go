package models

import "time"

// Notification represents a notification in the system
type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"createdAt"`
}

// NotificationUpdate represents a new notification
type NotificationEvent struct {
	NotificationID string    `json:"notificationId"`
	UserID         string    `json:"userId"`
	ParkingSpaceId string    `json:"parkingSpaceId"`
	ReservationId  string    `json:"reservationId"`
	Type           string    `json:"type"`
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"createdAt"`
}
