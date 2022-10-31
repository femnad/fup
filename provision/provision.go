package provision

import (
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
	precheck "github.com/femnad/fup/unless"
)

type Provisioner struct {
	Config base.Config
}

func (p Provisioner) Apply() {
	p.downloadArchives()
}

func (p Provisioner) downloadArchives() {
	for _, archive := range p.Config.Archives {
		if !precheck.ShouldRun(archive.Unless, archive.Version) {
			internal.Log.Infof("Skipping downloading archive %s", archive)
			continue
		}
		err := Extract(archive, p.Config.Settings.ArchiveDir)
		if err != nil {
			internal.Log.Errorf("Error downloading archive %v: %v", archive.Url, err)
		}
	}
}
