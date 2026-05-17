package provision

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/mare/cmd"
)

func getToolVersion(tool string) (string, error) {
	out, err := cmd.Run(cmd.Input{Command: "uv tool list"})
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(out.Stdout))
	for scanner.Scan() {
		line := scanner.Text()
		toolVersion := strings.Split(line, " ")
		if len(toolVersion) != 2 {
			return "", fmt.Errorf("unable to determine installed version for %s", tool)
		}
		installedTool := toolVersion[0]
		installedVersion := toolVersion[1]

		if installedTool != tool {
			continue
		}

		return installedVersion, nil
	}

	return "", fmt.Errorf("unable to determine installed version for %s", tool)
}

func installTool(tool entity.UvTool) error {
	version := tool.Version
	if version == "" {
		version = defaultVersion
	}
	name := tool.Name

	_, err := common.Which(name)
	if err == nil {
		if version == defaultVersion {
			return nil
		}
	
		var installedVersion string
		installedVersion, err = getToolVersion(name)
		if err != nil {
			return err
		}

		if installedVersion == version {
			return nil
		}
	}

	internal.Logger.Debug().Str("tool", name).Str("version", version).Msg("Installing uv tool")
	return cmd.RunErrOnly(cmd.Input{Command: fmt.Sprintf("uv tool install %s@%s", name, version)})
}

func uvInstallTools(cfg entity.Config) error {
	var errs []error
	for _, tool := range cfg.UvTools {
		err := installTool(tool)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
