package server

import (
	"LogWatcher/pkg/requests"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

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
				client: requests.NewHTTPDoerMock(mc).DoMock.Return(&http.Response{
					StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(gamesRawJSON)),
				}, nil),
			},
			want: &GameInfo{
				Map: "cp_granary_pro_rc8",
				Players: []*requests.PickupPlayer{
					{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"},
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