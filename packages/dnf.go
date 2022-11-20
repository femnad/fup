package packages

type Dnf struct {
}

func (d Dnf) ListPkgsHeader() string {
	return "Installed Packages"
}

func (Dnf) PkgExec() string {
	return "dnf"
}

func (Dnf) PkgNameSeparator() string {
	return "."
}

func (Dnf) RemoveCmd() string {
	return "remove"
}
