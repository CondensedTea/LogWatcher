package server

import (
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/stats"

	"github.com/sirupsen/logrus"
)

func StartWorker(log *logrus.Logger, lm LogFiler, r requests.LogProcessor, gi stats.MatchDater) {
	for msg := range lm.Channel() {
		processLogLine(msg, log, lm, r, gi)
	}
}

func processLogLine(msg string, log *logrus.Logger, lm LogFiler, r requests.LogProcessor, gi stats.MatchDater) {
	lm.Lock()
	defer lm.Unlock()
	switch lm.State() {
	case Pregame:
		gi.TryParseGameMap(msg)
		if roundStart.MatchString(msg) {
			processGameStartedEvent(msg, log, lm, r, gi)
		}
	case Game:
		processGameLogLine(msg, log, lm, gi)
		if logClosed.MatchString(msg) || gameOver.MatchString(msg) {
			processGameOverEvent(msg, log, lm, r, gi)
		}
	}
}

func processGameOverEvent(msg string, log *logrus.Logger, lm LogFiler, r requests.LogProcessor, gi stats.MatchDater) {
	lm.SetState(Pregame)
	gi.SetLength(msg)
	payload := r.MakeMultipartMap(gi.Map(), gi.Domain(), gi.PickupID(), lm.Buffer())
	if err := r.UploadLogFile(payload); err != nil {
		log.WithFields(logrus.Fields{"server": gi.String()}).
			Errorf("Failed to upload file to logs.tf: %s", err)
	}
	playersStats := stats.ExtractPlayerStats(gi)
	if err := stats.InsertGameStats(lm.GetConn(), playersStats); err != nil { // need to make databaser interface
		log.WithFields(logrus.Fields{"server": gi.String()}).
			Errorf("Failed to insert stats to db: %s", err)
	}
	log.WithFields(logrus.Fields{
		"server":    gi.String(),
		"pickup_id": gi.PickupID(),
		"map":       gi.Map(),
	}).Info("Pickup has ended")
	Flush(lm, gi)
}

func processGameLogLine(msg string, log *logrus.Logger, lm LogFiler, gi stats.MatchDater) {
	lm.WriteLine(msg)
	if err := stats.UpdateStatsMap(msg, gi.PlayerStatsMap()); err != nil {
		log.WithFields(logrus.Fields{
			"server": gi.String(),
			"state":  lm.State().String(),
			"msg":    msg,
		}).Errorf("Error on updating player stats: %s", err)
	}
}

func processGameStartedEvent(msg string, log *logrus.Logger, lm LogFiler, r requests.LogProcessor, gi stats.MatchDater) {
	lm.SetState(Game)
	gi.SetStartTime(msg)
	lm.WriteLine(msg)
	if err := UpdatePickupInfo(r, gi); err != nil {
		log.WithFields(logrus.Fields{"server": gi.String()}).
			Errorf("Failed to get pickup id from API: %s", err)
	}
	err := r.ResolvePlayers(gi.Domain(), gi.PickupPlayers())
	if err != nil {
		log.WithFields(logrus.Fields{"server": gi.String()}).
			Errorf("Failed to resolve pickup player ids through API: %s", err)
	}
	log.WithFields(logrus.Fields{
		"server":    gi.String(),
		"pickup_id": gi.PickupID(),
		"map":       gi.Map(),
	}).Infof("Pickup has started")
}
