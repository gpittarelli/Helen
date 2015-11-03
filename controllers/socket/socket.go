// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package socket

import (
	"encoding/json"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	chelpers "github.com/TF2Stadium/Helen/controllers/controllerhelpers"
	"github.com/TF2Stadium/Helen/controllers/socket/internal/handler"
	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/models"
	"github.com/vibhavp/wsevent"
)

func onDisconnect(id string) {
	chelpers.DeauthenticateSocket(id)
	if chelpers.IsLoggedInSocket(id) {
		steamid := chelpers.GetSteamId(id)
		broadcaster.RemoveSocket(steamid)
	}
	helpers.Logger.Debug("Disconnected from Socket")
}

func getEvent(data string) string {
	var js struct {
		Request string
	}
	json.Unmarshal([]byte(data), &js)
	return js.Request
}

func ServerInit(server *wsevent.Server) {
	server.OnDisconnect = onDisconnect
	server.Extractor = getEvent

	server.On("authenticationTest", func(server *wsevent.Server, so *wsevent.Client, data string) string {
		reqerr := chelpers.FilterRequest(so, 0, true)

		if reqerr != nil {
			bytes, _ := reqerr.ErrorJSON().Encode()
			return string(bytes)
		}

		return "authenticated"
	})
	//Lobby Handlers
	server.On("lobbyCreate", handler.LobbyCreate)
	server.On("serverVerify", handler.ServerVerify)
	server.On("lobbyClose", handler.LobbyClose)
	server.On("lobbyJoin", handler.LobbyJoin)
	server.On("lobbySpectatorJoin", handler.LobbySpectatorJoin)
	server.On("lobbyKick", handler.LobbyKick)
	server.On("requestLobbyListData", handler.RequestLobbyListData)
	//Player Handlers
	server.On("playerReady", handler.PlayerReady)
	server.On("playerNotReady", handler.PlayerNotReady)
	server.On("playerSettingsGet", handler.PlayerSettingsGet)
	server.On("playerSettingsSet", handler.PlayerSettingsSet)
	server.On("playerProfile", handler.PlayerProfile)
	//Chat Handlers
	server.On("chatSend", handler.ChatSend)
	//Admin Handlers
	server.On("adminChangeRole", handler.AdminChangeRole)
	//Debugging handlers
	if config.Constants.ServerMockUp {
		server.On("debugLobbyFill", handler.DebugLobbyFill)
		server.On("debugLobbyReady", handler.DebugLobbyReady)
		server.On("debugGetAllLobbies", handler.DebugRequestAllLobbies)
		server.On("debugRequestLobbyStart", handler.DebugRequestLobbyStart)
		server.On("debugUpdateStatsFilter", handler.DebugUpdateStatsFilter)
	}
}

func SocketInit(server *wsevent.Server, so *wsevent.Client) {
	chelpers.AfterConnect(server, so)
	loggedIn := chelpers.IsLoggedInSocket(so.Id())
	if loggedIn {
		player, err := models.GetPlayerBySteamId(chelpers.GetSteamId(so.Id()))
		if err != nil {
			helpers.Logger.Warning(
				"User has a cookie with but a matching player record doesn't exist: %s",
				chelpers.GetSteamId(so.Id()))
			return
		}

		chelpers.AfterConnectLoggedIn(server, so, player)
	} else {
		so.EmitJSON(helpers.NewRequest("playerSettings", []byte("{}")))
		so.EmitJSON(helpers.NewRequest("playerProfile", []byte("{}")))
	}

	so.EmitJSON(helpers.NewRequest("socketInitialized", []byte{}))
}
