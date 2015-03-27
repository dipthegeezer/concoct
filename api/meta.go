package api

import (
	"github.com/mitchellh/packer/packer"
)

type Meta struct {
	EnvConfig *packer.EnvironmentConfig
}

func (m *Meta) Environment() (packer.Environment, error) {
	return packer.NewEnvironment(m.EnvConfig)
}
