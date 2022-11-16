package internal

import (
	"io"
	"log"
	"os"
	"path"

	"github.com/op/go-logging"
)

var Log = logging.MustGetLogger("fup")

const (
	stderrLogLevel = logging.INFO
	fileFormat     = `%{color}%{time:2006-01-02 15:04:05} %{level:.6s} %{shortfunc} %{message} %{color:reset}`
	stderrFormat   = `%{color}%{message}%{color:reset}`
)

func initBackend(format string, writer io.Writer, level logging.Level) logging.LeveledBackend {
	backendFormat := logging.MustStringFormatter(format)
	backend := logging.NewLogBackend(writer, "", 0)
	formattedBackend := logging.NewBackendFormatter(backend, backendFormat)
	leveledBackend := logging.AddModuleLevel(formattedBackend)

	leveledBackend.SetLevel(level, "")
	return leveledBackend
}

func InitLogging(logFile string, level int) {
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

	fileBackend := initBackend(fileFormat, f, logging.Level(level))
	stderrBackend := initBackend(stderrFormat, os.Stderr, stderrLogLevel)

	logging.SetBackend(fileBackend, stderrBackend)
}
