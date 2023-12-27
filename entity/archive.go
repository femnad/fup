package entity

import (
	"fmt"
	"os"

	"github.com/antchfx/htmlquery"

	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/remote"
	"github.com/femnad/fup/settings"
)

type NamedLink struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
}

type VersionLookupSpec struct {
	FollowURL bool   `yaml:"follow_url"`
	PostProc  string `yaml:"post_proc"`
	Query     string `yaml:"query"`
	URL       string `yaml:"url"`
}

type Archive struct {
	DontLink      bool              `yaml:"dont_link"`
	DontUpdate    bool              `yaml:"dont_update"`
	ExecuteAfter  []string          `yaml:"execute_after"`
	NamedLink     []NamedLink       `yaml:"named_link"`
	Ref           string            `yaml:"name"`
	Symlink       []string          `yaml:"link"`
	Target        string            `yaml:"target"`
	Unless        unless.Unless     `yaml:"unless"`
	Url           string            `yaml:"url"`
	Version       string            `yaml:"version"`
	VersionLookup VersionLookupSpec `yaml:"version_lookup"`
	When          string            `yaml:"when"`
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

func (a Archive) hasVersionLookup() bool {
	return a.VersionLookup.URL != ""
}

func (a Archive) version(s settings.Settings) (string, error) {
	if a.Version != "" {
		return a.Version, nil
	}

	storedVersion := s.Versions[a.Name()]
	if storedVersion != "" {
		return storedVersion, nil
	}

	if a.hasVersionLookup() {
		return lookupVersion(a.VersionLookup)
	}

	return "", nil
}

func (a Archive) DefaultVersionCmd() string {
	return fmt.Sprintf("%s --version", a.Name())
}

func (a Archive) ExpandURL(s settings.Settings) (string, error) {
	version, err := a.version(s)
	if err != nil {
		return "", err
	}

	return settings.ExpandStringWithLookup(s, a.Url, map[string]string{"version": version}), nil
}

func (a Archive) ExpandSymlinks(maybeExec string) []NamedLink {
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

func resolveQuery(spec VersionLookupSpec) (string, error) {
	doc, err := htmlquery.LoadURL(spec.URL)
	if err != nil {
		return "", err
	}

	node, err := htmlquery.Query(doc, spec.Query)
	if err != nil {
		return "", err
	}

	if node == nil {
		return "", fmt.Errorf("error looking up version from spec %+v", spec)
	}

	return htmlquery.InnerText(node), nil
}

func lookupVersion(spec VersionLookupSpec) (text string, err error) {
	if spec.Query != "" {
		text, err = resolveQuery(spec)
		if err != nil {
			if err != nil {
				return "", err
			}
		}
	} else if spec.FollowURL {
		text, err = remote.FollowRedirects(spec.URL)
		if err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("version lookup requires either a query or follow_url to be set")
	}

	if spec.PostProc != "" {
		text, err = internal.RunTemplateFn(text, spec.PostProc)
		if err != nil {
			return "", err
		}
	}

	return text, nil
}

func (a Archive) GetVersion(s settings.Settings) (string, error) {
	return a.version(s)
}

func (a Archive) KeepUpToDate() bool {
	return !a.DontUpdate
}

func (a Archive) Name() string {
	return a.Ref
}

func (a Archive) RunWhen() string {
	return a.When
}
