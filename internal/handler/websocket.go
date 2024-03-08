package handler

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
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

type WsClient struct {
	connection *websocket.Conn
	wsHandler  *WsHandler
	egress     chan []byte // is used to avoid concurrent writes to the websocket connection
}

type ClientList map[*WsClient]bool

func NewWsHandler() *WsHandler {
	return &WsHandler{
		clients: make(ClientList),
	}
}

func NewWsClient(conn *websocket.Conn, wsHandler *WsHandler) *WsClient {
	return &WsClient{
		connection: conn,
		wsHandler:  wsHandler,
		egress:     make(chan []byte),
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

func (c *WsClient) ReadMessages() {
	defer func() {
		c.wsHandler.RemoveClient(c)
	}()

	for {
		messageType, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("error reading message: %v\n", err.Error())
			}
			break
		}
		c.wsHandler.Broadcast(payload, messageType)
	}
}

func (c *WsClient) WriteMessages() {
	defer func() {
		c.wsHandler.RemoveClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					fmt.Printf("error writing close message: %v\n", err.Error())
					return
				}
			}
			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				fmt.Printf("error writing message: %v\n", err.Error())
				return
			}

			logrus.Printf("message sent: %v\n", string(message))
		}
	}
}
