package entity

import (
	"fmt"
	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck"
	"github.com/femnad/fup/precheck/unless"
	marecmd "github.com/femnad/mare/cmd"
	"strings"
)

const pluginsCore = "dnf-plugins-core"

type installer struct {
	isRoot bool
}

func (i installer) runMaybeSudo(cmd string) error {
	_, err := marecmd.RunFormatError(marecmd.Input{Command: cmd, Sudo: !i.isRoot})
	return err
}

type DnfRepo struct {
	Packages []string `yaml:"packages"`
	Repo     string   `yaml:"repo"`
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

func (DnfRepo) GetVersion() string {
	return ""
}

func (DnfRepo) HasPostProc() bool {
	return false
}

func (DnfRepo) Name() string {
	return ""
}

func (d DnfRepo) RunWhen() string {
	return d.When
}

func (i installer) installCorePlugins() error {
	cmd := fmt.Sprintf("dnf install -y %s", pluginsCore)
	return i.runMaybeSudo(cmd)
}

func (i installer) ensureCorePlugins() error {
	cmd := fmt.Sprintf("dnf list installed %s", pluginsCore)
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

	cmd := fmt.Sprintf("dnf config-manager --add-repo %s", repo)
	return i.runMaybeSudo(cmd)
}

func (i installer) releasePackagesInstall(packages []string, osId string) error {
	packageList := strings.Join(packages, " ")
	cmd := fmt.Sprintf("rpm -E %%%s", osId)
	out, err := marecmd.RunFormatError(marecmd.Input{Command: cmd})
	if err != nil {
		return err
	}

	packageList = settings.Expand(packageList, map[string]string{"version_id": out.Stdout})
	cmd = fmt.Sprintf("dnf install -y %s", packageList)
	return i.runMaybeSudo(cmd)
}

func (d DnfRepo) Install() error {
	isRoot, err := internal.IsUserRoot()
	if err != nil {
		return err
	}

	i := installer{isRoot: isRoot}
	if d.Repo != "" {
		return i.configManagerInstall(d.Repo)
	}

	if len(d.Packages) > 0 {
		osId, err := precheck.GetOsId()
		if err != nil {
			return err
		}
		return i.releasePackagesInstall(d.Packages, osId)
	}

	return fmt.Errorf("unable to determine install method for repo %+v", d)
}
