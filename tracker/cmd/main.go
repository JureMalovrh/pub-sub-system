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
	"github.com/globalsign/mgo/bson"
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

	type Person struct {
		ID    bson.ObjectId `bson:"_id,omitempty"`
		Name  string
		Phone string
	}
	/*
					c := session.DB("davidim").C("people")
					err := c.Insert(&Person{Name: "Ale", Phone: "+55 53 8116 9639"},
						&Person{Name: "Cla", Phone: "+55 53 8402 8510"})
					if err != nil {
						log.Fatal(err)
					}

					result := Person{}
					err = c.Find(bson.M{"name": "Ale"}).One(&result)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Println("Phone:", result)
					result = Person{}
					err = c.Find(bson.M{"_id": bson.ObjectIdHex("5a638e99ed12aa438f5fef20")}).One(&result)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Println("Phone:", result)
					fmt.Println("Phone:", result.Phone)
					err = c.FindId(bson.ObjectIdHex("5a638e99ed12aa438f5fef20")).One(&result)
					if err != nil {
						log.Fatal(err)
					}

					fmt.Println("Phone:", result)
					fmt.Println("Phone:", result.Phone)


						type Person struct {
							Name  string
							Phone string
						}

							c := session.DB("davidim").C("people")
							err = c.Insert(&Person{"Ale", "+55 53 8116 9639"},
								&Person{"Cla", "+55 53 8402 8510"})
							if err != nil {
								log.Fatal(err)
							}

							result := Person{}
							err = c.Find(bson.M{"name": "Ale"}).One(&result)
							if err != nil {
								log.Fatal(err)
							}

							fmt.Println("Phone:", result.Phone)

				r := mux.NewRouter()
				r.HandleFunc("/", testHandler).Methods("POST")
				r.HandleFunc("/user", testHandler).Methods("POST")
				r.HandleFunc("/articles", testHandler)

				server := http.Server{
					Addr:    ":8080",
					Handler: r,
				}

			log.Println("Serving on", server.Addr)

		errS := server.ListenAndServe(userDatabase)
		if errS != nil {
			log.Fatal(errS)
		}
	*/
}
