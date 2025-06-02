package utils

import (
	"os"
	"sync"
	"time"

	"github.com/holycan/smart-parking-system/ws"
)

// Global variables for concurrency control
var (
	// Mutex for synchronizing access to shared resources
	Mutex sync.RWMutex

	// WaitGroup for managing goroutines
	Wg sync.WaitGroup

	// Channel for ngrok ready signal
	NgrokReady = make(chan struct{})

	// Channel for graceful shutdown
	ShutdownChan = make(chan os.Signal, 1)

	// Server start time for uptime tracking
	StartTime = time.Now()

	// WebSocket manager
	WsManager *ws.WebSocketManager

	// Ngrok URL
	NgrokURL string

	// Channel untuk buffer request parkir
	ParkingQueue = make(chan map[string]interface{}, 100)

	// Semaphore untuk membatasi jumlah akses paralel ke spot parkir
	Semaphore = make(chan struct{}, 50) // Max 5 permintaan yang diproses bersamaan
)
