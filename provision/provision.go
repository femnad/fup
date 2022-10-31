package provision

import (
	"github.com/femnad/fup/base"
	"log"
)

type Provisioner struct {
	Config base.Config
}

func (p Provisioner) Apply() {
	p.downloadArchives()
}

func (p Provisioner) downloadArchives() {
	for _, archive := range p.Config.Archives {
		err := Extract(archive, p.Config.Settings.ArchiveDir)
		if err != nil {
			log.Fatalf("Error downloading archive %v: %v", archive.Url, err)
		}
	}
}
