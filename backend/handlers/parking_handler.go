package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/holycan/smart-parking-system/database"
	"github.com/holycan/smart-parking-system/models"
)

// GetParkingLots handles fetching all parking lots
func GetParkingLots(c *gin.Context) {
	// Parse query parameters for filtering and pagination
	city := c.Query("city")
	state := c.Query("state")

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build the query based on filters
	query := "SELECT id, name, address, city, state, zip_code, latitude, longitude, " +
		"total_spaces, hourly_rate, open_time, close_time, is_open_24h, created_at, updated_at " +
		"FROM parking_lots WHERE 1=1"

	countQuery := "SELECT COUNT(*) FROM parking_lots WHERE 1=1"

	var queryParams []interface{}
	var paramIndex int = 1

	if city != "" {
		query += " AND city = $" + strconv.Itoa(paramIndex)
		countQuery += " AND city = $" + strconv.Itoa(paramIndex)
		queryParams = append(queryParams, city)
		paramIndex++
	}

	if state != "" {
		query += " AND state = $" + strconv.Itoa(paramIndex)
		countQuery += " AND state = $" + strconv.Itoa(paramIndex)
		queryParams = append(queryParams, state)
		paramIndex++
	}

	// Add order by and pagination
	query += " ORDER BY name ASC LIMIT $" + strconv.Itoa(paramIndex) + " OFFSET $" + strconv.Itoa(paramIndex+1)
	queryParams = append(queryParams, limit, offset)

	// Get total count for pagination
	var totalCount int
	err := database.DB.QueryRow(countQuery, queryParams[:paramIndex-1]...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting parking lots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parking lots"})
		return
	}

	// Execute the main query
	rows, err := database.DB.Query(query, queryParams...)
	if err != nil {
		log.Printf("Error fetching parking lots: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parking lots"})
		return
	}
	defer rows.Close()

	// Process the result set
	parkingLots := []models.ParkingLot{}
	for rows.Next() {
		var lot models.ParkingLot
		err := rows.Scan(
			&lot.ID,
			&lot.Name,
			&lot.Address,
			&lot.City,
			&lot.State,
			&lot.ZipCode,
			&lot.Latitude,
			&lot.Longitude,
			&lot.TotalSpaces,
			&lot.HourlyRate,
			&lot.OpenTime,
			&lot.CloseTime,
			&lot.IsOpen24H,
			&lot.CreatedAt,
			&lot.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning parking lot row: %v", err)
			continue
		}
		parkingLots = append(parkingLots, lot)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating parking lot rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing parking lots"})
		return
	}

	// Calculate pagination metadata
	totalPages := (totalCount + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"parkingLots": parkingLots,
		"pagination": gin.H{
			"total":      totalCount,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
		},
	})
}

// GetParkingLotByID handles fetching a specific parking lot by ID
func GetParkingLotByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parking lot ID is required"})
		return
	}

	// Query the database for the parking lot
	var lot models.ParkingLot
	err := database.DB.QueryRow(
		"SELECT id, name, address, city, state, zip_code, latitude, longitude, "+
			"total_spaces, hourly_rate, open_time, close_time, is_open_24h, created_at, updated_at "+
			"FROM parking_lots WHERE id = $1",
		id,
	).Scan(
		&lot.ID,
		&lot.Name,
		&lot.Address,
		&lot.City,
		&lot.State,
		&lot.ZipCode,
		&lot.Latitude,
		&lot.Longitude,
		&lot.TotalSpaces,
		&lot.HourlyRate,
		&lot.OpenTime,
		&lot.CloseTime,
		&lot.IsOpen24H,
		&lot.CreatedAt,
		&lot.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error fetching parking lot: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Parking lot not found"})
		return
	}

	c.JSON(http.StatusOK, lot)
}

