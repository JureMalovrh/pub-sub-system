package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"pub-sub/tracker/database"
	"pub-sub/tracker/handler"
	"pub-sub/tracker/socket"
	"time"

	"github.com/globalsign/mgo"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"a": "a"}`))
}

func connectToDatabase(databaseConfig databaseConfig) *mgo.Session {
	databaseURL := fmt.Sprintf("%s:%s", databaseConfig.Server, databaseConfig.Port)
	session, err := mgo.Dial(databaseURL)
	if err != nil {
		panic(err)
	}
	return session
}

func publisherConnection(publisherConfig publisherConfig) *websocket.Conn {
	host := fmt.Sprintf("%s:%s", publisherConfig.URL, publisherConfig.Port)
	u := url.URL{Scheme: publisherConfig.Method, Host: host}
	log.Printf("connecting to %s", u.String())

	var c *websocket.Conn
	for {
		var err error
		c, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err == nil {
			log.Print("Successfully connected to publisher")
			break
		}
		log.Print("error", err)
		time.Sleep(3 * time.Second)
	}
	return c
}

func startServer(address string, database database.Storage, publisher socket.Client) error {
	r := mux.NewRouter()
	r.HandleFunc("/{accountId}", handler.NewAccountHandler(database, publisher)).Methods("POST")

	server := http.Server{
		Addr:    address,
		Handler: r,
	}

	log.Println("Serving on", server.Addr)
	return server.ListenAndServe()
}

func main() {
	var configPath = flag.String("config", "config.toml", "Path for config file")
	flag.Parse()

	config := LoadConfig(*configPath)
	_ = config

	session := connectToDatabase(config.Database)
	defer session.Close()

	socketConnection := publisherConnection(config.Publisher)
	defer socketConnection.Close()

	userDatabase := database.NewUserStorage(session, config.Database.Table, config.Database.Collection)
	userActionNotifier := socket.NewSocketSender(socketConnection)

	startServer(config.Address, userDatabase, userActionNotifier)
}
