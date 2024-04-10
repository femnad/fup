package entity

import (
	"fmt"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/settings"
	"os"
)

type NamedLink struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
}

type ExecuteAfterSpec struct {
	Cmd    []string `yaml:"cmd"`
	SetPwd bool     `yaml:"set_pwd"`
}

type Release struct {
	Cleanup       bool              `yaml:"cleanup"`
	DontLink      bool              `yaml:"dont_link"`
	DontUpdate    bool              `yaml:"dont_update"`
	ExecuteAfter  ExecuteAfterSpec  `yaml:"execute_after"`
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

func (r Release) GetVersion() string {
	return r.Version
}

func (r Release) GetVersionLookup() VersionLookupSpec {
	return r.VersionLookup
}

func (r Release) GetLookupID() string {
	return r.VersionLookup.URL
}

func (r Release) String() string {
	return r.Url
}

func (r Release) expand(property string) string {
	if property == "version" {
		return r.Version
	}

	return ""
}

func (r Release) DefaultVersionCmd() string {
	return fmt.Sprintf("%s --version", r.Name())
}

func (r Release) ExpandURL(s settings.Settings) (string, error) {
	version, err := getVersion(r, s)
	if err != nil {
		return "", err
	}

	return settings.ExpandStringWithLookup(s, r.Url, map[string]string{"version": version}), nil
}

func (r Release) ExpandSymlinks(execCandidate string) []NamedLink {
	var links []NamedLink
	var expanded []NamedLink

	name := execCandidate
	if name == "" {
		name = r.Name()
	}

	symlinks := r.Symlink
	if len(r.NamedLink) == 0 && len(symlinks) == 0 && !r.DontLink && name != "" {
		symlinks = []string{name}
	}

	links = append(links, r.NamedLink...)
	for _, symlink := range symlinks {
		links = append(links, NamedLink{
			Target: symlink,
		})
	}

	for _, symlink := range links {
		symlink.Target = os.Expand(symlink.Target, r.expand)
		expanded = append(expanded, symlink)
	}

	return expanded
}

func (r Release) GetUnless() unless.Unless {
	return r.Unless
}

func (r Release) LookupVersion(s settings.Settings) (string, error) {
	return getVersion(r, s)
}

func (r Release) KeepUpToDate() bool {
	return !r.DontUpdate
}

func (r Release) Name() string {
	return r.Ref
}

func (r Release) RunWhen() string {
	return r.When
}
