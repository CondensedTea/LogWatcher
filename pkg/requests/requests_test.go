package requests_test

import (
	"LogWatcher/pkg/mocks"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/stats"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

const (
	playersRawJSON = `[{"steamId":"76561198011558250","name":"supra","avatar":{"small":""},"ID":"6133487c4573f9001cdc0abb","_links":[{"href":"/Players/6133487c4573f9001cdc0abb/linked-profiles","title":"Linked profiles"}]}]`
	gamesRawJSON   = `{"results":[{"connectInfoVersion":1,"state":"started","number":391,"map":"cp_granary_pro_rc8","slots":[{"connectionStatus":"","status":"","gameClass":"soldier","team":"red","player":"6133487c4573f9001cdc0abb"}],"launchedAt":"2021-09-29T21:42:54.745Z","gameServer":"","stvConnectString":"","logsUrl":"","ID":"6154dddef56b5b0013b269a3"}]}`
)

func TestNewRequestManager(t *testing.T) {
	log := logrus.New()

	type args struct {
		apiKey string
		client requests.HTTPDoer
	}
	tests := []struct {
		name string
		args args
		want *requests.Client
	}{
		{
			name: "default",
			args: args{
				apiKey: "test",
				client: &http.Client{},
			},
			want: &requests.Client{
				Client: &http.Client{},
				ApiKey: "test",
				Log:    log,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requests.NewClient(tt.args.apiKey, tt.args.client, log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequester_MakeMultipartMap(t *testing.T) {
	mc := minimock.NewController(t)

	type fields struct {
		client requests.HTTPDoer
		apiKey string
	}
	type args struct {
		match stats.Matcher
		buf   bytes.Buffer
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]io.Reader
	}{
		{
			name:   "default",
			fields: fields{apiKey: "test"},
			args: args{
				match: mocks.NewMatcherMock(mc).
					MapMock.Return("cp_granary_rc8").
					DomainMock.Return("test").
					PickupIDMock.Return(123),
				buf: bytes.Buffer{},
			},
			want: map[string]io.Reader{
				"title":    strings.NewReader("tf2pickup.test #123"),
				"map":      strings.NewReader("cp_granary_rc8"),
				"key":      strings.NewReader("test"),
				"logfile":  &bytes.Buffer{},
				"uploader": strings.NewReader("LogWatcher dev"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &requests.Client{
				Client: tt.fields.client,
				ApiKey: tt.fields.apiKey,
			}
			if got := r.MakeMultipartMap(tt.args.match, tt.args.buf); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeMultipartMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequester_ResolvePlayers(t *testing.T) {
	mc := minimock.NewController(t)
	type fields struct {
		client requests.HTTPDoer
		apiKey string
	}
	type args struct {
		domain  string
		players []*stats.PickupPlayer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*stats.PickupPlayer
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				client: mocks.NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(playersRawJSON))}, nil),
				apiKey: "test",
			},
			args: args{
				domain: "test",
				players: []*stats.PickupPlayer{
					{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"},
				},
			},
			want: []*stats.PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Name: "supra", Class: "soldier", SteamID: "76561198011558250"}},
		},
		{
			name: "non 200 http response",
			fields: fields{
				client: mocks.NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 404, Body: nil}, nil),
				apiKey: "test",
			},
			args: args{
				domain:  "test",
				players: []*stats.PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"}},
			},
			want:    []*stats.PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"}},
			wantErr: true,
		},
		{
			name: "error on Client.Do",
			fields: fields{
				client: mocks.NewHTTPDoerMock(mc).DoMock.Return(nil, errors.New("test error")),
				apiKey: "test",
			},
			args: args{
				domain:  "test",
				players: []*stats.PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"}},
			},
			want:    []*stats.PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"}},
			wantErr: true,
		},
		{
			name: "invalid json response",
			fields: fields{
				client: mocks.NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(`{"bad: `))}, nil),
				apiKey: "test",
			},
			args: args{
				domain:  "test",
				players: []*stats.PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"}},
			},
			want:    []*stats.PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &requests.Client{
				Client: tt.fields.client,
				ApiKey: tt.fields.apiKey,
			}
			if err := r.ResolvePlayers(tt.args.domain, tt.args.players); (err != nil) != tt.wantErr {
				t.Errorf("ResolvePlayers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(tt.args.players, tt.want) {
				t.Errorf("resolvePlayers() got = %#v, want = %#v", tt.args.players, tt.want)
			}
		})
	}
}

