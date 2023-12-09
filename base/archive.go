package base

import (
	"fmt"
	"os"

	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/precheck/unless"
)

type NamedLink struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
}

type Archive struct {
	DontLink     bool          `yaml:"dont_link"`
	ExecuteAfter []string      `yaml:"execute_after"`
	Ref          string        `yaml:"name"`
	NamedLink    []NamedLink   `yaml:"named_link"`
	Symlink      []string      `yaml:"link"`
	Unless       unless.Unless `yaml:"unless"`
	Url          string        `yaml:"url"`
	Version      string        `yaml:"version"`
	When         string        `yaml:"when"`
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

func (a Archive) version(s settings.Settings) string {
	if a.Version != "" {
		return a.Version
	}

	return s.Versions[a.Name()]
}

func (a Archive) DefaultVersionCmd() string {
	return fmt.Sprintf("%s --version", a.Name())
}

func (a Archive) ExpandURL(s settings.Settings) string {
	a.Version = a.version(s)
	return os.Expand(a.Url, a.expand)
}

func (a Archive) ExpandSymlinks(s settings.Settings, maybeExec string) []NamedLink {
	var links []NamedLink
	var expanded []NamedLink

	name := a.Name()
	if name == "" && maybeExec != "" {
		name = maybeExec
	}
	symlinks := a.Symlink
	if len(a.NamedLink) == 0 && len(symlinks) == 0 && !a.DontLink && name != "" {
		symlinks = []string{name}
	}

	links = append(links, a.NamedLink...)
	for _, symlink := range symlinks {
		links = append(links, NamedLink{
			Target: symlink,
		})
	}

	for _, symlink := range links {
		a.Version = a.version(s)
		symlink.Target = os.Expand(symlink.Target, a.expand)
		expanded = append(expanded, symlink)
	}

	return expanded
}

func (a Archive) ExpandStat(settings settings.Settings) string {
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

func (a Archive) GetUnless() unless.Unless {
	return a.Unless
}

func (a Archive) GetVersion() string {
	return a.Version
}

func (a Archive) HasPostProc() bool {
	return a.Unless.HasPostProc()
}

func (a Archive) Name() string {
	if a.Ref != "" {
		return a.Ref
	}
	return ""
}

func (a Archive) RunWhen() string {
	return a.When
}
