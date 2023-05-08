package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alexflint/go-arg"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/provision"
)

const (
	version = "0.3.0"
)

type args struct {
	DebugToStderr bool     `arg:"-b,--debug-to-stderr"`
	File          string   `arg:"-f,--file,env:FUP_CONFIG" default:"~/.config/fup/fup.yml"`
	LogFile       string   `arg:"--logfile" default:"~/.local/share/fup/fup.log"`
	LogLevel      int      `arg:"-l,--loglevel" default:"5"`
	Provisioners  []string `arg:"-p,--provisioners"`
	WriteLogs     bool     `arg:"-w,--writelogs"`
}

func (args) Version() string {
	return fmt.Sprintf("fup %s", version)
}

func determineConfigFile(a args) (cfg string) {
	cfg = a.File
	_, err := os.Stat("fup.yml")
	if err == nil {
		wd, err := os.Getwd()
		if err != nil {
			internal.Log.Errorf("error determining current dir: %v", err)
			return
		}
		return path.Join(wd, "fup.yml")
	}

	return cfg
}

func main() {
	var parsed args
	arg.MustParse(&parsed)

	logFile := internal.ExpandUser(parsed.LogFile)
	if !parsed.WriteLogs {
		logFile = ""
	}
	internal.InitLogging(parsed.LogLevel, logFile, parsed.DebugToStderr)

	cfg := determineConfigFile(parsed)

	internal.Log.Debugf("Reading config file %s", cfg)
	config, err := base.ReadConfig(cfg)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	p := provision.NewProvisioner(config, parsed.Provisioners)
	p.Apply()
}
