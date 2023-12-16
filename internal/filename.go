package internal

import "strings"

func FilenameWithoutSuffix(filename, suffix string) string {
	tokens := strings.Split(filename, "/")
	filename = tokens[len(tokens)-1]
	return strings.TrimSuffix(filename, suffix)
}
