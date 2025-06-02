package models

import (
	"database/sql"
	"time"
)

// Reservation represents a parking reservation
type Reservation struct {
	ID              string    `json:"id"`
	UserID          string    `json:"userId"`
	ParkingLotID    string    `json:"parkingLotId"`
	ParkingSpaceID  string    `json:"parkingSpaceId"`
	VehicleType     string    `json:"vehicleTyoe"`
	LicensePlate    string    `json:"licensePlate"`
	ReservationDate time.Time `json:"reservationDate"`
	ExpiredAt       time.Time `json:"expiredAt"`
	CheckinTime     time.Time `json:"checkInTime"`
	Duration        int       `json:"duration"`
	Status          string    `json:"status"`
	TotalCost       float64   `json:"totalCost"`
	PaymentStatus   string    `json:"paymentStatus"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// ReservationRequest represents the request body for creating a reservation
type ReservationRequest struct {
	ParkingLotID    string         `json:"parkingLotId" binding:"required"`
	ParkingSpaceID  string         `json:"parkingSpaceId" binding:"required"`
	VehicleType     string         `json:"vehicleType" binding:"required"`
	LicensePlate    string         `json:"licensePlate" binding:"required"`
	ReservationDate string         `json:"reservationDate" binding:"required"`
	CheckinTime     sql.NullString `json:"checkInTime"`
	ExpiredAt       time.Time      `json:"expiredAt"`
	Duration        int            `json:"duration" binding:"required"`
	TotalCost       float64        `json:"totalCost"`
}

// ReservationUpdate represents an update to a reservation
type ReservationUpdate struct {
	ReservationID  string    `json:"reservationId"`
	UserId         string    `json:"userId"`
	ParkingSpaceId string    `json:"parkingSpaceId"`
	Status         string    `json:"status"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ReservationStatus struct {
	ParkingLotID string    // ID tempat parkir yang dipesan
	Paid         bool      // Status pembayaran
	PaymentTime  time.Time // Waktu pembayaran
}

// ReservationInfo holds basic information about a reservation
type ReservationInfo struct {
	ReservationID string `json:"reservation_id"`
	Timestamp     int64  `json:"timestamp"`
}

type ReservationEvent struct {
	ReservationID  string `json:"reservationId"`
	PaymentID      string `json:"paymentId"`
	UserID         string `json:"userId"`
	ParkingSpaceID string `json:"parkingSpaceId"`
	Status         string `json:"status"`
	Timestamp      int64  `json:"timestamp"`
}
