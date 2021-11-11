package stats

import (
	"LogWatcher/pkg/config"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/leighmacdonald/steamid/steamid"
)

func TestMatchData_Domain(t *testing.T) {
	type fields struct {
		domain string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "default",
			fields: fields{domain: "test"},
			want:   "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{
				domain: tt.fields.domain,
			}
			if got := gi.Domain(); got != tt.want {
				t.Errorf("Domain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchData_FlushPlayerStatsMap(t *testing.T) {
	s := make(PlayerStatsCollection)
	s[steamid.SID64FromString("")] = &PlayerStats{Kills: 1}
	type fields struct {
		stats PlayerStatsCollection
	}
	tests := []struct {
		name   string
		fields fields
		want   PlayerStatsCollection
	}{
		{
			name:   "default",
			fields: fields{stats: s},
			want:   make(PlayerStatsCollection),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{stats: tt.fields.stats}
			gi.Flush()
			if !reflect.DeepEqual(gi.stats, tt.want) {
				t.Errorf("Flush(), got = %#v, want = %#v", gi.stats, tt.want)
			}
		})
	}
}

func TestMatchData_LengthSeconds(t *testing.T) {
	type fields struct {
		matchLength time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "default",
			fields: fields{matchLength: time.Minute},
			want:   60,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{
				matchLength: tt.fields.matchLength,
			}
			if got := gi.LengthSeconds(); got != tt.want {
				t.Errorf("LengthSeconds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchData_Map(t *testing.T) {
	type fields struct {
		_map string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "default",
			fields: fields{"cp_granary_pro_rc8"},
			want:   "cp_granary_pro_rc8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{
				_map: tt.fields._map,
			}
			if got := gi.Map(); got != tt.want {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchData_PickupID(t *testing.T) {
	type fields struct {
		pickupID int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "default",
			fields: fields{pickupID: 123},
			want:   123,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{
				pickupID: tt.fields.pickupID,
			}
			if got := gi.PickupID(); got != tt.want {
				t.Errorf("PickupID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchData_PickupPlayers(t *testing.T) {
	type fields struct {
		players []*PickupPlayer
	}
	tests := []struct {
		name   string
		fields fields
		want   []*PickupPlayer
	}{
		{
			name: "default",
			fields: fields{players: []*PickupPlayer{
				{PlayerID: "test", Class: "soldier", SteamID: "test", Team: "red"},
			}},
			want: []*PickupPlayer{{PlayerID: "test", Class: "soldier", SteamID: "test", Team: "red"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{
				players: tt.fields.players,
			}
			if got := gi.PickupPlayers(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PickupPlayers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchData_PlayerStatsMap(t *testing.T) {
	type fields struct {
		stats PlayerStatsCollection
	}
	tests := []struct {
		name   string
		fields fields
		want   PlayerStatsCollection
	}{
		{
			name: "default",
			fields: fields{PlayerStatsCollection{
				steamid.SID64FromString("76561198061825334"): {Kills: 1},
			}},
			want: PlayerStatsCollection{
				steamid.SID64FromString("76561198061825334"): {Kills: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{
				stats: tt.fields.stats,
			}
			if got := gi.PlayerStats(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PlayerStatsCollection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchData_SetLength(t *testing.T) {
	type fields struct {
		launchedAt  time.Time
		matchLength time.Duration
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name:   "default",
			fields: fields{launchedAt: time.Unix(1633217416, 0).UTC()},
			args: args{
				msg: `L 10/02/2021 - 23:31:56: \"Eshka<72><[U:1:183918108]><Red>\" triggered \"damage\" against \"slowtown<77><[U:1:148548823]><Blue>\"`,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{
				launchedAt: tt.fields.launchedAt,
			}
			gi.SetLength(tt.args.msg)
			if gi.LengthSeconds() != tt.want {
				t.Errorf("SetLength() = %v, want %v", gi.matchLength, tt.want)
			}
		})
	}
}

func TestMatchData_SetMap(t *testing.T) {
	type fields struct {
		_map string
	}
	type args struct {
		m string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "default",
			fields: fields{_map: ""},
			args:   args{m: "cp_granary_pro_rc8"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{_map: tt.fields._map}
			gi.SetMap(tt.args.m)
			if gi._map != tt.args.m {
				t.Errorf("SetMap() = %v, want %v", gi.matchLength, tt.args.m)
			}
		})
	}
}

func TestMatchData_SetPickupID(t *testing.T) {
	type fields struct {
		pickupID int
	}
	type args struct {
		id int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "default",
			fields: fields{pickupID: 0},
			args:   args{id: 123},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{pickupID: tt.fields.pickupID}
			gi.SetPickupID(tt.args.id)
			if gi.pickupID != tt.args.id {
				t.Errorf("SetPickupID() = %v, want %v", gi.matchLength, tt.args.id)
			}
		})
	}
}

func TestMatchData_SetPlayers(t *testing.T) {
	type fields struct {
		players []*PickupPlayer
	}
	type args struct {
		players []*PickupPlayer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "default",
			fields: fields{players: []*PickupPlayer{}},
			args: args{[]*PickupPlayer{
				{PlayerID: "123"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{
				players: tt.fields.players,
			}
			gi.SetPlayers(tt.args.players)
			if !cmp.Equal(gi.players, tt.args.players) {
				t.Errorf("SetPlayers() = %v, want %v", gi.matchLength, &tt.args.players)
			}
		})
	}
}

func TestMatchData_SetStartTime(t *testing.T) {
	type fields struct {
		launchedAt time.Time
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   time.Time
	}{
		{
			name:   "default",
			fields: fields{launchedAt: time.Time{}},
			args:   args{msg: `L 10/02/2021 - 23:31:56: \"Eshka<72><[U:1:183918108]><Red>\" triggered \"damage\" against \"slowtown<77><[U:1:148548823]><Blue>\"`},
			want:   time.Unix(1633217516, 0).UTC(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{
				launchedAt: tt.fields.launchedAt,
			}
			gi.SetStartTime(tt.args.msg)
			if gi.launchedAt != tt.want {
				t.Errorf("SetStartTime() = %v, want %v", gi.launchedAt, tt.want)

			}
		})
	}
}

func TestMatchData_String(t *testing.T) {
	type fields struct {
		serverID int
		domain   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "default",
			fields: fields{serverID: 1, domain: "test"},
			want:   "test#1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &Match{
				serverID: tt.fields.serverID,
				domain:   tt.fields.domain,
			}
			if got := gi.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchData_TryParseGameMap(t *testing.T) {
	type fields struct {
		_map string
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "default",
			fields: fields{_map: ""},
			args:   args{msg: `: Loading map "cp_granary_pro_rc8"`},
			want:   "cp_granary_pro_rc8",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := Match{
				_map: tt.fields._map,
			}
			gi.TryParseGameMap(tt.args.msg)
			if gi._map != tt.want {
				t.Errorf("TryParseGameMap() = %v, want %v", gi._map, tt.want)
			}
		})
	}
}

func TestNewMatchData(t *testing.T) {
	type args struct {
		host config.Client
	}
	tests := []struct {
		name string
		args args
		want *Match
	}{
		{
			name: "default",
			args: args{
				config.Client{
					Server: 1,
					Domain: "test",
				},
			},
			want: &Match{
				domain:   "test",
				serverID: 1,
				stats:    PlayerStatsCollection{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMatch(tt.args.host); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatch_SetPlayerStats(t *testing.T) {
	type fields struct {
		stats PlayerStatsCollection
	}
	type args struct {
		stats PlayerStatsCollection
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "default",
			fields: fields{map[steamid.SID64]*PlayerStats{}},
			args: args{map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {Kills: 1},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := &Match{
				stats: tt.fields.stats,
			}
			md.SetPlayerStats(tt.args.stats)
			if !cmp.Equal(md.stats, tt.args.stats) {
				t.Errorf("NewMatch() = %v, want %v", tt.fields.stats, tt.args.stats)
			}
		})
	}
}

func TestMatch_SetRedScore(t *testing.T) {
	type fields struct {
		Scores CurrentScores
	}
	type args struct {
		score int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "default",
			fields: fields{},
			args:   args{score: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Match{Scores: tt.fields.Scores}
			m.SetRedScore(tt.args.score)
			if !cmp.Equal(m.Scores.Red, tt.args.score) {
				t.Errorf("NewMatch() = %v, want %v", tt.fields.Scores.Red, tt.args.score)
			}
		})
	}
}

func TestMatch_SetBlueScore(t *testing.T) {
	type fields struct {
		Scores CurrentScores
	}
	type args struct {
		score int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "default",
			fields: fields{},
			args:   args{score: 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Match{Scores: tt.fields.Scores}
			m.SetBlueScore(tt.args.score)
			if !cmp.Equal(m.Scores.Blue, tt.args.score) {
				t.Errorf("NewMatch() = %v, want %v", tt.fields.Scores.Red, tt.args.score)
			}
		})
	}
}