func TestRequester_UploadLogFile(t *testing.T) {
	mc := minimock.NewController(t)
	type fields struct {
		client requests.HTTPDoer
		apiKey string
	}
	type args struct {
		payload map[string]io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				client: mocks.NewHTTPDoerMock(mc).DoMock.Return(&http.Response{StatusCode: 200}, nil),
				apiKey: "test",
			},
			args: args{
				payload: map[string]io.Reader{
					"logfile": strings.NewReader("file"),
					"map":     strings.NewReader("map"),
				},
			},
		},
		{
			name: "non 200 http status",
			fields: fields{
				client: mocks.NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{
						StatusCode: 404,
						Body:       ioutil.NopCloser(strings.NewReader(`{"error": "yes"}`)),
					}, nil,
				),
				apiKey: "test",
			},
			args: args{
				payload: map[string]io.Reader{},
			},
			wantErr: true,
		},
		{
			name: "error on http.Do",
			fields: fields{
				client: mocks.NewHTTPDoerMock(mc).DoMock.Return(nil, errors.New("test error")),
				apiKey: "test",
			},
			args: args{
				payload: map[string]io.Reader{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &requests.Client{
				Client: tt.fields.client,
				ApiKey: tt.fields.apiKey,
			}
			if err := r.UploadLogFile(tt.args.payload); (err != nil) != tt.wantErr {
				t.Errorf("UploadLogFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_GetPickupGames(t *testing.T) {
	mc := minimock.NewController(t)
	ts, _ := time.Parse("2006-01-02T15:04:05.999Z", "2021-09-29T21:42:54.745Z")

	type args struct {
		domain string
		client requests.HTTPDoer
	}
	tests := []struct {
		name    string
		args    args
		want    requests.GamesResponse
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				domain: "test",
				client: mocks.NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{
						StatusCode: 200,
						Body:       ioutil.NopCloser(strings.NewReader(gamesRawJSON)),
					}, nil),
			},
			want: requests.GamesResponse{
				Results: []requests.Result{
					{
						ConnectInfoVersion: 1,
						State:              "started",
						Number:             391,
						Map:                "cp_granary_pro_rc8",
						Slots: []requests.Slot{
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
			name: "err on client.Do",
			args: args{
				domain: "test",
				client: mocks.NewHTTPDoerMock(mc).
					DoMock.Return(nil, errors.New("test err")),
			},
			want:    requests.GamesResponse{},
			wantErr: true,
		},
		{
			name: "non-2xx status code",
			args: args{
				domain: "test",
				client: mocks.NewHTTPDoerMock(mc).
					DoMock.Return(&http.Response{StatusCode: 500}, nil),
			},
			want:    requests.GamesResponse{},
			wantErr: true,
		},
		{
			name: "invalid json response",
			args: args{
				domain: "test",
				client: mocks.NewHTTPDoerMock(mc).
					DoMock.Return(&http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(`{"bad: "json"`))}, nil),
			},
			want:    requests.GamesResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := requests.GetPickupGames(tt.args.domain, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPickupGames() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPickupGames() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_FindMatchingPickup(t *testing.T) {
	mc := minimock.NewController(t)

	httoDoerMock := mocks.NewHTTPDoerMock(mc)
	httoDoerMock.DoMock.Return(
		&http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(strings.NewReader(gamesRawJSON)),
		}, nil)
	type fields struct {
		Client requests.HTTPDoer
		ApiKey string
	}
	type args struct {
		domain  string
		gameMap string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *requests.Pickup
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				Client: httoDoerMock,
			},
			args: args{domain: "test", gameMap: "cp_granary_pro_rc8"},
			want: &requests.Pickup{Players: []*stats.PickupPlayer{
				{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier", Team: "red"},
			},
				ID: 391,
			},
		},
		{
			name: "error on GetPickupGames",
			fields: fields{Client: mocks.NewHTTPDoerMock(mc).
				DoMock.Return(&http.Response{StatusCode: 500}, nil),
			},
			args:    args{domain: "test", gameMap: "cp_granary_pro_rc8"},
			want:    &requests.Pickup{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &requests.Client{
				Client: tt.fields.Client,
				ApiKey: tt.fields.ApiKey,
				Log:    logrus.New(),
			}
			got, err := c.FindMatchingPickup(tt.args.domain, tt.args.gameMap)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindMatchingPickup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("FindMatchingPickup() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
