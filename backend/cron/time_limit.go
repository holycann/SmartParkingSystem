package cron

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/holycan/smart-parking-system/database"
	"github.com/holycan/smart-parking-system/models"
	"github.com/holycan/smart-parking-system/utils"
)

func TimeLimit() {
	db := database.DB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query untuk mendapatkan data reservasi aktif
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, checkin_time, duration, total_cost FROM reservations WHERE status = 'active'")
	if err != nil {
		log.Println("Error fetching active reservations:", err)
		return
	}
	defer rows.Close()

	var reservations []models.Reservation
	for rows.Next() {
		var reservation models.Reservation
		err := rows.Scan(&reservation.ID, &reservation.UserID, &reservation.CheckinTime, &reservation.Duration, &reservation.TotalCost)
		if err != nil {
			log.Println("Error scanning reservation:", err)
			continue
		}
		reservations = append(reservations, reservation)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error reading rows:", err)
		return
	}

	// Mengecek setiap reservasi apakah waktu limitnya kurang dari 15 menit
	for _, reservation := range reservations {
		// Hitung waktu batas
		timeLimit := reservation.CheckinTime.Add(time.Duration(reservation.Duration) * time.Minute)
		timeRemaining := timeLimit.Sub(time.Now())

		// Jika waktu yang tersisa kurang dari 15 menit, kirimkan notifikasi
		if timeRemaining <= 15*time.Minute && timeRemaining > 0 {
			// Kirimkan notifikasi
			utils.WsManager.HandleNotificationUpdate(models.NotificationEvent{
				Type:    "time_limit",
				Message: fmt.Sprintf("Your reservation time limit is about to expire in 15 minutes and cost you $%.2f", reservation.TotalCost),
			})
			log.Printf("Sent notification for reservation ID %s about time limit expiration\n", reservation.ID)
		}
	}

}

func ExpiredTime() {
	db := database.DB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Query untuk mendapatkan data reservasi dengan status "pending"
	rows, err := db.QueryContext(ctx, "SELECT id, user_id, checkin_time, duration, reservation_date FROM reservations WHERE status = 'pending'")
	if err != nil {
		log.Println("Error fetching pending reservations:", err)
		return
	}
	defer rows.Close()

	var reservations []models.Reservation
	for rows.Next() {
		var reservation models.Reservation
		err := rows.Scan(&reservation.ID, &reservation.UserID, &reservation.CheckinTime, &reservation.Duration, &reservation.ReservationDate)
		if err != nil {
			log.Println("Error scanning reservation:", err)
			continue
		}
		reservations = append(reservations, reservation)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error reading rows:", err)
		return
	}

	// Mengecek setiap reservasi apakah waktu batasnya lebih dari 1 hari dari reservation_date
	for _, reservation := range reservations {
		// Hitung selisih antara reservation_date dan waktu saat ini
		reservationDate, err := time.Parse(time.RFC3339, reservation.ReservationDate.Format("2006-01-02T15:04:05Z"))
		if err != nil {
			log.Println("Error parsing reservation date:", err)
			continue
		}

		timeElapsed := time.Now().Sub(reservationDate)

		// Jika lebih dari 1 hari, update status menjadi expired
		if timeElapsed > 24*time.Hour {
			// Update status menjadi expired
			_, err := db.ExecContext(ctx, "UPDATE reservations SET status = 'expired' WHERE id = ?", reservation.ID)
			if err != nil {
				log.Println("Error updating reservation status:", err)
				continue
			}

			// Kirimkan notifikasi bahwa reservasi telah expired
			utils.WsManager.HandleNotificationUpdate(models.NotificationEvent{
				Type:    "expired",
				Message: "Your reservation has expired due to time exceeding 1 day.",
			})
			log.Printf("Reservation ID %s has expired and notification sent\n", reservation.ID)
		}
	}
}
