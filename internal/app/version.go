package app

import (
	"os"
	"runtime"
	"strings"

	"github.com/KoreanCode/orange-hyper-run/internal/buildinfo"
)

func versionHyper() (commandOutput, *hyperError) {
	executable, err := os.Executable()
	if err != nil {
		executable = "unknown"
	}
	return stdout(strings.Join([]string{
		"Hyper Run",
		"Version: " + buildinfo.Version,
		"Commit: " + buildinfo.Commit,
		"Build date: " + buildinfo.BuildDate,
		"Go: " + runtime.Version(),
		"Platform: " + runtime.GOOS + "/" + runtime.GOARCH,
		"Executable: " + executable,
		"Update source: github:" + defaultUpdateRepo,
		"",
	}, "\n")), nil
}
