package provision

import (
	"os"
	"path"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
	precheck "github.com/femnad/fup/unless"
)

const binPath = "~/bin"

type Provisioner struct {
	Config base.Config
}

func (p Provisioner) Apply() {
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
	if precheck.ShouldSkip(archive, settings) {
		internal.Log.Infof("Skipping download: %s", archive.ShortURL())
		return
	}

	err := Extract(archive, settings.ExtractDir)
	if err != nil {
		internal.Log.Errorf("Error downloading archive %s: %v", archive.ShortURL(), err)
		return
	}

	for _, symlink := range archive.ExpandSymlinks() {
		createSymlink(symlink, settings.ExtractDir)
	}
}

func (p Provisioner) extractArchives() {
	for _, archive := range p.Config.Archives {
		extractArchive(archive, p.Config.Settings)
	}
}
