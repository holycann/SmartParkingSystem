package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/holycan/smart-parking-system/database"
	"github.com/holycan/smart-parking-system/models"
)

func UpdateBookingWithSpot(status string, userID string, bookingID string, parkingLotID string, spotID string) (map[string]interface{}, error) {
	// First check if the parking space is already occupied
	if status == "active" {
		isOccupied, err := IsParkingSpaceOccupied(spotID)
		if err != nil {
			return nil, fmt.Errorf("failed to check parking space status: %w", err)
		}

		if isOccupied {
			return nil, fmt.Errorf("parking space %s is already occupied", spotID)
		}
	}

	// Define the SQL query to update the reservation
	query := `
        UPDATE reservations
        SET status = $1, 
            parking_lot_id = $2, 
            parking_space_id = $3, 
            checkin_time = CASE WHEN $6 = 'active' THEN NOW() ELSE checkin_time END, 
            updated_at = NOW()
        WHERE id = $4 AND user_id = $5
        RETURNING id, parking_space_id
    `

	// Execute the query and scan the returned values
	var id, parkingSpaceID string
	err := database.DB.QueryRow(query, status, parkingLotID, spotID, bookingID, userID, status).Scan(&id, &parkingSpaceID)

	// Handle potential errors
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no booking found with ID %s for user %s: %w", bookingID, userID, err)
		}
		return nil, fmt.Errorf("failed to update booking %s: %w", bookingID, err)
	}

	// Return the updated booking information
	return map[string]interface{}{
		"id":               id,
		"status":           status,
		"parking_space_id": parkingSpaceID,
	}, nil
}

// RevertBookingSpot reverts a booking to its previous state if parking space update fails
func RevertBookingSpot(bookingID string, userID string) error {
	query := `
        UPDATE reservations
        SET status = 'confirmed', 
            checkin_time = NULL,
            updated_at = NOW()
        WHERE id = $1 AND user_id = $2
    `

	_, err := database.DB.Exec(query, bookingID, userID)
	if err != nil {
		return fmt.Errorf("failed to revert booking %s: %w", bookingID, err)
	}

	return nil
}

func UpdateBookingPaymentStatus(paymentStatus string, bookingID string) error {
	query := "UPDATE reservations SET payment_status = $1 WHERE id = $2"
	_, err := database.DB.Exec(query, paymentStatus, bookingID)
	return err
}

func GetBookingByID(ReservationID string, userID string) (*models.Reservation, error) {
	var r models.Reservation
	err := database.DB.QueryRow(`
		SELECT id, user_id, parking_lot_id, parking_space_id, vehicle_type, license_plate,
		       duration, status, total_cost, payment_status,
		       created_at, updated_at
		FROM reservations
		WHERE id = $1
	`, ReservationID).Scan(
		&r.ID,
		&r.UserID,
		&r.ParkingLotID,
		&r.ParkingSpaceID,
		&r.VehicleType,
		&r.LicensePlate,
		&r.Duration,
		&r.Status,
		&r.TotalCost,
		&r.PaymentStatus,
		&r.CreatedAt,
		&r.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &r, nil
}
