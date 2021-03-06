// Copyright (C) 2015  TF2Stadium
// Use of this source code is governed by the GPLv3
// that can be found in the COPYING file.

package models

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/TF2Stadium/Helen/config"
	"github.com/TF2Stadium/Helen/controllers/broadcaster"
	db "github.com/TF2Stadium/Helen/database"
)

type SlotDetails struct {
	Filled       bool           `json:"filled"`
	Player       *PlayerSummary `json:"player,omitempty"`
	Ready        *bool          `json:"ready,omitempty"`
	InGame       *bool          `json:"ingame,omitempty"`
	Requirements *Requirement   `json:"requirements,omitempty"`
}

type ClassDetails struct {
	Blu   SlotDetails `json:"blu"`
	Class string      `json:"class"`
	Red   SlotDetails `json:"red"`
}

type SpecDetails struct {
	Name    string `json:"name,omitempty"`
	SteamID string `json:"steamid,omitempty"`
}

type LobbyData struct {
	ID         uint   `json:"id"`
	Mode       string `json:"gamemode"`
	Type       string `json:"type"`
	Players    int    `json:"players"`
	Map        string `json:"map"`
	League     string `json:"league"`
	Mumble     bool   `json:"mumbleRequired"`
	MaxPlayers int    `json:"maxPlayers"`

	PlayerWhitelist bool `json:"whitelisted"`
	Password        bool `json:"password"`

	Region struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"region"`

	Classes []ClassDetails `json:"classes"`

	Leader      PlayerSummary `json:"leader"`
	CreatedAt   int64         `json:"createdAt"`
	State       int           `json:"state"`
	WhitelistID string        `json:"whitelistId"`

	Spectators []SpecDetails `json:"spectators,omitempty"`
}

type LobbyListData struct {
	Lobbies []LobbyData `json:"lobbies,omitempty"`
}

type LobbyConnectData struct {
	ID   uint   `json:"id"`
	Time int64  `json:"time"`
	Pass string `json:"password"`

	Game struct {
		Host string `json:"host"`
	} `json:"game"`

	Mumble struct {
		Address  string `json:"address"`
		Nick     string `json:"nick"`
		Port     string `json:"port"`
		Password string `json:"password"`
		Channel  string `json:"channel"`
	} `json:"mumble"`
}

type SubstituteData struct {
	LobbyID uint   `json:"id"`
	Format  string `json:"type"`
	MapName string `json:"map"`

	Region struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"region"`

	Mumble bool   `json:"mumbleRequired"`
	Team   string `sql:"-" json:"team"`
	Class  string `sql:"-" json:"class"`
}

type LobbyEvent struct {
	ID uint `json:"id"`
}

func decorateSlotDetails(lobby *Lobby, slot int, includeDetails bool) SlotDetails {
	playerId, err := lobby.GetPlayerIDBySlot(slot)
	j := SlotDetails{Filled: err == nil}

	if err == nil && includeDetails {
		var player Player
		db.DB.First(&player, playerId)
		db.DB.Preload("Stats").First(&player, player.ID)

		summary := DecoratePlayerSummary(&player)
		j.Player = &summary

		ready, _ := lobby.IsPlayerReady(&player)
		j.Ready = &ready

		ingame, _ := lobby.IsPlayerInGame(&player)
		j.InGame = &ingame

		if lobby.HasSlotRequirement(slot) {
			j.Requirements, _ = lobby.GetSlotRequirement(slot)
		} else if lobby.HasGeneralRequirement() {
			j.Requirements, _ = lobby.GetGeneralRequirement()
		}
	}

	return j
}

