package server

import (
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/stats"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

const gamesRawJSON = `{"results":[{"connectInfoVersion":1,"state":"started","number":391,"map":"cp_granary_pro_rc8","slots":[{"connectionStatus":"","status":"","gameClass":"soldier","team":"red","player":"6133487c4573f9001cdc0abb"}],"launchedAt":"2021-09-29T21:42:54.745Z","gameServer":"","stvConnectString":"","logsUrl":"","id":"6154dddef56b5b0013b269a3"}]}`

func TestServer_updatePickupInfo(t *testing.T) {
	mc := minimock.NewController(t)
	type fields struct {
		log  *logrus.Logger
		Game *GameInfo
	}
	type args struct {
		client requests.HTTPDoer
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
				log:  logrus.New(),
				Game: &GameInfo{Map: "cp_granary_pro_rc8"},
			},
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(&http.Response{
					StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(gamesRawJSON)),
				}, nil),
			},
			want: &GameInfo{
				Map: "cp_granary_pro_rc8",
				Players: []*stats.PickupPlayer{
					{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier", Team: "red"},
				},
				PickupID: 391,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				log:  tt.fields.log,
				Game: tt.fields.Game,
			}
			if err := s.updatePickupInfo(tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("updatePickupInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(s.Game, tt.want) {
				t.Errorf("updatePickupInfo() got = %v, want = %v", s.Game, tt.want)
			}
		})
	}
}

func TestServer_Origin(t *testing.T) {
	type fields struct {
		Server stats.ServerInfo
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "default",
			fields: fields{
				Server: stats.ServerInfo{
					ID:     1,
					Domain: "test",
				},
			},
			want: "test#1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				Server: tt.fields.Server,
			}
			if got := s.Origin(); got != tt.want {
				t.Errorf("Origin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseTimeStamp(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "default",
			args: args{msg: `L 10/02/2021 - 23:31:56: \"Eshka<72><[U:1:183918108]><Red>\" triggered \"damage\" against \"slowtown<77><[U:1:148548823]><Blue>\"`},
			want: time.Unix(1633217516, 0).UTC(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTimeStamp(tt.args.msg)
			if got.Sub(tt.want) != 0 {
				t.Errorf("parseTimeStamp() got = %v, want %v", got, tt.want)
			}
		})
	}
}
