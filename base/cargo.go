package base

import (
	"fmt"

	"github.com/femnad/fup/precheck/unless"
)

type CargoPkg struct {
	Bins    bool          `yaml:"bins"`
	Crate   string        `yaml:"name"`
	Unless  unless.Unless `yaml:"unless"`
	Version string        `yaml:"version"`
	When    string        `yaml:"when"`
}

func (c CargoPkg) DefaultVersionCmd() string {
	return fmt.Sprintf("%s --version", c.Name())
}

func (c CargoPkg) GetUnless() unless.Unless {
	return c.Unless
}

func (c CargoPkg) GetVersion() (string, error) {
	return c.Version, nil
}

func (c CargoPkg) HasPostProc() bool {
	return false
}

func (c CargoPkg) Name() string {
	return c.Crate
}

func (c CargoPkg) RunWhen() string {
	return c.When
}
