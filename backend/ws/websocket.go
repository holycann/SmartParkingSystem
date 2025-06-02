package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/holycan/smart-parking-system/models"
)

// Client represents a connected WebSocket client
type Client struct {
	ID           string
	Conn         *websocket.Conn
	Send         chan []byte
	UserID       string
	ParkingSpace map[string]models.ParkingEvent
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebSocketManager handles WebSocket connections and broadcasts
type WebSocketManager struct {
	// Registered clients
	clients map[string]*Client

	// Mutex for synchronizing access to the clients map
	clientsMutex sync.RWMutex

	// Channel for notification updates
	reservationAdded chan models.Reservation

	// Channel for notification updates
	ReservationUpdates chan models.Reservation

	// Channel for notification updates
	parkingSpaceUpdates chan models.ParkingEvent

	// Channel for notification updates
	notificationUpdates chan models.NotificationEvent

	// Channel for gate updates
	gateEvents chan models.GateEvent

	// Channel for registering new clients
	register chan *Client

	// Channel for unregistering clients
	unregister chan *Client

	// Channel for graceful shutdown
	shutdown chan struct{}

	// Map to track interest clients for each parking space
	interestClients map[string]map[string]*models.ReservationInfo // parkingSpaceID -> userID -> reservationInfo
}

// WebSocket connection upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// NewWebSocketManager creates a new WebSocketManager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		clients:             make(map[string]*Client),
		reservationAdded:    make(chan models.Reservation, 1000),
		ReservationUpdates:  make(chan models.Reservation, 1000),
		parkingSpaceUpdates: make(chan models.ParkingEvent, 1000),
		notificationUpdates: make(chan models.NotificationEvent, 100),
		gateEvents:          make(chan models.GateEvent, 100),
		register:            make(chan *Client),
		unregister:          make(chan *Client),
		shutdown:            make(chan struct{}),
		interestClients:     make(map[string]map[string]*models.ReservationInfo),
	}
}

// Start begins the WebSocketManager's main loop
func (manager *WebSocketManager) Start() {
	log.Println("WebSocketManager starting...")
	for {
		select {
		case <-manager.shutdown:
			return
		case client := <-manager.register:
			manager.registerClient(client)
		case client := <-manager.unregister:
			manager.unregisterClient(client)
		case update := <-manager.parkingSpaceUpdates:
			manager.parkingUpdate(update)
		case added := <-manager.reservationAdded:
			manager.reservationAdd(added)
		case updated := <-manager.ReservationUpdates:
			manager.reservationUpdate(updated)
		case notification := <-manager.notificationUpdates:
			manager.handleNotificationUpdate(notification)
		case event := <-manager.gateEvents:
			manager.gateEvent(event)
		}
	}
}

// Stop gracefully shuts down the WebSocketManager
func (manager *WebSocketManager) Stop() {
	close(manager.shutdown)
}

// HandleWebSocket upgrades HTTP connections to WebSocket
func (manager *WebSocketManager) HandleWebSocket(c *gin.Context) {
	userID := c.GetString("userId")
	if userID == "" {
		http.Error(c.Writer, "User Not Authenticated", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	clientID := uuid.New().String()
	client := &Client{
		ID:           clientID,
		Conn:         conn,
		Send:         make(chan []byte, 256),
		UserID:       userID,
		ParkingSpace: make(map[string]models.ParkingEvent),
	}

	// Register new client
	manager.register <- client

	// Start client routines
	go client.readPump(manager)
	go client.writePump()
}

// registerClient adds a new client to the manager
func (manager *WebSocketManager) registerClient(client *Client) {
	manager.clientsMutex.Lock()
	manager.clients[client.ID] = client
	manager.clientsMutex.Unlock()

	log.Printf("Client registered: %s (User: %s)", client.ID, client.UserID)
}

// unregisterClient removes a client from the manager
func (manager *WebSocketManager) unregisterClient(client *Client) {
	manager.clientsMutex.Lock()
	if _, ok := manager.clients[client.ID]; ok {
		delete(manager.clients, client.ID)
		close(client.Send)
	}
	manager.clientsMutex.Unlock()

	log.Printf("Client unregistered: %s (User: %s)", client.ID, client.UserID)
}

// broadcastToUser sends a message to all connections of a specific user
func (manager *WebSocketManager) broadcastToUser(userID string, message []byte) {
	manager.clientsMutex.RLock()
	defer manager.clientsMutex.RUnlock()

	for _, client := range manager.clients {
		if client.UserID == userID {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(manager.clients, client.ID)
			}
		}
	}
}

// broadcastToAll sends a message to all connected clients
func (manager *WebSocketManager) BroadcastToAll(message []byte) {
	manager.clientsMutex.RLock()
	defer manager.clientsMutex.RUnlock()

	for _, client := range manager.clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(manager.clients, client.ID)
		}
	}
}

