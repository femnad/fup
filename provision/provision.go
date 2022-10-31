package provision

import (
	"github.com/femnad/fup/base"
	precheck "github.com/femnad/fup/unless"
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
		if !precheck.ShouldRun(archive.Unless, archive.Version) {
			log.Printf("Skipping archive based on unless eval: `%s`", archive.Unless)
			continue
		}
		err := Extract(archive, p.Config.Settings.ArchiveDir)
		if err != nil {
			log.Fatalf("Error downloading archive %v: %v", archive.Url, err)
		}
	}
}
