package provision

import (
	"fmt"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
)

const binPath = "~/bin"

type Provisioner struct {
	Config       base.Config
	Packager     packager
	provisioners provisioners
}

type provisionFn struct {
	name string
	fn   func()
}

type provisioners struct {
	provMap map[string]func()
	order   []string
}

func (p provisioners) apply() {
	for _, fnName := range p.order {
		fn := p.provMap[fnName]
		fn()
	}
}

func newProvisioners(allProvisioners []provisionFn, filter []string) (provisioners, error) {
	provMap := make(map[string]func())
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
		{"archive", p.extractArchives},
		{"binary", p.downloadBinaries},
		{"package", p.installPackages},
		{"rm-package", p.removePackages},
		{"known-hosts", p.acceptHostKeys},
		{"github-key", p.githubUserKey},
		{"cargo", p.cargoInstall},
		{"go", p.goInstall},
		{"python", p.pythonInstall},
		{"task", p.runTasks},
		{"template", p.applyTemplates},
		{"service", p.initServices},
		{"ensure-dir", p.ensureDirs},
		{"ensure-line", p.ensureLines},
		{"ssh-clone", p.sshClone},
		{"unwanted-dir", p.unwantedDirs},
		{"user-in-group", p.userInGroup},
		{"postflight", p.runPostflightTasks},
	}

	provs, err := newProvisioners(all, filter)
	if err != nil {
		return p, err
	}

	p.provisioners = provs
	return p, nil
}

func (p Provisioner) Apply() {
	p.provisioners.apply()
}

func (p Provisioner) extractArchives() {
	internal.Log.Notice("Extracting archives")

	for _, archive := range p.Config.Archives {
		extractArchive(archive, p.Config.Settings)
	}
}

func (p Provisioner) runPreflightTasks() {
	internal.Log.Notice("Running preflight tasks")

	for _, task := range p.Config.PreflightTasks {
		runTask(task, p.Config)
	}
}

func (p Provisioner) runPostflightTasks() {
	internal.Log.Notice("Running postflight tasks")

	for _, task := range p.Config.PostflightTasks {
		runTask(task, p.Config)
	}
}

func (p Provisioner) installPackages() {
	internal.Log.Notice("Installing packages")

	err := p.Packager.installPackages(p.Config.Packages)
	if err != nil {
		internal.Log.Errorf("error installing packages: %v", err)
	}

	err = p.Packager.installRemotePackages(p.Config.RemotePackages, p.Config.Settings)
	if err != nil {
		internal.Log.Errorf("error installing remote packages: %v", err)
	}
}

func (p Provisioner) removePackages() {
	internal.Log.Notice("Removing unwanted packages")

	err := p.Packager.removePackages(p.Config.UnwantedPackages)
	if err != nil {
		internal.Log.Errorf("error removing packages: %v", err)
	}
}

func (p Provisioner) cargoInstall() {
	internal.Log.Noticef("Installing Rust packages")

	cargoInstallPkgs(p.Config)
}

func (p Provisioner) githubUserKey() {
	internal.Log.Noticef("Adding GitHub user keys")

	addGithubUserKeys(p.Config)
}

func (p Provisioner) goInstall() {
	internal.Log.Noticef("Installing Go packages")

	goInstallPkgs(p.Config)
}

func (p Provisioner) acceptHostKeys() {
	internal.Log.Noticef("Adding known hosts")

	acceptHostKeys(p.Config)
}

func (p Provisioner) initServices() {
	internal.Log.Noticef("Initializing services")

	for _, s := range p.Config.Services {
		initService(s, p.Config)
	}
}

func (p Provisioner) pythonInstall() {
	internal.Log.Notice("Installing Python packages")

	pythonInstallPkgs(p.Config)
}

func (p Provisioner) runTasks() {
	internal.Log.Noticef("Running tasks")

	runTasks(p.Config)
}

func (p Provisioner) applyTemplates() {
	internal.Log.Noticef("Applying templates")

	applyTemplates(p.Config)
}

func (p Provisioner) ensureDirs() {
	internal.Log.Noticef("Creating desired dirs")

	ensureDirs(p.Config)
}

func (p Provisioner) ensureLines() {
	internal.Log.Noticef("Ensuring lines in files")

	ensureLines(p.Config)
}

func (p Provisioner) sshClone() {
	internal.Log.Noticef("Cloning repos via SSH")

	sshClone(p.Config)
}

func (p Provisioner) unwantedDirs() {
	internal.Log.Noticef("Removing unwanted dirs")

	removeUnwantedDirs(p.Config)
}

func (p Provisioner) userInGroup() {
	internal.Log.Noticef("Ensuring user is in desired groups")

	userInGroup(p.Config)
}
