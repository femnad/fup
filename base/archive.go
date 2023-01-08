package base

import (
	"fmt"
	"os"
)

type Archive struct {
	ExecuteAfter []string `yaml:"execute_after"`
	Ref          string   `yaml:"name"`
	Symlink      []string `yaml:"link"`
	Target       string   `yaml:"target"`
	Unless       Unless   `yaml:"unless"`
	Url          string   `yaml:"url"`
	Version      string   `yaml:"version"`
	When         string   `yaml:"when"`
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

func (a Archive) version(s Settings) string {
	if a.Version != "" {
		return a.Version
	}

	return s.Versions[a.Name()]
}

func (a Archive) ExpandURL(s Settings) string {
	a.Version = a.version(s)
	return os.Expand(a.Url, a.expand)
}

func (a Archive) ExpandSymlinks(s Settings) []string {
	var expanded []string
	for _, symlink := range a.Symlink {
		a.Version = a.version(s)
		expanded = append(expanded, os.Expand(symlink, a.expand))
	}

	return expanded
}

func (a Archive) ExpandStat(settings Settings) string {
	return os.Expand(a.Unless.Stat, func(s string) string {
		if IsExpandable(s) {
			return ExpandSettings(settings, a.Unless.Stat)
		}
		if s == "version" {
			return a.Version
		}
		return fmt.Sprintf("${%s}", s)
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

func (a Archive) Name() string {
	return a.Ref
}

func (a Archive) RunWhen() string {
	return a.When
}
