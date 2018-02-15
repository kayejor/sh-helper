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

type Game struct {
	players     []*Player
	plaersMutex sync.Mutex
	gameStarted bool
}

type Player struct {
	ws   *websocket.Conn
	name string
	role string
}

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

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if !game.gameStarted {
				removePlayer(game, thisPlayer)
				broadcastNames(game)
			} else {
				//wait for reconnect
				removeConnection(game, thisPlayer)
			}
			return
		}

		if string(msg) == "start" {
			if len(game.players) >= 5 {
				startGame(game)
			} else {
				sendMessageToClient(conn, "Not enough players")
			}
		} else {
			invIndex, _ := strconv.Atoi(string(msg))
			htmlList := createHTMLListForPlayerWithInv(game, thisPlayer, invIndex)
			sendMessageToClient(conn, htmlList)
		}
	}
}

func joinGame(game *Game, playerName string, conn *websocket.Conn) *Player {
	var player *Player
	if game.gameStarted {
		player = findPlayerInGame(game, playerName)
		if player == nil || player.ws != nil {
			sendMessageToClient(conn, "Game started")
			return nil
		}
		player.ws = conn
		sendMessageToClient(conn, "Joined")
		htmlList := createHTMLListForPlayer(game, player)
		sendMessageToClient(conn, htmlList)
	} else {
		player = addNewPlayerToGame(game, playerName, conn)
		if player != nil {
			sendMessageToClient(conn, "Joined")
			//for now send "First" to everyone, should eventually only go to the creator
			sendMessageToClient(conn, "First")
			broadcastNames(game)
		}
	}
	return player
}

func sendMessageToClient(conn *websocket.Conn, message string) {
	conn.WriteMessage(websocket.TextMessage, []byte(message))
}

func addNewPlayerToGame(game *Game, playerName string, conn *websocket.Conn) *Player {
	if len(game.players) == 10 {
		sendMessageToClient(conn, "Game full")
		return nil
	}
	if findPlayerInGame(game, playerName) != nil {
		sendMessageToClient(conn, "Duplicate name")
		return nil
	}
	player := Player{
		ws:   conn,
		name: playerName,
		role: "",
	}
	game.players = append(game.players, &player)
	return &player
}

func findPlayerInGame(game *Game, playerName string) *Player {
	for _, player := range game.players {
		if player.name == playerName {
			return player
		}
	}
	return nil
}

func startGame(game *Game) {
	game.gameStarted = true
	for _, player := range game.players {
		sendMessageToClient(player.ws, "Started")
	}
	assignRoles(game)
	broadcastNames(game)
}

func assignRoles(game *Game) {
	game.plaersMutex.Lock()

	numberOfPlayers := len(game.players)
	randPerm := rand.Perm(numberOfPlayers)
	numberOfLiberals := numberOfPlayers/2 + 1
	for i, player := range game.players {
		if randPerm[i] == 0 {
			player.role = "hitler"
		} else if randPerm[i] > numberOfLiberals {
			player.role = "fascist"
		} else {
			player.role = "liberal"
		}
	}

	game.plaersMutex.Unlock()
}

func broadcastNames(game *Game) {
	game.plaersMutex.Lock()
	for _, player := range game.players {
		if player.ws != nil {
			htmlList := createHTMLListForPlayer(game, player)
			sendMessageToClient(player.ws, htmlList)
		}
	}
	game.plaersMutex.Unlock()
}

func removePlayer(game *Game, player *Player) {
	game.plaersMutex.Lock()
	//len(players) has a max of 10
	for index := 0; index < len(game.players); index++ {
		if player == game.players[index] {
			game.players = append(game.players[:index], game.players[index+1:]...)
			break
		}
	}
	game.plaersMutex.Unlock()
}

func removeConnection(game *Game, playerToRemove *Player) {
	game.plaersMutex.Lock()

	for _, player := range game.players {
		if player == playerToRemove {
			player.ws = nil
			break
		}
	}

	game.plaersMutex.Unlock()
}

func createHTMLListForPlayer(game *Game, currentPlayer *Player) string {
	return createHTMLListForPlayerWithInv(game, currentPlayer, -1)
}

func createHTMLListForPlayerWithInv(game *Game, currentPlayer *Player, invIndex int) string {
	var result = ""
	players := game.players
	currentPlayerRole := currentPlayer.role
	seeAll := currentPlayerRole == "fascist" || currentPlayerRole == "hitler" && len(players) < 7
	for index, player := range players {
		if game.gameStarted && (seeAll || player == currentPlayer || index == invIndex) {
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

func createListElement(name string, class string, index int) string {
	onclick := "javascript:sendInvestigationMessage(" + strconv.Itoa(index) + ")"
	nameDiv := fmt.Sprintf("<div class=\"listName\">%s</div>", name)
	logoDiv := createLogoDiv(class)
	return fmt.Sprintf("<li class=\"%s\" onclick=\"%s\">%s%s</li>",
		class, onclick, logoDiv, nameDiv)
}

func createLogoDiv(class string) string {
	party := class
	if party == "hitler" {
		party = "fascist"
	}
	return fmt.Sprintf("<div class=\"logo %s\"></div>", party)
}
