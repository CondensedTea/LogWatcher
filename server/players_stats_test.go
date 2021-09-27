package main

import (
	"reflect"
	"testing"

	"github.com/leighmacdonald/steamid/steamid"
)

func TestGameInfo_updatePlayerStats(t *testing.T) {
	type fields struct {
		PickupID int
		Map      string
		Players  []PickupPlayer
		Stats    map[steamid.SID64]*PlayerStats
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[steamid.SID64]*PlayerStats
		wantErr bool
	}{
		{
			name: "triggered healed plus",
			fields: fields{
				Stats: map[steamid.SID64]*PlayerStats{
					steamid.SID64FromString("76561198439712695"): {Healed: 49},
					steamid.SID64FromString("76561198821399014"): {HealsReceived: 49},
				},
			},
			args: args{`"jel<62><[U:1:479446967]><Blue>" triggered "healed" against "KEYREAL<65><[U:1:861133286]><Blue>" (healing "51")"`},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {Healed: 100},
				steamid.SID64FromString("76561198821399014"): {HealsReceived: 100},
			},
		},
		{
			name: "triggered healed new",
			fields: fields{
				Stats: map[steamid.SID64]*PlayerStats{},
			},
			args: args{`"jel<62><[U:1:479446967]><Blue>" triggered "healed" against "KEYREAL<65><[U:1:861133286]><Blue>" (healing "51")"`},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {Healed: 51},
				steamid.SID64FromString("76561198821399014"): {HealsReceived: 51},
			},
		},
		{
			name: "triggered damage new",
			fields: fields{
				Stats: map[steamid.SID64]*PlayerStats{},
			},
			args: args{`"jel<62><[U:1:479446967]><Blue>" triggered "damage" against "KEYREAL<65><[U:1:861133286]><Red>" (damage "30")"`},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {DamageDone: 30},
				steamid.SID64FromString("76561198821399014"): {DamageTaken: 30},
			},
		},
		{
			name: "triggered damage plus",
			fields: fields{
				Stats: map[steamid.SID64]*PlayerStats{
					steamid.SID64FromString("76561198439712695"): {DamageDone: 70},
					steamid.SID64FromString("76561198821399014"): {DamageTaken: 70},
				},
			},
			args: args{`"jel<62><[U:1:479446967]><Blue>" triggered "damage" against "KEYREAL<65><[U:1:861133286]><Red>" (damage "30")"`},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {DamageDone: 100},
				steamid.SID64FromString("76561198821399014"): {DamageTaken: 100},
			},
		},
		{
			name: "killed new",
			fields: fields{
				Stats: map[steamid.SID64]*PlayerStats{},
			},
			args: args{`"jel<62><[U:1:479446967]><Blue>" killed "KEYREAL<65><[U:1:861133286]><Red>" with "sniperrifle""`},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {Kills: 1},
				steamid.SID64FromString("76561198821399014"): {Deaths: 1},
			},
		},
		{
			name: "killed plus",
			fields: fields{
				Stats: map[steamid.SID64]*PlayerStats{
					steamid.SID64FromString("76561198439712695"): {Kills: 1},
					steamid.SID64FromString("76561198821399014"): {Deaths: 1},
				},
			},
			args: args{`"jel<62><[U:1:479446967]><Blue>" killed "KEYREAL<65><[U:1:861133286]><Red>" with "sniperrifle""`},
			want: map[steamid.SID64]*PlayerStats{
				steamid.SID64FromString("76561198439712695"): {Kills: 2},
				steamid.SID64FromString("76561198821399014"): {Deaths: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gi := &GameInfo{
				PickupID: tt.fields.PickupID,
				Map:      tt.fields.Map,
				Players:  tt.fields.Players,
				Stats:    tt.fields.Stats,
			}
			if err := gi.updatePlayerStats(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("updatePlayerStats() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(gi.Stats, tt.want) {
				t.Errorf("updatePlayerStats() got = %v, want = %v", gi.Stats, tt.want)
			}
		})
	}
}
