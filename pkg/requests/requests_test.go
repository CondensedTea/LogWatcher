package requests

import (
	"io"
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
	gamesRawJSON   = `{"results":[{"connectInfoVersion":1,"state":"started","number":391,"map":"cp_granary_pro_rc8","slots":[{"connectionStatus":"","status":"","gameClass":"soldier","team":"red","player":"6133487c4573f9001cdc0abb"}],"launchedAt":"","gameServer":"","stvConnectString":"","logsUrl":"","id":"6154dddef56b5b0013b269a3"}]}`
)

func TestLogFile_ResolvePlayers(t *testing.T) {
	mc := minimock.NewController(t)
	type args struct {
		client  HTTPDoer
		domain  string
		players []*PickupPlayer
	}
	tests := []struct {
		name    string
		args    args
		want    []*PickupPlayer
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				client: NewHTTPDoerMock(mc).DoMock.Return(
					&http.Response{StatusCode: 200, Body: ioutil.NopCloser(strings.NewReader(playersRawJSON))}, nil),
				domain: "test",
				players: []*PickupPlayer{
					{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier"},
				},
			},
			want: []*PickupPlayer{{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier", SteamID: "76561198011558250"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ResolvePlayers(tt.args.client, tt.args.domain, tt.args.players); (err != nil) != tt.wantErr {
				t.Errorf("resolvePlayers() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(tt.args.players, tt.want) {
				t.Errorf("resolvePlayers() got = %v, want = %v", tt.args.players, tt.want)

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UploadLogFile(tt.args.client, tt.args.payload); (err != nil) != tt.wantErr {
				t.Errorf("UploadLogFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
