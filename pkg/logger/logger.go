package logger

import (
	"github.com/sirupsen/logrus"
)

func NewLogger(level string) (*logrus.Logger, error) {
	l := logrus.New()
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	l.SetLevel(lvl)
	//l.SetFormatter(&logrus.JSONFormatter{DisableHTMLEscape: true, TimestampFormat: "06/01/02 15:04:05"})
	l.SetFormatter(&logrus.TextFormatter{})
	return l, nil
}
