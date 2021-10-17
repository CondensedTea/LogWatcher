package stateMachine_test

import (
	"LogWatcher/pkg/mocks"
	"LogWatcher/pkg/mongo"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/server"
	"LogWatcher/pkg/stateMachine"
	"LogWatcher/pkg/stats"
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/leighmacdonald/steamid/steamid"
	"github.com/sirupsen/logrus"
)

//func TestStateMachine_ProcessGameLogLine(t *testing.T) {
//	msg := `: World triggered "Round_Start"`
//	mc := minimock.NewController(t)
//	log := logrus.New()
//
//	file := mocks.NewLogFilerMock(mc)
//	uploader := mocks.NewLogUploaderMock(mc)
//	matcherMock := mocks.NewMatcherMock(mc)
//	inserterMock := &mongo.Mongo{}
//
//	_ = stateMachine.NewStateMachine(log, file, uploader, matcherMock, inserterMock)
//
//	stateMachineMock := mocks.NewStaterMock(mc)
//
//	sm := stateMachine.StateMachine{}
//
//	//stateMachineMock.
//	//	ProcessLogLineMock.Expect(msg).Return().
//	//	StateMock.Expect().Return(stateMachine.Pregame)
//	stateMachineMock.
//		//ProcessLogLineMock.Expect(msg).Return().
//		StateMock.Expect().Return(stateMachine.Pregame).
//		MatchMock.Expect().Return(matcherMock).
//		ProcessGameStartedEventMock.Expect(msg).Return()
//	//matcherMock.TryParseGameMapMock.Expect(msg).Return()
//
//	defer mc.Finish()
//
//	type args struct {
//		msg string
//	}
//	tests := []struct {
//		name     string
//		args     args
//		receiver stateMachine.Stater
//	}{
//		{
//			name: "game started",
//			args: args{msg: msg},
//			receiver: stateMachineMock.
//				ProcessLogLineMock.Expect(msg).Return().
//				StateMock.Expect().Return(stateMachine.Pregame).
//				MatchMock.Expect().Return(matcherMock).
//				ProcessGameStartedEventMock.Expect(msg).Return(),
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			sm.ProcessLogLine(tt.args.msg)
//		})
//	}
//}

