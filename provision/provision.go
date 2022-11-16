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
}

func createSymlink(symlink, extractDir string) {
	symlinkTarget := path.Join(extractDir, symlink)
	symlinkTarget = internal.ExpandUser(symlinkTarget)

	_, symlinkBasename := path.Split(symlink)
	symlinkName := path.Join(binPath, symlinkBasename)
	symlinkName = internal.ExpandUser(symlinkName)

	_, err := os.Stat(symlinkName)
	if err == nil {
		internal.Log.Debugf("Symlink %s already exists", symlinkName)
		return
	}

	internal.Log.Debugf("Creating symlink target=%s, name=%s", symlinkTarget, symlinkName)
	err = os.Symlink(symlinkTarget, symlinkName)
	if err != nil {
		internal.Log.Errorf("Error creating symlink target=%s, name=%s: %v", symlinkTarget, symlinkName, err)
	}
}

func extractArchive(archive base.Archive, settings base.Settings) {
	archive.Unless.Stat = archive.ExpandStat(settings)

	if precheck.ShouldSkip(archive) {
		internal.Log.Debugf("Skipping download: %s", archive.ExpandURL())
		return
	}

	err := Extract(archive, settings.ExtractDir)
	if err != nil {
		internal.Log.Errorf("Error downloading archive %s: %v", archive.ExpandURL(), err)
		return
	}

	for _, symlink := range archive.ExpandSymlinks() {
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
		internal.Log.Debugf("Skipping running task %s as when condition %s evaluated to false", task.Name, task.When)
		return
	}

	if precheck.ShouldSkip(task) {
		internal.Log.Debugf("Skipping running task %s as unless condition %s evaluated to true", task.Name, task.Unless)
		return
	}

	internal.Log.Debugf("Running task %s", task.Name)
	task.Run()
}

func (p Provisioner) runPreflightTasks() {
	for _, task := range p.Config.PreflightTasks {
		p.runPreflightTask(task)
	}
}
