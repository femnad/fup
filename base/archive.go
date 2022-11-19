package base

import (
	"os"

	"github.com/femnad/fup/internal"
)

type Archive struct {
	Url     string   `yaml:"url"`
	Unless  Unless   `yaml:"unless"`
	Version string   `yaml:"version"`
	Symlink []string `yaml:"symlink"`
	Binary  string   `yaml:"binary"`
	When    string   `yaml:"when"`
}

func (a Archive) String() string {
	return a.Url
}

func (a Archive) expand(property string) string {
	if property == "version" {
		return a.Version
	}

	return ""
}

func (a Archive) ExpandURL() string {
	return os.Expand(a.Url, a.expand)
}

func (a Archive) ExpandSymlinks() []string {
	var expanded []string
	for _, symlink := range a.Symlink {
		expanded = append(expanded, os.Expand(symlink, a.expand))
	}

	return expanded
}

func (a Archive) ExpandStat(settings Settings) string {
	return os.Expand(a.Unless.Stat, func(s string) string {
		if s == "extract_dir" {
			extractDir := settings.ExtractDir
			return internal.ExpandUser(extractDir)
		}
		if s == "version" {
			return a.Version
		}
		return s
	})
}

func (a Archive) GetUnless() Unless {
	return a.Unless
}

func (a Archive) GetVersion() string {
	return a.Version
}

func (a Archive) HasPostProc() bool {
	return a.Unless.HasPostProc()
}

func (a Archive) RunWhen() string {
	return a.When
}
