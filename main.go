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

//Game holds the player info, and other game related variables
type Game struct {
	players     []*Player
	plaersMutex sync.Mutex
	gameStarted bool
	gameName    string
}

//Player holds a pointer to the websocket connection, the player's name, and the player's role if the game has started
type Player struct {
	ws   *websocket.Conn
	name string
	role string
}

//JoinMessage is type that the client will send when trying to join a game
type JoinMessage struct {
	GameName string `json:"gameName"`
	Name     string `json:"name"`
}

var games = make(map[string]*Game)

func main() {
	port := os.Getenv("PORT")
	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)
	http.HandleFunc("/websocket", handleConnections)
	http.HandleFunc("/end", handleGameEnd)
	http.HandleFunc("/create", handleCreateGame)
	rand.Seed(time.Now().UnixNano())
	log.Println("Server starting on port " + port)
	http.ListenAndServe(":"+port, nil)
	log.Println("Server stopped")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleCreateGame(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println("Error parsing request")
		return
	}
	gameName := r.Form.Get("gameName")

	if _, exists := games[gameName]; exists {
		w.Write([]byte("Game already exists"))
	} else {
		log.Printf("Creating game %s", gameName)
		game := Game{
			gameStarted: false,
			gameName:    gameName,
		}
		games[gameName] = &game
	}
}

func handleGameEnd(w http.ResponseWriter, r *http.Request) {
	//for now just use this to end ALL game
	for game := range games {
		delete(games, game)
	}
	http.Redirect(w, r, "/", 302)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	var joinMessage JoinMessage
	err = conn.ReadJSON(&joinMessage)
	if err != nil {
		log.Println("problem with initial message")
		return
	}

	game, exists := games[joinMessage.GameName]
	if !exists {
		sendMessageToClient(conn, "Game does not exist")
		return
	}

	var thisPlayer *Player
	thisPlayer = joinGame(game, joinMessage.Name, conn)
	if thisPlayer == nil {
		return
	}
	log.Printf("Player %s joined game %s", joinMessage.Name, joinMessage.GameName)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if !game.gameStarted {
				removePlayer(game, thisPlayer)
				log.Printf("Player %s left game %s", thisPlayer.name, game.gameName)
				if len(game.players) == 0 {
					log.Printf("All players gone from game %s before starting", game.gameName)
					endGame(game)
				} else {
					broadcastNames(game)
				}
			} else {
				removeConnection(game, thisPlayer) //wait for reconnect
				log.Printf("Player %s left started game %s", thisPlayer.name, game.gameName)
			}
			return
		}

		if string(msg) == "start" {
			if len(game.players) >= 5 {
				log.Printf("Game %s started", game.gameName)
				startGame(game)
				go pollGameForPlayers(game, 30) //delete game if all players gone for 30 minutes
			} else {
				sendMessageToClient(conn, "Not enough players")
			}
		} else {
			invIndex, _ := strconv.Atoi(string(msg))
			log.Printf("In game %s: %s investigated player %d", game.gameName, thisPlayer.name, invIndex)
			htmlList := createHTMLListForPlayerWithInv(game, thisPlayer, invIndex)
			sendMessageToClient(conn, htmlList)
		}
	}
}

func sendMessageToClient(conn *websocket.Conn, message string) {
	conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func pollGameForPlayers(game *Game, idleMinutesToEnd int) {
	var counter = 0
	for counter < idleMinutesToEnd {
		if atLeastOnePlayerConnected(game) {
			counter = 0
		} else {
			counter = counter + 1
		}
		time.Sleep(time.Minute)
	}
	log.Printf("Game %s idle for %d minutes, ending game", game.gameName, idleMinutesToEnd)
	endGame(game)
}

func atLeastOnePlayerConnected(game *Game) bool {
	for _, player := range game.players {
		if player.ws != nil {
			return true
		}
	}
	return false
}

func endGame(game *Game) {
	delete(games, game.gameName)
}
