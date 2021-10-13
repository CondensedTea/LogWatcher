package stats

import (
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

// MatchData represents data from all players in single game,
// including game info and player stats
type MatchData struct {
	pickupID    int
	serverID    int
	domain      string
	_map        string
	players     []*PickupPlayer
	stats       PlayerStatsCollection
	launchedAt  time.Time
	matchLength time.Duration
}

// MatchDater is interface for MatchData object
type MatchDater interface {
	String() string
	PickupPlayers() []*PickupPlayer
	PlayerStatsCollection() PlayerStatsCollection
	Domain() string
	PickupID() int
	SetPickupID(id int)
	SetStartTime(msg string)
	SetLength(msg string)
	LengthSeconds() int
	SetMap(m string)
	Map() string
	SetPlayers(players []*PickupPlayer)
	FlushPlayerStatsCollection()
	TryParseGameMap(msg string)
}

// PlayerStatsCollection represents game stats for all players from single game
type PlayerStatsCollection map[steamid.SID64]*PlayerStats

// NewMatchData is a factory for MatchData
func NewMatchData(domain string, serverID int) *MatchData {
	return &MatchData{
		domain:   domain,
		serverID: serverID,
		stats:    make(PlayerStatsCollection),
	}
}

func (md *MatchData) PickupPlayers() []*PickupPlayer {
	return md.players
}

func (md *MatchData) PlayerStatsCollection() PlayerStatsCollection {
	return md.stats
}

func (md *MatchData) PickupID() int {
	return md.pickupID
}

func (md *MatchData) SetLength(msg string) {
	ts := ParseTimeStamp(msg)
	md.matchLength = ts.Sub(md.launchedAt)
}

func (md *MatchData) LengthSeconds() int {
	return int(md.matchLength.Seconds())
}

func (md *MatchData) Domain() string {
	return md.domain
}

func (md *MatchData) String() string {
	return fmt.Sprintf("%s#%d", md.domain, md.serverID)
}

func (md *MatchData) SetPickupID(id int) {
	md.pickupID = id
}

func (md *MatchData) SetPlayers(players []*PickupPlayer) {
	md.players = players
}

func (md *MatchData) SetStartTime(msg string) {
	ts := ParseTimeStamp(msg)
	md.launchedAt = ts
}

func (md *MatchData) SetMap(m string) {
	md._map = m
}

func (md *MatchData) Map() string {
	return md._map
}

func (md *MatchData) FlushPlayerStatsCollection() {
	md.stats = make(PlayerStatsCollection)
}

// TryParseGameMap tries to find "Loading map" with regexp in message
// and sets it to go._map if succeed
func (md *MatchData) TryParseGameMap(msg string) {
	if match := mapLoaded.FindStringSubmatch(msg); len(match) > 0 {
		md._map = match[1]
	}
}
