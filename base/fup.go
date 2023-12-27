package base

import (
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
)

func ReadConfig(filename string) (entity.Config, error) {
	filename = internal.ExpandUser(filename)
	return entity.UnmarshalConfig(filename)
}
