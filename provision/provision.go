package provision

import (
	"errors"
	"fmt"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
)

type Provisioner struct {
	Config       base.Config
	Packager     packager
	provisioners provisioners
}

type provisionFn struct {
	name string
	fn   func() error
}

type provisioners struct {
	provMap map[string]func() error
	order   []string
}

func (p provisioners) apply() error {
	var provErrs []error
	for _, fnName := range p.order {
		fn := p.provMap[fnName]
		err := fn()
		provErrs = append(provErrs, err)
	}

	var uniqErrs []error
	seenErr := make(map[string]error)
	for _, err := range provErrs {
		msg := err.Error()
		_, ok := seenErr[msg]
		if !ok {
			seenErr[msg] = err
			uniqErrs = append(uniqErrs, err)
		}
	}

	return errors.Join(uniqErrs...)
}

func newProvisioners(allProvisioners []provisionFn, filter []string) (provisioners, error) {
	provMap := make(map[string]func() error)
	var order []string
	var hasFilter = len(filter) > 0

	for _, prov := range allProvisioners {
		provMap[prov.name] = prov.fn
		if !hasFilter {
			order = append(order, prov.name)
		}
	}

	if hasFilter {
		for _, fnName := range filter {
			_, ok := provMap[fnName]
			if !ok {
				return provisioners{}, fmt.Errorf("%s is not a provisioning function", fnName)
			}

			order = append(order, fnName)
		}
	}

	return provisioners{
		provMap: provMap,
		order:   order,
	}, nil
}

func NewProvisioner(cfg base.Config, filter []string) (Provisioner, error) {
	pkgr, err := newPackager()
	if err != nil {
		return Provisioner{}, err
	}

	p := Provisioner{Config: cfg, Packager: pkgr}

	all := []provisionFn{
		{"preflight", p.runPreflightTasks},
		{"repo", p.AddOSRepos},
		{"archive", p.extractArchives},
		{"binary", p.downloadBinaries},
		{"package", p.installPackages},
		{"rm-package", p.removePackages},
		{"known-hosts", p.acceptHostKeys},
		{"github-key", p.githubUserKey},
		{"go", p.goInstall},
		{"python", p.pythonInstall},
		{"rust", p.rustInstall},
		{"task", p.runTasks},
		{"template", p.applyTemplates},
		{"service", p.initServices},
		{"ensure-dir", p.ensureDirs},
		{"ensure-line", p.ensureLines},
		{"flatpak", p.flatpakInstall},
		{"snap", p.snapInstall},
		{"ssh-clone", p.sshClone},
		{"unwanted-dir", p.unwantedDirs},
		{"user-in-group", p.userInGroup},
		{"postflight", p.runPostFlightTasks},
	}

	provs, err := newProvisioners(all, filter)
	if err != nil {
		return p, err
	}

	p.provisioners = provs
	return p, nil
}

func (p Provisioner) Apply() error {
	return p.provisioners.apply()
}

func (p Provisioner) AddOSRepos() error {
	internal.Log.Notice("Adding OS repos")

	return addRepos(p.Config)
}

func (p Provisioner) extractArchives() error {
	internal.Log.Notice("Extracting archives")

	if p.Config.Settings.ExtractDir == "" {
		return errors.New("empty archive extraction directory")
	}

	return extractArchives(p.Config.Archives, p.Config.Settings)
}

func (p Provisioner) runPreflightTasks() error {
	internal.Log.Notice("Running preflight tasks")

	return runTasks(p.Config, p.Config.PreflightTasks)
}

func (p Provisioner) runPostFlightTasks() error {
	internal.Log.Notice("Running postflight tasks")

	return runTasks(p.Config, p.Config.PostflightTasks)
}

func (p Provisioner) installPackages() error {
	internal.Log.Notice("Installing packages")

	return installPackages(p)
}

func (p Provisioner) removePackages() error {
	internal.Log.Notice("Removing unwanted packages")

	err := p.Packager.removePackages(p.Config.UnwantedPackages)
	if err != nil {
		internal.Log.Errorf("error removing packages: %v", err)
		return err
	}

	return nil
}

func (p Provisioner) rustInstall() error {
	internal.Log.Noticef("Installing Rust packages")

	return cargoInstallPkgs(p.Config)
}

func (p Provisioner) githubUserKey() error {
	internal.Log.Noticef("Adding GitHub user keys")

	return addGithubUserKeys(p.Config)
}

func (p Provisioner) goInstall() error {
	internal.Log.Noticef("Installing Go packages")

	return goInstallPkgs(p.Config)
}

func (p Provisioner) acceptHostKeys() error {
	internal.Log.Noticef("Adding known hosts")

	return acceptHostKeys(p.Config)
}

func (p Provisioner) initServices() error {
	internal.Log.Noticef("Initializing services")

	return initServices(p.Config)
}

func (p Provisioner) pythonInstall() error {
	internal.Log.Notice("Installing Python packages")

	return pythonInstallPkgs(p.Config)
}

func (p Provisioner) runTasks() error {
	internal.Log.Noticef("Running tasks")

	return runTasks(p.Config, p.Config.Tasks)
}

func (p Provisioner) applyTemplates() error {
	internal.Log.Noticef("Applying templates")

	return applyTemplates(p.Config)
}

func (p Provisioner) ensureDirs() error {
	internal.Log.Noticef("Creating desired dirs")

	return ensureDirs(p.Config)
}

func (p Provisioner) ensureLines() error {
	internal.Log.Noticef("Ensuring lines in files")

	return ensureLines(p.Config)
}

func (p Provisioner) flatpakInstall() error {
	internal.Log.Noticef("Installing flatpak packages")

	_, err := common.Which("flatpak")
	if err != nil {
		internal.Log.Debug("Skipping Flatpak packages installation as Flatpak is not available")
		return nil
	}

	return flatpakInstall(p.Config)
}

func (p Provisioner) snapInstall() error {
	internal.Log.Noticef("Installing snap packages")

	return snapInstall(p.Config)
}

func (p Provisioner) sshClone() error {
	internal.Log.Noticef("Cloning repos via SSH")

	return sshClone(p.Config)
}

func (p Provisioner) unwantedDirs() error {
	internal.Log.Noticef("Removing unwanted dirs")

	return removeUnwantedDirs(p.Config)
}

func (p Provisioner) userInGroup() error {
	internal.Log.Noticef("Ensuring user is in desired groups")

	return userInGroup(p.Config)
}
