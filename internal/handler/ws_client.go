package handler

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type WsClient struct {
	connection *websocket.Conn
	wsHandler  *WsHandler
	egress     chan []byte // is used to avoid concurrent writes to the websocket connection
}
type ClientList map[*WsClient]bool

func NewWsClient(conn *websocket.Conn, wsHandler *WsHandler) *WsClient {
	return &WsClient{
		connection: conn,
		wsHandler:  wsHandler,
		egress:     make(chan []byte),
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
