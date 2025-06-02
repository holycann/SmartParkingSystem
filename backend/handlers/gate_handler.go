package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/holycan/smart-parking-system/lock"
	"github.com/holycan/smart-parking-system/models"
	"github.com/holycan/smart-parking-system/services"
	"github.com/holycan/smart-parking-system/utils"
)

// Error definitions
var (
	ErrUserNotAuthenticated = errors.New("user not authenticated")
	ErrBookingNotFound      = errors.New("booking not found")
	ErrNoAvailableSpot      = errors.New("no available parking spots")
	ErrUpdateBooking        = errors.New("failed to update booking")
	ErrUpdateParkingSpace   = errors.New("failed to update parking space status")
	ErrFetchParkingData     = errors.New("failed to fetch parking space data")
)

// CheckInHandler handles user check-in requests
func CheckInHandler(c *gin.Context) {
	id := c.Param("id")
	userID, exists := getUserIDFromContext(c)
	if !exists {
		respondWithError(c, http.StatusUnauthorized, ErrUserNotAuthenticated)
		return
	}

	// Queue the check-in request for processing
	utils.ParkingQueue <- map[string]interface{}{
		"user_id":        userID,
		"reservation_id": id,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Check-in request accepted and being processed",
	})
}

// PaymentHandler handles payment for parking reservations
func PaymentHandler(c *gin.Context) {
	id := c.Param("id")
	userID, exists := getUserIDFromContext(c)
	if !exists {
		respondWithError(c, http.StatusUnauthorized, ErrUserNotAuthenticated)
		return
	}

	// Get booking information
	booking, err := services.GetBookingByID(id, userID)
	if err != nil || booking == nil {
		log.Printf("Booking not found: %v", err)
		respondWithError(c, http.StatusNotFound, ErrBookingNotFound)
		return
	}

	// Update payment status
	if err := services.UpdateBookingPaymentStatus("completed", id); err != nil {
		log.Printf("Failed to update payment status for booking %s: %v", id, err)
		respondWithError(c, http.StatusInternalServerError, ErrUpdateBooking)
		return
	}

	// Notify clients about the payment
	notifyParkingUpdate(booking.ParkingLotID, booking.ParkingSpaceID, true, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment completed successfully",
	})
}

// CheckOutHandler handles user check-out requests
func CheckOutHandler(c *gin.Context) {
	id := c.Param("id")
	userID, exists := getUserIDFromContext(c)
	if !exists {
		respondWithError(c, http.StatusUnauthorized, ErrUserNotAuthenticated)
		return
	}

	// Get booking information
	booking, err := services.GetBookingByID(id, userID)
	if err != nil || booking == nil {
		log.Printf("Booking not found: %v", err)
		respondWithError(c, http.StatusNotFound, ErrBookingNotFound)
		return
	}

	// Mark booking as completed
	if _, err := services.UpdateBookingWithSpot("completed", userID, id,
		booking.ParkingLotID, booking.ParkingSpaceID); err != nil {
		log.Printf("Failed to update booking %s: %v", id, err)
		respondWithError(c, http.StatusInternalServerError, ErrUpdateBooking)
		return
	}

	// Free up the parking space
	if err := services.UpdateParkingSpaceOccupied(false, booking.ParkingSpaceID); err != nil {
		log.Printf("Failed to update parking space %s: %v", booking.ParkingSpaceID, err)
		respondWithError(c, http.StatusInternalServerError, ErrUpdateParkingSpace)
		return
	}

	// Notify clients about the space becoming available
	notifyParkingUpdate(booking.ParkingLotID, booking.ParkingSpaceID, false, false)

	// Send availability notification
	if err := notifySpaceAvailability(booking.ParkingSpaceID); err != nil {
		log.Printf("Warning: Failed to send availability notification: %v", err)
		// Continue execution - this is not a critical error
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Check-out processed successfully",
	})
}

