package towel

import (
	"fmt"
	"net/http"
	"time"

	"github.com/marco-m/clim"
	"github.com/marco-m/jira-towel/internal"
)

type App struct {
	ConfigDir string
	Server    string
	Timeout   time.Duration
	//
	HttpClient *http.Client // Overridable for tests.
}

func MainErr(args []string) error {
	defaultConfigDir, err := defaultConfigDir()
	if err != nil {
		return fmt.Errorf("user configuration directory: %w", err)
	}

	app := App{
		HttpClient: &http.Client{},
	}
	cli := clim.New[App]("jira-towel", "attempt to make life with Jira bearable", nil)

	cli.AddFlag(&clim.Flag{
		Value: clim.String(&app.ConfigDir, defaultConfigDir),
		Long:  "config-dir", Label: "DIR", Help: "Configuration directory",
	})
	cli.AddFlag(&clim.Flag{
		Value: clim.String(&app.Server, "FIXME"),
		Long:  "server", Help: "Jira server URL",
	})
	cli.AddFlag(&clim.Flag{
		Value: clim.Duration(&app.Timeout, 5*time.Second),
		Long:  "timeout", Help: "Timeout for network operations (eg: 5m7s)",
	})

	cli.SetFooter("For more information visit https://github.com/marco-m/jira-towel")

	versionCmd := clim.New("version", "display the version",
		func(app App) error {
			fmt.Println(internal.Version())
			return nil
		})

	cli.AddCLI(newInitCLI())
	cli.AddCLI(newGraphCLI())
	cli.AddCLI(newGanttCLI())
	cli.AddCLI(newQueryCLI())
	cli.AddCLI(newDotCLI())
	cli.AddCLI(versionCmd)

	action, err := cli.Parse(args)
	if err != nil {
		return err
	}

	return action(app)
}