// GetParkingSpacesByLotID handles fetching all parking spaces for a given lot ID
func GetParkingSpacesByLotID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parking Lot ID is required"})
		return
	}

	// Parse query parameters for filtering and pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build the query based on filters
	query := "SELECT id, parking_lot_id, space_number, floor, type, is_occupied, last_updated, created_at, updated_at " +
		"FROM parking_spaces WHERE parking_lot_id = $1 ORDER BY space_number ASC LIMIT $2 OFFSET $3"

	var queryParams []interface{}
	queryParams = append(queryParams, id, limit, offset)

	// Execute the query
	rows, err := database.DB.Query(query, queryParams...)
	if err != nil {
		log.Printf("Error fetching parking spaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching parking spaces"})
		return
	}

	defer rows.Close()

	// Create a slice to store the results
	var parkingSpaces []models.ParkingSpace

	for rows.Next() {
		var space models.ParkingSpace
		err = rows.Scan(
			&space.ID,
			&space.ParkingLotID,
			&space.SpaceNumber,
			&space.Floor,
			&space.Type,
			&space.IsOccupied,
			&space.LastUpdated,
			&space.CreatedAt,
			&space.UpdatedAt,
		)

		if err != nil {
			log.Printf("Error scanning row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching parking spaces"})
			return
		}

		parkingSpaces = append(parkingSpaces, space)
	}

	c.JSON(http.StatusOK, parkingSpaces)
}

// GetParkingSpaceByLotID handles fetching all parking spaces by Lot ID
func GetParkingSpaceByLotID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parking Lot ID is required"})
		return
	}

	// Query all parking spaces for the lot
	rows, err := database.DB.Query(
		`SELECT id, parking_lot_id, space_number, floor, type, is_occupied, last_updated, created_at, updated_at 
		FROM parking_spaces WHERE parking_lot_id = $1`, id)
	if err != nil {
		log.Printf("Error querying parking spaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parking spaces"})
		return
	}
	defer rows.Close()

	var spaces []models.ParkingSpace

	for rows.Next() {
		var space models.ParkingSpace
		err := rows.Scan(
			&space.ID,
			&space.ParkingLotID,
			&space.SpaceNumber,
			&space.Floor,
			&space.Type,
			&space.IsOccupied,
			&space.LastUpdated,
			&space.CreatedAt,
			&space.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning parking space: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing parking spaces"})
			return
		}
		spaces = append(spaces, space)
	}

	if len(spaces) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No parking spaces found"})
		return
	}

	// Get parking lot name (pakai id param langsung)
	var lotName string
	err = database.DB.QueryRow(
		"SELECT name FROM parking_lots WHERE id = $1", id,
	).Scan(&lotName)
	if err != nil {
		log.Printf("Error fetching parking lot name: %v", err)
		lotName = "" // default jika error
	}

	// Optional: Ambil reservation info untuk setiap space
	type SpaceWithReservation struct {
		ParkingSpace         models.ParkingSpace `json:"parkingSpace"`
		HasActiveReservation bool                `json:"hasActiveReservation"`
		ReservationEndTime   time.Time           `json:"reservationEndTime"`
	}

	var result []SpaceWithReservation

	now := time.Now()

	for _, s := range spaces {
		var hasActiveReservation bool
		var reservationEndTime time.Time

		err = database.DB.QueryRow(
			`SELECT EXISTS(
				SELECT 1 FROM reservations 
				WHERE parking_space_id = $1 AND status = 'active' AND end_time > $2
			),
			COALESCE((
				SELECT end_time FROM reservations 
				WHERE parking_space_id = $1 AND status = 'active' AND end_time > $2 
				ORDER BY end_time ASC LIMIT 1
			), $2)`,
			s.ID, now,
		).Scan(&hasActiveReservation, &reservationEndTime)

		if err != nil {
			log.Printf("Error checking reservation for space ID %v: %v", s.ID, err)
			// Default jika error
			hasActiveReservation = false
			reservationEndTime = time.Time{}
		}

		result = append(result, SpaceWithReservation{
			ParkingSpace:         s,
			HasActiveReservation: hasActiveReservation,
			ReservationEndTime:   reservationEndTime,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"parkingLotName": lotName,
		"spaces":         result,
	})
}

