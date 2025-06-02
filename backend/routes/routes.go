package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/holycan/smart-parking-system/handlers"
	"github.com/holycan/smart-parking-system/middleware"
	"github.com/holycan/smart-parking-system/utils"
)

// RegisterRoutes sets up all the routes for the application
func RegisterRoutes(router *gin.Engine) {

	// Add WebSocket route with authentication middleware
	router.GET("/ws", middleware.AuthWebSocketMiddleware(), func(c *gin.Context) {
		utils.WsManager.HandleWebSocket(c)
	})

	// API routes
	api := router.Group("/api")
	{
		// Public routes
		api.POST("/users/register", handlers.RegisterUser)
		api.POST("/users/login", handlers.LoginUser)
		api.POST("/users/validate-token", handlers.ValidateToken)
		api.POST("/users/request-password-reset", handlers.RequestPasswordReset)
		api.POST("/users/reset-password", handlers.ResetPassword)
		api.GET("/users/verify-email/:token", handlers.VerifyEmail)
		api.GET("/parking-lots", handlers.GetParkingLots)
		api.GET("/parking-lots/:id", handlers.GetParkingLotByID)
		api.GET("/parking-lots/:id/space", handlers.GetParkingSpaceByLotID)

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// Gate routes
			protected.POST("/checkin/:id", handlers.CheckInHandler)
			protected.POST("/checkout/:id", handlers.CheckOutHandler)
			protected.POST("/payment/:id", handlers.PaymentHandler)

			// User routes
			protected.GET("/users/profile", handlers.GetUserProfile)
			protected.PUT("/users/profile", handlers.UpdateUserProfile)
			protected.POST("/users/logout", handlers.LogoutUser)
			protected.POST("/users/refresh-token", handlers.RefreshToken)
			protected.POST("/users/setup-mfa", handlers.SetupMFA)
			protected.POST("/users/verify-mfa", handlers.VerifyMFA)

			// Parking space routes
			protected.GET("/parking-spaces", handlers.GetParkingSpaces)
			protected.GET("/parking-spaces/:id", handlers.GetParkingSpaceByID)
			protected.GET("/parking-spaces/filter", handlers.FilterParkingSpaces)

			// Reservation routes
			reservations := protected.Group("/reservations")
			{
				reservations.GET("/user", handlers.GetUserReservations)
				reservations.GET("/details/:id", handlers.GetReservationDetails)
				reservations.POST("/create", handlers.CreateReservation)
				reservations.POST("/cancel/:id", handlers.CancelReservation)
			}
		}
	}
}
