package provision

import (
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
	precheck "github.com/femnad/fup/unless"
	"os"
	"path"
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
	_, symlinkBasename := path.Split(symlink)
	symlinkName := path.Join(binPath, symlinkBasename)
	err := os.Symlink(symlinkTarget, symlinkName)
	if err != nil {
		internal.Log.Errorf("Error creating symlink target=%s, name=%s: %v", symlinkTarget, symlinkName, err)
	}
}

func extractArchive(archive base.Archive, extractDir string) {
	if !precheck.ShouldRun(archive.Unless, archive.Version) {
		internal.Log.Infof("Skipping download: %s", archive.ShortURL())
		return
	}

	err := Extract(archive, extractDir)
	if err != nil {
		internal.Log.Errorf("Error downloading archive %s: %v", archive.Url, err)
		return
	}

	for _, symlink := range archive.Symlink {
		createSymlink(extractDir, symlink)
	}
}

func (p Provisioner) extractArchives() {
	for _, archive := range p.Config.Archives {
		if !precheck.ShouldRun(archive.Unless, archive.Version) {
			internal.Log.Infof("Skipping download: %s", archive.ShortURL())
			continue
		}

		extractArchive(archive, p.Config.Settings.ExtractDir)
	}
}
