package provision

import (
	"fmt"
	"strings"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
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

func goInstall(pkg base.GoPkg) {
	name, err := qualifyPkg(pkg)
	if err != nil {
		internal.Log.Errorf("error in installing go package %v", err)
	}

	cmd := fmt.Sprintf("go install %s", name)
	out, err := common.RunCmd(cmd)

	if err != nil {
		internal.Log.Errorf("error in installing go package %s: %v, output: %s", name, err, out)
	}
}

func goInstallPkgs(cfg base.Config) {
	for _, pkg := range cfg.Go {
		goInstall(pkg)
	}
}
