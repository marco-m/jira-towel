package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/marco-m/clim"
	"github.com/marco-m/jira-towel/pkg/towel"
)

func main() {
	os.Exit(mainInt())
}

func mainInt() int {
	err := towel.MainErr(os.Args[1:])
	if err == nil {
		return 0
	}
	fmt.Println(err)
	if errors.Is(err, clim.ErrHelp) {
		return 0
	}
	if errors.Is(err, clim.ErrParse) {
		return 2
	}
	return 1
}
