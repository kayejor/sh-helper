package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"os"

	"github.com/gorilla/websocket"
)

type Player struct {
	ws   *websocket.Conn
	name string
	team int
}

var players []Player
var playersMux sync.Mutex

func main() {
	port := os.Getenv("PORT")
	fs := http.FileServer(http.Dir("public"))
	http.HandleFunc("/websocket", handleConnections)
	http.Handle("/", fs)
	rand.Seed(time.Now().UnixNano())
	http.ListenAndServe(":" + port, nil)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type PlayerRole struct {
	Name string `json:"name"`
	Role int    `json:"role"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	var msg PlayerRole
	err = conn.ReadJSON(&msg)
	if err != nil {
		log.Fatal("problem with message")
		log.Fatal(msg)
		return
	}
	player := Player{
		ws:   conn,
		name: string(msg.Name),
		team: msg.Role,
	}
	players = append(players, player)
	broadcastNames()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			removeConnection(conn)
			broadcastNames()
		}
		if string(msg) == "start" {
			presentInformationToAllPlayers()
		}
	}
}

func broadcastNames() {
	playersMux.Lock()
	var namesHTML = createHTMLListFromNames()
	for _, player := range players {
		player.ws.WriteMessage(websocket.TextMessage, []byte(namesHTML))
	}
	playersMux.Unlock()
}

func removeConnection(conn *websocket.Conn) {
	playersMux.Lock()
	for index := 0; index < len(players); index++ {
		if conn == players[index].ws {
			players = append(players[:index], players[index+1:]...)
			break
		}
	}
	playersMux.Unlock()
}

func createHTMLListFromNames() string {
	var result = ""
	for _, player := range players {
		result += "<li>" + player.name + "</li>"
	}
	return result
}

func createAllKnowingList() string {
	var result = ""
	for _, player := range players {
		result += "<li class=" + getClass(player.team) + ">" + player.name + "</li>"
	}
	return result
}

func getClass(i int) string {
	if i == 0 {
		return "liberal"
	} else if i == 1 {
		return "fascist"
	} else {
		return "hitler"
	}
}

func presentInformationToAllPlayers() {
	playersMux.Lock()
	var namesHTML = createHTMLListFromNames()
	var allKnowingNamesHTML = createAllKnowingList()
	for _, player := range players {
		if player.team == 1 || player.team == 2 && len(players) < 7 {
			player.ws.WriteMessage(websocket.TextMessage, []byte(allKnowingNamesHTML))
		} else {
			player.ws.WriteMessage(websocket.TextMessage, []byte(namesHTML))
		}
	}
	playersMux.Unlock()
}
