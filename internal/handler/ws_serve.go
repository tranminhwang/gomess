package handler

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var WsUpGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WsHandler struct {
	clients ClientList
	sync.RWMutex
}

func NewWsHandler() *WsHandler {
	return &WsHandler{
		clients: make(ClientList),
	}
}

func (wsh *WsHandler) WsServe(w http.ResponseWriter, r *http.Request) {
	conn, err := WsUpGrader.Upgrade(w, r, nil)
	if err != nil {
		println(err.Error())
		return
	}
	client := NewWsClient(conn, wsh)
	wsh.AddClient(client)
	// start client processing
	go client.ReadMessages()
	go client.WriteMessages()
}

func (wsh *WsHandler) AddClient(client *WsClient) {
	wsh.Lock()
	defer wsh.Unlock()

	if _, ok := wsh.clients[client]; !ok {
		wsh.clients[client] = true
	}
}

func (wsh *WsHandler) RemoveClient(client *WsClient) {
	wsh.Lock()
	defer wsh.Unlock()

	if _, ok := wsh.clients[client]; ok {
		delete(wsh.clients, client)
	}
}

func (wsh *WsHandler) Broadcast(message []byte, messageType int) {
	wsh.Lock()
	defer wsh.Unlock()

	fmt.Printf("message type: %v, payload: %v\n", messageType, string(message))

	for client := range wsh.clients {
		client.egress <- message
	}
}
