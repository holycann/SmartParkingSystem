package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"

	"github.com/holycan/smart-parking-system/database"
	"github.com/holycan/smart-parking-system/lock"
)

func FindAndLockAvailableSpot(parkingLotID string) (*redsync.Mutex, map[string]interface{}, error) {
	query := `
        SELECT ps.id
        FROM parking_spaces ps
        WHERE ps.parking_lot_id = $1
          AND NOT EXISTS (
              SELECT 1 FROM reservations r
              WHERE r.parking_space_id = ps.id
                AND r.status IN ('active', 'checked-in')
          )
        LIMIT 1
    `

	rows, err := database.DB.Query(query, parkingLotID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var spotID string
	for rows.Next() {
		err := rows.Scan(&spotID)
		if err != nil {
			continue
		}

		mutex, err := lock.AcquireLock("spot-lock:"+spotID, 10*time.Second)
		if err == nil {
			return mutex, map[string]interface{}{"spot_id": spotID}, nil
		}
	}

	return nil, nil, fmt.Errorf("no available spot could be locked in parking lot %s", parkingLotID)
}

func UpdateParkingSpaceOccupied(status bool, spotID string) error {
	query := "UPDATE parking_spaces SET is_occupied = $1 WHERE id = $2"
	_, err := database.DB.Exec(query, status, spotID)
	return err
}

func GetParkingDataById(id string) (map[string]interface{}, error) {
	query := `
		SELECT ps.id, ps.parking_lot_id, pl.name, ps.space_number, ps.floor
		FROM parking_spaces ps
		JOIN parking_lots pl ON ps.parking_lot_id = pl.id
		WHERE ps.id = $1
	`
	rows, err := database.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result map[string]interface{}
	if rows.Next() {
		var spotID, parkingLotID, zoneName, spaceNumber, floor string
		err := rows.Scan(&spotID, &parkingLotID, &zoneName, &spaceNumber, &floor)
		if err != nil {
			return nil, err
		}
		// Menyusun map hasil query
		result = map[string]interface{}{
			"spot_id":        spotID,
			"parking_lot_id": parkingLotID,
			"zone_name":      zoneName,
			"space_number":   spaceNumber, // Ini seharusnya spaceNumber, bukan name
			"floor":          floor,       // Ini seharusnya floor, bukan name
		}
	} else {
		// Jika tidak ada data ditemukan
		return nil, fmt.Errorf("parking space not found")
	}

	return result, nil
}

// IsParkingSpaceOccupied checks if a parking space is currently occupied
func IsParkingSpaceOccupied(spotID string) (bool, error) {
	query := "SELECT is_occupied FROM parking_spaces WHERE id = $1"

	var isOccupied bool
	err := database.DB.QueryRow(query, spotID).Scan(&isOccupied)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("parking space %s not found", spotID)
		}
		return false, fmt.Errorf("database error: %w", err)
	}

	return isOccupied, nil
}
