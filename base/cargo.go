package base

type CargoPkg struct {
	Crate     string `yaml:"name"`
	Unless   Unless `yaml:"unless"`
	Bins     bool   `yaml:"bins"`
	MultiBin bool   `yaml:"multibin"`
}

func (c CargoPkg) GetUnless() Unless {
	return c.Unless
}

func (c CargoPkg) GetVersion() string {
	return ""
}

func (c CargoPkg) HasPostProc() bool {
	return false
}

func (c CargoPkg) Name() string {
    return ""
}
