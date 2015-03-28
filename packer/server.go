package packer

import (
	"./api"
	"github.com/mitchellh/packer/packer"
	"fmt"
)

type Packer struct {
	api.Meta
}

func (t *Packer) Build(args []string, reply *string) error {
	*reply = "Build:";
	b := api.Build{t.Meta}
	byteArray := []byte(args[0])
	userVars := make(map[string]string)
	tpl, err := packer.ParseTemplate(byteArray, userVars)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to parse template: %s", err))
		*reply = "Error"
		return nil
	}
	b.Run(tpl)
	return nil
}


func (t *Packer) Validate(args []string, reply *string) error {
	*reply = "Validate:";
	return nil
}


func (t *Packer) Version(args []string, reply *string) error {
	*reply = "Validate:";
	return nil
}

func (t *Packer) Inspect(args []string, reply *string) error {
	*reply = "Inspect:";
	return nil
}