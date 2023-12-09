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
	version = "0.18.2"
)

type args struct {
	DebugToStderr bool     `arg:"-b,--debug-to-stderr" help:"Write logs to stderr"`
	File          string   `arg:"-f,--file,env:FUP_CONFIG" default:"~/.config/fup/fup.yml" help:"Config file path"`
	LogFile       string   `arg:"--logfile" default:"~/.local/share/fup/fup.log" help:"Log file path"`
	LogLevel      int      `arg:"-l,--loglevel" default:"5" help:"Log level as integer, 0 least, 5 most"`
	NoLogs        bool     `arg:"-n,--nologs" help:"Don't write logs to a file"`
	Provisioners  []string `arg:"-p,--provisioners" help:"List of provisioners to run"`
}

func (args) Version() string {
	return fmt.Sprintf("fup v%s", version)
}

func main() {
	var parsed args
	arg.MustParse(&parsed)

	logFile := internal.ExpandUser(parsed.LogFile)
	if parsed.NoLogs {
		logFile = ""
	}
	internal.InitLogging(parsed.LogLevel, logFile, parsed.DebugToStderr)

	cfg := parsed.File
	internal.Log.Debugf("Reading config file %s", cfg)
	config, err := base.ReadConfig(cfg)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	p, err := provision.NewProvisioner(config, parsed.Provisioners)
	if err != nil {
		internal.Log.Errorf("error building provisioner: %v", err)
		return
	}

	p.Apply()
}
