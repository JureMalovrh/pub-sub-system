package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"
)

type Message struct {
	AccountID string `json:"accountId"`
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
}

func messageReceiverHandler(messages chan []byte, messageReceiver Receiver) {
	for {
		message := messageReceiver.ReadMessage()
		messages <- message
	}
}

func messageParserHandler(messages chan []byte, parsedMessages chan Message) {
	for {
		msg := <-messages
		messageObject := Message{}
		err := json.Unmarshal(msg, &messageObject)
		if err != nil {
			continue
		}
		parsedMessages <- messageObject
	}
}

func messageFilterHandler(parsedMessages chan Message, filteredMessages chan Message, idFilter string) {
	for {
		msg := <-parsedMessages
		if idFilter != "" {
			if msg.AccountID == idFilter {
				filteredMessages <- msg
			}
			continue
		}
		filteredMessages <- msg
	}
}

func multiplexerHandler(filteredMessages chan Message, aggregatedMessages chan Message, printedMessages chan Message, aggregateMessages bool) {
	for {
		msg := <-filteredMessages
		if aggregateMessages {
			aggregatedMessages <- msg
			continue
		}
		printedMessages <- msg
	}
}

func messagePrinterHandler(printedMessages chan Message, interrupt chan os.Signal, messageReceiver Receiver) {
	defer messageReceiver.Close()
	for {
		select {
		case msg := <-printedMessages:
			log.Printf("Received a data from active account id %s: data: %s, time: %d", msg.AccountID, msg.Data, msg.Timestamp)
		case <-interrupt:
			log.Println("interrupt")
			err := messageReceiver.CloseMessage()
			if err != nil {
				log.Println("write close:", err)
				return
			}
			messageReceiver.Close()
			return
		}

	}
}

func messageAggregatorHandler(aggregatedMessages chan Message, aggregateFrequency int, interrupt chan os.Signal, messageReceiver Receiver) {
	ticker := time.NewTicker(time.Duration(aggregateFrequency) * time.Second)

	defer ticker.Stop()
	defer messageReceiver.Close()

	kv := map[string]int{}
	for {
		select {
		case msg := <-aggregatedMessages:
			kv[msg.AccountID] = kv[msg.AccountID] + 1

		case <-ticker.C:
			log.Print("Aggregated messages received for accounts\n")
			for key, val := range kv {
				log.Printf("ID: %s, number of messages %d", key, val)
			}

		case <-interrupt:
			log.Println("interrupt")
			err := messageReceiver.CloseMessage()
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-time.After(time.Second):
			}
			messageReceiver.Close()
			return
		}
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

	log.Printf("connecting to %s", *addr)
	messageReceiver := NewMessageReceiver(*addr)
	messageReceiver.Connect()
	defer messageReceiver.Close()

	messages := make(chan []byte)
	parsedMessages := make(chan Message)
	filteredMessages := make(chan Message)
	aggregatedMessages := make(chan Message)
	printedMessages := make(chan Message)

	go messageReceiverHandler(messages, messageReceiver)
	go messageParserHandler(messages, parsedMessages)
	go messageFilterHandler(parsedMessages, filteredMessages, *filter)
	go multiplexerHandler(filteredMessages, aggregatedMessages, printedMessages, *aggregate)

	if *aggregate {
		go messageAggregatorHandler(aggregatedMessages, *aggregateFrequency, interrupt, messageReceiver)
	} else {
		go messagePrinterHandler(printedMessages, interrupt, messageReceiver)
	}

	//keep some dummy code otherwise the process ends
	for {
	}
}
