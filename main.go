package main

import (
	"fmt"
	"log"

	"github.com/alexflint/go-arg"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/provision"
)

const (
	version = "0.2.2"
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

func main() {
	var parsed args
	arg.MustParse(&parsed)

	logFile := internal.ExpandUser(parsed.LogFile)
	if !parsed.WriteLogs {
		logFile = ""
	}
	internal.InitLogging(parsed.LogLevel, logFile, parsed.DebugToStderr)

	config, err := base.ReadConfig(parsed.File)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	p := provision.NewProvisioner(config, parsed.Provisioners)
	p.Apply()
}
