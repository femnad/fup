package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexflint/go-arg"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/provision"
)

const (
	version = "0.26.1"
)

type args struct {
	DebugToStderr  bool     `arg:"-b,--debug-to-stderr" help:"Write logs to stderr"`
	File           string   `arg:"-f,--file,env:FUP_CONFIG" default:"~/.config/fup/fup.yml" help:"Config file path"`
	LogFile        string   `arg:"--logfile" default:"~/.local/share/fup/fup.log" help:"Log file path"`
	LogLevel       int      `arg:"-l,--loglevel" default:"4" help:"Log level as integer, 0 least, 5 most"`
	NoLogs         bool     `arg:"-n,--nologs" help:"Don't write logs to a file"`
	PrintConfig    bool     `arg:"-r,--print-config" help:"Print final config and exit"`
	Provisioners   []string `arg:"-p,--provisioners" help:"List of provisioners to run"`
	ValidateConfig bool     `arg:"-c,--validate-config" help:"Validate config and exit"`
}

func (args) Version() string {
	return fmt.Sprintf("fup v%s", version)
}

func printConfig(configFile string) {
	out, err := base.FinalizeConfig(configFile)
	if err != nil {
		internal.Log.Errorf("error printing config from %s: %v", configFile, err)
		os.Exit(1)
	}

	fmt.Println(out)
	os.Exit(0)
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

	if parsed.PrintConfig {
		printConfig(cfg)
	}

	config, err := base.ReadConfig(cfg)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	if parsed.ValidateConfig {
		os.Exit(0)
	}

	p, err := provision.NewProvisioner(config, parsed.Provisioners)
	if err != nil {
		internal.Log.Errorf("error building provisioner: %v", err)
		return
	}

	err = p.Apply()
	if err != nil {
		fmt.Printf("Some provisioners had errors:\n%v\n", err)
		os.Exit(1)
	}
}
