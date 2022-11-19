package packages

type Dnf struct {
}

func (f Dnf) PkgExec() string {
	return "dnf"
}

func (f Dnf) PkgNameSeparator() string {
	return "."
}
