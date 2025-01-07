package internal

import (
	"fmt"
	"runtime/debug"
)

// These variables must be set by the linker (see Taskfile or .goreleaser.yaml).
var (
	version = "unknown"
	// commit  = "unknown"
)

// Version reports the version of the main package of the binary.
//
// As of Go 1.21, we still need to use two different approaches to be able to
// report version information:
// 1. to support "go build", we must use the linker.
// 2. to support "go install with remote path", we must use the `debug` package.
// 3. as far as I understand it, "go install with local path" does not work.
// See https://github.com/golang/go/issues/50603 to be able to use the `debug`
// package for all use cases.
// It seems that finally Go 1.24 will do the right thing, after 6 years.
// It was discussed at least since Go1.13 (see https://github.com/golang/go/issues/29228)
// See https://github.com/caddyserver/caddy/pull/4931
func Version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Sprintf("%s (stripped)", version)
	}
	mod := &info.Main
	if mod.Replace != nil {
		mod = mod.Replace
	}
	if mod.Version != "" && mod.Version != "(devel)" {
		return fmt.Sprintf("%s (%s)", mod.Version, mod.Path)
	}
	return fmt.Sprintf("%s (%s)", version, mod.Path)
}
