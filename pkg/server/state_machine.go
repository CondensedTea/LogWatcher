package server

import (
	"LogWatcher/pkg/mongo"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/stats"

	"github.com/sirupsen/logrus"
)

func StartWorker(log *logrus.Logger, lm LogFiler, r requests.LogUploader, gi stats.MatchDater, i mongo.Inserter) {
	for msg := range lm.Channel() {
		processLogLine(msg, log, lm, r, gi, i)
	}
}

func processLogLine(
	msg string,
	log *logrus.Logger,
	lm LogFiler,
	r requests.LogUploader,
	gi stats.MatchDater,
	i mongo.Inserter,
) {
	lm.Lock()
	defer lm.Unlock()
	switch lm.State() {
	case Pregame:
		gi.TryParseGameMap(msg)
		if roundStart.MatchString(msg) {
			ProcessGameStartedEvent(msg, log, lm, r, gi)
		}
	case Game:
		ProcessGameLogLine(msg, lm, gi)
		if logClosed.MatchString(msg) || gameOver.MatchString(msg) {
			processGameOverEvent(msg, log, lm, r, gi, i)
		}
	}
}

func ProcessGameStartedEvent(msg string, log *logrus.Logger, lm LogFiler, lu requests.LogUploader, md stats.MatchDater) {
	lm.SetState(Game)
	md.SetStartTime(msg)
	lm.WriteLine(msg)
	err := UpdatePickupInfo(lu, md)
	if err != nil {
		log.WithFields(logrus.Fields{"server": md.String()}).
			Errorf("Failed to get pickup id from API: %s", err)
	}
	err = lu.ResolvePlayers(md.Domain(), md.PickupPlayers())
	if err != nil {
		log.WithFields(logrus.Fields{"server": md.String()}).
			Errorf("Failed to resolve pickup player ids through API: %s", err)
	}
	log.WithFields(logrus.Fields{
		"server":    md.String(),
		"pickup_id": md.PickupID(),
		"map":       md.Map(),
	}).Infof("Pickup has started")
}

func ProcessGameLogLine(msg string, lf LogFiler, gi stats.MatchDater) {
	lf.WriteLine(msg)
	stats.UpdateStatsMap(msg, gi.PlayerStatsCollection())
}

func processGameOverEvent(
	msg string,
	log *logrus.Logger,
	lm LogFiler,
	r requests.LogUploader,
	gi stats.MatchDater,
	i mongo.Inserter,
) {
	lm.SetState(Pregame)
	gi.SetLength(msg)
	payload := r.MakeMultipartMap(gi.Map(), gi.Domain(), gi.PickupID(), lm.Buffer())
	if err := r.UploadLogFile(payload); err != nil {
		log.WithFields(logrus.Fields{"server": gi.String()}).Errorf("Failed to upload file to logs.tf: %s", err)
	}
	playersStats := stats.ExtractPlayerStats(gi)
	if err := i.InsertGameStats(playersStats); err != nil {
		log.WithFields(logrus.Fields{"server": gi.String()}).Errorf("Failed to insert stats to db: %s", err)
	}
	log.WithFields(logrus.Fields{
		"server":    gi.String(),
		"pickup_id": gi.PickupID(),
		"map":       gi.Map(),
	}).Info("Pickup has ended")
	Flush(lm, gi)
}
