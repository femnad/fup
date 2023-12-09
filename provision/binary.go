package provision

import (
	"fmt"
	"os"
	"path"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/remote"
)

const defaultBinaryPerms = 0o755

func downloadBinary(binary entity.Binary, config base.Config) error {
	s := config.Settings
	url := binary.Url
	version := binary.Version
	if version == "" {
		version = config.Settings.Versions[binary.Name()]
	}

	url = settings.ExpandStringWithLookup(s, url, map[string]string{"version": version})

	if unless.ShouldSkip(binary, s) {
		internal.Log.Debug("skipping downloading binary %s", url)
		return nil
	}

	name := binary.Name()
	if name == "" {
		_, name = path.Split(url)
	}

	dir := binary.Dir
	if dir == "" {
		dir = name
	}

	binaryDir := fmt.Sprintf("%s/%s", internal.ExpandUser(config.Settings.ExtractDir), dir)
	binaryPath := fmt.Sprintf("%s/%s", binaryDir, name)

	internal.Log.Infof("Downloading binary %s", url)
	err := remote.Download(url, binaryPath)
	if err != nil {
		return err
	}

	err = os.Chmod(binaryPath, defaultBinaryPerms)
	if err != nil {
		return err
	}

	return createSymlink(base.NamedLink{Target: name}, binaryDir, s.GetBinPath())
}

func (p Provisioner) downloadBinaries() {
	internal.Log.Notice("Downloading binaries")

	for _, binary := range p.Config.Binaries {
		err := downloadBinary(binary, p.Config)
		if err != nil {
			internal.Log.Errorf("error downloading binary %s: %v", binary.Url, err)
		}
	}
}
