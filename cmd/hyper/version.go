package main

import (
	"os"
	"runtime"
	"strings"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func versionHyper() (commandOutput, *hyperError) {
	executable, err := os.Executable()
	if err != nil {
		executable = "unknown"
	}
	return stdout(strings.Join([]string{
		"Hyper Run",
		"Version: " + version,
		"Commit: " + commit,
		"Build date: " + buildDate,
		"Go: " + runtime.Version(),
		"Platform: " + runtime.GOOS + "/" + runtime.GOARCH,
		"Executable: " + executable,
		"Update source: github:" + defaultUpdateRepo,
		"",
	}, "\n")), nil
}
