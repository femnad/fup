package packages

type Apt struct {
}

func (a Apt) ListPkgsHeader() string {
	return "Listing..."
}

func (Apt) PkgExec() string {
	return "apt"
}

func (Apt) PkgNameSeparator() string {
	return "/"
}

func (Apt) RemoveCmd() string {
	return "purge"
}