// handleParkingUpdate processes parking status updates
func (manager *WebSocketManager) HandleParkingUpdate(update models.ParkingEvent) {
	select {
	case manager.parkingSpaceUpdates <- update:
		log.Println("Sent parking update")
	default:
		log.Println("Parking update channel is full!")
	}
}

// handleParkingUpdate processes parking status updates
func (manager *WebSocketManager) parkingUpdate(update models.ParkingEvent) {
	message := WebSocketMessage{
		Type:    "PARKING_UPDATE",
		Payload: update,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling parking update: %v", err)
		return
	}

	// Broadcast to all clients
	manager.BroadcastToAll(data)
}

// handleParkingUpdate processes parking status updates
func (manager *WebSocketManager) HandleReservationAdded(add models.Reservation) {
	select {
	case manager.reservationAdded <- add:
		log.Println("Sent reservation add")
	default:
		log.Println("Reservation add channel is full!")
	}
}

// handleParkingUpdate processes parking status updates
func (manager *WebSocketManager) reservationAdd(update models.Reservation) {
	message := WebSocketMessage{
		Type:    "RESERVATION_ADD",
		Payload: update,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling reservation add: %v", err)
		return
	}

	// Broadcast to all clients
	manager.broadcastToUser(update.UserID, data)
}

// handleParkingUpdate processes parking status updates
func (manager *WebSocketManager) HandleReservationUpdated(update models.Reservation) {
	select {
	case manager.ReservationUpdates <- update:
		log.Println("Sent reservation update")
	default:
		log.Println("Reservation update channel is full!")
	}
}

// handleParkingUpdate processes parking status updates
func (manager *WebSocketManager) reservationUpdate(update models.Reservation) {
	message := WebSocketMessage{
		Type:    "RESERVATION_UPDATE",
		Payload: update,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling reservation update: %v", err)
		return
	}

	// Broadcast to all clients
	manager.broadcastToUser(update.UserID, data)
}

// handleParkingUpdate processes parking status updates
func (manager *WebSocketManager) HandleNotificationUpdate(update models.NotificationEvent) {
	select {
	case manager.notificationUpdates <- update:
		log.Println("Sent notification update:")
	default:
		log.Println("Notification update channel is full!")
	}
}

// handleNotificationUpdate sends notifications to users
func (manager *WebSocketManager) handleNotificationUpdate(notification models.NotificationEvent) {
	message := WebSocketMessage{
		Type:    "NOTIFICATION_UPDATE",
		Payload: notification,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling notification: %v", err)
		return
	}

	if (notification.Type == "time_limit" || notification.Type == "expired") && notification.UserID != "" {
		manager.broadcastToUser(notification.UserID, data)
	} else {
		manager.BroadcastToAll(data)
	}
}

// handleGateEvent processes gate events
func (manager *WebSocketManager) HandleGateEvent(event models.GateEvent) {
	manager.gateEvents <- event
}

// handleGateEvent processes gate events
func (manager *WebSocketManager) gateEvent(event models.GateEvent) {
	message := WebSocketMessage{
		Type:    "GATE_EVENT",
		Payload: event,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling gate event: %v", err)
		return
	}

	// Broadcast to all clients
	manager.BroadcastToAll(data)
}

// readPump pumps messages from the WebSocket connection to the manager
func (client *Client) readPump(manager *WebSocketManager) {
	defer func() {
		manager.unregister <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadLimit(512)
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading WebSocket message: %v", err)
			}
			break
		}

		log.Printf("Received message from client %s: %s", client.ID, string(message))

		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		log.Printf("Received message from client 2 %s: %s", client.ID, string(wsMsg.Type))

		// switch wsMsg.Type {
		// case "VEHICLE_UPDATE":
		// 	var update models.VehicleEvent
		// 	payloadBytes, _ := json.Marshal(wsMsg.Payload)
		// 	if err := json.Unmarshal(payloadBytes, &update); err != nil {
		// 		log.Printf("Error unmarshaling VEHICLE_UPDATE payload: %v", err)
		// 		continue
		// 	}

		// 	log.Printf("Received VEHICLE_UPDATE from client %s: %v", client.ID, update)
		// 	manager.vehicleUpdates <- update
		// default:
		// 	log.Printf("Unknown message type: %s", wsMsg.Type)
		// }
	}

}

// writePump pumps messages from the hub to the WebSocket connection
func (client *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The manager closed the channel
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func mapToStruct(input interface{}, output interface{}) error {
	tmp, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(tmp, output)
}
