package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/holycan/smart-parking-system/database"
	"github.com/holycan/smart-parking-system/handlers"
	"github.com/holycan/smart-parking-system/lock"
	"github.com/holycan/smart-parking-system/services"
)

func TestFullCheckInFlow(t *testing.T) {
	lock.InitializeRedisLock()
	err := database.Initialize()
	assert.NoError(t, err)

	router := gin.Default()
	router.POST("/checkin/:id", handlers.CheckInHandler)

	// Data user & reservasi
	users := []struct {
		userID      string
		reservation string
	}{
		{"118101c7-d496-451b-bb81-56ee2c011b3d", "3df7cd7b-ff74-4c74-81b7-aaaacfacf8e0"},
		{"18a183af-e32f-475a-8f74-50dbb969d0e7", "6b27fc8b-40ba-46ec-9fef-1de8a98e0d2b"},
		{"1ae71ef7-d69a-4bc5-9952-a7701aa8716f", "abe3dbe5-a847-4965-8c0d-70883e8364d1"},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	spotUsed := map[string]string{}

	for _, u := range users {
		wg.Add(1)
		go func(userID, bookingID string) {
			defer wg.Done()

			req, _ := http.NewRequest(http.MethodPost, "/checkin/"+bookingID, nil)
			rec := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(rec)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: bookingID}}
			c.Set("userId", userID)

			handlers.CheckInHandler(c)

			t.Logf("🧪 User: %s | Status: %d | Response: %s", userID, rec.Code, rec.Body.String())

			// Tunggu proses async selesai
			time.Sleep(2 * time.Second)

			// Ambil reservasi hasil assign
			res, err := services.GetBookingByID(bookingID, userID)
			if err != nil {
				t.Errorf("❌ Gagal ambil reservasi %s: %v", bookingID, err)
				return
			}

			if res.ParkingSpaceID != "" {
				spotID := res.ParkingSpaceID

				mu.Lock()
				if prev, exists := spotUsed[spotID]; exists {
					t.Errorf("❌ RACE: Spot %s digunakan oleh %s & %s!", spotID, prev, userID)
				} else {
					spotUsed[spotID] = userID
					t.Logf("✅ %s diarahkan ke Spot: %s at %s", userID, spotID, time.Now().Format(time.RFC3339Nano))
				}
				mu.Unlock()
			} else {
				t.Logf("⚠️ %s gagal diarahkan ke spot mana pun", userID)
			}
		}(u.userID, u.reservation)
	}

	wg.Wait()

	t.Logf("📊 Final Spot Usage: %+v", spotUsed)
}
