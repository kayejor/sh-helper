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

var players []*Player
var playersMux sync.Mutex
var gameStarted bool

func main() {
	gameStarted = false
	port := os.Getenv("PORT")
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
	http.HandleFunc("/websocket", handleConnections)
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Server starting on port " + port)
	http.ListenAndServe(":"+port, nil)
	fmt.Println("Server stopped")
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

	//when websocket is opened, client sends us the player name
	_, playerName, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("problem with name message")
		return
	}

	var myIndex = -1

	if gameStarted {
		myIndex = handleReconnect(string(playerName), conn)
		if myIndex == -1 {
			conn.WriteMessage(websocket.TextMessage, []byte("Game started"))
			return
		}
		conn.WriteMessage(websocket.TextMessage, []byte("Joined"))
		htmlList := createHTMLListForPlayer(myIndex)
		conn.WriteMessage(websocket.TextMessage, []byte(htmlList))
	} else {
		index, successful := handleNewPlayer(string(playerName), conn)
		if successful {
			myIndex = index
			conn.WriteMessage(websocket.TextMessage, []byte("Joined"))
			broadcastNames()
		} else {
			return
		}
	}

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if !gameStarted {
				removePlayer(conn)
				broadcastNames()
			} else if connectionCount() > 1 {
				//wait for reconnect
				removeConnection(conn)
			} else {
				endGame()
			}
			return
		}

		if string(msg) == "start" {
			if len(players) >= 5 {
				startGame()
			} else {
				conn.WriteMessage(websocket.TextMessage, []byte("Not enough players"))
			}
		} else {
			invIndex, _ := strconv.Atoi(string(msg))
			htmlList := createHTMLListForPlayerWithInv(myIndex, invIndex)
			conn.WriteMessage(websocket.TextMessage, []byte(htmlList))
		}
	}
}

func handleNewPlayer(playerName string, conn *websocket.Conn) (index int, successful bool) {
	if len(players) == 10 {
		conn.WriteMessage(websocket.TextMessage, []byte("Game full"))
		return -1, false
	}
	if playersContains(playerName) {
		conn.WriteMessage(websocket.TextMessage, []byte("Duplicate name"))
		return -1, false
	}
	player := Player{
		ws:   conn,
		name: playerName,
		role: "",
	}
	index = len(players)
	players = append(players, &player)
	return index, true
}

func handleReconnect(playerName string, ws *websocket.Conn) int {
	for index, player := range players {
		if player.ws == nil && player.name == playerName {
			player.ws = ws
			return index
		}
	}
	return -1
}

func playersContains(name string) bool {
	for _, player := range players {
		if player.name == name {
			return true
		}
	}
	return false
}

func connectionCount() int {
	var result = 0
	for _, player := range players {
		if player.ws != nil {
			result++
		}
	}
	return result
}

func startGame() {
	gameStarted = true
	assignRoles()
	broadcastNames()
}

func endGame() {
	players = players[:0]
	gameStarted = false
}

func assignRoles() {
	playersMux.Lock()

	numberOfPlayers := len(players)
	randPerm := rand.Perm(numberOfPlayers)
	numberOfLiberals := numberOfPlayers/2 + 1
	for i, player := range players {
		if randPerm[i] == 0 {
			player.role = "hitler"
		} else if randPerm[i] > numberOfLiberals {
			player.role = "fascist"
		} else {
			player.role = "liberal"
		}
	}

	playersMux.Unlock()
}

func broadcastNames() {
	playersMux.Lock()
	for index, player := range players {
		if player.ws != nil {
			htmlList := createHTMLListForPlayer(index)
			player.ws.WriteMessage(websocket.TextMessage, []byte(htmlList))
		}
	}
	playersMux.Unlock()
}

func removePlayer(conn *websocket.Conn) {
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

func removeConnection(conn *websocket.Conn) {
	playersMux.Lock()

	for _, player := range players {
		if conn == player.ws {
			player.ws = nil
			break
		}
	}

	playersMux.Unlock()
}

func createHTMLListForPlayer(currentPlayerIndex int) string {
	return createHTMLListForPlayerWithInv(currentPlayerIndex, -1)
}

func createHTMLListForPlayerWithInv(currentPlayerIndex int, invIndex int) string {
	var result = ""
	currentPlayerRole := players[currentPlayerIndex].role
	seeAll := currentPlayerRole == "fascist" || currentPlayerRole == "hitler" && len(players) < 7
	for index, player := range players {
		if gameStarted && (seeAll || index == currentPlayerIndex || index == invIndex) {
			party := player.role
			if index == invIndex && party == "hitler" {
				party = "fascist"
			}
			result += createListElement(player.name, party, index)
		} else {
			result += createListElement(player.name, "", index)
		}
	}
	return result
}

func createListElement(value string, class string, index int) string {
	onclick := "javascript:sendInvestigationMessage(" + strconv.Itoa(index) + ")"
	return "<li class=\"" + class + "\" onclick=\"" + onclick + "\">" + value + "</li>"
}
