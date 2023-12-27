package provision

import (
	"path"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
)

func createSymlink(symlink entity.NamedLink, linkDir, binPath string) error {
	symlinkTarget := path.Join(linkDir, symlink.Target)
	symlinkTarget = internal.ExpandUser(symlinkTarget)

	symlinkBasename := symlink.Name
	if symlinkBasename == "" {
		name := symlink.Name
		if name == "" {
			name = symlinkTarget
		}
		_, symlinkBasename = path.Split(name)
	}
	symlinkName := path.Join(binPath, symlinkBasename)
	symlinkName = internal.ExpandUser(symlinkName)

	return common.Symlink(symlinkName, symlinkTarget)
}
