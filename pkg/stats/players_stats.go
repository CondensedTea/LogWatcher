package stats

import (
	"LogWatcher/pkg/requests"

	//"LogWatcher/pkg/app"
	"context"
	"regexp"
	"strconv"

	"github.com/leighmacdonald/steamid/steamid"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	mongoDatabase   = "log_watcher"
	mongoCollection = "player_stats"
)

var (
	killRegexp   = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+killed.+(\[U:\d:\d{1,10}])`)
	damageRegexp = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+triggered "damage" against.+(\[U:\d:\d{1,10}]).+\(damage "(\d+)"\)`)
	healsRegexp  = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+triggered "healed" against.+(\[U:\d:\d{1,10}]).+\(healing "(\d+)"\)`)
)

type PlayerStats struct {
	Kills         int
	Deaths        int
	DamageDone    int `bson:"damage_done"`
	DamageTaken   int `bson:"damage_taken"`
	Healed        int
	HealsReceived int `bson:"heals_received"`
}

type GameStats struct {
	Player   requests.PickupPlayer
	Stats    PlayerStats
	Server   ServerInfo
	PickupID int
}

func UpdatePlayerStats(msg string, stats map[steamid.SID64]*PlayerStats) error {
	switch {
	case killRegexp.MatchString(msg):
		match := killRegexp.FindStringSubmatch(msg)
		p1 := steamid.SID3ToSID64(steamid.SID3(match[1]))
		p2 := steamid.SID3ToSID64(steamid.SID3(match[2]))
		_, ok := stats[p1]
		if !ok {
			stats[p1] = &PlayerStats{Kills: 1}
		} else {
			stats[p1].Kills += 1
		}
		_, ok = stats[p2]
		if !ok {
			stats[p2] = &PlayerStats{Deaths: 1}
		} else {
			stats[p2].Deaths += 1
		}
	case damageRegexp.MatchString(msg):
		match := damageRegexp.FindStringSubmatch(msg)
		p1 := steamid.SID3ToSID64(steamid.SID3(match[1]))
		p2 := steamid.SID3ToSID64(steamid.SID3(match[2]))
		dmg, _ := strconv.Atoi(match[3])
		_, ok := stats[p1]
		if !ok {
			stats[p1] = &PlayerStats{DamageDone: dmg}
		} else {
			stats[p1].DamageDone += dmg
		}
		_, ok = stats[p2]
		if !ok {
			stats[p2] = &PlayerStats{DamageTaken: dmg}
		} else {
			stats[p2].DamageTaken += dmg
		}
	case healsRegexp.MatchString(msg):
		match := healsRegexp.FindStringSubmatch(msg)
		p1 := steamid.SID3ToSID64(steamid.SID3(match[1]))
		p2 := steamid.SID3ToSID64(steamid.SID3(match[2]))
		heals, _ := strconv.Atoi(match[3])
		h, ok := stats[p1]
		if !ok {
			stats[p1] = &PlayerStats{Healed: heals}
		} else {
			h.Healed += heals
		}
		h, ok = stats[p2]
		if !ok {
			stats[p2] = &PlayerStats{HealsReceived: heals}
		} else {
			h.HealsReceived += heals
		}
	}
	return nil
}

func ExtractPlayerStats(gameStats map[steamid.SID64]*PlayerStats, server ServerInfo, pickupID int) []interface{} {
	s := make([]interface{}, 0)
	for steamID, stats := range gameStats {
		gs := GameStats{
			Player:   requests.PickupPlayer{SteamID: steamID.String()},
			Stats:    *stats,
			Server:   server,
			PickupID: pickupID,
		}
		s = append(s, gs)
	}
	return s
}

func InsertGameStats(ctx context.Context, conn *mongo.Client, documents []interface{}) error {
	_, err := conn.
		Database(mongoDatabase).
		Collection(mongoCollection).
		InsertMany(ctx, documents)
	return err
}

type ServerInfo struct {
	ID     int
	Domain string
	IP     string
}
