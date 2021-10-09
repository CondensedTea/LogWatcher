package stats_test

import (
	"LogWatcher/pkg/mocks"
	"LogWatcher/pkg/stats"
	"reflect"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/leighmacdonald/steamid/steamid"
)

func TestUpdateStatsMap(t *testing.T) {
	type args struct {
		msg   string
		stats map[steamid.SID64]*stats.PlayerGameStats
	}
	tests := []struct {
		name string
		args args
		want map[steamid.SID64]*stats.PlayerGameStats
	}{
		{
			name: "triggered healed plus",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" triggered "healed" against "KEYREAL<65><[U:1:861133286]><Blue>" (healing "51")"`,
				map[steamid.SID64]*stats.PlayerGameStats{
					steamid.SID64FromString("76561198439712695"): {Healed: 49},
					steamid.SID64FromString("76561198821399014"): {HealsReceived: 49},
				},
			},
			want: map[steamid.SID64]*stats.PlayerGameStats{
				steamid.SID64FromString("76561198439712695"): {Healed: 100},
				steamid.SID64FromString("76561198821399014"): {HealsReceived: 100},
			},
		},
		{
			name: "triggered healed new",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" triggered "healed" against "KEYREAL<65><[U:1:861133286]><Blue>" (healing "51")"`,
				map[steamid.SID64]*stats.PlayerGameStats{},
			},
			want: map[steamid.SID64]*stats.PlayerGameStats{
				steamid.SID64FromString("76561198439712695"): {Healed: 51},
				steamid.SID64FromString("76561198821399014"): {HealsReceived: 51},
			},
		},
		{
			name: "triggered damage new",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" triggered "damage" against "KEYREAL<65><[U:1:861133286]><Red>" (damage "30")"`,
				map[steamid.SID64]*stats.PlayerGameStats{},
			},
			want: map[steamid.SID64]*stats.PlayerGameStats{
				steamid.SID64FromString("76561198439712695"): {DamageDone: 30},
				steamid.SID64FromString("76561198821399014"): {DamageTaken: 30},
			},
		},
		{
			name: "triggered damage plus",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" triggered "damage" against "KEYREAL<65><[U:1:861133286]><Red>" (damage "30")"`,
				map[steamid.SID64]*stats.PlayerGameStats{
					steamid.SID64FromString("76561198439712695"): {DamageDone: 70},
					steamid.SID64FromString("76561198821399014"): {DamageTaken: 70},
				},
			},
			want: map[steamid.SID64]*stats.PlayerGameStats{
				steamid.SID64FromString("76561198439712695"): {DamageDone: 100},
				steamid.SID64FromString("76561198821399014"): {DamageTaken: 100},
			},
		},
		{
			name: "killed new",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" killed "KEYREAL<65><[U:1:861133286]><Red>" with "sniperrifle""`,
				map[steamid.SID64]*stats.PlayerGameStats{},
			},
			want: map[steamid.SID64]*stats.PlayerGameStats{
				steamid.SID64FromString("76561198439712695"): {Kills: 1},
				steamid.SID64FromString("76561198821399014"): {Deaths: 1},
			},
		},
		{
			name: "killed plus",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" killed "KEYREAL<65><[U:1:861133286]><Red>" with "sniperrifle""`,
				map[steamid.SID64]*stats.PlayerGameStats{
					steamid.SID64FromString("76561198439712695"): {Kills: 1},
					steamid.SID64FromString("76561198821399014"): {Deaths: 1},
				},
			},
			want: map[steamid.SID64]*stats.PlayerGameStats{
				steamid.SID64FromString("76561198439712695"): {Kills: 2},
				steamid.SID64FromString("76561198821399014"): {Deaths: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats.UpdateStatsMap(tt.args.msg, tt.args.stats)
			if !cmp.Equal(tt.args.stats, tt.want) {
				t.Errorf("updatePlayerStats() got = %v, want = %v", tt.args.stats, tt.want)
			}
		})
	}
}

func TestParseTimeStamp(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "default",
			args: args{msg: `L 10/02/2021 - 23:31:56: \"Eshka<72><[U:1:183918108]><Red>\" triggered \"damage\" against \"slowtown<77><[U:1:148548823]><Blue>\"`},
			want: time.Unix(1633217516, 0).UTC(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stats.ParseTimeStamp(tt.args.msg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTimeStamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractPlayerStats(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	matchDaterMock := mocks.NewMatchDaterMock(mc)

	type args struct {
		md stats.MatchDater
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "default",
			args: args{
				md: matchDaterMock.
					PickupPlayersMock.Return([]*stats.PickupPlayer{
					{SteamID: "76561198011558250"},
				}).
					PlayerStatsMapMock.Return(map[steamid.SID64]*stats.PlayerGameStats{
					steamid.SID64FromString("76561198011558250"): {Kills: 1},
				}).
					DomainMock.Return("test").
					PickupIDMock.Return(123).
					LengthSecondsMock.Return(100),
			},
			want: []interface{}{
				stats.MongoPlayerInfo{
					Player:   &stats.PickupPlayer{SteamID: "76561198011558250"},
					Stats:    stats.PlayerGameStats{Kills: 1},
					Domain:   "test",
					PickupID: 123,
					Length:   100,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stats.ExtractPlayerStats(tt.args.md); !cmp.Equal(got, tt.want) {
				t.Errorf("ExtractPlayerStats() = %v, want %v", got, tt.want)
			}
		})
	}
}
