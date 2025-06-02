package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/holycan/smart-parking-system/database"
	"github.com/holycan/smart-parking-system/models"
	"github.com/holycan/smart-parking-system/utils"
)

// GetUserReservations handles fetching reservations for the current user
func GetUserReservations(c *gin.Context) {
	// Get current user ID from context
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Query upcoming reservations
	rows, err := database.DB.Query(`
		SELECT r.id, r.parking_lot_id, r.parking_space_id, r.duration, vehicle_type, license_plate, reservation_date, expired_at, checkin_time,
		       r.status, r.total_cost, r.payment_status, r.created_at, r.updated_at,
		       pl.name AS parking_lot_name, ps.space_number
		FROM reservations r
		JOIN parking_lots pl ON r.parking_lot_id = pl.id
		JOIN parking_spaces ps ON r.parking_space_id = ps.id
		WHERE r.user_id = $1
		ORDER BY r.created_at ASC
	`, userID)

	if err != nil {
		log.Printf("Error querying upcoming reservations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch upcoming reservations"})
		return
	}
	defer rows.Close()

	// Parse results
	upcomingReservations := []gin.H{}
	for rows.Next() {
		var id, parkingLotID, parkingSpaceID, status, parkingLotName, spaceNumber string
		var createdAt, updatedAt time.Time
		var duration int
		var vehicleType, licensePlate, paymentStatus string
		var totalCost float64
		var reservationDate, expiredAt string
		var checkinTime sql.NullTime

		err := rows.Scan(
			&id, &parkingLotID, &parkingSpaceID, &duration, &vehicleType, &licensePlate, &reservationDate, &expiredAt, &checkinTime,
			&status, &totalCost, &paymentStatus, &createdAt, &updatedAt,
			&parkingLotName, &spaceNumber,
		)

		if err != nil {
			log.Printf("Error scanning reservation row: %v", err)
			continue
		}

		upcomingReservations = append(upcomingReservations, gin.H{
			"id":               id,
			"parking_lot_id":   parkingLotID,
			"parking_space_id": parkingSpaceID,
			"vehicle_type":     vehicleType,
			"license_plate":    licensePlate,
			"reservation_date": reservationDate,
			"expired_at":       expiredAt,
			"checkin_time": func() interface{} {
				if checkinTime.Valid {
					return checkinTime.Time.Format(time.RFC3339) // atau format lain
				}
				return nil
			}(),
			"payment_status":   paymentStatus,
			"duration":         duration,
			"status":           status,
			"total_cost":       totalCost,
			"created_at":       createdAt,
			"updated_at":       updatedAt,
			"parking_lot_name": parkingLotName,
			"space_number":     spaceNumber,
		})
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating reservation rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing reservation data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reservations": upcomingReservations,
		"count":        len(upcomingReservations),
	})
}

// GetUserReservations handles fetching reservations for the current user
func GetReservationDetails(c *gin.Context) {
	// Get user ID from the context (set by auth middleware)
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Build the query based on filters
	query := `
		SELECT r.id, r.user_id, r.parking_lot_id, r.parking_space_id, r.vehicle_type, r.license_plate,
		       r.duration, r.status, r.total_cost, r.payment_status, 
		       r.created_at, r.updated_at,
		       pl.name AS parking_lot_name,
		       ps.space_number,
		       v.license_plate
		FROM reservations r
		JOIN parking_lots pl ON r.parking_lot_id = pl.id
		JOIN parking_spaces ps ON r.parking_space_id = ps.id
		WHERE r.user_id = $1 AND r.id = $2
	`

	var queryParams []interface{}
	queryParams = append(queryParams, userID, c.Param("id"))

	// Execute the main query
	rows, err := database.DB.Query(query, queryParams...)
	if err != nil {
		log.Printf("Error fetching user reservations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
		return
	}
	defer rows.Close()

	// Process the result set
	type ReservationWithDetails struct {
		models.Reservation
		ParkingLotName string `json:"parkingLotName"`
		SpaceNumber    string `json:"spaceNumber"`
		LicensePlate   string `json:"licensePlate"`
	}

	var reservation ReservationWithDetails
	for rows.Next() {
		err := rows.Scan(
			&reservation.ID,
			&reservation.UserID,
			&reservation.ParkingLotID,
			&reservation.ParkingSpaceID,
			&reservation.VehicleType,
			&reservation.LicensePlate,
			&reservation.Duration,
			&reservation.Status,
			&reservation.TotalCost,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
			&reservation.ParkingLotName,
			&reservation.SpaceNumber,
			&reservation.LicensePlate,
		)
		if err != nil {
			log.Printf("Error scanning reservation row: %v", err)
			continue
		}
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating reservation rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing reservations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reservation": reservation,
	})
}

