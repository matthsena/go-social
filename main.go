package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type ChatMessage struct {
	Username string `json:"username"`
	Text     string `json:"text"`
	Room     string  `json:"room"`
}

var clients = make(map[*websocket.Conn]string)
var broadcaster = make(chan ChatMessage)
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	room := params["room"]

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	
	clients[ws] = room

	for {
		var msg ChatMessage

		err := ws.ReadJSON(&msg)
		msg.Room = room
		
		if err != nil {
			delete(clients, ws)
			break
		}
		broadcaster <- msg
	}
}

func handleMessages() {
	for {
		messageClients(<-broadcaster)
	}
}

func messageClients(msg ChatMessage) {
	for client := range clients {
		
		room := clients[client]

		if room == msg.Room {
			err := client.WriteJSON(msg)

			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	r := mux.NewRouter()

	r.Handle("/", http.FileServer(http.Dir("./public")))
	r.HandleFunc("/ws/{room}", handleConnections)
	
	go handleMessages()



	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowCredentials: true,
	})
	
	log.Print("running http://localhost:8080")


	handler := c.Handler(r)	
	PORT := ":8080"

	log.Fatal(http.ListenAndServe(PORT, handler))
}