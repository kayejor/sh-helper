package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	ws   *websocket.Conn
	name string
	role string
}

var players []Player
var playersMux sync.Mutex
var gameStarted bool

func main() {
	gameStarted = false
	port := os.Getenv("PORT")
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
	http.HandleFunc("/websocket", handleConnections)
	rand.Seed(time.Now().UnixNano())
	http.ListenAndServe(":"+port, nil)
	log.Println("Server started on port " + port)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	if gameStarted {
		conn.WriteMessage(websocket.TextMessage, []byte("Game started"))
		return
	}

	if len(players) == 10 {
		conn.WriteMessage(websocket.TextMessage, []byte("Game full"))
		return
	}

	//when websocket is opened, client sends us the player name
	_, playerName, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("problem with name message")
		return
	}
	player := Player{
		ws:   conn,
		name: string(playerName),
		role: "",
	}
	players = append(players, player)
	broadcastNames()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			removeConnection(conn)
			broadcastNames()
			if len(players) == 0 {
				gameStarted = false
			}
		}

		if string(msg) == "start" {
			if len(players) >= 5 {
				startGame()
			} else {
				conn.WriteMessage(websocket.TextMessage, []byte("Not enough players"))
			}
		} else {
			invIndex, _ := strconv.Atoi(string(msg))
			htmlList := createHTMLListForPlayerWithInv(player, invIndex)
			conn.WriteMessage(websocket.TextMessage, []byte(htmlList))
		}
	}
}

func startGame() {
	gameStarted = true
	assignRoles()
	broadcastNames()
}

func assignRoles() {
	playersMux.Lock()

	numberOfPlayers := len(players)
	randPerm := rand.Perm(numberOfPlayers)
	numberOfLiberals := numberOfPlayers/2 + 1
	for i := 0; i < numberOfPlayers; i++ {
		if randPerm[i] == 0 {
			players[i].role = "hitler"
		} else if randPerm[i] > numberOfLiberals {
			players[i].role = "fascist"
		} else {
			players[i].role = "liberal"
		}
	}

	playersMux.Unlock()
}

func broadcastNames() {
	playersMux.Lock()
	for _, player := range players {
		htmlList := createHTMLListForPlayer(player)
		player.ws.WriteMessage(websocket.TextMessage, []byte(htmlList))
	}
	playersMux.Unlock()
}

func removeConnection(conn *websocket.Conn) {
	playersMux.Lock()
	//len(players) has a max of 10
	for index := 0; index < len(players); index++ {
		if conn == players[index].ws {
			players = append(players[:index], players[index+1:]...)
			break
		}
	}
	playersMux.Unlock()
}

func createHTMLListForPlayer(currentPlayer Player) string {
	return createHTMLListForPlayerWithInv(currentPlayer, -1)
}

func createHTMLListForPlayerWithInv(currentPlayer Player, invIndex int) string {
	var result = ""
	seeAll := currentPlayer.role == "fascist" || currentPlayer.role == "hitler" && len(players) < 7
	for index, player := range players {
		if gameStarted && (seeAll || currentPlayer == player || index == invIndex) {
			result += createListElement(player.name, player.role, index)
		} else {
			result += createListElement(player.name, "", index)
		}
	}
	return result
}

func createListElement(value string, class string, index int) string {
	onclick := "javascript:sendInvestigationMessage(" + string(index) + ")"
	return "<li class=\"" + class + "\" onclick=\"" + onclick + "\">" + value + "</li>"
}
