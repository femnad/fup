package entity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/settings"
	marecmd "github.com/femnad/mare/cmd"
)

const (
	pluginsCore      = "dnf-plugins-core"
	repoFileTemplate = `[{{ .Name }}]
name="{{ .Description }}"
baseurl={{ .URL }}
enabled=1
gpgcheck=1
gpgkey="{{ .GPGKey }}"
`
)

type installer struct {
	isRoot bool
}

type repoListEntry struct {
	ID string `json:"id"`
}

func (i installer) runMaybeSudo(cmd string) error {
	_, err := marecmd.RunFmtErr(marecmd.Input{Command: cmd, Sudo: !i.isRoot})
	return err
}

type repoSpec struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	URL         string `yaml:"url"`
	GPGKey      string `yaml:"gpg_key"`
}

type DnfRepo struct {
	unless.BasicUnlessable
	RepoName string   `yaml:"name"`
	Packages []string `yaml:"packages"`
	Repo     string   `yaml:"repo"`
	RepoSpec repoSpec `yaml:"repo_spec"`
	Url      []string `yaml:"url"`
	When     string   `yaml:"when"`
}

func (d DnfRepo) DefaultVersionCmd() string {
	return ""
}

func (d DnfRepo) GetUnless() unless.Unless {
	if d.Repo != "" {
		filename := internal.FilenameWithoutSuffix(d.Repo, ".repo")
		return unless.Unless{
			Stat: fmt.Sprintf("/etc/yum.repos.d/%s.repo", filename),
		}
	}

	if len(d.Packages) > 0 {
		packages := strings.Join(d.Packages, " ")
		return unless.Unless{
			Cmd: fmt.Sprintf("dnf list --installed %s", packages),
		}
	}

	return unless.Unless{}
}

func (DnfRepo) LookupVersion(_ settings.Settings) (string, error) {
	return "", nil
}

func (d DnfRepo) Name() string {
	return d.RepoName
}

func (d DnfRepo) RunWhen() string {
	return d.When
}

func (i installer) installCorePlugins() error {
	cmd := fmt.Sprintf("dnf install -qy %s", pluginsCore)
	return i.runMaybeSudo(cmd)
}

func (i installer) ensureCorePlugins() error {
	cmd := fmt.Sprintf("dnf list --installed %s", pluginsCore)
	out, err := marecmd.Run(marecmd.Input{Command: cmd})
	if out.Code == 1 {
		return i.installCorePlugins()
	}

	return err
}

func (i installer) configManagerInstall(repo string) error {
	err := i.ensureCorePlugins()
	if err != nil {
		return err
	}

	cmd := fmt.Sprintf("dnf config-manager addrepo --from-repofile=%s", repo)
	return i.runMaybeSudo(cmd)
}

func (i installer) releasePackagesInstall(url []string, osId string) error {
	packageList := strings.Join(url, " ")
	cmd := fmt.Sprintf("rpm -E %%%s", osId)
	out, err := marecmd.RunFmtErr(marecmd.Input{Command: cmd})
	if err != nil {
		return err
	}

	packageList = settings.Expand(packageList, map[string]string{"version_id": strings.TrimSpace(out.Stdout)})
	cmd = fmt.Sprintf("dnf install -y %s", packageList)
	return i.runMaybeSudo(cmd)
}

func writeRepoSpec(spec repoSpec) error {
	tmpl, err := template.New("repo").Parse(repoFileTemplate)
	if err != nil {
		return err
	}

	repoFile := path.Join("/etc/yum.repos.d", fmt.Sprintf("%s.repo", spec.Name))
	if _, err = os.Stat(repoFile); err == nil {
		return nil
	}

	out := bytes.Buffer{}

	err = tmpl.Execute(&out, spec)
	if err != nil {
		return err
	}

	err = internal.MaybeRunWithSudo(fmt.Sprintf("rpm --import %s", spec.GPGKey))
	if err != nil {
		return err
	}

	_, err = internal.WriteContent(internal.ManagedFile{
		Content: out.String(),
		Path:    repoFile,
		Mode:    0644,
		User:    "root",
		Group:   "root",
	})
	return nil
}

func (d DnfRepo) Exists() (bool, error) {
	out, err := marecmd.RunOutBuffer(marecmd.Input{Command: "dnf repolist --enabled --json"})
	if err != nil {
		return false, err
	}

	var repoListOut []repoListEntry
	err = json.NewDecoder(&out.Stdout).Decode(&repoListOut)
	if err != nil {
		return false, err
	}

	for _, entry := range repoListOut {
		if entry.ID == d.RepoName {
			return true, nil
		}
	}

	return false, nil
}

func (d DnfRepo) Install() error {
	isRoot, err := internal.IsUserRoot()
	if err != nil {
		return err
	}

	empty := repoSpec{}
	if d.RepoSpec != empty {
		return writeRepoSpec(d.RepoSpec)
	}

	i := installer{isRoot: isRoot}
	if d.Repo != "" {
		return i.configManagerInstall(d.Repo)
	}

	if len(d.Packages) > 0 {
		var osId string
		osId, err = precheck.GetOSId()
		if err != nil {
			return err
		}
		return i.releasePackagesInstall(d.Url, osId)
	}

	return fmt.Errorf("unable to determine install method for repo %+v", d)
}
