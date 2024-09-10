package towel

import "fmt"

func cmdGraph(global Global, graph GraphCmd) error {
	_, err := loadConfig(global.ConfigDir)
	if err != nil {
		return fmt.Errorf("graph: %w", err)
	}
	return fmt.Errorf("graph: not implemented")
}
