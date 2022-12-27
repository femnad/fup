package base

import (
	"fmt"
	"os"

	"github.com/femnad/fup/internal"
)

var (
	expandables = []string{
		"clone_dir",
		"extract_dir",
	}
)

func IsExpandable(prop string) bool {
	return internal.Contains(expandables, prop)
}

func ExpandSettings(settings Settings, s string) string {
	return ExpandSettingsWithLookup(settings, s, map[string]string{})
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
