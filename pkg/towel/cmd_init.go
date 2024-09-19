package towel

import (
	"fmt"

	"github.com/marco-m/clim"
)

type initCmd struct{}

func newInitCLI() *clim.CLI[App] {
	initCmd := initCmd{}

	cli := clim.New("init",
		"create a configuration directory (to be filled by hand)",
		initCmd.Run)

	return cli
}

func (cmd *initCmd) Run(app App) error {
	if err := initConfig(app.ConfigDir); err != nil {
		return fmt.Errorf("init: %s", err)
	}
	return nil
}
