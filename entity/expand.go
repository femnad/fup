package entity

import (
	"github.com/femnad/fup/settings"
)

func ExpandSettings(stg settings.Settings, s string) string {
	return settings.ExpandStringWithLookup(stg, s, map[string]string{})
}
