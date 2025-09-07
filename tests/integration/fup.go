package integration

import (
	"fmt"
	"os"
	"path"

	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	marecmd "github.com/femnad/mare/cmd"
	"gopkg.in/yaml.v3"
)

func writeConfig(cfg entity.Config, configFile string) error {
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	configDir, _ := path.Split(configFile)
	err = internal.EnsureDirExists(configDir)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFile, out, 0o600)
	return err
}

func runFup(provisioner, configFile string) error {
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = internal.ExpandUser("~/go")
	}
	fup := path.Join(goPath, "bin", "fup", "apply")

	err := marecmd.RunErrOnly(marecmd.Input{Command: fmt.Sprintf("%s -p %s -f %s", fup, provisioner, configFile)})
	return err
}
