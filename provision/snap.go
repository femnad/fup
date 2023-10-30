package provision

import (
	"fmt"
	"github.com/femnad/fup/base"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	marecmd "github.com/femnad/mare/cmd"
)

func installSnap(snap entity.Snap) {
	cmd := fmt.Sprintf("snap install %s", snap.Name)
	if snap.Classic {
		cmd += " --classic"
	}
	in := marecmd.Input{Command: cmd}

	_, err := marecmd.RunFormatError(in)
	if err != nil {
		internal.Log.Errorf("error installing snap %s: %v", snap.Name, err)
	}
}

func snapInstall(config base.Config) {
	for _, snap := range config.SnapPackages {
		installSnap(snap)
	}
}
