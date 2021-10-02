package server

//
//import (
//	"LogWatcher/pkg/requests"
//	"LogWatcher/pkg/stats"
//	"bytes"
//	"context"
//	"testing"
//
//	"github.com/sirupsen/logrus"
//	"go.mongodb.org/mongo-driver/mongo"
//)
//
//func TestServer_updatePickupInfo(t *testing.T) {
//	type fields struct {
//		log     *logrus.Logger
//		ctx     context.Context
//		Server  stats.ServerInfo
//		State   StateType
//		Channel chan string
//		buffer  bytes.Buffer
//		Game    *GameInfo
//		apiKey  string
//		conn    *mongo.Client
//	}
//	type args struct {
//		client requests.HTTPDoer
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		{
//			name: "default",
//			fields: fields{
//				Game: &GameInfo{Map: "cp_badlands"},
//			},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				log:     tt.fields.log,
//				ctx:     tt.fields.ctx,
//				Server:  tt.fields.Server,
//				State:   tt.fields.State,
//				Channel: tt.fields.Channel,
//				buffer:  tt.fields.buffer,
//				Game:    tt.fields.Game,
//				apiKey:  tt.fields.apiKey,
//				conn:    tt.fields.conn,
//			}
//			if err := s.updatePickupInfo(tt.args.client); (err != nil) != tt.wantErr {
//				t.Errorf("updatePickupInfo() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
//}
