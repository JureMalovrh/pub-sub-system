package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"
)

//Message struct definition
type Message struct {
	AccountID string `json:"accountId"`
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

func messageReceiverHandler(messages chan []byte, close chan bool, messageReceiver Receiver) {
	for {
		if messageReceiver.IsClosed() {
			close <- true
			return
		}

		message := messageReceiver.ReadMessage()
		if len(message) == 0 {
			continue
		}

		messages <- message
	}
}

func messageParserHandler(messages chan []byte, close chan bool, parsedMessages chan Message) {
	for {
		select {
		case msg := <-messages:
			messageObject := Message{}
			err := json.Unmarshal(msg, &messageObject)
			if err != nil {
				continue
			}
			parsedMessages <- messageObject
		case <-close:
			return
		}
	}
}

func messageFilterHandler(parsedMessages chan Message, filteredMessages chan Message, close chan bool, idFilter string) {
	for {
		select {
		case msg := <-parsedMessages:
			if idFilter != "" {
				if msg.AccountID == idFilter {
					filteredMessages <- msg
				}
				continue
			}
			filteredMessages <- msg
		case <-close:
			return
		}
	}
}

func multiplexerHandler(filteredMessages chan Message, aggregatedMessages chan Message, printedMessages chan Message, close chan bool, aggregateMessages bool) {
	for {
		select {
		case msg := <-filteredMessages:
			if aggregateMessages {
				aggregatedMessages <- msg
				continue
			}
			printedMessages <- msg
		case <-close:
			return
		}

	}
}

func messagePrinterHandler(printedMessages chan Message, interrupt chan os.Signal, done chan bool, close chan bool, messageReceiver Receiver) {
	for {
		select {
		case msg := <-printedMessages:
			log.Printf("Received a data from active account id %s: data: %s, time: %d", msg.AccountID, msg.Data, msg.Timestamp)
		case <-interrupt:
			messageReceiver.Close()
			messageReceiver.CloseMessage()
			done <- true
			return
		case <-close:
			return
		}

	}
}

func messageAggregatorHandler(aggregatedMessages chan Message, aggregateFrequency int, interrupt chan os.Signal, done chan bool, close chan bool, messageReceiver Receiver) {
	ticker := time.NewTicker(time.Duration(aggregateFrequency) * time.Second)
	defer ticker.Stop()

	aggregateCounter := map[string]int{}
	for {
		select {
		case msg := <-aggregatedMessages:
			aggregateCounter[msg.AccountID] = aggregateCounter[msg.AccountID] + 1
		case <-ticker.C:
			log.Print("Aggregated messages received for accounts\n")
			for key, val := range aggregateCounter {
				log.Printf("ID: %s, number of messages %d", key, val)
			}
		case <-interrupt:
			//log.Println("interrupt")
			messageReceiver.CloseMessage()

			select {
			case <-time.After(time.Second):
			}
			messageReceiver.Close()
			done <- true
			return
		case <-close:
			return
		}
	}
}

func createMessageHandler(messageReceiver Receiver, filter string, aggregate bool, aggregateFrequency int, interrupt chan os.Signal, done chan bool) {
	messages := make(chan []byte, 5)
	close := make(chan bool, 1)
	parsedMessages := make(chan Message, 5)
	filteredMessages := make(chan Message, 5)
	aggregatedMessages := make(chan Message, 5)
	printedMessages := make(chan Message, 5)

	go messageReceiverHandler(messages, close, messageReceiver)
	go messageParserHandler(messages, close, parsedMessages)
	go messageFilterHandler(parsedMessages, filteredMessages, close, filter)
	go multiplexerHandler(filteredMessages, aggregatedMessages, printedMessages, close, aggregate)

	if aggregate {
		go messageAggregatorHandler(aggregatedMessages, aggregateFrequency, interrupt, done, close, messageReceiver)
	} else {
		go messagePrinterHandler(printedMessages, interrupt, done, close, messageReceiver)
	}
}

func main() {
	var (
		addr               = flag.String("addr", "0.0.0.0:8000", "http service address")
		filter             = flag.String("filter", "", "AccountID to filter data")
		aggregate          = flag.Bool("agg", false, "Print messages of aggregated amount of messages")
		aggregateFrequency = flag.Int("aggfreq", 3, "Only if agg=true, set time for updation of screen for aggregated data")
	)
	flag.Parse()

	interrupt := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	log.Printf("connecting to %s", *addr)
	messageReceiver := NewMessageReceiver(*addr)
	messageReceiver.Connect()
	defer messageReceiver.Close()

	createMessageHandler(messageReceiver, *filter, *aggregate, *aggregateFrequency, interrupt, done)

	for {
		select {
		case <-done:
			os.Exit(0)
		}
	}
}
