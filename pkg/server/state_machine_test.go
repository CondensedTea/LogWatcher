package server_test

import (
	"LogWatcher/pkg/mocks"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/server"
	"LogWatcher/pkg/stats"
	"errors"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/leighmacdonald/steamid/steamid"
	"github.com/sirupsen/logrus"
)

func TestProcessGameStartedEvent(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	logger := logrus.New()
	logFilerMock := mocks.NewLogFilerMock(t)
	logProcessorMock := mocks.NewLogProcessorMock(t)
	matchDaterMock := mocks.NewMatchDaterMock(t)

	matchDaterMock.DomainMock.Return("test_domain")

	type args struct {
		msg string
		log *logrus.Logger
		lf  server.LogFiler
		lp  requests.LogUploader
		md  stats.MatchDater
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default",
			args: args{
				msg: `test`,
				log: logger,
				lf: logFilerMock.
					WriteLineMock.Expect("test").Return().
					SetStateMock.Expect(server.Game).Return().
					GetConnMock.Return(nil),
				lp: logProcessorMock.
					GetPickupGamesMock.Return(requests.GamesResponse{}, nil).
					ResolvePlayersMock.Return(nil),
				md: matchDaterMock.
					SetStartTimeMock.Expect("test").Return().
					StringMock.Return("").
					PickupPlayersMock.Return([]*stats.PickupPlayer{}).
					PickupIDMock.Return(0).
					MapMock.Return(""),
			},
		},
		{
			name: "UpdatePickupInfo error",
			args: args{
				msg: `test`,
				log: logger,
				lf: logFilerMock.
					WriteLineMock.Expect("test").Return().
					SetStateMock.Expect(server.Game).Return().
					GetConnMock.Return(nil),
				md: matchDaterMock.
					SetStartTimeMock.Expect("test").Return().
					StringMock.Return("test#1"),
				lp: logProcessorMock.
					GetPickupGamesMock.Expect("test_domain").Return(requests.GamesResponse{}, errors.New("error")),
			},
		},
		{
			name: "ResolverPlayers error",
			args: args{
				msg: `test`,
				log: logger,
				lf: logFilerMock.
					WriteLineMock.Expect("test").Return().
					SetStateMock.Expect(server.Game).Return().
					GetConnMock.Return(nil),
				lp: logProcessorMock.
					GetPickupGamesMock.Return(requests.GamesResponse{}, nil).
					ResolvePlayersMock.Return(errors.New("error")),
				md: matchDaterMock.
					SetStartTimeMock.Expect("test").Return().
					StringMock.Return("").
					PickupPlayersMock.Return([]*stats.PickupPlayer{}).
					PickupIDMock.Return(0).
					MapMock.Return(""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.ProcessGameStartedEvent(tt.args.msg, tt.args.log, tt.args.lf, tt.args.lp, tt.args.md)
		})
	}
}

func Test_processGameLogLine(t *testing.T) {
	mc := minimock.NewController(t)
	defer mc.Finish()

	logFilerMock := mocks.NewLogFilerMock(mc)
	matchDaterMock := mocks.NewMatchDaterMock(mc)

	type args struct {
		msg string
		lm  server.LogFiler
		gi  stats.MatchDater
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default",
			args: args{
				msg: `test`,
				lm: logFilerMock.
					WriteLineMock.Expect("test").Return(),
				gi: matchDaterMock.
					PlayerStatsMapMock.Return(stats.PlayerStatsCollection{
					steamid.SID64FromString("76561198439712695"): {Kills: 1},
				}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server.ProcessGameLogLine(tt.args.msg, tt.args.lm, tt.args.gi)
		})
	}
}
