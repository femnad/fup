package internal

import (
	"github.com/op/go-logging"
	"os"
)

var Log = logging.MustGetLogger("fup")

func InitLogging(level int) {
	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02 15:04:05} %{level:.5s} %{message} %{color:reset}`,
	)
	stderrLogs := logging.NewLogBackend(os.Stderr, "", 0)
	formattedLogs := logging.NewBackendFormatter(stderrLogs, format)

	leveledLogs := logging.AddModuleLevel(formattedLogs)
	leveledLogs.SetLevel(logging.Level(level), "")
	logging.SetBackend(leveledLogs)
}