// GetParkingSpaces handles fetching all parking spaces
func GetParkingSpaces(c *gin.Context) {
	// Parse query parameters for filtering and pagination
	lotID := c.Query("lotId")
	floor := c.Query("floor")
	spaceType := c.Query("type")
	availability := c.Query("availability") // "all", "available", "occupied"

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build the query based on filters
	query := "SELECT id, parking_lot_id, space_number, floor, type, is_occupied, last_updated, created_at, updated_at " +
		"FROM parking_spaces WHERE 1=1"

	countQuery := "SELECT COUNT(*) FROM parking_spaces WHERE 1=1"

	var queryParams []interface{}
	var paramIndex int = 1

	if lotID != "" {
		query += " AND parking_lot_id = $" + strconv.Itoa(paramIndex)
		countQuery += " AND parking_lot_id = $" + strconv.Itoa(paramIndex)
		queryParams = append(queryParams, lotID)
		paramIndex++
	}

	if floor != "" {
		floorNum, err := strconv.Atoi(floor)
		if err == nil {
			query += " AND floor = $" + strconv.Itoa(paramIndex)
			countQuery += " AND floor = $" + strconv.Itoa(paramIndex)
			queryParams = append(queryParams, floorNum)
			paramIndex++
		}
	}

	if spaceType != "" {
		query += " AND type = $" + strconv.Itoa(paramIndex)
		countQuery += " AND type = $" + strconv.Itoa(paramIndex)
		queryParams = append(queryParams, spaceType)
		paramIndex++
	}

	if availability == "available" {
		query += " AND is_occupied = false"
		countQuery += " AND is_occupied = false"
	} else if availability == "occupied" {
		query += " AND is_occupied = true"
		countQuery += " AND is_occupied = true"
	}

	// Add order by and pagination
	query += " ORDER BY space_number ASC LIMIT $" + strconv.Itoa(paramIndex) + " OFFSET $" + strconv.Itoa(paramIndex+1)
	queryParams = append(queryParams, limit, offset)

	// Get total count for pagination
	var totalCount int
	err := database.DB.QueryRow(countQuery, queryParams[:paramIndex-1]...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting parking spaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parking spaces"})
		return
	}

	// Execute the main query
	rows, err := database.DB.Query(query, queryParams...)
	if err != nil {
		log.Printf("Error fetching parking spaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parking spaces"})
		return
	}
	defer rows.Close()

	// Process the result set
	parkingSpaces := []models.ParkingSpace{}
	for rows.Next() {
		var space models.ParkingSpace
		err := rows.Scan(
			&space.ID,
			&space.ParkingLotID,
			&space.SpaceNumber,
			&space.Floor,
			&space.Type,
			&space.IsOccupied,
			&space.LastUpdated,
			&space.CreatedAt,
			&space.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning parking space row: %v", err)
			continue
		}
		parkingSpaces = append(parkingSpaces, space)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating parking space rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing parking spaces"})
		return
	}

	// Calculate pagination metadata
	totalPages := (totalCount + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"parkingSpaces": parkingSpaces,
		"pagination": gin.H{
			"total":      totalCount,
			"page":       page,
			"limit":      limit,
			"totalPages": totalPages,
		},
	})
}

// GetParkingSpaceByID handles fetching a specific parking space by ID
func GetParkingSpaceByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parking space ID is required"})
		return
	}

	// Query the database for the parking space
	var space models.ParkingSpace
	err := database.DB.QueryRow(
		"SELECT id, parking_lot_id, space_number, floor, type, is_occupied, last_updated, created_at, updated_at "+
			"FROM parking_spaces WHERE id = $1",
		id,
	).Scan(
		&space.ID,
		&space.ParkingLotID,
		&space.SpaceNumber,
		&space.Floor,
		&space.Type,
		&space.LastUpdated,
		&space.CreatedAt,
		&space.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error fetching parking space: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Parking space not found"})
		return
	}

	// Get parking lot information for this space
	var lotName string
	err = database.DB.QueryRow(
		"SELECT name FROM parking_lots WHERE id = $1",
		space.ParkingLotID,
	).Scan(&lotName)

	if err != nil {
		log.Printf("Error fetching parking lot name: %v", err)
		// Continue anyway, just won't have the lot name
	}

	// Check if there's an active reservation for this space
	var hasActiveReservation bool
	var reservationEndTime time.Time
	err = database.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM reservations WHERE parking_space_id = $1 AND status = 'active' AND end_time > $2), "+
			"COALESCE((SELECT end_time FROM reservations WHERE parking_space_id = $1 AND status = 'active' AND end_time > $2 ORDER BY end_time ASC LIMIT 1), $2)",
		id, time.Now(),
	).Scan(&hasActiveReservation, &reservationEndTime)

	if err != nil {
		log.Printf("Error checking for active reservation: %v", err)
		// Continue anyway, just won't have the reservation info
	}

	c.JSON(http.StatusOK, gin.H{
		"parkingSpace":   space,
		"parkingLotName": lotName,
		"reservation": gin.H{
			"hasActiveReservation": hasActiveReservation,
			"reservationEndTime":   reservationEndTime,
		},
	})
}

