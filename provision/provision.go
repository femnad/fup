package provision

import (
	"os"
	"path"

	"github.com/femnad/fup/base"
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
}

func shouldUpdateSymlink(name, target string) (bool, bool) {
	_, err := os.Lstat(name)
	if err != nil {
		return false, true
	}

	currLink, err := os.Readlink(name)
	if err != nil {
		return true, true
	}

	return true, currLink != target
}

func createSymlink(symlink, extractDir string) {
	symlinkTarget := path.Join(extractDir, symlink)
	symlinkTarget = internal.ExpandUser(symlinkTarget)

	_, symlinkBasename := path.Split(symlink)
	symlinkName := path.Join(binPath, symlinkBasename)
	symlinkName = internal.ExpandUser(symlinkName)

	exists, update := shouldUpdateSymlink(symlinkName, symlinkTarget)
	if !update {
		internal.Log.Debugf("Symlink %s already exists", symlinkName)
		return
	}

	symlinkDir, _ := path.Split(symlinkName)
	err := mkdirAll(symlinkDir, dirMode)
	if err != nil {
		internal.Log.Errorf("Error creating symlink dir %s: %v", symlinkDir, err)
		return
	}

	internal.Log.Debugf("Creating symlink target=%s, name=%s", symlinkTarget, symlinkName)
	if exists {
		if err = os.Remove(symlinkName); err != nil {
			internal.Log.Errorf("Error removing existing symlink %s: %v", symlinkName, err)
			return
		}
	}

	err = os.Symlink(symlinkTarget, symlinkName)
	if err != nil {
		internal.Log.Errorf("Error creating symlink target=%s, name=%s: %v", symlinkTarget, symlinkName, err)
	}
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

func (p Provisioner) runPreflightTask(task base.Task) {
	if !when.ShouldRun(task) {
		internal.Log.Debugf("Skipping running task %s as when condition %s evaluated to false", task.Desc, task.When)
		return
	}

	if precheck.ShouldSkip(task, p.Config.Settings) {
		internal.Log.Debugf("Skipping running task %s as unless condition %s evaluated to true", task.Desc, task.Unless)
		return
	}

	internal.Log.Infof("Running task: %s", task.Name)
	task.Run()
}

func (p Provisioner) runPreflightTasks() {
	internal.Log.Notice("Running preflight tasks")

	for _, task := range p.Config.PreflightTasks {
		p.runPreflightTask(task)
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
