package server_test

//func Test_processGameStartedEvent(t *testing.T) {
//	mc := minimock.NewController(t)
//	type args struct {
//		msg string
//		log *logrus.Logger
//		lm  server.LogFiler
//		r   requests.LogProcessor
//		gi  stats.MatchDater
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		{
//			name: "default",
//			args: args{
//				msg: `test`,
//				log: logrus.New(),
//				lm:  mocks.NewLogFilerMock(mc).SetStateMock.Return().GetConnMock.Return(nil),
//				r:   mocks.NewLogProcessorMock(mc).ResolvePlayersMock.Return(nil),
//				gi: mocks.NewMatchDaterMock(mc).
//					SetStartTimeMock.Return().
//					StringMock.Return("").
//					DomainMock.Return("").
//					PickupPlayersMock.Return([]*stats.PickupPlayer{}),
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			//if tt.args.lm.State() != Game {
//			//	t.Errorf("processGameStartedEvent(), state = %v, want %v", tt.args.lm.State(), Game)
//			//}
//			//if cmp.Equal(tt.args.lm.Buffer(), updatedBuf) {
//			//	t.Errorf("processGameStartedEvent(), buffer = %v, want %v", tt.args.lm.Buffer(), updatedBuf)
//			//}
//		})
//	}
//}
