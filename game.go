package main

import (
	"log"
	"math/rand"

	"github.com/gorilla/websocket"
)

func joinGame(game *Game, playerName string, conn *websocket.Conn) *Player {
	var player *Player
	if game.gameStarted {
		player = findPlayerInGame(game, playerName)
		if player == nil || player.ws != nil {
			sendErrorMessageToClient(conn, "Game already in progress")
		} else {
			sendPlayerListToClient(conn, createPlayerInfoList(game))
		}
		return nil //either way, disconnect
	}

	game.playersMutex.Lock()
	player = addNewPlayerToGame(game, playerName, conn)
	if player != nil {
		sendControlMessageToClient(conn, "Joined")
		broadcastNames(game)
	}
	game.playersMutex.Unlock()
	return player
}

func addNewPlayerToGame(game *Game, playerName string, conn *websocket.Conn) *Player {
	if len(game.players) == 10 {
		sendErrorMessageToClient(conn, "Game full")
		return nil
	}
	if findPlayerInGame(game, playerName) != nil {
		sendErrorMessageToClient(conn, "Duplicate name")
		return nil
	}
	player := Player{
		ws:   conn,
		name: playerName,
		role: " ",
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
	game.playersMutex.Lock()
	assignRoles(game)
	broadcastNames(game)
	game.gameStarted = true
	for _, player := range game.players {
		sendControlMessageToClient(player.ws, "Start")
		player.ws.Close()
		player.ws = nil
	}
	game.playersMutex.Unlock()
	endGame(game)
}

func handleRemovedPlayer(game *Game) {
	game.playersMutex.Lock()
	if len(game.players) == 0 {
		log.Printf("Removing game %s", game.gameName)
		endGame(game)
	} else {
		broadcastNames(game)
	}
	game.playersMutex.Unlock()
}

func endGame(game *Game) {
	delete(games, game.gameName)
}

func assignRoles(game *Game) {
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
}

func broadcastNames(game *Game) {
	playerList := createPlayerInfoList(game)
	for _, player := range game.players {
		if player.ws != nil {
			sendPlayerListToClient(player.ws, playerList)
		}
	}
}

func removePlayer(game *Game, player *Player) {
	game.playersMutex.Lock()
	//len(players) has a max of 10
	for index := 0; index < len(game.players); index++ {
		if player == game.players[index] {
			game.players = append(game.players[:index], game.players[index+1:]...)
			break
		}
	}
	game.playersMutex.Unlock()
}

//PlayerInfo is used for sending the list of players to the client
type PlayerInfo struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

func createPlayerInfoList(game *Game) []PlayerInfo {
	var players []PlayerInfo
	for _, player := range game.players {
		playerInfo := PlayerInfo{
			Name: player.name,
			Role: player.role,
		}
		players = append(players, playerInfo)
	}
	return players
}
