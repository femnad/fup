package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alexflint/go-arg"

	"github.com/femnad/fup/base"
	basecmd "github.com/femnad/fup/cmd/base"
	"github.com/femnad/fup/cmd/printspec"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/provision"
)

const (
	version = "0.42.1"
)

type ApplyCmd struct {
	Provisioners   []string `arg:"-p,--provisioners" help:"List of provisioners to run"`
	PrintConfig    bool     `arg:"-r,--print-config" help:"Print final config and exit"`
	ValidateConfig bool     `arg:"-c,--validate-config" help:"Validate config and exit"`
}

type VersionLookupCmd struct {
	AssetURL    string `arg:"-a,--asset-url"`
	FollowURL   bool   `arg:"-o,--follow-redirect" help:"Follow redirects"`
	GetRedirect bool   `arg:"-r,--get-redirect" help:"Get redirect URL"`
	LookupURL   string `arg:"positional,required" help:"Version lookup URL"`
	PostProc    string `arg:"-p,--post-proc" help:"Post processing function"`
	Query       string `arg:"positional,required" help:"Version lookup query"`
}

type VersionPrintSpecCmd struct {
	Owner string `arg:"positional,required" help:"Repo owner"`
	Repo  string `arg:"positional,required" help:"Repo name"`
}

type args struct {
	Apply            *ApplyCmd            `arg:"subcommand:apply" help:"Apply a configuration"`
	LogLevel         string               `arg:"-l,--loglevel" default:"debug" help:"Log level: trace, debug, info, warn, error, fatal, panic"`
	VersionPrintSpec *VersionPrintSpecCmd `arg:"subcommand:github-spec" help:"Print a GitHub release spec based on a URL"`
	VersionLookup    *VersionLookupCmd    `arg:"subcommand:lookup" help:"Lookup a version based on a URL and query"`
	File             string               `arg:"-f,--file,env:FUP_CONFIG" default:"~/.config/fup/fup.yml" help:"Config file path"`
}

func (args) Version() string {
	return fmt.Sprintf("fup v%s", version)
}

func printConfig(configFile string) {
	out, err := base.FinalizeConfig(configFile)
	if err != nil {
		internal.Logger.Error().Str("file", configFile).Err(err).Msg("Error printing config")
		os.Exit(1)
	}

	fmt.Println(out)
	os.Exit(0)
}

func apply(parsed args) {
	applyCfg := parsed.Apply
	internal.InitLogging(parsed.LogLevel)

	cfg := parsed.File
	internal.Logger.Trace().Str("path", cfg).Msg("Reading config file")

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
		internal.Logger.Fatal().Err(err).Msg("Error creating provisioner")
	}

	err = p.Apply()
	if err != nil {
		internal.Logger.Fatal().Err(err).Msg("Error applying provisioner")
	}
}

func lookup(parsed args) {
	versionLookup := parsed.VersionLookup
	config, err := base.ReadConfig(parsed.File)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	spec := entity.VersionLookupSpec{
		FollowURL:   versionLookup.FollowURL,
		GetRedirect: versionLookup.GetRedirect,
		PostProc:    versionLookup.PostProc,
		Query:       versionLookup.Query,
		URL:         versionLookup.LookupURL,
	}

	var out string
	out, err = entity.LookupVersion(spec, versionLookup.AssetURL, config.Settings)
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	fmt.Println(out)
}

func printSpec(parsed args) {
	input := printspec.Input{
		Input: basecmd.Input{LogLevel: parsed.LogLevel},
		Owner: parsed.VersionPrintSpec.Owner,
		Repo:  parsed.VersionPrintSpec.Repo,
	}
	printspec.PrintSpec(input)
}

func main() {
	var parsed args
	p := arg.MustParse(&parsed)

	switch {
	case parsed.Apply != nil:
		apply(parsed)
	case parsed.VersionLookup != nil:
		lookup(parsed)
	case parsed.VersionPrintSpec != nil:
		printSpec(parsed)
	}
	if p.Subcommand() == nil {
		p.Fail("Missing subcommand")
	}
}
