package main

import (
	"regexp"
	"strconv"

	"github.com/leighmacdonald/steamid/steamid"
)

const insertPlayerStatsQuery = `insert into stats(domain,
												server_id,
												pickup_id,
												steamid64,
												class,
												kills,
												death,
												damage_done,
												damage_taken,
												heals,
												heals_received) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

var (
	killRegexp   = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+killed.+(\[U:\d:\d{1,10}])`)
	damageRegexp = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+triggered "damage" against.+(\[U:\d:\d{1,10}]).+\(damage "(\d+)"\)`)
	healsRegexp  = regexp.MustCompile(`(\[U:\d:\d{1,10}]).+triggered "healed" against.+(\[U:\d:\d{1,10}]).+\(healing "(\d+)"\)`)
)

type PlayerStats struct {
	Kills         int
	Deaths        int
	DamageDone    int
	DamageTaken   int
	Healed        int
	HealsReceived int
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

func (lf *LogFile) insertPlayerStats() error {
	tx, err := lf.conn.Begin()
	if err != nil {
		return err
	}
	for steamID64, stat := range lf.Game.Stats {
		for _, player := range lf.Game.Players {
			if steamID64.String() == player.SteamID64 {
				_, err := tx.Exec(
					insertPlayerStatsQuery,
					lf.Server.Domain,
					lf.Server.ID,
					lf.Game.PickupID,
					player.SteamID64,
					player.Class,
					stat.Kills,
					stat.Deaths,
					stat.DamageDone,
					stat.DamageTaken,
					stat.Healed,
					stat.HealsReceived,
				)
				if err != nil {
					return err
				}
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
