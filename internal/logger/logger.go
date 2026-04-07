package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init(environment string) {
	Log = logrus.New()

	Log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
		ForceColors:     true,
		DisableQuote:    true,
	})

	Log.SetOutput(os.Stdout)
	Log.SetLevel(logrus.DebugLevel)
}
