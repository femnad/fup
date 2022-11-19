package packages

type Apt struct {
}

func (a Apt) PkgExec() string {
	return "apt"
}

func (a Apt) PkgNameSeparator() string {
	return "/"
}
