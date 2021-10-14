package server_test

import (
	"LogWatcher/pkg/mocks"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/server"
	"LogWatcher/pkg/stats"
	"errors"
	"testing"

	"github.com/gojuno/minimock/v3"
)

func TestUpdatePickupInfo(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	logProcessorMock := mocks.NewLogProcessorMock(mc)
	matchDaterMock := mocks.NewMatchDaterMock(mc)

	type args struct {
		r  requests.LogUploader
		gi stats.MatchDater
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				r: logProcessorMock.
					GetPickupGamesMock.Expect("test").Return(requests.GamesResponse{
					Results: []requests.Result{
						{
							State: "started",
							Map:   "cp_granary_pro_rc8",
							Slots: []requests.Slot{
								{
									Status:    "connected",
									Player:    "6133487c4573f9001cdc0abb",
									Team:      "red",
									GameClass: "soldier",
								},
							},
							Number: 123,
						},
					},
				}, nil),
				gi: matchDaterMock.
					DomainMock.Return("test").
					MapMock.Return("cp_granary_pro_rc8").
					SetPlayersMock.Expect([]*stats.PickupPlayer{
					{PlayerID: "6133487c4573f9001cdc0abb", Class: "soldier", Team: "red"},
				}).Return().
					SetPickupIDMock.Expect(123).Return(),
			},
		},
		{
			name: "get pickup games error",
			args: args{
				r:  mocks.NewLogProcessorMock(mc).GetPickupGamesMock.Expect("test").Return(requests.GamesResponse{}, errors.New("test")),
				gi: mocks.NewMatchDaterMock(mc).DomainMock.Return("test"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := server.UpdatePickupInfo(tt.args.r, tt.args.gi); (err != nil) != tt.wantErr {
				t.Errorf("updatePickupInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStateType_String(t *testing.T) {
	tests := []struct {
		name string
		st   server.StateType
		want string
	}{
		{
			name: "pregame",
			st:   server.Pregame,
			want: "pregame",
		},
		{
			name: "game",
			st:   server.Game,
			want: "game",
		},
		{
			name: "unknown",
			st:   server.StateType(3),
			want: "unknown state",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.st.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlush(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	logProcessorMock := mocks.NewLogFilerMock(mc)
	matchDaterMock := mocks.NewMatchDaterMock(mc)

	type args struct {
		lf server.LogFiler
		md stats.MatchDater
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default",
			args: args{
				lf: logProcessorMock.FlushBufferMock.
					Return(),
				md: matchDaterMock.
					SetPickupIDMock.Expect(0).Return().
					SetMapMock.Expect("").Return().
					FlushPlayerStatsMapMock.Return(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.Flush(tt.args.lf, tt.args.md)
		})
	}
}
