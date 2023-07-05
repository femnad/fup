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
		"binary",
		"packages",
		"remove-packages",
		"known-hosts",
		"github-key",
		"cargo",
		"go",
		"python",
		"template",
		"services",
		"tasks",
		"ensure-dirs",
		"ensure-lines",
		"self-clone",
		"unwanted-dirs",
		"user-in-group",
		"postflight",
	}
)

type Provisioner struct {
	Config       base.Config
	Provisioners map[string]func()
}

func NewProvisioner(cfg base.Config, provs []string) Provisioner {
	p := Provisioner{Config: cfg}
	provisioners := map[string]func(){
		"archive":         p.extractArchives,
		"binary":          p.downloadBinaries,
		"cargo":           p.cargoInstall,
		"ensure-dirs":     p.ensureDirs,
		"ensure-lines":    p.ensureLines,
		"github-key":      p.githubUserKey,
		"go":              p.goInstall,
		"known-hosts":     p.acceptHostKeys,
		"packages":        p.installPackages,
		"postflight":      p.runPostflightTasks,
		"preflight":       p.runPreflightTasks,
		"python":          p.pythonInstall,
		"remove-packages": p.removePackages,
		"self-clone":      p.selfClone,
		"services":        p.initServices,
		"tasks":           p.runTasks,
		"template":        p.applyTemplates,
		"unwanted-dirs":   p.unwantedDirs,
		"user-in-group":   p.userInGroup,
	}

	if len(provs) == 0 {
		return Provisioner{Config: cfg, Provisioners: provisioners}
	}

	filtered := map[string]func(){}
	for _, desired := range provs {
		prov, ok := provisioners[desired]
		if !ok {
			internal.Log.Warningf("Unknown provisioner %s", desired)
			return p
		}

		filtered[desired] = prov
	}

	if len(filtered) == 0 {
		internal.Log.Warningf("Ignoring provisioner filter which returned no results")
		return p
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

func (p Provisioner) runPostflightTasks() {
	internal.Log.Notice("Running postlight tasks")

	for _, task := range p.Config.PostflightTasks {
		runTask(task, p.Config)
	}
}

func (p Provisioner) installPackages() {
	internal.Log.Notice("Installing packages")

	err := installPackages(p.Config.Packages)
	if err != nil {
		internal.Log.Errorf("error installing packages: %v", err)
	}

	err = installRemotePackages(p.Config.RemotePackages)
	if err != nil {
		internal.Log.Errorf("error installing remote packages: %v", err)
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

func (p Provisioner) selfClone() {
	internal.Log.Noticef("Cloning own repos")

	selfClone(p.Config)
}

func (p Provisioner) unwantedDirs() {
	internal.Log.Noticef("Removing unwanted dirs")

	removeUnwantedDirs(p.Config)
}

func (p Provisioner) userInGroup() {
	internal.Log.Noticef("Ensuring user is in desired groups")

	userInGroup(p.Config)
}