func TestStateMachine_ProcessLogLine(t *testing.T) {
	log := logrus.New()

	mc := minimock.NewController(t)
	defer mc.Finish()

	type fields struct {
		State    stateMachine.StateType
		log      *logrus.Logger
		File     server.LogFiler
		Uploader requests.LogUploader
		Match    stats.Matcher
		Mongo    mongo.Inserter
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "pregame-Round_Start",
			args: args{msg: `: World triggered "Round_Start"`},
			fields: fields{
				State: stateMachine.Pregame,
				log:   log,
				File: mocks.NewLogFilerMock(mc).
					WriteLineMock.Expect(`: World triggered "Round_Start"`).Return(),
				Uploader: mocks.NewLogUploaderMock(mc).
					ResolvePlayersMock.Expect("test",
					[]*stats.PickupPlayer{
						{PlayerID: "123", Class: "soldier", SteamID: "76561198011558250", Team: "red"},
					}).Return(nil).
					GetPickupGamesMock.Expect("test").Return(requests.GamesResponse{
					Results: []requests.Result{
						{
							State: stateMachine.StartedState,
							Map:   "cp_granary_pro_rc8",
							Slots: []requests.Slot{
								{Player: "123", GameClass: "soldier", Team: "red"},
							},
						},
					},
				}, nil),
				Match: mocks.NewMatcherMock(mc).
					TryParseGameMapMock.Expect(`: World triggered "Round_Start"`).Return().
					SetStartTimeMock.Expect(`: World triggered "Round_Start"`).Return().
					DomainMock.Return("test").PickupPlayersMock.Return(
					[]*stats.PickupPlayer{
						{PlayerID: "123", Class: "soldier", SteamID: "76561198011558250", Team: "red"},
					}).
					StringMock.Return("test#1").
					PickupIDMock.Return(0).
					MapMock.Return("cp_granary_pro_rc8").
					SetPlayersMock.Expect([]*stats.PickupPlayer{
					{PlayerID: "123", Class: "soldier", Team: "red"},
				}).Return().
					SetPickupIDMock.Expect(0).Return(),
				Mongo: mocks.NewInserterMock(mc),
			},
		},
		{
			name: "game-default",
			args: args{
				msg: `"jel<62><[U:1:479446967]><Blue>" killed "KEYREAL<65><[U:1:861133286]><Red>" with "sniperrifle""`,
			},
			fields: fields{
				State: stateMachine.Game,
				log:   log,
				File: mocks.NewLogFilerMock(mc).
					WriteLineMock.Expect(`"jel<62><[U:1:479446967]><Blue>" killed "KEYREAL<65><[U:1:861133286]><Red>" with "sniperrifle""`).Return(),
				Uploader: mocks.NewLogUploaderMock(mc),
				Match: mocks.NewMatcherMock(mc).
					PlayerStatsMock.Return(stats.PlayerStatsCollection{}).
					SetPlayerStatsMock.Expect(stats.PlayerStatsCollection{
					steamid.SID64FromString("76561198439712695"): {Kills: 1},
					steamid.SID64FromString("76561198821399014"): {Deaths: 1},
				}).Return(),
				Mongo: mocks.NewInserterMock(mc),
			},
		},
		{
			name: "pregame with err on ResolvePlayers",
			args: args{msg: `: World triggered "Round_Start"`},
			fields: fields{
				State: stateMachine.Pregame,
				log:   log,
				File: mocks.NewLogFilerMock(mc).
					WriteLineMock.Expect(`: World triggered "Round_Start"`).Return(),
				Uploader: mocks.NewLogUploaderMock(mc).
					ResolvePlayersMock.Expect("test",
					[]*stats.PickupPlayer{
						{PlayerID: "123", Class: "soldier", SteamID: "76561198011558250", Team: "red"},
					}).Return(errors.New("failed to resolve players")).
					GetPickupGamesMock.Expect("test").Return(requests.GamesResponse{
					Results: []requests.Result{
						{
							State: stateMachine.StartedState,
							Map:   "cp_granary_pro_rc8",
							Slots: []requests.Slot{
								{Player: "123", GameClass: "soldier", Team: "red"},
							},
						},
					},
				}, nil),
				Match: mocks.NewMatcherMock(mc).
					TryParseGameMapMock.Expect(`: World triggered "Round_Start"`).Return().
					SetStartTimeMock.Expect(`: World triggered "Round_Start"`).Return().
					DomainMock.Return("test").PickupPlayersMock.Return(
					[]*stats.PickupPlayer{
						{PlayerID: "123", Class: "soldier", SteamID: "76561198011558250", Team: "red"},
					}).
					StringMock.Return("test#1").
					PickupIDMock.Return(0).
					MapMock.Return("cp_granary_pro_rc8").
					SetPlayersMock.Expect([]*stats.PickupPlayer{
					{PlayerID: "123", Class: "soldier", Team: "red"},
				}).Return().
					SetPickupIDMock.Expect(0).Return(),
				Mongo: mocks.NewInserterMock(mc),
			},
		},
		{
			name: "game-Game_Over",
			args: args{
				msg: `: World triggered "Game_Over" reason "`,
			},
			fields: fields{
				State: stateMachine.Game,
				log:   log,
				File: mocks.NewLogFilerMock(mc).
					WriteLineMock.Expect(`: World triggered "Game_Over" reason "`).Return().
					BufferMock.Return(bytes.Buffer{}).
					FlushBufferMock.Return(),
				Uploader: mocks.NewLogUploaderMock(mc).
					MakeMultipartMapMock.Expect("cp_granary_pro_rc8", "test", 0, bytes.Buffer{}).Return(map[string]io.Reader{}).
					UploadLogFileMock.Expect(map[string]io.Reader{}).Return(nil),
				Match: mocks.NewMatcherMock(mc).
					MapMock.Return("cp_granary_pro_rc8").
					DomainMock.Return("test").
					PickupIDMock.Return(0).
					PlayerStatsMock.Return(stats.PlayerStatsCollection{
					steamid.SID64FromString("76561198439712695"): {Kills: 1},
				}).
					SetPlayerStatsMock.Expect(stats.PlayerStatsCollection{
					steamid.SID64FromString("76561198439712695"): {Kills: 1},
				}).Return().
					SetLengthMock.Expect(`: World triggered "Game_Over" reason "`).Return().
					PickupPlayersMock.Return([]*stats.PickupPlayer{
					{SteamID: "76561198439712695"},
				}).
					LengthSecondsMock.Return(0).
					StringMock.Return("test#1").
					FlushMock.Return(),
				Mongo: mocks.NewInserterMock(mc).InsertGameStatsMock.Expect([]interface{}{
					stats.MongoPlayerInfo{
						Player:   &stats.PickupPlayer{SteamID: "76561198439712695"},
						Stats:    stats.PlayerStats{Kills: 1},
						Domain:   "test",
						PickupID: 0,
						Length:   0,
					},
				}).Return(nil),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &stateMachine.StateMachine{
				State:    tt.fields.State,
				Log:      tt.fields.log,
				File:     tt.fields.File,
				Uploader: tt.fields.Uploader,
				Match:    tt.fields.Match,
				Mongo:    tt.fields.Mongo,
			}
			sm.ProcessLogLine(tt.args.msg)
		})
	}
}

func TestStateType_String(t *testing.T) {
	tests := []struct {
		name string
		st   stateMachine.StateType
		want string
	}{
		{
			name: "pregame",
			st:   stateMachine.Pregame,
			want: "pregame",
		},
		{
			name: "game",
			st:   stateMachine.Game,
			want: "game",
		},
		{
			name: "unknown",
			st:   stateMachine.StateType(3),
			want: "unknown State",
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
