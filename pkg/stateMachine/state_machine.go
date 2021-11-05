package stateMachine

import (
	"LogWatcher/pkg/mongo"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/server"
	"LogWatcher/pkg/stats"
	"regexp"
	"strconv"

	"github.com/sirupsen/logrus"
)

var (
	roundStart   = regexp.MustCompile(`: World triggered "Round_Start"`)
	roundWin     = regexp.MustCompile(`: World triggered "Round_Win"`)
	gameOver     = regexp.MustCompile(`: World triggered "Game_Over" reason "`)
	logClosed    = regexp.MustCompile(`: Log File closed.`)
	currentScore = regexp.MustCompile(`: Team "(Red|Blue)" current score "(\d)" with "\d" players`)
)

type StateType int

const (
	Pregame StateType = iota
	Game
	RoundReset
)

func (st StateType) String() string {
	switch st {
	case Pregame:
		return "pregame"
	case Game:
		return "game"
	case RoundReset:
		return "round reset"
	default:
		return "unknown State"
	}
}

type StateMachine struct {
	State    StateType
	Log      *logrus.Logger
	File     server.LogFiler
	Uploader requests.LogUploader
	Match    stats.Matcher
	Mongo    mongo.Inserter
	Channel  chan string
}

type Stater interface {
	Channel() chan string
	StartWorker()
	ProcessLogLine(msg string)
	ProcessGameStartedEvent(msg string)
	ProcessGameLogLine(msg string)
	ProcessGameOverEvent(msg string)
	UpdatePickupInfo() error
	Match() stats.Matcher
	Uploader() requests.LogUploader
	Inserter() mongo.Inserter
	SetState(state StateType)
}

func NewStateMachine(
	log *logrus.Logger,
	file server.LogFiler,
	uploader requests.LogUploader,
	matchData stats.Matcher,
	inserter mongo.Inserter,
) *StateMachine {
	return &StateMachine{
		State:    Pregame,
		Log:      log,
		File:     file,
		Uploader: uploader,
		Match:    matchData,
		Mongo:    inserter,
		Channel:  make(chan string),
	}
}

func (sm *StateMachine) StartWorker() {
	for msg := range sm.Channel {
		sm.ProcessLogLine(msg)
	}
}

func (sm *StateMachine) ProcessLogLine(msg string) {
	switch sm.State {
	case Pregame:
		sm.Match.TryParseGameMap(msg)
		if roundStart.MatchString(msg) {
			sm.ProcessGameStartedEvent(msg)
		}
	case Game:
		if roundWin.MatchString(msg) {
			sm.State = RoundReset
			break
		}
		sm.ProcessGameLogLine(msg)
		if logClosed.MatchString(msg) || gameOver.MatchString(msg) {
			sm.ProcessGameOverEvent(msg)
		}
	case RoundReset:
		if currentScore.MatchString(msg) {
			sm.ProcessCurrentScore(msg)
		}
		if roundStart.MatchString(msg) {
			sm.State = Game
		}
	}
}

func (sm *StateMachine) ProcessGameStartedEvent(msg string) {
	sm.State = Game
	sm.Match.SetStartTime(msg)
	sm.File.WriteLine(msg)

	pickup, err := sm.Uploader.FindMatchingPickup(sm.Match.Domain(), sm.Match.Map())
	if err != nil {
		sm.Log.WithFields(logrus.Fields{"server": sm.Match.String()}).
			Errorf("Failed to get pickup id from API: %s", err)
		return
	}

	sm.Match.SetPlayers(pickup.Players)
	sm.Match.SetPickupID(pickup.ID)

	if err := sm.Uploader.ResolvePlayers(sm.Match.Domain(), sm.Match.PickupPlayers()); err != nil {
		sm.Log.WithFields(logrus.Fields{"server": sm.Match.String()}).
			Errorf("Failed to resolve pickup player ids through API: %s", err)
	}
	sm.Log.WithFields(logrus.Fields{
		"server":    sm.Match.String(),
		"pickup_id": sm.Match.PickupID(),
		"map":       sm.Match.Map(),
	}).Infof("Pickup has started")
}

func (sm *StateMachine) ProcessGameLogLine(msg string) {
	sm.File.WriteLine(msg)
	playerStats := stats.UpdateStatsMap(msg, sm.Match.PlayerStats())
	sm.Match.SetPlayerStats(playerStats)
}

func (sm *StateMachine) ProcessGameOverEvent(msg string) {
	sm.State = Pregame
	sm.Match.SetLength(msg)
	payload := sm.Uploader.MakeMultipartMap(sm.Match, sm.File.Buffer())
	if err := sm.Uploader.UploadLogFile(payload); err != nil {
		sm.Log.WithFields(logrus.Fields{"server": sm.Match.String()}).Errorf("Failed to upload File to logs.tf: %s", err)
	}
	playersStats := stats.ExtractPlayerStats(sm.Match)
	if err := sm.Mongo.InsertGameStats(playersStats); err != nil {
		sm.Log.WithFields(logrus.Fields{"server": sm.Match.String()}).Errorf("Failed to insert stats to db: %s", err)
	}
	sm.Log.WithFields(logrus.Fields{
		"server":    sm.Match.String(),
		"pickup_id": sm.Match.PickupID(),
		"map":       sm.Match.Map(),
	}).Info("Pickup has ended")
	sm.Flush()
}

// Flush is used to empty all game data
func (sm *StateMachine) Flush() {
	sm.File.FlushBuffer()
	sm.Match.Flush()
}

func (sm *StateMachine) ProcessCurrentScore(msg string) {
	match := currentScore.FindStringSubmatch(msg)
	score, _ := strconv.Atoi(match[2])
	if match[1] == "Red" {
		sm.Match.SetRedScore(score)
	} else if match[1] == "Blue" {
		sm.Match.SetBlueScore(score)
	}
}
