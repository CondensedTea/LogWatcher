package server

import (
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/stats"
)

type StateType int

const (
	StartedState           = "started"
	Pregame      StateType = iota
	Game
)

func (st StateType) String() string {
	switch st {
	case Pregame:
		return "pregame"
	case Game:
		return "game"
	default:
		return "unknown state"
	}
}

// Flush is used to empty all game data
func Flush(lf LogFiler, gi stats.MatchDater) {
	lf.FlushBuffer()
	gi.SetPickupID(0)
	gi.SetMap("")
	gi.FlushPlayerStatsMap()
}

// UpdatePickupInfo is used for finding current game on tf2pickup API
// and loading to LogFile list of its players and pickup ID
func UpdatePickupInfo(r requests.LogProcessor, gi stats.MatchDater) error {
	gr, err := r.GetPickupGames(gi.Domain())
	if err != nil {
		return err
	}
	for _, game := range gr.Results {
		if game.State == StartedState && game.Map == gi.Map() {
			players := make([]*stats.PickupPlayer, 0)
			for _, player := range game.Slots {
				p := &stats.PickupPlayer{
					PlayerID: player.Player, Class: player.GameClass, Team: player.Team,
				}
				players = append(players, p)
			}
			gi.SetPlayers(players)
			gi.SetPickupID(game.Number)
			break
		}
	}
	return nil
}
