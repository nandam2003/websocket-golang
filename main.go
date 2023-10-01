package main

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/websocket"
)

type Server struct {
	conn map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		conn: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleConn(ws *websocket.Conn) {
	s.conn[ws] = true
	s.readMsgs(ws)

}

func (s *Server) readMsgs(ws *websocket.Conn) {
	buff := make([]byte, 1024)
	for {
		n, err := ws.Read(buff)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Error in reading message:", err)
			continue
		}
		msg := buff[:n]
		fmt.Println(string(msg))
		ws.Write([]byte("Pong"))
	}
}

func main() {
	server := NewServer()
	http.Handle("/ws", websocket.Handler(server.handleConn))

	http.ListenAndServe(":3000", nil)

}
