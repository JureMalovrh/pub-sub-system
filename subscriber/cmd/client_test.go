package main

import (
	"bufio"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

func setLoggerToFile(fileName string) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file %s, %s", fileName, err)
	}
	log.SetFlags(0)
	log.SetOutput(file)
}

func Test_messageReceiverHandler(t *testing.T) {
	sendMessageString := `{"accountId": "test", "data": "data", "timestamp": 1}`
	receiveMessage := `{"accountId": "test", "data": "data", "timestamp": 1}`
	t.Run("Should receive and resend the data he gets", func(t *testing.T) {
		mr := connectWS()
		defer closeWS(mr)
		messages := make(chan []byte)
		close := make(chan bool)
		go messageReceiverHandler(messages, close, mr)

		//firt message should be successfully connected one
		successMsg := <-messages
		if string(successMsg) != ConnectedMessage {
			t.Errorf("Expected successfully connected message: %s, got %s", ConnectedMessage, successMsg)
		}

		go sendMessage(sendMessageString)

		msg := <-messages
		if string(msg) != receiveMessage {
			t.Errorf("Expected %s, got %s", receiveMessage, msg)
		}
	})
}
func Test_messageParserHandler(t *testing.T) {
	sendMessageString := `{"accountId": "test", "data": "data", "timestamp": 1}`
	testCases := []struct {
		desc           string
		sendMessages   []string
		expectedObject Message
	}{
		{
			desc:           "Should parse correct JSON data",
			sendMessages:   []string{sendMessageString},
			expectedObject: Message{"test", "data", 1},
		},
		{
			desc:           "Should skip incorrect JSON data",
			sendMessages:   []string{"wrong", "wrong2", "wrong3", sendMessageString},
			expectedObject: Message{"test", "data", 1},
		},
	}
	for _, tC := range testCases {
		messages := make(chan []byte)
		close := make(chan bool)
		parsedData := make(chan Message)
		go messageParserHandler(messages, close, parsedData)

		for _, msg := range tC.sendMessages {
			messages <- []byte(msg)
		}

		parsedObject := <-parsedData

		eq := reflect.DeepEqual(parsedObject, tC.expectedObject)
		if !eq {
			t.Errorf("Expected %v, got %v", tC.expectedObject, parsedObject)
		}
		close <- true
	}
}

func Test_messageFilterHandler(t *testing.T) {
	testCases := []struct {
		desc           string
		filter         string
		sendMessages   []Message
		expectedObject Message
	}{
		{
			desc:           "Should send every data if filter is empty",
			sendMessages:   []Message{Message{"test", "data", 1}},
			expectedObject: Message{"test", "data", 1},
		},
		{
			desc:           "Should send only data with correct id",
			sendMessages:   []Message{Message{"test", "data", 1}, Message{"test2", "data", 1}, Message{"test3", "data", 1}},
			filter:         "test3",
			expectedObject: Message{"test3", "data", 1},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			parsedData := make(chan Message)
			filteredData := make(chan Message)
			close := make(chan bool)
			go messageFilterHandler(parsedData, filteredData, close, tC.filter)

			for _, msg := range tC.sendMessages {
				parsedData <- msg
			}

			filteredObject := <-filteredData

			eq := reflect.DeepEqual(filteredObject, tC.expectedObject)
			if !eq {
				t.Errorf("Expected %v, got %v", tC.expectedObject, filteredObject)
			}
		})
	}
}

func Test_multiplexerHandler(t *testing.T) {
	testCases := []struct {
		desc           string
		sendMessage    Message
		expectedObject Message
		isAggregator   bool
	}{
		{
			desc:           "Should send data to printed data channel",
			sendMessage:    Message{"test", "data", 1},
			expectedObject: Message{"test", "data", 1},
		},
		{
			desc:           "Should send data to aggregated data channel",
			sendMessage:    Message{"test", "data", 1},
			isAggregator:   true,
			expectedObject: Message{"test", "data", 1},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			filteredData := make(chan Message)
			printedData := make(chan Message)
			aggregatedData := make(chan Message)
			close := make(chan bool)
			go multiplexerHandler(filteredData, aggregatedData, printedData, close, tC.isAggregator)

			filteredData <- tC.sendMessage

			if tC.isAggregator {
				aggregatedObject := <-aggregatedData
				eq := reflect.DeepEqual(aggregatedObject, tC.expectedObject)
				if !eq {
					t.Errorf("Expected %v, got %v", tC.expectedObject, aggregatedObject)
				}
			} else {
				printedData := <-printedData
				eq := reflect.DeepEqual(printedData, tC.expectedObject)
				if !eq {
					t.Errorf("Expected %v, got %v", tC.expectedObject, printedData)
				}
			}

		})
	}
}

