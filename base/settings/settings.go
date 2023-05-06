package settings

import (
	"bytes"
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

func expand(s string, lookup map[string]string) (string, error) {
	var backspace bool
	var cur bytes.Buffer
	var out bytes.Buffer
	consuming := false
	dollar := false

	for _, c := range s {
		if c == '$' && !backspace {
			dollar = true
			continue
		}
		if c == '\\' {
			backspace = true
			continue
		}
		backspace = c == '\\'
		if dollar {
			if c == '{' {
				dollar = false
				consuming = true
				continue
			} else {
				out.WriteRune('$')
			}
		}
		if c == '}' && consuming {
			consuming = false
			curStr := cur.String()
			val, ok := lookup[curStr]
			if ok {
				out.WriteString(val)
			} else {
				orig := fmt.Sprintf("${%s}", curStr)
				out.WriteString(orig)
			}
			cur.Reset()
			continue
		}
		if consuming {
			cur.WriteRune(c)
		} else {
			out.WriteRune(c)
		}
	}

	return out.String(), nil
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
