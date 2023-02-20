package settings

import (
	"fmt"
	"os"

	"github.com/femnad/fup/internal"
)

type Settings struct {
	CloneDir      string                    `yaml:"clone_dir"`
	ExtractDir    string                    `yaml:"extract_dir"`
	HostFacts     map[string]map[string]any `yaml:"host_facts"`
	TemplateDir   string                    `yaml:"template_dir"`
	Versions      map[string]string         `yaml:"versions"`
	VirtualEnvDir string                    `yaml:"virtualenv_dir"`
}

func ExpandSettingsWithLookup(settings Settings, s string, lookup map[string]string) string {
	expanded := os.Expand(s, func(prop string) string {
		val, ok := lookup[prop]
		if ok {
			return val
		}

		switch prop {
		case "clone_dir":
			return settings.CloneDir
		case "extract_dir":
			return settings.ExtractDir
		default:
			return fmt.Sprintf("${%s}", prop)
		}
	})
	return internal.ExpandUser(expanded)
}
