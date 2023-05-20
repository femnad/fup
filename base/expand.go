package base

import (
	"github.com/femnad/fup/base/settings"
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

func ExpandSettings(stg settings.Settings, s string) string {
	return settings.ExpandStringWithLookup(stg, s, map[string]string{})
}
