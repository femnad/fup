package common

import "os"

func IsExecutableFile(info os.FileInfo) bool {
	return !info.IsDir() && info.Mode().Perm()&0100 != 0
}