func Test_messagePrinterHandler(t *testing.T) {
	testCases := []struct {
		desc         string
		fileName     string
		sendMessage  Message
		expected     string
		expectedLogs int
	}{
		{
			desc:         "Should print received data",
			fileName:     "test2.txt",
			sendMessage:  Message{"test", "data", 1},
			expected:     "Received a data from active account id test: data: data, time: 1",
			expectedLogs: 1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			setLoggerToFile(tC.fileName)

			printedData := make(chan Message)
			close := make(chan bool)

			go messagePrinterHandler(printedData, nil, nil, close, nil)
			printedData <- tC.sendMessage

			//wait for aggregator to log something
			time.Sleep(500 * time.Millisecond)

			fileWritten, err := os.OpenFile(tC.fileName, os.O_RDONLY, 0666)
			if err != nil {
				t.Fatal(err)
			}
			scanner := bufio.NewScanner(fileWritten)
			counter := 0
			for scanner.Scan() {
				counter += 1
				readLine := scanner.Text()
				if readLine != tC.expected {
					t.Errorf("Wanted logged line %s  got: %s", tC.expected, readLine)
				}
			}
			if counter != tC.expectedLogs {
				t.Errorf("Expected number of logs: %d, got %d", tC.expectedLogs, counter)
			}
			os.Remove(tC.fileName)

		})
	}
}

func Test_messageAggregatorHandler(t *testing.T) {
	aggregateHeader := "Aggregated messages received for accounts"
	testCases := []struct {
		desc        string
		fileName    string
		sendMessage Message
		expected    string
	}{
		{
			desc:        "Should print aggregated data",
			fileName:    "test1.txt",
			sendMessage: Message{"test", "data", 1},
			expected:    "ID: test, number of messages 1",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			setLoggerToFile(tC.fileName)

			aggregatedData := make(chan Message)
			close := make(chan bool)

			go messageAggregatorHandler(aggregatedData, 1, nil, nil, close, nil)
			aggregatedData <- tC.sendMessage
			//wait for aggregator to log something
			time.Sleep(2 * time.Second)
			close <- true
			fileWritten, err := os.OpenFile(tC.fileName, os.O_RDONLY, 0666)
			if err != nil {
				t.Fatal(err)
			}
			scanner := bufio.NewScanner(fileWritten)
			for scanner.Scan() {
				readLine := scanner.Text()
				if readLine != aggregateHeader && readLine != tC.expected {
					t.Errorf("Wanted logged line %s or %s, got: %s", aggregateHeader, tC.expected, readLine)
				}
			}
			os.Remove(tC.fileName)
		})
	}
}

func Test_createMessageHandler_printer_filtering(t *testing.T) {
	aggregateHeader := "Aggregated messages received for accounts"
	testCases := []struct {
		desc               string
		fileName           string
		filter             string
		aggregate          bool
		aggregateFrequency int
		sendMessage        string
		sendMessageFalse   string
		expected           string
		count              int
	}{
		{
			desc:             "Should print received data, with filtering",
			fileName:         "test3b.txt",
			filter:           "test",
			sendMessage:      `{"accountId": "test", "data": "data", "timestamp": 1}`,
			sendMessageFalse: `{"accountId": "test1", "data": "data", "timestamp": 1}`,
			expected:         "Received a data from active account id test: data: data, time: 1",
			count:            1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mr := connectWS()
			setLoggerToFile(tC.fileName)

			interrupt := make(chan os.Signal, 1)
			createMessageHandler(mr, tC.filter, tC.aggregate, tC.aggregateFrequency, interrupt, nil)

			go sendMessage(tC.sendMessageFalse)
			go sendMessage(tC.sendMessageFalse)
			go sendMessage(tC.sendMessage)

			//wait for aggregator to log something
			time.Sleep(500 * time.Millisecond)
			closeWS(mr)

			fileWritten, err := os.OpenFile(tC.fileName, os.O_RDONLY, 0666)
			if err != nil {
				t.Fatal(err)
			}
			scanner := bufio.NewScanner(fileWritten)
			counter := 0
			for scanner.Scan() {
				counter++
				readLine := scanner.Text()
				if readLine != aggregateHeader && readLine != tC.expected {
					t.Errorf("Wanted logged line %s or %s, got: %s", aggregateHeader, tC.expected, readLine)
				}
			}
			if counter != tC.count {
				t.Errorf("Wanted %d messages, got %d", tC.count, counter)
			}
			os.Remove(tC.fileName)
		})
	}
}
func Test_createMessageHandler_printer(t *testing.T) {
	aggregateHeader := "Aggregated messages received for accounts"
	testCases := []struct {
		desc               string
		fileName           string
		filter             string
		aggregate          bool
		aggregateFrequency int
		sendMessage        string
		sendMessageFalse   string
		expected           string
		count              int
	}{
		{
			desc:        "Should print received data",
			fileName:    "test3a.txt",
			sendMessage: `{"accountId": "test", "data": "data", "timestamp": 1}`,
			expected:    "Received a data from active account id test: data: data, time: 1",
			count:       1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mr := connectWS()
			setLoggerToFile(tC.fileName)
			interrupt := make(chan os.Signal, 1)
			createMessageHandler(mr, tC.filter, tC.aggregate, tC.aggregateFrequency, interrupt, nil)

			go sendMessage(tC.sendMessageFalse)
			go sendMessage(tC.sendMessageFalse)
			go sendMessage(tC.sendMessage)
			//wait to log something
			time.Sleep(500 * time.Millisecond)
			closeWS(mr)

			fileWritten, err := os.OpenFile(tC.fileName, os.O_RDONLY, 0666)
			if err != nil {
				t.Fatal(err)
			}
			scanner := bufio.NewScanner(fileWritten)
			counter := 0
			for scanner.Scan() {
				counter++
				readLine := scanner.Text()
				if readLine != tC.expected {
					t.Errorf("Wanted logged line %s or %s, got: %s", aggregateHeader, tC.expected, readLine)
				}
			}
			if counter != tC.count {
				t.Errorf("Wanted %d messages, got %d", tC.count, counter)
			}

			os.Remove(tC.fileName)
		})
	}
}

