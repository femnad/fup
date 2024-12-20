package entity

import (
	"fmt"

	"github.com/femnad/fup/settings"
)

type versioned interface {
	GetVersion() string
	GetVersionLookup() VersionLookupSpec
	GetLookupID() string
	Name() string
}

func hasVersionLookup(v versioned) bool {
	spec := v.GetVersionLookup()
	return spec.URL != "" || spec.Strategy != ""
}

func getVersion(v versioned, s settings.Settings) (string, error) {
	version := v.GetVersion()
	if version != "" {
		return version, nil
	}

	storedVersion := s.Versions[v.Name()]
	if storedVersion != "" {
		return storedVersion, nil
	}

	if hasVersionLookup(v) {
		lookupId := v.GetLookupID()
		if lookupId == "" {
			return "", fmt.Errorf("lookup ID for versioned %+v is empty", v)
		}
		return LookupVersion(v.GetVersionLookup(), v.GetLookupID(), s)
	}

	return "", nil
}