func DecorateLobbyData(lobby *Lobby, includeDetails bool) LobbyData {
	lobbyData := LobbyData{
		ID:      lobby.ID,
		Mode:    lobby.Mode,
		Type:    formatMap[lobby.Type],
		Players: lobby.GetPlayerNumber(),
		Map:     lobby.MapName,
		League:  lobby.League,
		Mumble:  lobby.Mumble,

		PlayerWhitelist: lobby.PlayerWhitelist != "",
		Password:        lobby.SlotPassword != "",
	}

	lobbyData.Region.Name = lobby.RegionName
	lobbyData.Region.Code = lobby.RegionCode

	var classList = typeClassList[lobby.Type]

	classes := make([]ClassDetails, len(classList))
	lobbyData.MaxPlayers = NumberOfClassesMap[lobby.Type] * 2

	for slot, className := range classList {
		class := ClassDetails{
			Red:   decorateSlotDetails(lobby, slot, includeDetails),
			Blu:   decorateSlotDetails(lobby, slot+NumberOfClassesMap[lobby.Type], includeDetails),
			Class: className,
		}

		classes[slot] = class
	}

	lobbyData.Classes = classes
	lobbyData.WhitelistID = lobby.Whitelist

	if !includeDetails {
		return lobbyData
	}

	var leader Player
	db.DB.Where("steam_id = ?", lobby.CreatedBySteamID).First(&leader)

	lobbyData.Leader = DecoratePlayerSummary(&leader)
	lobbyData.CreatedAt = lobby.CreatedAt.Unix()
	lobbyData.State = int(lobby.State)

	var specIDs []uint
	db.DB.Table("spectators_players_lobbies").Where("lobby_id = ?", lobby.ID).Pluck("player_id", &specIDs)

	spectators := make([]SpecDetails, len(specIDs))

	for i, spectatorID := range specIDs {
		specPlayer := &Player{}
		db.DB.First(specPlayer, spectatorID)

		specJs := SpecDetails{
			Name:    specPlayer.Alias(),
			SteamID: specPlayer.SteamID,
		}

		spectators[i] = specJs
	}

	lobbyData.Spectators = spectators

	return lobbyData
}

func (l LobbyData) Send() {
	broadcaster.SendMessageToRoom(fmt.Sprintf("%d_public", l.ID), "lobbyData", l)
}

func (l LobbyData) SendToPlayer(steamid string) {
	broadcaster.SendMessage(steamid, "lobbyData", l)
}

func DecorateLobbyListData(lobbies []Lobby) []LobbyData {
	var lobbyList = make([]LobbyData, len(lobbies))

	for i, lobby := range lobbies {
		lobbyData := DecorateLobbyData(&lobby, false)
		lobbyList[i] = lobbyData
	}

	return lobbyList
}

func DecorateLobbyConnect(lobby *Lobby, name string, slot int) LobbyConnectData {
	l := LobbyConnectData{}
	l.ID = lobby.ID
	l.Time = lobby.CreatedAt.Unix()
	l.Pass = lobby.ServerInfo.ServerPassword

	l.Game.Host = lobby.ServerInfo.Host

	l.Mumble.Address = config.Constants.MumbleAddr
	l.Mumble.Password = config.Constants.MumblePassword
	l.Mumble.Channel = "match" + strconv.FormatUint(uint64(lobby.ID), 10)
	team, class, _ := LobbyGetSlotInfoString(lobby.Type, slot)
	nick := strings.ToUpper(team) + "_" + strings.ToUpper(class)
	l.Mumble.Nick = nick

	return l
}

func DecorateLobbyJoin(lobby *Lobby) LobbyEvent {
	return LobbyEvent{lobby.ID}
}

func DecorateLobbyLeave(lobby *Lobby) LobbyEvent {
	return LobbyEvent{lobby.ID}
}

func DecorateLobbyClosed(lobby *Lobby) LobbyEvent {
	return LobbyEvent{lobby.ID}
}

func DecorateSubstitute(slot *LobbySlot) SubstituteData {
	lobby := &Lobby{}
	db.DB.First(lobby, slot.LobbyID)
	substitute := SubstituteData{
		LobbyID: lobby.ID,
		Format:  formatMap[lobby.Type],
		MapName: lobby.MapName,
		Mumble:  lobby.Mumble,
	}

	substitute.Region.Name = lobby.RegionName
	substitute.Region.Code = lobby.RegionCode

	substitute.Team, substitute.Class, _ = LobbyGetSlotInfoString(lobby.Type, slot.Slot)

	return substitute
}

func DecorateSubstituteList() []SubstituteData {
	slots := []*LobbySlot{}
	subList := []SubstituteData{}

	db.DB.Table("lobby_slots").Joins("INNER JOIN lobbies ON lobbies.id = lobby_slots.lobby_id").Where("lobby_slots.needs_sub = ? AND lobbies.state = ?", true, LobbyStateInProgress).Find(&slots)

	for _, slot := range slots {
		subList = append(subList, DecorateSubstitute(slot))
	}

	return subList
}
