package towel

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexflint/go-arg"

	"github.com/marco-m/jira-towel/internal"
)

type Args struct {
	Global
	//
	Graph *GraphCmd `arg:"subcommand:graph" help:"generate the dependency graph of a set of tickets"`
}

type Global struct {
	Server  string        `arg:"required" help:"Jira server URL"`
	Timeout time.Duration `help:"timeout for network operations (eg: 1h32m7s)"`
	Version bool          `help:"display version and exit"`
	//
	HttpClient *http.Client `arg:"-"` // Overridable for tests.
}

func (Args) Description() string {
	return "This program attempts to make life with Jira bearable"
}

func (Args) Epilogue() string {
	return "For more information visit https://github.com/marco-m/jira-towel"
}

type GraphCmd struct {
	Pipeline string `arg:"required"`
}

func Main() int {
	if err := run(os.Args[1:]); err != nil {
		fmt.Println("error:", err)
		return 1
	}
	return 0
}

func run(cmdLine []string) error {
	args := Args{
		Global: Global{
			Timeout: 5 * time.Second,
		},
	}
	argParser, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		return fmt.Errorf("init cli parsing: %s", err)
	}
	argParser.MustParse(cmdLine)
	if args.Version {
		fmt.Printf("jira-towel version %s\n", internal.Version())
		return nil
	}
	if argParser.Subcommand() == nil {
		argParser.Fail("missing subcommand")
	}

	switch {
	default:
		return fmt.Errorf("internal error: unwired subcommand: %s", argParser.Subcommand())
	}
}
