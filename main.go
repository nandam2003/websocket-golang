package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketRegistry struct {
	Conn        map[int]*websocket.Conn
	RegistryLoc sync.RWMutex
}

func NewWebSocketRegistry() *WebSocketRegistry {
	return &WebSocketRegistry{
		Conn: make(map[int]*websocket.Conn),
	}
}

func (r *WebSocketRegistry) Register(driverID int, conn *websocket.Conn) {
	r.RegistryLoc.Lock()
	defer r.RegistryLoc.Unlock()
	r.Conn[driverID] = conn
}

func (r *WebSocketRegistry) Unregister(driverID int) {
	r.RegistryLoc.Lock()
	defer r.RegistryLoc.Unlock()
	delete(r.Conn, driverID)
}

func (r *WebSocketRegistry) Broadcast(message []byte) {
	r.RegistryLoc.RLock()
	defer r.RegistryLoc.RUnlock()
	for _, conn := range r.Conn {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error broadcasting message: %v", err)
		}
	}

}

var (
	registry = NewWebSocketRegistry()
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins
			return true
		},
	}
)

func handleDriverConnections(c *gin.Context) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	driverID := rand.Intn(101)

	registry.Register(driverID, conn)

	defer func() {
		registry.Unregister(driverID)
	}()

	for {

	}
}

func requestRaid(c *gin.Context) {
	var request struct {
		PickUp      string `json:"pick_up"`
		Destination string `json:"destination"`
		TotalDist   string `json:"total_dist"`
	}
	if err := c.BindJSON(&request); err != nil {
		print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	json, err := json.Marshal(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to convert the data into json"})
		return
	}
	registry.Broadcast(json)

	c.JSON(http.StatusOK, gin.H{"message": "Resquest made"})

}

func main() {

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}  // Specify the allowed HTTP methods
	config.AllowHeaders = []string{"Content-Type", "Authorization"} // Specify the allowed headers

	// Use the CORS middleware with the configured options
	r.Use(cors.New(config))

	r.POST("/raid", requestRaid)
	r.GET("/ws", handleDriverConnections)

	go func() {
		if err := r.Run(); err != nil {
			fmt.Println("API server error:", err)
		}
	}()

	select {}

}