// FilterParkingSpaces handles filtering and pagination for parking spaces
func FilterParkingSpaces(c *gin.Context) {
	// Get filter parameters
	parkingLotID := c.Query("parking_lot_id")
	availableOnly := c.Query("available_only") == "true"
	vehicleType := c.Query("vehicle_type")

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Build query
	query := `
		SELECT ps.id, ps.parking_lot_id, ps.space_number, ps.space_type, ps.is_occupied, 
		       ps.is_reserved, ps.is_disabled_only, ps.hourly_rate, ps.created_at, ps.updated_at,
		       pl.name AS parking_lot_name, pl.address
		FROM parking_spaces ps
		JOIN parking_lots pl ON ps.parking_lot_id = pl.id
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	// Add filters
	if parkingLotID != "" {
		query += " AND ps.parking_lot_id = $" + strconv.Itoa(argIndex)
		args = append(args, parkingLotID)
		argIndex++
	}

	if availableOnly {
		query += " AND ps.is_occupied = false AND ps.is_reserved = false"
	}

	if vehicleType != "" {
		query += " AND ps.space_type = $" + strconv.Itoa(argIndex)
		args = append(args, vehicleType)
		argIndex++
	}

	// Add sorting and pagination
	query += " ORDER BY pl.name, ps.space_number LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := database.DB.Query(query, args...)
	if err != nil {
		log.Printf("Error querying parking spaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch parking spaces"})
		return
	}
	defer rows.Close()

	// Parse results
	parkingSpaces := []gin.H{}
	for rows.Next() {
		var id, parkingLotID, spaceNumber, spaceType, parkingLotName, address string
		var isOccupied, isReserved, isDisabledOnly bool
		var hourlyRate float64
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&id, &parkingLotID, &spaceNumber, &spaceType, &isOccupied,
			&isReserved, &isDisabledOnly, &hourlyRate, &createdAt, &updatedAt,
			&parkingLotName, &address,
		)

		if err != nil {
			log.Printf("Error scanning parking space row: %v", err)
			continue
		}

		parkingSpaces = append(parkingSpaces, gin.H{
			"id":               id,
			"parking_lot_id":   parkingLotID,
			"space_number":     spaceNumber,
			"space_type":       spaceType,
			"is_occupied":      isOccupied,
			"is_reserved":      isReserved,
			"is_disabled_only": isDisabledOnly,
			"hourly_rate":      hourlyRate,
			"created_at":       createdAt,
			"updated_at":       updatedAt,
			"parking_lot_name": parkingLotName,
			"address":          address,
		})
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating parking space rows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing parking space data"})
		return
	}

	// Get total count for pagination
	countQuery := strings.Replace(query, `
		SELECT ps.id, ps.parking_lot_id, ps.space_number, ps.space_type, ps.is_occupied, 
		       ps.is_reserved, ps.is_disabled_only, ps.hourly_rate, ps.created_at, ps.updated_at,
		       pl.name AS parking_lot_name, pl.address`,
		"SELECT COUNT(*)", 1)

	// Remove ORDER BY and LIMIT clauses for count query
	countQuery = countQuery[:strings.LastIndex(countQuery, "ORDER BY")]

	var totalCount int
	err = database.DB.QueryRow(countQuery, args[:len(args)-2]...).Scan(&totalCount)
	if err != nil {
		log.Printf("Error counting parking spaces: %v", err)
		totalCount = len(parkingSpaces)
	}

	c.JSON(http.StatusOK, gin.H{
		"parking_spaces": parkingSpaces,
		"total":          totalCount,
		"page":           page,
		"limit":          limit,
		"total_pages":    (totalCount + limit - 1) / limit,
	})
}
