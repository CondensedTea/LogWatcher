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
	Name     string `bson:"name"`
	Class    string `bson:"class"`
	SteamID  string `bson:"steam_id"`
	Team     string `bson:"team"`
}

// MongoPlayerInfo represents single player's data from single game,
// used as model for mongo entries.
// Schema version used for tracking new features.
type MongoPlayerInfo struct {
	Player        *PickupPlayer
	Stats         PlayerStats
	Domain        string
	PickupID      int
	Length        int
	SchemaVersion int
}

// CurrentScores represents teams score in single round
type CurrentScores struct {
	Red  int
	Blue int
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
	Scores      CurrentScores
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
	SetRedScore(score int)
	SetBlueScore(score int)
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

func (m *Match) PickupPlayers() []*PickupPlayer {
	return m.players
}

func (m *Match) PlayerStats() PlayerStatsCollection {
	return m.stats
}

func (m *Match) PickupID() int {
	return m.pickupID
}

func (m *Match) SetLength(msg string) {
	ts := ParseTimeStamp(msg)
	m.matchLength = ts.Sub(m.launchedAt)
}

func (m *Match) LengthSeconds() int {
	return int(m.matchLength.Seconds())
}

func (m *Match) Domain() string {
	return m.domain
}

func (m *Match) String() string {
	return fmt.Sprintf("%s#%d", m.domain, m.serverID)
}

func (m *Match) SetPickupID(id int) {
	m.pickupID = id
}

func (m *Match) SetPlayers(players []*PickupPlayer) {
	m.players = players
}

func (m *Match) SetStartTime(msg string) {
	ts := ParseTimeStamp(msg)
	m.launchedAt = ts
}

func (m *Match) SetMap(_map string) {
	m._map = _map
}

func (m *Match) Map() string {
	return m._map
}

func (m *Match) Flush() {
	m.pickupID = 0
	m._map = ""
	m.stats = make(PlayerStatsCollection)
}

// TryParseGameMap tries to find "Loading map" with regexp in message
// and sets it to Match._map if succeed
func (m *Match) TryParseGameMap(msg string) {
	if match := mapLoaded.FindStringSubmatch(msg); len(match) > 0 {
		m._map = match[1]
	}
}

func (m *Match) SetPlayerStats(stats PlayerStatsCollection) {
	m.stats = stats
}

func (m *Match) SetRedScore(score int) {
	m.Scores.Red = score
}

func (m *Match) SetBlueScore(score int) {
	m.Scores.Blue = score
}
