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
	fileFormat     = `%{time:2006-01-02 15:04:05} %{level:.6s} %{shortfunc} %{message}`
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

func getFileBackend(level int, logFile string) logging.LeveledBackend {
	dirName, _ := path.Split(logFile)

	_, err := os.Stat(dirName)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dirName, 0755)
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

	return initBackend(fileFormat, f, logging.Level(level))
}

func InitLogging(level int, logFile string, debugToStderr bool) {
	stderrBackend := initBackend(stderrFormat, os.Stderr, stderrLogLevel)
	backends := []logging.Backend{stderrBackend}

	if logFile != "" {
		fileBackend := getFileBackend(level, logFile)
		backends = append(backends, fileBackend)
	}

	if debugToStderr {
		stderrDebugBackend := initBackend(fileFormat, os.Stderr, logging.DEBUG)
		backends = append(backends, stderrDebugBackend)
	}

	logging.SetBackend(backends...)
}
