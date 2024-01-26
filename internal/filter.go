package internal

import (
	"os"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

// Contains checks if needle exists in elems.
// Yoinked from https://gosamples.dev/generics-slice-contains/.
func Contains[T comparable](elems []T, needle T) bool {
	for _, elem := range elems {
		if needle == elem {
			return true
		}
	}
	return false
}

func ExpandUser(path string) string {
	return strings.Replace(path, "~", os.Getenv("HOME"), 1)
}

func SetFromList[T comparable](items []T) mapset.Set[T] {
	set := mapset.NewSet[T]()
	for _, item := range items {
		set.Add(item)
	}
	return set
}
