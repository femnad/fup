package provision

import (
	"fmt"
	"strings"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/internal"
	precheck "github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/run"
)

const (
	defaultHost    = "github.com"
	defaultVersion = "latest"
)

func qualifyPkg(pkg base.GoPkg) (string, error) {
	name := pkg.Name()
	tokens := strings.Split(name, "/")
	if len(tokens) == 0 {
		return "", fmt.Errorf("unable to qualify package: %s", name)
	}

	version := pkg.GetVersion()
	if version == "" {
		version = defaultVersion
	}

	maybeHost := tokens[0]
	if strings.Index(maybeHost, ".") > 0 {
		return fmt.Sprintf("%s@%s", name, version), nil
	}

	return fmt.Sprintf("%s/%s@%s", defaultHost, name, version), nil
}

func goInstall(pkg base.GoPkg, s settings.Settings) {
	if precheck.ShouldSkip(pkg, s) {
		internal.Log.Debugf("Skipping go install for %s", pkg.Name())
		return
	}

	internal.Log.Infof("Installing go package %s", pkg.Name())

	name, err := qualifyPkg(pkg)
	if err != nil {
		internal.Log.Errorf("error in installing go package %v", err)
	}

	cmd := fmt.Sprintf("go install %s", name)
	resp, err := run.Cmd(s, marecmd.Input{Command: cmd})

	if err != nil {
		internal.Log.Errorf("error in installing go package %s: %v, output: %s", name, err, resp.Stderr)
	}
}

func goInstallPkgs(cfg base.Config) {
	for _, pkg := range cfg.Go {
		goInstall(pkg, cfg.Settings)
	}
}
