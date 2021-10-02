package stats

import (
	//"LogWatcher/pkg/app"
	//"bytes"
	//"context"
	//"reflect"
	//"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/leighmacdonald/steamid/steamid"
	//"go.mongodb.org/mongo-driver/mongo"
)

func TestGameInfo_UpdatePlayerStats(t *testing.T) {
	type args struct {
		msg   string
		stats map[steamid.SID64]*PlayerStats
	}
	tests := []struct {
		name    string
		args    args
		want    map[steamid.SID64]*PlayerStats
		wantErr bool
	}{
		{
			name: "triggered healed plus",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" triggered "healed" against "KEYREAL<65><[U:1:861133286]><Blue>" (healing "51")"`,
				map[steamid.SID64]*PlayerStats{
					steamid.SID64FromString("76561198439712695"): {Healed: 49},
					steamid.SID64FromString("76561198821399014"): {HealsReceived: 49},
				},
			},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {Healed: 100},
				steamid.SID64FromString("76561198821399014"): {HealsReceived: 100},
			},
		},
		{
			name: "triggered healed new",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" triggered "healed" against "KEYREAL<65><[U:1:861133286]><Blue>" (healing "51")"`,
				map[steamid.SID64]*PlayerStats{},
			},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {Healed: 51},
				steamid.SID64FromString("76561198821399014"): {HealsReceived: 51},
			},
		},
		{
			name: "triggered damage new",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" triggered "damage" against "KEYREAL<65><[U:1:861133286]><Red>" (damage "30")"`,
				map[steamid.SID64]*PlayerStats{},
			},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {DamageDone: 30},
				steamid.SID64FromString("76561198821399014"): {DamageTaken: 30},
			},
		},
		{
			name: "triggered damage plus",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" triggered "damage" against "KEYREAL<65><[U:1:861133286]><Red>" (damage "30")"`,
				map[steamid.SID64]*PlayerStats{
					steamid.SID64FromString("76561198439712695"): {DamageDone: 70},
					steamid.SID64FromString("76561198821399014"): {DamageTaken: 70},
				},
			},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {DamageDone: 100},
				steamid.SID64FromString("76561198821399014"): {DamageTaken: 100},
			},
		},
		{
			name: "killed new",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" killed "KEYREAL<65><[U:1:861133286]><Red>" with "sniperrifle""`,
				map[steamid.SID64]*PlayerStats{},
			},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {Kills: 1},
				steamid.SID64FromString("76561198821399014"): {Deaths: 1},
			},
		},
		{
			name: "killed plus",
			args: args{
				`"jel<62><[U:1:479446967]><Blue>" killed "KEYREAL<65><[U:1:861133286]><Red>" with "sniperrifle""`,
				map[steamid.SID64]*PlayerStats{
					steamid.SID64FromString("76561198439712695"): {Kills: 1},
					steamid.SID64FromString("76561198821399014"): {Deaths: 1},
				},
			},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {Kills: 2},
				steamid.SID64FromString("76561198821399014"): {Deaths: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdatePlayerStats(tt.args.msg, tt.args.stats); (err != nil) != tt.wantErr {
				t.Errorf("updatePlayerStats() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(tt.args.stats, tt.want) {
				t.Errorf("updatePlayerStats() got = %v, want = %v", tt.args.stats, tt.want)
			}
		})
	}
}

func TestLogFile_ExtractPlayerStats(t *testing.T) {
	type args struct {
		players   []*PickupPlayer
		gameStats map[steamid.SID64]*PlayerStats
		server    ServerInfo
		pickupID  int
	}
	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{
			name: "default",
			args: args{
				players: []*PickupPlayer{
					{SteamID: "76561198011558250", Class: "soldier", PlayerID: "0"},
				},
				gameStats: map[steamid.SID64]*PlayerStats{
					steamid.SID64FromString("76561198011558250"): {Kills: 1},
				},
				server: ServerInfo{
					ID:     1,
					Domain: "test",
					IP:     "test",
				},
				pickupID: 123,
			},
			want: []interface{}{
				GameStats{
					Player:   &PickupPlayer{SteamID: "76561198011558250", Class: "soldier", PlayerID: "0"},
					Stats:    PlayerStats{Kills: 1},
					PickupID: 123,
					Server: ServerInfo{
						ID:     1,
						Domain: "test",
						IP:     "test",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractPlayerStats(tt.args.players, tt.args.gameStats, tt.args.server, tt.args.pickupID); !cmp.Equal(got, tt.want) {
				t.Errorf("ExtractPlayerStats() = %v, want %v", got, tt.want)
			}
		})
	}
}