// ProcessCheckIn handles the asynchronous check-in process with improved spot availability checking
func ProcessCheckIn(req map[string]interface{}) error {
	// Acquire semaphore to limit concurrent processing
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	select {
	case utils.Semaphore <- struct{}{}:
		defer func() { <-utils.Semaphore }()
	case <-ctx.Done():
		return fmt.Errorf("timed out waiting for processing resources")
	}

	userID := req["user_id"].(string)
	reservationID := req["reservation_id"].(string)

	// Get booking information
	booking, err := services.GetBookingByID(reservationID, userID)
	if err != nil || booking == nil {
		log.Printf("Booking not found: %v", err)
		return ErrBookingNotFound
	}

	// Try to lock the originally assigned spot
	spaceID := booking.ParkingSpaceID
	parkingLotID := booking.ParkingLotID

	// Try acquiring lock first
	mutex, err := lock.AcquireLock("spot-lock:"+spaceID, 10*time.Second)

	// First fallback: If lock acquisition fails, spot is being processed by another request
	if err != nil {
		log.Println("Original spot is locked, finding alternative spot...")
		goto FindAlternative
	}

	// Even with successful lock, double-check if spot is physically occupied
	{
		isOccupied, checkErr := services.IsParkingSpaceOccupied(spaceID)
		if checkErr != nil {
			lock.ReleaseLock(mutex)
			log.Printf("Failed to check spot status: %v", checkErr)
			goto FindAlternative
		}

		// Second fallback: If spot is occupied despite successful lock acquisition
		if isOccupied {
			lock.ReleaseLock(mutex)
			notifyFindAlternative(spaceID)
			log.Printf("Spot %s is already occupied despite successful lock, finding alternative", spaceID)
			goto FindAlternative
		}
	}

	// Successfully locked and verified as unoccupied - proceed with this spot
	goto ProcessSpot

FindAlternative:
	// Find and lock an available alternate spot
	{
		var space map[string]interface{}
		mutex, space, err = services.FindAndLockAvailableSpot(parkingLotID)

		if err != nil {
			notifyNoAvailableAlternativeSpot(spaceID)
			log.Println("No available parking spots or failed to lock any spot")
			return ErrNoAvailableSpot
		}

		// Update with new spot details
		spaceID = space["spot_id"].(string)

		// Double-check that this spot is indeed unoccupied
		isOccupied, checkErr := services.IsParkingSpaceOccupied(spaceID)
		if checkErr != nil || isOccupied {
			lock.ReleaseLock(mutex)
			notifyNoAvailableAlternativeSpot(spaceID)
			log.Printf("Alternative spot %s is unavailable or occupied: %v", spaceID, checkErr)
			return ErrNoAvailableSpot
		}

		notifyAvailableAlternativeSpot(spaceID)
	}

ProcessSpot:
	defer lock.ReleaseLock(mutex)

	// Update booking with the assigned spot
	_, err = services.UpdateBookingWithSpot("active", userID, reservationID, parkingLotID, spaceID)
	if err != nil {
		log.Printf("Failed to update booking with spot %s: %v", spaceID, err)
		return ErrUpdateBooking
	}

	// Mark parking space as occupied
	if err := services.UpdateParkingSpaceOccupied(true, spaceID); err != nil {
		log.Printf("Failed to update parking space %s: %v", spaceID, err)
		// Attempt to revert booking update on failure
		if revertErr := services.RevertBookingSpot(reservationID, userID); revertErr != nil {
			log.Printf("Failed to revert booking after space update failure: %v", revertErr)
		}
		return ErrUpdateParkingSpace
	}

	// Notify clients about space being occupied
	notifyParkingUpdate(parkingLotID, spaceID, true, false)

	// Send occupancy notification
	if err := notifySpaceOccupancy(spaceID); err != nil {
		log.Printf("Warning: Failed to send occupancy notification: %v", err)
		// Continue execution - this is not a critical error
	}

	return nil
}

// getUserIDFromContext extracts the user ID from the context
func getUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userId")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// respondWithError sends a standardized error response
func respondWithError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, gin.H{"error": err.Error()})
}

// notifyParkingUpdate broadcasts parking space updates
func notifyParkingUpdate(parkingLotID, spaceID string, isOccupied, isPaid bool) {
	utils.WsManager.HandleParkingUpdate(models.ParkingEvent{
		ParkingLotID: parkingLotID,
		SpaceID:      spaceID,
		IsOccupied:   isOccupied,
		IsPaid:       isPaid,
		Timestamp:    time.Now().Unix(),
	})
}

// notifySpaceAvailability sends notification about space becoming available
func notifySpaceAvailability(spaceID string) error {
	parkingData, err := services.GetParkingDataById(spaceID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFetchParkingData, err)
	}

	message := fmt.Sprintf(
		"Parking Space %s Floor %s Zone %s is available now!",
		parkingData["space_number"].(string),
		parkingData["floor"].(string),
		parkingData["zone_name"].(string),
	)

	utils.WsManager.HandleNotificationUpdate(models.NotificationEvent{
		Type:    "availability_update",
		Message: message,
	})

	return nil
}

func notifyNoAvailableAlternativeSpot(spaceID string) error {
	parkingData, err := services.GetParkingDataById(spaceID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFetchParkingData, err)
	}

	message := fmt.Sprintf(
		"Alternative Space %s Floor %s Zone %s is unavailable or occupied!",
		parkingData["space_number"].(string),
		parkingData["floor"].(string),
		parkingData["zone_name"].(string),
	)

	utils.WsManager.HandleNotificationUpdate(models.NotificationEvent{
		Type:    "availability_update",
		Message: message,
	})

	return nil
}

func notifyAvailableAlternativeSpot(spaceID string) error {
	parkingData, err := services.GetParkingDataById(spaceID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFetchParkingData, err)
	}

	message := fmt.Sprintf(
		"Alternative Space %s Floor %s Zone %s is available, changing your parking spot",
		parkingData["space_number"].(string),
		parkingData["floor"].(string),
		parkingData["zone_name"].(string),
	)

	utils.WsManager.HandleNotificationUpdate(models.NotificationEvent{
		Type:    "availability_update",
		Message: message,
	})

	return nil
}

func notifyFindAlternative(spaceID string) error {
	parkingData, err := services.GetParkingDataById(spaceID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFetchParkingData, err)
	}

	message := fmt.Sprintf(
		"Parking Space %s Floor %s Zone %s is occupied!, Finding alternative spot",
		parkingData["space_number"].(string),
		parkingData["floor"].(string),
		parkingData["zone_name"].(string),
	)

	utils.WsManager.HandleNotificationUpdate(models.NotificationEvent{
		Type:    "availability_update",
		Message: message,
	})

	return nil
}

// notifySpaceOccupancy sends notification about space becoming occupied
func notifySpaceOccupancy(spaceID string) error {
	parkingData, err := services.GetParkingDataById(spaceID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrFetchParkingData, err)
	}

	message := fmt.Sprintf(
		"Parking Space %s Floor %s Zone %s is occupied",
		parkingData["space_number"].(string),
		parkingData["floor"].(string),
		parkingData["zone_name"].(string),
	)

	utils.WsManager.HandleNotificationUpdate(models.NotificationEvent{
		Type:    "availability_update",
		Message: message,
	})

	return nil
}
