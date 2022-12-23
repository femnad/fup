package base

import (
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
	expanded := os.Expand(s, func(prop string) string {
		switch prop {
		case "clone_dir":
			return settings.CloneDir
		case "extract_dir":
			return settings.ExtractDir
		default:
			return s
		}
	})
	return internal.ExpandUser(expanded)
}
