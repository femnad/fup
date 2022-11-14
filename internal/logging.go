package internal

import (
	"log"
	"os"
	"path"

	"github.com/op/go-logging"
)

var Log = logging.MustGetLogger("fup")

func InitLogging(level int, logFile string) {
	dirName, _ := path.Split(logFile)
	_, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirName, 0755)
		if err != nil {
			log.Fatalf("error creating directory %s: %v", dirName, err)
		}
	} else if err != nil {
		log.Fatalf("error checking logging directory %s: %v", dirName, err)
	}

	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}

	format := logging.MustStringFormatter(
		`%{color}%{time:2006-01-02 15:04:05} %{level:.6s} %{shortfunc} %{message} %{color:reset}`,
	)

	backendLogs := logging.NewLogBackend(f, "", 0)
	formattedLogs := logging.NewBackendFormatter(backendLogs, format)
	leveledLogs := logging.AddModuleLevel(formattedLogs)

	leveledLogs.SetLevel(logging.Level(level), "")
	logging.SetBackend(leveledLogs)
}
