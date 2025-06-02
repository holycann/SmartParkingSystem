package models

import "time"

type GateEvent struct {
	ParkingLotID   string
	ParkingSpaceID string
	UserID         string
	ReservationID  string
	Status         string
	Timestamp      time.Time
}
