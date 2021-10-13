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
func Flush(lf LogFiler, md stats.MatchDater) {
	lf.FlushBuffer()
	md.SetPickupID(0)
	md.SetMap("")
	md.FlushPlayerStatsCollection()
}

// UpdatePickupInfo is used for finding current game on tf2pickup API
// and loading to LogFile list of its players and pickup ID
func UpdatePickupInfo(lp requests.LogProcessor, md stats.MatchDater) error {
	gr, err := lp.GetPickupGames(md.Domain())
	if err != nil {
		return err
	}
	for _, game := range gr.Results {
		if game.State == StartedState && game.Map == md.Map() {
			players := make([]*stats.PickupPlayer, 0)
			for _, player := range game.Slots {
				p := &stats.PickupPlayer{
					PlayerID: player.Player, Class: player.GameClass, Team: player.Team,
				}
				players = append(players, p)
			}
			md.SetPlayers(players)
			md.SetPickupID(game.Number)
			break
		}
	}
	return nil
}
