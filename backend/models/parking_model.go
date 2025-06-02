package models

import "time"

// ParkingLot represents a parking lot
type ParkingLot struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	ZipCode     string    `json:"zipCode"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	TotalSpaces int       `json:"totalSpaces"`
	HourlyRate  float64   `json:"hourlyRate"`
	OpenTime    string    `json:"openTime"`
	CloseTime   string    `json:"closeTime"`
	IsOpen24H   bool      `json:"isOpen24h"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ParkingSpace represents a parking space
type ParkingSpace struct {
	ID           string    `json:"id"`
	ParkingLotID string    `json:"parkingLotId"`
	SpaceNumber  string    `json:"spaceNumber"`
	Floor        int       `json:"floor"`
	Type         string    `json:"type"`
	IsOccupied   bool      `json:"is_occupied"`
	LastUpdated  time.Time `json:"lastUpdated"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ParkingUpdate represents a change in parking space status
type ParkingEvent struct {
	ParkingLotID string `json:"parkingLotId"`
	SpaceID      string `json:"spaceId"`
	IsOccupied   bool   `json:"isOccupied"`
	IsPaid       bool   `json:"isPaid"`
	Timestamp    int64  `json:"timestamp"`
}
