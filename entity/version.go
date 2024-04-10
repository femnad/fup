package entity

import "github.com/femnad/fup/settings"

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
		return lookupVersion(v.GetVersionLookup(), v.GetLookupID())
	}

	return "", nil
}
