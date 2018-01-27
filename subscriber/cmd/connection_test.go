package main

import (
	"log"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

//ConnectedMessage connection confirmation
const ConnectedMessage = "Successfully connected to publisher"

func connectWS() Receiver {
	mr := NewMessageReceiver("localhost:8000")
	mr.Connect()
	return mr
}

func closeWS(mr Receiver) {
	mr.Close()
}

// connect to WS server and send message
func sendMessage(msg string) {
	u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/"}

	var c *websocket.Conn
	for {
		var err error
		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err == nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	err := c.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		log.Println("write:", err)
		return
	}
	c.Close()
	return
}

func TestReadMessage(t *testing.T) {
	testCases := []struct {
		desc            string
		sendMessage     string
		receivedMessage string
	}{
		{
			desc:            "Should receive message",
			sendMessage:     "test",
			receivedMessage: "test",
		},
	}
	mr := connectWS()
	defer closeWS(mr)

	successfullyConnectedMessage := mr.ReadMessage()
	if string(successfullyConnectedMessage) != ConnectedMessage {
		t.Errorf("Expected successfully connected message: %s, got %s", ConnectedMessage, successfullyConnectedMessage)
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			go sendMessage(tC.sendMessage)
			msg := mr.ReadMessage()
			if string(msg) != tC.receivedMessage {
				t.Errorf("Expected %s, got %s", tC.receivedMessage, msg)
			}
		})
	}
}
