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
	File         string   `arg:"-f,--file,env:FUP_CONFIG" default:"~/.config/fup/fup.yml"`
	LogFile      string   `arg:"--logfile" default:"~/.local/share/fup/fup.log"`
	LogLevel     int      `arg:"-l,--loglevel" default:"5"`
	Provisioners []string `arg:"-p,--provisioners"`
	WriteLogs    bool     `arg:"-w,--writelogs" default:"false"`
}

func (args) Version() string {
	return fmt.Sprintf("fup %s", version)
}

func main() {
	var args args
	arg.MustParse(&args)

	logFile := internal.ExpandUser(args.LogFile)
	if !args.WriteLogs {
		logFile = ""
	}
	internal.InitLogging(args.LogLevel, logFile)

	config, err := base.ReadConfig(args.File)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	p := provision.NewProvisioner(config, args.Provisioners)
	p.Apply()
}
