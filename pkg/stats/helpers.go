package stats

import (
	"regexp"
	"strconv"
	"time"

	"github.com/leighmacdonald/steamid/steamid"
)

const CurrentStatsSchemaVersion = 1

var (
	timeStamp = regexp.MustCompile(`\d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}`)

	// spaces are need for not mixing up messages with 'killedobject'
	killRegexp   = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+ killed .+(\[U:\d:\d{1,10}])`)
	damageRegexp = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+triggered "damage" against.+(\[U:\d:\d{1,10}]).+\(damage "(\d+)"\)`)
	healsRegexp  = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+triggered "healed" against.+(\[U:\d:\d{1,10}]).+\(healing "(\d+)"\)`)
)

// UpdateStatsMap parses log line and updates PlayerStatsCollection
func UpdateStatsMap(msg string, stats PlayerStatsCollection) PlayerStatsCollection {
	switch {
	case killRegexp.MatchString(msg):
		match := killRegexp.FindStringSubmatch(msg)
		fromPlayer := steamid.SID3ToSID64(steamid.SID3(match[1]))
		toPlayer := steamid.SID3ToSID64(steamid.SID3(match[2]))
		playerStats, ok := stats[fromPlayer]
		if !ok {
			stats[fromPlayer] = &PlayerStats{Kills: 1}
		} else {
			playerStats.Kills += 1
		}
		playerStats, ok = stats[toPlayer]
		if !ok {
			stats[toPlayer] = &PlayerStats{Deaths: 1}
		} else {
			playerStats.Deaths += 1
		}
	case damageRegexp.MatchString(msg):
		match := damageRegexp.FindStringSubmatch(msg)
		fromPlayer := steamid.SID3ToSID64(steamid.SID3(match[1]))
		toPlayer := steamid.SID3ToSID64(steamid.SID3(match[2]))
		dmg, _ := strconv.Atoi(match[3])
		playerStats, ok := stats[fromPlayer]
		if !ok {
			stats[fromPlayer] = &PlayerStats{DamageDone: dmg}
		} else {
			playerStats.DamageDone += dmg
		}
		playerStats, ok = stats[toPlayer]
		if !ok {
			stats[toPlayer] = &PlayerStats{DamageTaken: dmg}
		} else {
			playerStats.DamageTaken += dmg
		}
	case healsRegexp.MatchString(msg):
		match := healsRegexp.FindStringSubmatch(msg)
		fromPlayer := steamid.SID3ToSID64(steamid.SID3(match[1]))
		toPlayer := steamid.SID3ToSID64(steamid.SID3(match[2]))
		heals, _ := strconv.Atoi(match[3])
		playerStats, ok := stats[fromPlayer]
		if !ok {
			stats[fromPlayer] = &PlayerStats{Healed: heals}
		} else {
			playerStats.Healed += heals
		}
		playerStats, ok = stats[toPlayer]
		if !ok {
			stats[toPlayer] = &PlayerStats{HealsReceived: heals}
		} else {
			playerStats.HealsReceived += heals
		}
	}
	return stats
}

func ExtractPlayerStats(md Matcher) []interface{} {
	s := make([]interface{}, 0)
	for _, player := range md.PickupPlayers() {
		for steamID, stats := range md.PlayerStats() {
			if player.SteamID == steamID.String() {
				gs := MongoPlayerInfo{
					Player:        player,
					Stats:         *stats,
					Domain:        md.Domain(),
					PickupID:      md.PickupID(),
					Length:        md.LengthSeconds(),
					SchemaVersion: CurrentStatsSchemaVersion,
				}
				s = append(s, gs)
			}
		}
	}
	return s
}

func ParseTimeStamp(msg string) time.Time {
	match := timeStamp.FindString(msg)
	t, _ := time.Parse(`01/2/2006 - 15:04:05`, match) // err is always nil
	return t
}
