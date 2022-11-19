package packages

type Apt struct {
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
