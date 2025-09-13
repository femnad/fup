package entity

import (
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
)

type OSRepo interface {
	unless.Unlessable
	when.Whenable
	Install() error
}
