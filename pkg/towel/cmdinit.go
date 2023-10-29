package towel

import (
	"fmt"
)

func cmdInit(global Global, init InitCmd) error {
	if err := initConfig(global.ConfigDir); err != nil {
		return fmt.Errorf("init: %s", err)
	}
	return nil
}