// CreateReservation handles creating a new reservation
func CreateReservation(c *gin.Context) {
	// Get user ID from the context (set by auth middleware)
	userID, exists := c.Get("userId")
	if !exists {
		log.Println("User Not Authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse request body
	var req models.ReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("Error Binding Json")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate duration
	if req.Duration <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Duration time must be a number greater than 0"})
		return
	}

	// Validate that the parking space exists and belongs to the specified parking lot
	var spaceExists bool
	err := database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM parking_spaces WHERE id = $1 AND parking_lot_id = $2)",
		req.ParkingSpaceID, req.ParkingLotID,
	).Scan(&spaceExists)

	if err != nil {
		log.Printf("Error checking parking space: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate parking space"})
		return
	}

	if !spaceExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parking space or parking lot"})
		return
	}

	// Begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		log.Printf("Error beginning transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}

	// Generate a new UUID for the reservation
	reservationID := uuid.New().String()
	parsedDate, err := time.Parse("2006-01-02", req.ReservationDate)
	if err != nil {
		log.Println("Failed to parse reservation date:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse reservation date"})
	}

	req.ExpiredAt = parsedDate.AddDate(0, 0, 1)

	// Insert the reservation
	_, err = tx.Exec(`
		INSERT INTO reservations (
			id, user_id, parking_lot_id, parking_space_id, vehicle_type, license_plate, reservation_date, expired_at,
			duration, status, total_cost, payment_status,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`,
		reservationID, userID, req.ParkingLotID, req.ParkingSpaceID, req.VehicleType, req.LicensePlate, req.ReservationDate, req.ExpiredAt,
		req.Duration, "pending", req.TotalCost, "pending",
		time.Now(),
	)

	if err != nil {
		tx.Rollback()
		log.Printf("Error inserting reservation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}

	// Get parking space name
	var parkingSpaceName string
	err = database.DB.QueryRow("SELECT space_number FROM parking_spaces WHERE id = $1", req.ParkingSpaceID).Scan(&parkingSpaceName)
	if err != nil {
		log.Printf("Error fetching parking space name: %v", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	parsedDate, err = time.Parse("2006-01-02", req.ReservationDate)
	if err != nil {
		log.Printf("Error parsing reservation date: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error Parsing Reservation Date"})
		return
	}

	utils.WsManager.HandleReservationAdded(models.Reservation{
		ID:              reservationID,
		UserID:          userID.(string),
		ParkingLotID:    req.ParkingLotID,
		ParkingSpaceID:  req.ParkingSpaceID,
		VehicleType:     req.VehicleType,
		LicensePlate:    req.LicensePlate,
		ReservationDate: parsedDate,
		ExpiredAt:       req.ExpiredAt,
		Duration:        req.Duration,
		Status:          "pending",
		TotalCost:       req.TotalCost,
		PaymentStatus:   "pending",
	})

	// Return the created reservation
	c.JSON(http.StatusCreated, gin.H{
		"message": "Reservation created successfully",
		"reservation": gin.H{
			"id":               reservationID,
			"userId":           userID,
			"parkingLotId":     req.ParkingLotID,
			"parkingSpaceId":   req.ParkingSpaceID,
			"vehicleType":      req.VehicleType,
			"licensePlate":     req.LicensePlate,
			"reservation_date": req.ReservationDate,
			"expired_at":       req.ExpiredAt,
			"duration":         req.Duration,
			"status":           "pending",
			"totalCost":        req.TotalCost,
		},
	})
}

// CancelReservation handles canceling a reservation
func CancelReservation(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userId")
	if !exists {
		log.Println("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get reservation ID from URL params
	reservationID := c.Param("id")
	if reservationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reservation ID is required"})
		return
	}

	// Check if reservation exists and belongs to user
	var existsAndOwned bool
	err := database.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM reservations
			WHERE id = $1 AND user_id = $2 AND status != 'cancelled'
		)
	`, reservationID, userID).Scan(&existsAndOwned)

	if err != nil {
		log.Printf("Error checking reservation ownership: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !existsAndOwned {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found or already cancelled"})
		return
	}

	// Update reservation status to cancelled
	_, err = database.DB.Exec(`
		UPDATE reservations SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1
	`, reservationID)

	if err != nil {
		log.Printf("Error updating reservation status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel reservation"})
		return
	}

	// Notify via WebSocket
	utils.WsManager.HandleReservationUpdated(models.Reservation{
		ID:     reservationID,
		UserID: userID.(string),
		Status: "cancelled",
	})

	c.JSON(http.StatusOK, gin.H{
		"message":       "Reservation cancelled successfully",
		"reservationId": reservationID,
	})
}

// GetReservationByID handles fetching a specific reservation by ID
func GetReservationByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reservation ID is required"})
		return
	}

	// Get user ID from the context (set by auth middleware)
	_, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Query the database for the reservation
	var r models.Reservation
	err := database.DB.QueryRow(`
		SELECT id, user_id, parking_lot_id, parking_space_id, vehicle_type, license_plate,
		       duration, status, total_cost, payment_status,
		       created_at, updated_at
		FROM reservations
		WHERE id = $1
	`, id).Scan(
		&r.ID,
		&r.UserID,
		&r.ParkingLotID,
		&r.ParkingSpaceID,
		&r.VehicleType,
		&r.LicensePlate,
		&r.Duration,
		&r.Status,
		&r.TotalCost,
		&r.CreatedAt,
		&r.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reservation not found"})
		return
	} else if err != nil {
		log.Printf("Error fetching reservation: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservation"})
		return
	}

	// Get additional information about the reservation
	type ReservationDetails struct {
		models.Reservation
		ParkingLotName string `json:"parkingLotName"`
		SpaceNumber    string `json:"spaceNumber"`
		LicensePlate   string `json:"licensePlate"`
		VehicleMake    string `json:"vehicleMake"`
		VehicleModel   string `json:"vehicleModel"`
	}

	var details ReservationDetails
	details.Reservation = r

	err = database.DB.QueryRow(`
		SELECT pl.name, ps.space_number, v.license_plate, v.make, v.model
		FROM parking_lots pl
		JOIN parking_spaces ps ON pl.id = ps.parking_lot_id
		WHERE pl.id = $1 AND ps.id = $2
	`, r.ParkingLotID, r.ParkingSpaceID).Scan(
		&details.ParkingLotName,
		&details.SpaceNumber,
		&details.LicensePlate,
		&details.VehicleMake,
		&details.VehicleModel,
	)

	if err != nil {
		log.Printf("Error fetching reservation details: %v", err)
		// Continue anyway, just won't have the additional details
	}

	c.JSON(http.StatusOK, gin.H{"reservation": details})
}
