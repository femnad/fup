package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexflint/go-arg"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/provision"
)

const (
	version = "0.38.0"
)

type ApplyCmd struct {
	Provisioners   []string `arg:"-p,--provisioners" help:"List of provisioners to run"`
	LogFile        string   `arg:"--logfile" default:"~/.local/share/fup/fup.log" help:"Log file path"`
	LogLevel       int      `arg:"-l,--loglevel" default:"4" help:"Log level as integer, 0 least, 5 most"`
	NoLogs         bool     `arg:"-n,--nologs" help:"Don't write logs to a file"`
	PrintConfig    bool     `arg:"-r,--print-config" help:"Print final config and exit"`
	ValidateConfig bool     `arg:"-c,--validate-config" help:"Validate config and exit"`
	DebugToStderr  bool     `arg:"-b,--debug-to-stderr" help:"Write logs to stderr"`
}

type VersionLookupCmd struct {
	AssetURL  string `arg:"-a,--asset-url"`
	LookupURL string `arg:"positional,required" help:"Version lookup URL"`
	Query     string `arg:"positional,required" help:"Version lookup query"`
	PostProc  string `arg:"-p,--post-proc" help:"Post processing function"`
}

type args struct {
	Apply         *ApplyCmd         `arg:"subcommand:apply" help:"Apply a configuration"`
	VersionLookup *VersionLookupCmd `arg:"subcommand:lookup" help:"Lookup a version based on a URL and query"`
	File          string            `arg:"-f,--file,env:FUP_CONFIG" default:"~/.config/fup/fup.yml" help:"Config file path"`
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

func apply(parsed args) {
	applyCfg := parsed.Apply
	logFile := internal.ExpandUser(applyCfg.LogFile)
	if applyCfg.NoLogs {
		logFile = ""
	}
	internal.InitLogging(applyCfg.LogLevel, logFile, applyCfg.DebugToStderr)

	cfg := parsed.File
	internal.Log.Debugf("Reading config file %s", cfg)

	if applyCfg.PrintConfig {
		printConfig(cfg)
	}

	config, err := base.ReadConfig(cfg)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	if applyCfg.ValidateConfig {
		os.Exit(0)
	}

	p, err := provision.NewProvisioner(config, applyCfg.Provisioners)
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

func lookup(parsed args) {
	versionLookup := parsed.VersionLookup
	config, err := base.ReadConfig(parsed.File)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	spec := entity.VersionLookupSpec{
		PostProc: versionLookup.PostProc,
		Query:    versionLookup.Query,
		URL:      versionLookup.LookupURL,
	}

	var out string
	out, err = entity.LookupVersion(spec, versionLookup.AssetURL, config.Settings)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	fmt.Println(out)
}

func main() {
	var parsed args
	p := arg.MustParse(&parsed)

	switch {
	case parsed.Apply != nil:
		apply(parsed)
	case parsed.VersionLookup != nil:
		lookup(parsed)
	}
	if p.Subcommand() == nil {
		p.Fail("Missing subcommand")
	}
}
