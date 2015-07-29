package decorators

import (
	chelpers "github.com/TF2Stadium/Server/controllers/controllerhelpers"
	db "github.com/TF2Stadium/Server/database"
	"github.com/TF2Stadium/Server/models"
	"github.com/bitly/go-simplejson"
	"github.com/jinzhu/gorm"
)

func GetLobbyListData() (string, error) {
	count := 0
	db.DB.Where("state = ?", LobbyStateWaiting).Count(&count)

	if count == 0 {
		return "{}", nil
	}

	lobbyList := make([]*simplejson.Json, count)
	lobbies := make([]*Lobby, count)
	err := db.DB.Where("state = ?", LobbyStateWaiting).Find(&lobbies).Error

	if err != nil {
		return "{}", err
	}

	for lobbyIndex, lobby := range lobbies {
		lobbyJs := simplejson.New()
		lobbyJs.Set("id", lobby.ID)
		lobbyJs.Set("type", models.FormatMap[lobby.Type])
		lobbyJs.Set("createdAt", lobby.CreatedAt.String())
		lobbyJs.Set("players", lobby.GetPlayerNumber())
		classes := make([]*simplejson.Json, int(lobby.Type))

		for i := 0; i <= int(lobby.Type); i++ {
			slot := simplejson.New()
			class := simplejson.New()

			slot.Set("red", lobby.IsSlotFilled(i))
			slot.Set("blu", lobby.IsSlotFilled(i+6))

			class.Set(chelpers.PlayerSlotToString(i, lobby.Type), slot)
			classes[i] = class
		}

		lobbyList[lobbyIndex] = lobbyJs
	}

	bytes, _ := json.Marshal(lobbyList)
	return string(bytes), nil
}
