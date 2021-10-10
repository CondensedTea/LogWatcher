package stats

import (
	"fmt"
	"regexp"
	"time"

	"github.com/leighmacdonald/steamid/steamid"
)

const (
	mongoDatabase   = "log_watcher"
	mongoCollection = "player_stats"
)

var mapLoaded = regexp.MustCompile(`: Loading map "(.+?)"`)

// PlayerGameStats represents stats set
// from one player from one game
type PlayerGameStats struct {
	Kills         int
	Deaths        int
	DamageDone    int `bson:"damage_done"`
	DamageTaken   int `bson:"damage_taken"`
	Healed        int
	HealsReceived int `bson:"heals_received"`
}

// PickupPlayer represents information about player in single game
type PickupPlayer struct {
	PlayerID string `bson:"player_id"`
	Class    string `bson:"class"`
	SteamID  string `bson:"steam_id"`
	Team     string `bson:"team"`
}

// MongoPlayerInfo represents single player's data from one game,
// used as model for mongo entries
type MongoPlayerInfo struct {
	Player   *PickupPlayer
	Stats    PlayerGameStats
	Domain   string
	PickupID int
	Length   int
}

// MatchData represents data from all players in single game,
// including game info and player stats
type MatchData struct {
	pickupID    int
	serverID    int
	domain      string
	_map        string
	players     []*PickupPlayer
	stats       map[steamid.SID64]*PlayerGameStats
	launchedAt  time.Time
	matchLength time.Duration
}

// MatchDater is interface for MatchData object
type MatchDater interface {
	String() string
	PickupPlayers() []*PickupPlayer
	PlayerStatsMap() map[steamid.SID64]*PlayerGameStats
	Domain() string
	PickupID() int
	SetPickupID(id int)
	SetStartTime(msg string)
	SetLength(msg string)
	LengthSeconds() int
	SetMap(m string)
	Map() string
	SetPlayers(players []*PickupPlayer)
	FlushPlayerStatsMap()
	TryParseGameMap(msg string)
}

// NewMatchData is a factory for MatchData
func NewMatchData(domain string, serverID int) *MatchData {
	return &MatchData{
		domain:   domain,
		serverID: serverID,
		stats:    make(map[steamid.SID64]*PlayerGameStats),
	}
}

func (gi *MatchData) PickupPlayers() []*PickupPlayer {
	return gi.players
}

func (gi *MatchData) PlayerStatsMap() map[steamid.SID64]*PlayerGameStats {
	return gi.stats
}

func (gi *MatchData) PickupID() int {
	return gi.pickupID
}

func (gi *MatchData) SetLength(msg string) {
	ts := ParseTimeStamp(msg)
	gi.matchLength = ts.Sub(gi.launchedAt)
}

func (gi *MatchData) LengthSeconds() int {
	return int(gi.matchLength.Seconds())
}

func (gi *MatchData) Domain() string {
	return gi.domain
}

func (gi *MatchData) String() string {
	return fmt.Sprintf("%s#%d", gi.domain, gi.serverID)
}

func (gi *MatchData) SetPickupID(id int) {
	gi.pickupID = id
}

func (gi *MatchData) SetPlayers(players []*PickupPlayer) {
	gi.players = players
}

func (gi *MatchData) SetStartTime(msg string) {
	ts := ParseTimeStamp(msg)
	gi.launchedAt = ts
}

func (gi *MatchData) SetMap(m string) {
	gi._map = m
}

func (gi *MatchData) Map() string {
	return gi._map
}

func (gi *MatchData) FlushPlayerStatsMap() {
	gi.stats = make(map[steamid.SID64]*PlayerGameStats)
}

// TryParseGameMap tries to find "Loading map" with regexp in message
// and sets it to go._map if succeed
func (gi *MatchData) TryParseGameMap(msg string) {
	if match := mapLoaded.FindStringSubmatch(msg); len(match) > 0 {
		gi._map = match[1]
	}
}
