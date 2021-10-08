package stats

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/leighmacdonald/steamid/steamid"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	timeStamp    = regexp.MustCompile(`\d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}`)
	killRegexp   = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+killed.+(\[U:\d:\d{1,10}])`)
	damageRegexp = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+triggered "damage" against.+(\[U:\d:\d{1,10}]).+\(damage "(\d+)"\)`)
	healsRegexp  = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+triggered "healed" against.+(\[U:\d:\d{1,10}]).+\(healing "(\d+)"\)`)
)

// UpdateStatsMap is used for accumulating players stats in a map
func UpdateStatsMap(msg string, stats map[steamid.SID64]*PlayerGameStats) error {
	switch {
	case killRegexp.MatchString(msg):
		match := killRegexp.FindStringSubmatch(msg)
		p1 := steamid.SID3ToSID64(steamid.SID3(match[1]))
		p2 := steamid.SID3ToSID64(steamid.SID3(match[2]))
		_, ok := stats[p1]
		if !ok {
			stats[p1] = &PlayerGameStats{Kills: 1}
		} else {
			stats[p1].Kills += 1
		}
		_, ok = stats[p2]
		if !ok {
			stats[p2] = &PlayerGameStats{Deaths: 1}
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
			stats[p1] = &PlayerGameStats{DamageDone: dmg}
		} else {
			stats[p1].DamageDone += dmg
		}
		_, ok = stats[p2]
		if !ok {
			stats[p2] = &PlayerGameStats{DamageTaken: dmg}
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
			stats[p1] = &PlayerGameStats{Healed: heals}
		} else {
			h.Healed += heals
		}
		h, ok = stats[p2]
		if !ok {
			stats[p2] = &PlayerGameStats{HealsReceived: heals}
		} else {
			h.HealsReceived += heals
		}
	}
	return nil
}

func ExtractPlayerStats(gi MatchDater) []interface{} {
	s := make([]interface{}, 0)
	for _, player := range gi.PickupPlayers() {
		for steamID, stats := range gi.PlayerStatsMap() {
			if player.SteamID == steamID.String() {
				gs := MongoPlayerInfo{
					Player:   player,
					Stats:    *stats,
					Domain:   gi.Domain(),
					PickupID: gi.PickupID(),
					Length:   gi.LengthSeconds(),
				}
				s = append(s, gs)
			}
		}
	}
	return s
}

func InsertGameStats(conn *mongo.Client, documents []interface{}) error {
	_, err := conn.
		Database(mongoDatabase).
		Collection(mongoCollection).
		InsertMany(context.Background(), documents)
	return err
}

func ParseTimeStamp(msg string) time.Time {
	match := timeStamp.FindString(msg)
	t, _ := time.Parse(`01/2/2006 - 15:04:05`, match) // err is always nil
	return t
}
