package main

import (
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//Receiver interface definition
type Receiver interface {
	Connect() error
	ReadMessage() []byte
	Close() error
	CloseMessage() error
	IsClosed() bool
}

//MessageReceiver is a receiver for messages
type MessageReceiver struct {
	Connection *websocket.Conn
	URL        string
	sync.Mutex
	Closed bool
}

//NewMessageReceiver returns new MessageReceiver
func NewMessageReceiver(address string) Receiver {
	u := url.URL{Scheme: "ws", Host: address}
	return &MessageReceiver{
		URL: u.String(),
	}
}

//Connect connects MessageReceiver to socket. It tries forever.
func (mr *MessageReceiver) Connect() error {
	var connection *websocket.Conn

	for {
		var err error
		connection, _, err = websocket.DefaultDialer.Dial(mr.URL, nil)
		if err == nil {
			break
		}
		time.Sleep(3 * time.Second)
	}
	mr.Connection = connection
	return nil
}

//ReadMessage tries to read a message from socket connection. If he fails, he tries to reconect.
func (mr *MessageReceiver) ReadMessage() []byte {
	if mr.IsClosed() {
		return nil
	}
	_, msg, err := mr.Connection.ReadMessage()
	if err != nil && mr.IsClosed() == false {
		connErr := mr.Connect()
		if connErr == nil {
			_, msg, _ := mr.Connection.ReadMessage()
			return msg
		}
		panic(err)
	}
	return msg
}

//Close closes WS connection
func (mr *MessageReceiver) Close() error {
	mr.Lock()
	mr.Closed = true
	mr.Unlock()
	return mr.Connection.Close()
}

//CloseMessage sends close message to socket
func (mr *MessageReceiver) CloseMessage() error {
	mr.Lock()
	mr.Closed = true
	mr.Unlock()
	err := mr.Connection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	return err
}

func (mr *MessageReceiver) IsClosed() bool {
	mr.Lock()
	tmp := mr.Closed
	mr.Unlock()
	return tmp
}
