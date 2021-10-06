package requests

import (
	"LogWatcher/pkg/stats"
	"errors"
	"io"
	"time"

	//"LogWatcher/pkg/app"
	//"LogWatcher/pkg/stats"
	//"bytes"
	//"context"
	"io/ioutil"
	"net/http"
	"strings"
	//"sync"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/google/go-cmp/cmp"
	//"go.mongodb.org/mongo-driver/mongo"
)

const (
	playersRawJSON = `[{"steamId":"76561198011558250","name":"supra","avatar":{"small":""},"id":"6133487c4573f9001cdc0abb","_links":[{"href":"/players/6133487c4573f9001cdc0abb/linked-profiles","title":"Linked profiles"}]}]`
	gamesRawJSON   = `{"results":[{"connectInfoVersion":1,"state":"started","number":391,"map":"cp_granary_pro_rc8","slots":[{"connectionStatus":"","status":"","gameClass":"soldier","team":"red","player":"6133487c4573f9001cdc0abb"}],"launchedAt":"2021-09-29T21:42:54.745Z","gameServer":"","stvConnectString":"","logsUrl":"","id":"6154dddef56b5b0013b269a3"}]}`
)

func TestLogFile_ResolvePlayers(t *testing.T) {
	mc := minimock.NewController(t)
	type args struct {
		client  HTTPDoer
		domain  string
		players []*stats.PickupPlayer
	}
	tests := []struct {
		name    string
		args    args
		want    []*stats.PickupPlayer
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(playersRawJSON))}, nil),
				domain: "test",
				players: []*stats.PickupPlayer{
					{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"},
				},
			},
			want: []*stats.PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier", SteamID: "76561198011558250"}},
		},
		{
			name: "non 200 http response",
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 404, Body: nil}, nil),
				domain:  "test",
				players: []*stats.PickupPlayer{},
			},
			want:    []*stats.PickupPlayer{},
			wantErr: true,
		},
		{
			name: "error on client.Do",
			args: args{
				client:  NewHTTPDoerMock(mc).DoMock.Return(nil, errors.New("test error")),
				domain:  "test",
				players: []*stats.PickupPlayer{},
			},
			want:    []*stats.PickupPlayer{},
			wantErr: true,
		},
		{
			name: "invalid json response",
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(`{"bad: `))}, nil),
				domain:  "test",
				players: []*stats.PickupPlayer{},
			},
			want:    []*stats.PickupPlayer{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ResolvePlayers(tt.args.client, tt.args.domain, tt.args.players)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolvePlayers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(tt.args.players, tt.want) {
				t.Errorf("resolvePlayers() got = %#v, want = %#v", tt.args.players, tt.want)
			}
		})
	}
}

func TestUploadLogFile(t *testing.T) {
	mc := minimock.NewController(t)
	type args struct {
		client  HTTPDoer
		payload map[string]io.Reader
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 200}, nil,
				),
				payload: map[string]io.Reader{
					"logfile": strings.NewReader("file"),
					"map":     strings.NewReader("map"),
				},
			},
		},
		{
			name: "non 200 http status",
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 404, Body: ioutil.NopCloser(strings.NewReader(`{"error": "yes"}`))}, nil,
				),
				payload: map[string]io.Reader{},
			},
			wantErr: true,
		},
		{
			name: "error on http.Do",
			args: args{
				client:  NewHTTPDoerMock(mc).DoMock.Return(nil, errors.New("test error")),
				payload: map[string]io.Reader{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UploadLogFile(tt.args.client, tt.args.payload); (err != nil) != tt.wantErr {
				t.Errorf("UploadLogFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPickupGames(t *testing.T) {
	ts, _ := time.Parse("2006-01-02T15:04:05.999Z", "2021-09-29T21:42:54.745Z")
	mc := minimock.NewController(t)
	type args struct {
		client HTTPDoer
		domain string
	}
	tests := []struct {
		name    string
		args    args
		want    GamesResponse
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(&http.Response{
					StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(gamesRawJSON)),
				}, nil),
				domain: "test",
			},
			want: GamesResponse{
				Results: []Result{
					{
						ConnectInfoVersion: 1,
						State:              "started",
						Number:             391,
						Map:                "cp_granary_pro_rc8",
						Slots: []Slot{
							{
								GameClass: "soldier",
								Team:      "red",
								Player:    "6133487c4573f9001cdc0abb",
							},
						},
						LaunchedAt: ts,
						ID:         "6154dddef56b5b0013b269a3",
					},
				},
				ItemCount: 0,
			},
		},
		{
			name: "non 200 http response",
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 404, Body: nil}, nil),
				domain: "test",
			},
			want:    GamesResponse{},
			wantErr: true,
		},
		{
			name: "error on client.Do",
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(nil, errors.New("test error")),
				domain: "test",
			},
			want:    GamesResponse{},
			wantErr: true,
		},
		{
			name: "invalid json response",
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(`{"bad: "json"`))}, nil),
				domain: "test",
			},
			want:    GamesResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPickupGames(tt.args.client, tt.args.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPickupGames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("GetPickupGames() got = %v, want %v", got, tt.want)
			}
		})
	}
}
