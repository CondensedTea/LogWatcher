package main

import (
	"regexp"
	"strconv"

	"github.com/leighmacdonald/steamid/steamid"
)

const (
	mongoDatabase   = "logwatcher"
	mongoCollection = "playerstats"
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
	Player   PickupPlayer
	Stats    PlayerStats
	Server   ServerInfo
	PickupID int
}

func (gi *GameInfo) updatePlayerStats(msg string) error {
	switch {
	case killRegexp.MatchString(msg):
		match := killRegexp.FindStringSubmatch(msg)
		p1 := steamid.SID3ToSID64(steamid.SID3(match[1]))
		p2 := steamid.SID3ToSID64(steamid.SID3(match[2]))
		_, ok := gi.Stats[p1]
		if !ok {
			gi.Stats[p1] = &PlayerStats{Kills: 1}
		} else {
			gi.Stats[p1].Kills += 1
		}
		_, ok = gi.Stats[p2]
		if !ok {
			gi.Stats[p2] = &PlayerStats{Deaths: 1}
		} else {
			gi.Stats[p2].Deaths += 1
		}
	case damageRegexp.MatchString(msg):
		match := damageRegexp.FindStringSubmatch(msg)
		p1 := steamid.SID3ToSID64(steamid.SID3(match[1]))
		p2 := steamid.SID3ToSID64(steamid.SID3(match[2]))
		dmg, _ := strconv.Atoi(match[3])
		_, ok := gi.Stats[p1]
		if !ok {
			gi.Stats[p1] = &PlayerStats{DamageDone: dmg}
		} else {
			gi.Stats[p1].DamageDone += dmg
		}
		_, ok = gi.Stats[p2]
		if !ok {
			gi.Stats[p2] = &PlayerStats{DamageTaken: dmg}
		} else {
			gi.Stats[p2].DamageTaken += dmg
		}
	case healsRegexp.MatchString(msg):
		match := healsRegexp.FindStringSubmatch(msg)
		p1 := steamid.SID3ToSID64(steamid.SID3(match[1]))
		p2 := steamid.SID3ToSID64(steamid.SID3(match[2]))
		heals, _ := strconv.Atoi(match[3])
		h, ok := gi.Stats[p1]
		if !ok {
			gi.Stats[p1] = &PlayerStats{Healed: heals}
		} else {
			h.Healed += heals
		}
		h, ok = gi.Stats[p2]
		if !ok {
			gi.Stats[p2] = &PlayerStats{HealsReceived: heals}
		} else {
			h.HealsReceived += heals
		}
	}
	return nil
}

func (lf *LogFile) ExtractPlayerStats() []interface{} {
	s := make([]interface{}, 0)
	for steamID, stats := range lf.Game.Stats {
		//for _, player := range lf.Game.Players {
		//	if steamID.String() == player.SteamID {
		gs := GameStats{
			Player:   PickupPlayer{SteamID: steamID.String()},
			Stats:    *stats,
			Server:   lf.Server,
			PickupID: lf.Game.PickupID,
		}
		s = append(s, gs)
	}
	//}
	//}
	return s
}

func (lf *LogFile) insertGameStats(documents []interface{}) error {
	_, err := lf.conn.
		Database(mongoDatabase).
		Collection(mongoCollection).
		InsertMany(lf.ctx, documents)
	return err
}
