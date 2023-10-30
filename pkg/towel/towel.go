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
	Init  *InitCmd  `arg:"subcommand:init" help:"create a configuration directory (to be filled by hand)"`
	Graph *GraphCmd `arg:"subcommand:graph" help:"generate the dependency graph of a set of tickets"`
	Query *QueryCmd `arg:"subcommand:query" help:"issue a JQL query and dump its contents"`
	Dot   *DotCmd   `arg:"subcommand:dot" help:"generate a graphviz DOT file (WIP)"`
}

type Global struct {
	ConfigDir string        `help:"configuration directory"`
	Timeout   time.Duration `help:"timeout for network operations (eg: 5m7s)"`
	Version   bool          `help:"display version and exit"`
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
	JQL string `arg:"required" help:"JQL query, for example: 'project = \"MY PROJECT\"''. An empty string is not accepted because it would query ALL the projects in the Jira instance"`
}

type InitCmd struct {
}

func Main() int {
	if err := run(os.Args[1:]); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func run(cmdLine []string) error {
	defaultConfigDir, err := defaultConfigDir()
	if err != nil {
		return fmt.Errorf("user configuration directory: %w", err)
	}

	args := Args{
		Global: Global{
			ConfigDir: defaultConfigDir,
			Timeout:   5 * time.Second,
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
	case args.Init != nil:
		return cmdInit(args.Global, *args.Init)
	case args.Graph != nil:
		return cmdGraph(args.Global, *args.Graph)
	case args.Query != nil:
		return cmdQuery(args.Global, *args.Query)
	case args.Dot != nil:
		return cmdDot(args.Global, *args.Dot)
	default:
		return fmt.Errorf("internal error: unwired subcommand: %s", argParser.Subcommand())
	}
}
