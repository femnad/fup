package provision

import (
	"errors"
	"fmt"
	"strings"

	marecmd "github.com/femnad/mare/cmd"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	precheck "github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/run"
	"github.com/femnad/fup/settings"
)

const (
	defaultHost    = "github.com"
	defaultVersion = "latest"
)

func qualifyPkg(pkg entity.GoPkg, s settings.Settings) (string, error) {
	name := pkg.Name()
	tokens := strings.Split(name, "/")
	if len(tokens) == 0 {
		return "", fmt.Errorf("unable to qualify package: %s", name)
	}

	version, err := pkg.GetVersion(s)
	if err != nil {
		return "", err
	}

	if version == "" {
		version = defaultVersion
	}

	maybeHost := tokens[0]
	if strings.Index(maybeHost, ".") > 0 {
		return fmt.Sprintf("%s@%s", name, version), nil
	}

	return fmt.Sprintf("%s/%s@%s", defaultHost, name, version), nil
}

func goInstall(pkg entity.GoPkg, s settings.Settings) error {
	name := pkg.Name()

	if precheck.ShouldSkip(pkg, s) {
		internal.Log.Debugf("Skipping go install for %s", name)
		return nil
	}

	internal.Log.Infof("Installing Go package %s", name)

	qualifiedName, err := qualifyPkg(pkg, s)
	if err != nil {
		internal.Log.Errorf("Error in installing Go package %s: %v", name, err)
		return err
	}

	cmd := fmt.Sprintf("go install %s", qualifiedName)
	resp, err := run.Cmd(s, marecmd.Input{Command: cmd})

	if err != nil {
		internal.Log.Errorf("error in installing go package %s: %v, output: %s", name, err, resp.Stderr)
		return err
	}

	return nil
}

func goInstallPkgs(cfg entity.Config) error {
	var goErrs []error
	for _, pkg := range cfg.Go {
		err := goInstall(pkg, cfg.Settings)
		goErrs = append(goErrs, err)
	}

	return errors.Join(goErrs...)
}
