package base

type CargoPkg struct {
	Bins    bool   `yaml:"bins"`
	Crate   string `yaml:"name"`
	Unless  Unless `yaml:"unless"`
	Version string `yaml:"version"`
}

func (c CargoPkg) GetUnless() Unless {
	return c.Unless
}

func (c CargoPkg) GetVersion() string {
	return c.Version
}

func (c CargoPkg) HasPostProc() bool {
	return false
}

func (c CargoPkg) Name() string {
	return c.Crate
}