func Test_createMessageHandler_aggregator_filtering(t *testing.T) {
	aggregateHeader := "Aggregated messages received for accounts"
	testCases := []struct {
		desc               string
		fileName           string
		filter             string
		aggregate          bool
		aggregateFrequency int
		sendMessage        string
		sendMessageFalse   string
		expected           string
		count              int
	}{
		{
			desc:               "Should print received data, with filtering",
			fileName:           "test4b.txt",
			filter:             "test",
			sendMessage:        `{"accountId": "test", "data": "data", "timestamp": 1}`,
			sendMessageFalse:   `{"accountId": "test1", "data": "data", "timestamp": 1}`,
			expected:           "ID: test, number of messages 1",
			count:              2,
			aggregate:          true,
			aggregateFrequency: 1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mr := connectWS()
			setLoggerToFile(tC.fileName)

			interrupt := make(chan os.Signal, 1)
			createMessageHandler(mr, tC.filter, tC.aggregate, tC.aggregateFrequency, interrupt, nil)

			go sendMessage(tC.sendMessageFalse)
			go sendMessage(tC.sendMessageFalse)
			go sendMessage(tC.sendMessage)
			time.Sleep(500 * time.Millisecond)
			closeWS(mr)
			//wait for aggregator to log something
			time.Sleep(1000 * time.Millisecond)
			interrupt <- os.Kill

			fileWritten, err := os.OpenFile(tC.fileName, os.O_RDONLY, 0666)
			if err != nil {
				t.Fatal(err)
			}

			scanner := bufio.NewScanner(fileWritten)
			counter := 0
			aggHeadCounter := 0
			for scanner.Scan() {
				counter++
				readLine := scanner.Text()
				if readLine == aggregateHeader {
					aggHeadCounter++
				}
				if readLine != aggregateHeader && readLine != tC.expected {
					t.Errorf("Wanted logged line %s or %s, got: %s", aggregateHeader, tC.expected, readLine)
				}
			}
			counter -= (aggHeadCounter - 1)
			if counter != tC.count {
				t.Errorf("Wanted %d messages, got %d", tC.count, counter)
			}
			os.Remove(tC.fileName)
		})
	}
}
func Test_createMessageHandler_aggregator(t *testing.T) {
	aggregateHeader := "Aggregated messages received for accounts"
	testCases := []struct {
		desc               string
		fileName           string
		filter             string
		aggregate          bool
		aggregateFrequency int
		sendMessage        string
		sendMessage2       string
		sendMessageFalse   string
		expected           string
		expected2          string
		count              int
	}{
		{
			desc:               "Should print received data",
			fileName:           "test4b.txt",
			sendMessage:        `{"accountId": "test", "data": "data", "timestamp": 1}`,
			sendMessage2:       `{"accountId": "test1", "data": "data", "timestamp": 1}`,
			expected:           "ID: test, number of messages 1",
			expected2:          "ID: test1, number of messages 1",
			count:              3,
			aggregate:          true,
			aggregateFrequency: 1,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mr := connectWS()
			setLoggerToFile(tC.fileName)
			interrupt := make(chan os.Signal, 1)
			createMessageHandler(mr, tC.filter, tC.aggregate, tC.aggregateFrequency, interrupt, nil)

			go sendMessage(tC.sendMessage2)
			go sendMessage(tC.sendMessage)
			time.Sleep(500 * time.Millisecond)
			closeWS(mr)
			//wait for aggregator to log something
			time.Sleep(1000 * time.Millisecond)
			interrupt <- os.Kill
			fileWritten, err := os.OpenFile(tC.fileName, os.O_RDONLY, 0666)
			if err != nil {
				t.Fatal(err)
			}
			scanner := bufio.NewScanner(fileWritten)
			counter := 0
			// because of undeterministic data, we only substract multiple aggregate header lines
			aggHeader := 0
			for scanner.Scan() {
				counter++
				readLine := scanner.Text()
				if readLine == aggregateHeader {
					aggHeader++
				}
				if readLine != aggregateHeader && readLine != tC.expected && readLine != tC.expected2 {
					t.Errorf("Wanted logged line %s or %s or %s, got: %s", aggregateHeader, tC.expected, tC.expected2, readLine)
				}
			}
			counter -= (aggHeader - 1)
			if counter != tC.count {
				t.Errorf("Wanted %d messages, got %d", tC.count, counter)
			}
			os.Remove(tC.fileName)
		})
	}
}
