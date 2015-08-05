package socket

import (
	"strconv"
	"time"

	db "github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/decorators"
	"github.com/TF2Stadium/Server/helpers"
	"github.com/TF2Stadium/Server/models"
	"github.com/googollee/go-socket.io"
)

type broadcastMessage struct {
	Room    string
	SteamId string
	Event   string
	Content string
}

var SteamIdSocketMap = make(map[string]*socketio.Socket)
var broadcasterTicker *time.Ticker
var broadcastStopChannel chan bool
var broadcastMessageChannel chan broadcastMessage
var socketServer *socketio.Server

func InitBroadcaster(server *socketio.Server) {
	broadcasterTicker = time.NewTicker(time.Millisecond * 500)
	broadcastStopChannel = make(chan bool)
	broadcastMessageChannel = make(chan broadcastMessage)
	socketServer = server
	go broadcaster()
}

func StopBroadcaster() {
	broadcasterTicker.Stop()
	broadcastStopChannel <- true
}

func SendMessage(steamid string, event string, content string) {
	broadcastMessageChannel <- broadcastMessage{
		Room:    "",
		SteamId: steamid,
		Event:   event,
		Content: content,
	}
}

func SendMessageToRoom(room string, steamid string, event string, content string) {
	broadcastMessageChannel <- broadcastMessage{
		Room:    room,
		SteamId: steamid,
		Event:   event,
		Content: content,
	}
}

func broadcaster() {
	for {
		select {
		case <-broadcasterTicker.C:
			var lobbies []models.Lobby
			db.DB.Where("state = ?", models.LobbyStateWaiting).Order("id desc").Find(&lobbies)
			list, err := decorators.GetLobbyListData(lobbies)
			if err != nil {
				helpers.Logger.Warning("Failed to send lobby list: %s", err.Error())
			} else {
				socketServer.BroadcastTo("-1", "lobbyListData", list)
			}

			var activeStates = []models.LobbyState{models.LobbyStateWaiting, models.LobbyStateInProgress}
			db.DB.Where("state IN (?)", activeStates).
				Find(&lobbies)
			for _, lobby := range lobbies {
				bytes, _ := decorators.GetLobbyDataJSON(lobby).Encode()
				socketServer.BroadcastTo(strconv.FormatUint(uint64(lobby.ID), 10), string(bytes))
			}

		case message := <-broadcastMessageChannel:
			socket, ok := SteamIdSocketMap[message.SteamId]
			if !ok {
				helpers.Logger.Warning("Failed to get user's socket: %d", message.SteamId)
				continue
			}
			if message.Room == "" {
				(*socket).Emit(message.Event, message.Content)
			} else {
				(*socket).BroadcastTo(message.Room, message.Event, message.Content)
			}
		case <-broadcastStopChannel:
			return
		}
	}
}
