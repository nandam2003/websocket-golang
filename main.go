package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

type WebSocketRegistry struct {
	Conn        map[string]*websocket.Conn
	RegistryLoc sync.RWMutex
}

func NewWebSocketRegistry() *WebSocketRegistry {
	return &WebSocketRegistry{
		Conn: make(map[string]*websocket.Conn),
	}
}

func (r *WebSocketRegistry) Register(driverID string, conn *websocket.Conn) {
	r.RegistryLoc.Lock()
	defer r.RegistryLoc.Unlock()
	r.Conn[driverID] = conn
}

func (r *WebSocketRegistry) Unregister(driverID string) {
	r.RegistryLoc.Lock()
	defer r.RegistryLoc.Unlock()
	delete(r.Conn, driverID)
}

func (r *WebSocketRegistry) Broadcast(message []byte) {
	r.RegistryLoc.RLock()
	defer r.RegistryLoc.RUnlock()
	for _, conn := range r.Conn {
		_, err := conn.Write(message)
		if err != nil {
			log.Printf("Error broadcasting message: %v", err)
		}
	}

}

var (
	registry = NewWebSocketRegistry()
)

func handleDriverConnections(ws *websocket.Conn) {
	driverID := "1"

	registry.Register(driverID, ws)

	defer func() {
		registry.Unregister(driverID)
	}()

	for {

	}
}

func requestRaid(c *gin.Context) {
	var request struct {
		Flat      string `json:"f_lat"`
		Flong     string `json:"f_long"`
		Tlat      string `json:"t_lat"`
		Tlong     string `json:"t_long"`
		TotalDist string `json:"total_dist"`
	}
	if err := c.BindJSON(&request); err != nil {
		print(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}
	json, err := json.Marshal(request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
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
	http.Handle("/ws", websocket.Handler(handleDriverConnections))

	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			fmt.Println("API server error:", err)
		}
	}()

	go func() {
		if err := r.Run(":3000"); err != nil {
			fmt.Println("API server error:", err)
		}
	}()

	select {}

}
