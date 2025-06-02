package main

import (
	"context"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	cronLib "github.com/robfig/cron/v3"

	"github.com/holycan/smart-parking-system/cron"
	"github.com/holycan/smart-parking-system/database"
	"github.com/holycan/smart-parking-system/handlers"
	"github.com/holycan/smart-parking-system/lock"
	"github.com/holycan/smart-parking-system/routes"
	"github.com/holycan/smart-parking-system/utils"
	"github.com/holycan/smart-parking-system/ws"
)

func initEnvironment() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}
}

func initDatabaseWithRetry() {
	err := database.Initialize()
	if err == nil {
		log.Println("Database initialized successfully")
		return
	}

	log.Printf("Error: Database initialization failed: %v", err)
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		retryDelay := time.Duration(math.Pow(2, float64(i))) * time.Second
		log.Printf("Retrying database connection in %v... (Attempt %d/%d)", retryDelay, i+1, maxRetries)
		time.Sleep(retryDelay)

		err = database.Initialize()
		if err == nil {
			log.Println("Database connection established successfully after retry")
			return
		}

		if i == maxRetries-1 {
			log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
		}
	}
}

func initWebSocketManager() {
	utils.WsManager = ws.NewWebSocketManager()
	go func() {
		log.Println("WebSocket manager starting...")
		utils.WsManager.Start()
		log.Println("WebSocket manager stopped.")
	}()
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(utils.Logger())
	router.Static("/qrcodes", "./static/qrcodes")

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve frontend
	frontendDir := filepath.Join("../", "frontend", "build")
	if _, err := os.Stat(frontendDir); !os.IsNotExist(err) {
		router.Static("/static", filepath.Join(frontendDir, "static"))
		router.StaticFile("/favicon.ico", filepath.Join(frontendDir, "favicon.ico"))
		router.StaticFile("/manifest.json", filepath.Join(frontendDir, "manifest.json"))
		router.StaticFile("/logo192.png", filepath.Join(frontendDir, "logo192.png"))
		router.StaticFile("/logo512.png", filepath.Join(frontendDir, "logo512.png"))
		router.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(frontendDir, "index.html"))
		})
	} else {
		router.NoRoute(func(c *gin.Context) {
			if c.Request.URL.Path == "/" {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error":   "Frontend application is not available",
					"message": "Frontend is not built/deployed yet.",
				})
			} else {
				c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
			}
		})
	}

	routes.RegisterRoutes(router)
	return router
}

func processParkingRequest() {
	for req := range utils.ParkingQueue {
		// Memproses request parkir secara asinkron dengan semaphore dan fault tolerance
		err := handlers.ProcessCheckIn(req)
		if err != nil {
			log.Println("Check-in failed:", err)
		}
	}
}

func cronJob() {
	c := cronLib.New()

	// Schedule the TimeLimit cron job (every 15 minutes)
	_, err := c.AddFunc("*/1 * * * *", cron.TimeLimit)
	if err != nil {
		log.Fatalf("Error scheduling TimeLimit cron job: %v", err)
	}

	// Schedule the Expired cron job (every day at midnight)
	_, err = c.AddFunc("0 0 * * *", cron.ExpiredTime)
	if err != nil {
		log.Fatalf("Error scheduling Expired cron job: %v", err)
	}

	// Start cron jobs in a separate goroutine so they don't block the main goroutine
	go func() {
		c.Start()
	}()
}

func startServer(router *gin.Engine) *http.Server {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s...\n", port)
		certFile := os.Getenv("TLS_CERT_FILE")
		keyFile := os.Getenv("TLS_KEY_FILE")

		var err error
		if certFile != "" && keyFile != "" {
			log.Println("Starting server with TLS enabled")
			err = server.ListenAndServeTLS(certFile, keyFile)
		} else {
			log.Println("Starting server without TLS")
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	return server
}

func gracefulShutdown(server *http.Server) {
	signal.Notify(utils.ShutdownChan, syscall.SIGINT, syscall.SIGTERM)
	<-utils.ShutdownChan
	log.Println("Shutting down server...")

	// Stop WebSocket manager
	utils.WsManager.Stop()

	// Gracefully shut down the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Close database connection during shutdown
	database.Close()

	// Attempt to gracefully shut down the HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

func main() {
	initEnvironment()
	initDatabaseWithRetry()
	defer database.Close()

	lock.InitializeRedisLock()

	initWebSocketManager()

	go processParkingRequest()
	cronJob()

	router := setupRouter()
	server := startServer(router)
	gracefulShutdown(server)
}
