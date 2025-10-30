package entity

import (
	"fmt"
	"os"

	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/settings"
)

type NamedLink struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
}

type ExecuteSpec struct {
	Cmd    []string `yaml:"cmd"`
	SetPwd bool     `yaml:"set_pwd"`
	Sudo   bool     `yaml:"sudo"`
}

type Release struct {
	ChromeSandbox string            `yaml:"chrome-sandbox,omitempty"`
	Cleanup       bool              `yaml:"cleanup,omitempty"`
	DontLink      bool              `yaml:"dont_link,omitempty"`
	DontUpdate    bool              `yaml:"dont_update,omitempty"`
	ExecuteAfter  ExecuteSpec       `yaml:"execute_after,omitempty"`
	ExecuteBefore ExecuteSpec       `yaml:"execute_before,omitempty"`
	NamedLink     []NamedLink       `yaml:"named_link,omitempty"`
	Ref           string            `yaml:"name,omitempty"`
	Symlink       []string          `yaml:"link,omitempty"`
	Target        string            `yaml:"target,omitempty"`
	Unless        unless.Unless     `yaml:"unless,omitempty"`
	Url           string            `yaml:"url,omitempty"`
	Version       string            `yaml:"version,omitempty"`
	VersionLookup VersionLookupSpec `yaml:"version_lookup,omitempty"`
	When          string            `yaml:"when,omitempty"`
}

func (r Release) GetVersion() string {
	return r.Version
}

func (r Release) GetVersionLookup() VersionLookupSpec {
	return r.VersionLookup
}

func (r Release) GetLookupID() string {
	return r.Url
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
