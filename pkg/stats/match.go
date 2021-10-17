package stats

import (
	"LogWatcher/pkg/config"
	"fmt"
	"regexp"
	"time"

	"github.com/leighmacdonald/steamid/steamid"
)

var mapLoaded = regexp.MustCompile(`: Loading map "(.+?)"`)

// PlayerStats represents game stats from one player from single game
type PlayerStats struct {
	Kills         int
	Deaths        int
	DamageDone    int `bson:"damage_done"`
	DamageTaken   int `bson:"damage_taken"`
	Healed        int
	HealsReceived int `bson:"heals_received"`
}

// PickupPlayer represents player's information from single game
type PickupPlayer struct {
	PlayerID string `bson:"player_id"`
	Class    string `bson:"class"`
	SteamID  string `bson:"steam_id"`
	Team     string `bson:"team"`
}

// MongoPlayerInfo represents single player's data from single game,
// used as model for mongo entries
type MongoPlayerInfo struct {
	Player   *PickupPlayer
	Stats    PlayerStats
	Domain   string
	PickupID int
	Length   int
}

// Match represents data from all players in single game,
// including game info and player stats
type Match struct {
	pickupID    int
	serverID    int
	domain      string
	_map        string
	players     []*PickupPlayer
	stats       PlayerStatsCollection
	launchedAt  time.Time
	matchLength time.Duration
}

// Matcher is interface for Match object
type Matcher interface {
	String() string
	PickupPlayers() []*PickupPlayer
	PlayerStats() PlayerStatsCollection
	SetPlayerStats(stats PlayerStatsCollection)
	Domain() string
	PickupID() int
	SetPickupID(id int)
	SetStartTime(msg string)
	SetLength(msg string)
	LengthSeconds() int
	SetMap(m string)
	Map() string
	SetPlayers(players []*PickupPlayer)
	Flush()
	TryParseGameMap(msg string)
}

// PlayerStatsCollection represents game stats for all players from single game
type PlayerStatsCollection map[steamid.SID64]*PlayerStats

// NewMatch is a factory for Match
func NewMatch(host config.Client) *Match {
	return &Match{
		domain:   host.Domain,
		serverID: host.Server,
		stats:    make(PlayerStatsCollection),
	}
}

func (md *Match) PickupPlayers() []*PickupPlayer {
	return md.players
}

func (md *Match) PlayerStats() PlayerStatsCollection {
	return md.stats
}

func (md *Match) PickupID() int {
	return md.pickupID
}

func (md *Match) SetLength(msg string) {
	ts := ParseTimeStamp(msg)
	md.matchLength = ts.Sub(md.launchedAt)
}

func (md *Match) LengthSeconds() int {
	return int(md.matchLength.Seconds())
}

func (md *Match) Domain() string {
	return md.domain
}

func (md *Match) String() string {
	return fmt.Sprintf("%s#%d", md.domain, md.serverID)
}

func (md *Match) SetPickupID(id int) {
	md.pickupID = id
}

func (md *Match) SetPlayers(players []*PickupPlayer) {
	md.players = players
}

func (md *Match) SetStartTime(msg string) {
	ts := ParseTimeStamp(msg)
	md.launchedAt = ts
}

func (md *Match) SetMap(m string) {
	md._map = m
}

func (md *Match) Map() string {
	return md._map
}

func (md *Match) Flush() {
	md.pickupID = 0
	md._map = ""
	md.stats = make(PlayerStatsCollection)
}

// TryParseGameMap tries to find "Loading map" with regexp in message
// and sets it to go._map if succeed
func (md *Match) TryParseGameMap(msg string) {
	if match := mapLoaded.FindStringSubmatch(msg); len(match) > 0 {
		md._map = match[1]
	}
}

func (md *Match) SetPlayerStats(stats PlayerStatsCollection) {
	md.stats = stats
}
