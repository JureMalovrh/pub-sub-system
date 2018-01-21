package socket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

//Message definition
type Message struct {
	AccountID string `json:"accountId"`
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

//Client interface definition
type Client interface {
	SendMessage(accountID string, data string) (bool, error)
}

//ClientSender definition
type ClientSender struct {
	Connection *websocket.Conn
}

//NewSocketSender returns new ClientSender object
func NewSocketSender(connection *websocket.Conn) Client {
	return &ClientSender{
		Connection: connection,
	}
}

//SendMessage sends a message to a socket
func (s *ClientSender) SendMessage(accountID string, data string) (bool, error) {
	message := Message{
		AccountID: accountID,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	messageToSend, err := json.Marshal(message)
	if err != nil {
		return false, err
	}
	log.Printf("Sending message to socket, %s", messageToSend)

	err = s.Connection.WriteMessage(websocket.TextMessage, messageToSend)
	if err != nil {
		return false, err
	}
	return true, nil
}
