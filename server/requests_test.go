package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	playersRawJSON = `[{"steamId":"76561198011558250","name":"supra","avatar":{"small":""},"id":"6133487c4573f9001cdc0abb","_links":[{"href":"/players/6133487c4573f9001cdc0abb/linked-profiles","title":"Linked profiles"}]}]`
	gamesRawJSON   = `{"results":[{"connectInfoVersion":1,"state":"started","number":391,"map":"cp_granary_pro_rc8","slots":[{"connectionStatus":"","status":"","gameClass":"soldier","team":"red","player":"6133487c4573f9001cdc0abb"}],"launchedAt":"2021-09-29T21:42:54.745Z","gameServer":"","stvConnectString":"","logsUrl":"","id":"6154dddef56b5b0013b269a3"}]}`
)

func TestLogFile_updatePickupInfo(t *testing.T) {
	mc := minimock.NewController(t)
	type fields struct {
		ctx     context.Context
		Mutex   sync.Mutex
		Server  ServerInfo
		State   StateType
		channel chan string
		buffer  bytes.Buffer
		Game    *GameInfo
		apiKey  string
		dryRun  bool
		conn    *mongo.Client
	}
	type args struct {
		client ClientInterface
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *GameInfo
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				Game: &GameInfo{
					Players: nil,
					Map:     "cp_granary_pro_rc8",
				},
			},
			args: args{
				NewClientInterfaceMock(mc).DoMock.Return(
					&http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(gamesRawJSON))}, nil,
				),
			},
			want: &GameInfo{
				PickupID: 391,
				Map:      "cp_granary_pro_rc8",
				Players: []PickupPlayer{
					{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lf := &LogFile{
				ctx:     tt.fields.ctx,
				Mutex:   tt.fields.Mutex,
				Server:  tt.fields.Server,
				State:   tt.fields.State,
				channel: tt.fields.channel,
				buffer:  tt.fields.buffer,
				Game:    tt.fields.Game,
				apiKey:  tt.fields.apiKey,
				dryRun:  tt.fields.dryRun,
				conn:    tt.fields.conn,
			}
			if err := lf.updatePickupInfo(tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("updatePickupInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(lf.Game, tt.want) {
				t.Errorf("updatePickupInfo() got = %v, want %v", lf.Game, tt.want)
			}
		})
	}
}

func TestLogFile_resolvePlayers(t *testing.T) {
	mc := minimock.NewController(t)
	type fields struct {
		ctx     context.Context
		Mutex   sync.Mutex
		Server  ServerInfo
		State   StateType
		channel chan string
		buffer  bytes.Buffer
		Game    *GameInfo
		apiKey  string
		dryRun  bool
		conn    *mongo.Client
	}
	type args struct {
		client ClientInterface
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []PickupPlayer
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				Game: &GameInfo{
					Players: []PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"}},
				},
			},
			args: args{NewClientInterfaceMock(mc).DoMock.Return(
				&http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(playersRawJSON))}, nil,
			)},
			want: []PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier", SteamID64: "76561198011558250"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lf := &LogFile{
				ctx:     tt.fields.ctx,
				Mutex:   tt.fields.Mutex,
				Server:  tt.fields.Server,
				State:   tt.fields.State,
				channel: tt.fields.channel,
				buffer:  tt.fields.buffer,
				Game:    tt.fields.Game,
				apiKey:  tt.fields.apiKey,
				dryRun:  tt.fields.dryRun,
				conn:    tt.fields.conn,
			}
			if err := lf.resolvePlayers(tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("resolvePlayers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(lf.Game.Players, tt.want) {
				t.Errorf("resolvePlayers() got = %v, want = %v", lf.Game.Players, tt.want)

			}
		})
	}
}
