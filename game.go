package main

import (
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

	player = addNewPlayerToGame(game, playerName, conn)
	if player != nil {
		sendControlMessageToClient(conn, "Joined")
		//for now send "First" to everyone, should eventually only go to the creator
		sendControlMessageToClient(conn, "First")
		broadcastNames(game)
	}
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
	assignRoles(game)
	broadcastNames(game)
	for _, player := range game.players {
		sendControlMessageToClient(player.ws, "Start")
		player.ws = nil
	}
	//this is where we set off some kind of timer to only save the game for so long, for now do nothing
}

func endGame(game *Game) {
	delete(games, game.gameName)
}

func assignRoles(game *Game) {
	game.playersMutex.Lock()

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

	game.playersMutex.Unlock()
}

func broadcastNames(game *Game) {
	game.playersMutex.Lock()
	playerList := createPlayerInfoList(game)
	for _, player := range game.players {
		if player.ws != nil {
			sendPlayerListToClient(player.ws, playerList)
		}
	}
	game.playersMutex.Unlock()
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
