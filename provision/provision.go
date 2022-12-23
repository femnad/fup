package provision

import (
	"path"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/common"
	"github.com/femnad/fup/internal"
	precheck "github.com/femnad/fup/unless"
	"github.com/femnad/fup/unless/when"
)

const binPath = "~/bin"

type Provisioner struct {
	Config base.Config
}

func (p Provisioner) Apply() {
	p.runPreflightTasks()
	p.extractArchives()
	p.installPackages()
	p.removePackages()
	p.cargoInstall()
	p.initServices()
	p.runRecipes()
}

func createSymlink(symlink, extractDir string) {
	symlinkTarget := path.Join(extractDir, symlink)
	symlinkTarget = internal.ExpandUser(symlinkTarget)

	_, symlinkBasename := path.Split(symlink)
	symlinkName := path.Join(binPath, symlinkBasename)
	symlinkName = internal.ExpandUser(symlinkName)

	common.Symlink(symlinkName, symlinkTarget)
}

func extractArchive(archive base.Archive, settings base.Settings) {
	archive.Unless.Stat = archive.ExpandStat(settings)
	url := archive.ExpandURL(settings)

	if !when.ShouldRun(archive) {
		internal.Log.Debugf("Skipping extracting archive %s due to when condition %s", url, archive.When)
	}

	if precheck.ShouldSkip(archive, settings) {
		internal.Log.Debugf("Skipping download: %s", url)
		return
	}

	err := Extract(archive, settings)
	if err != nil {
		internal.Log.Errorf("Error downloading archive %s: %v", url, err)
		return
	}

	for _, symlink := range archive.ExpandSymlinks(settings) {
		createSymlink(symlink, settings.ExtractDir)
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
	internal.Log.Noticef("Installing cargo packages")

	cargoInstallPkgs(p.Config)
}

func (p Provisioner) initServices() {
	internal.Log.Noticef("Initializing services")

	for _, s := range p.Config.Services {
		initService(s, p.Config)
	}
}

func (p Provisioner) runRecipes() {
	internal.Log.Noticef("Running tasks")

	runTasks(p.Config)
}
