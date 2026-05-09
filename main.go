package main

import "github.com/hieutapt/seals-subscription-cli/cmd"

// These are set by GoReleaser via ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	cmd.Execute()
}
