package packages

type Dnf struct {
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
