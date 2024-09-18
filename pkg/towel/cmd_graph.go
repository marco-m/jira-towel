package towel

import (
	"fmt"

	"github.com/marco-m/clim"
)

type graphCmd struct{}

func newGraphCLI() *clim.CLI[App] {
	graphCmd := graphCmd{}

	cli := clim.New("graph", "generate the dependency graph of a set of tickets",
		graphCmd.Run)

	return cli
}

func (cmd *graphCmd) Run(app App) error {
	_, err := loadConfig(app.ConfigDir)
	if err != nil {
		return fmt.Errorf("graph: %w", err)
	}
	return fmt.Errorf("graph: not implemented")
}
