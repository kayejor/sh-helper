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
			sendMessageToClient(conn, "Game already in progress")
			return nil
		}
		sendMessageToClient(conn, "Joined")
		//create list and send to client
	} else {
		player = addNewPlayerToGame(game, playerName, conn)
		if player != nil {
			sendMessageToClient(conn, "Joined")
			//for now send "First" to everyone, should eventually only go to the creator
			sendMessageToClient(conn, "First")
			broadcastNames(game) //maybe change this to just sending the list of players
		}
	}
	return player
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
	//send list of players with roles
	endGame(game) //for now just remove the game from the server
}

func endGame(game *Game) {
	delete(games, game.gameName)
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
			//get list but again we are just sending this once!
			//send list to clients, player.ws literally can never be nil here
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

/*
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
*/
