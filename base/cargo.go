package base

type CargoPkg struct {
	Crate    string `yaml:"name"`
	Unless   Unless `yaml:"unless"`
	Bins     bool   `yaml:"bins"`
	MultiBin bool   `yaml:"multibin"`
	Version  string `yaml:"version"`
}

func (c CargoPkg) GetUnless() Unless {
	return c.Unless
}

func (c CargoPkg) GetVersion() string {
	return Version
}

func (c CargoPkg) HasPostProc() bool {
	return false
}

func (c CargoPkg) Name() string {
	return Crate
}
