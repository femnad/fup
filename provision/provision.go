package provision

import (
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
)

const binPath = "~/bin"

var (
	provisionOrder = []string{
		"preflight",
		"archive",
		"packages",
		"remove-packages",
		"cargo",
		"go",
		"python",
		"services",
		"tasks",
		"template",
	}
)

type Provisioner struct {
	Config       base.Config
	Provisioners map[string]func()
}

func NewProvisioner(cfg base.Config, provs []string) Provisioner {
	p := Provisioner{Config: cfg}
	provisioners := map[string]func(){
		"preflight":       p.runPreflightTasks,
		"archive":         p.extractArchives,
		"packages":        p.installPackages,
		"remove-packages": p.removePackages,
		"cargo":           p.cargoInstall,
		"go":              p.goInstall,
		"python":          p.pythonInstall,
		"services":        p.initServices,
		"tasks":           p.runTasks,
		"template":        p.applyTemplates,
	}

	if len(provs) == 0 {
		return Provisioner{Config: cfg, Provisioners: provisioners}
	}

	filtered := map[string]func(){}
	for _, desired := range provs {
		prov, ok := provisioners[desired]
		if !ok {
			internal.Log.Warningf("Unknown provisioner %s", desired)
			continue
		}

		filtered[desired] = prov
	}

	if len(filtered) == 0 {
		internal.Log.Warningf("Ignoring provisioner filter which returned no results")
		return Provisioner{Config: cfg, Provisioners: provisioners}
	}

	return Provisioner{Config: cfg, Provisioners: filtered}
}

func (p Provisioner) Apply() {
	for _, prov := range provisionOrder {
		fn, ok := p.Provisioners[prov]
		if !ok {
			continue
		}

		fn()
	}
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

func (p Provisioner) installPackages() {
	internal.Log.Notice("Installing packages")

	err := installPackages(p.Config.Packages)
	if err != nil {
		internal.Log.Errorf("error installing packages: %v", err)
	}
}

func (p Provisioner) removePackages() {
	internal.Log.Notice("Removing unwanted packages")

	err := removePackages(p.Config.UnwantedPackages)
	if err != nil {
		internal.Log.Errorf("error removing packages: %v", err)
	}
}

func (p Provisioner) cargoInstall() {
	internal.Log.Noticef("Installing Cargo packages")

	cargoInstallPkgs(p.Config)
}

func (p Provisioner) goInstall() {
	internal.Log.Noticef("Installing Go packages")

	goInstallPkgs(p.Config)
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
