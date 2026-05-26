package app

import (
	"fmt"
	"os"
	"strings"
)

const (
	hyperDir          = ".hyper"
	planFile          = "plan.md"
	defaultUpdateRepo = "KoreanCode/orange-hyper-run"
)

// Main runs the Hyper Run CLI and returns the process exit code.
func Main(args []string) int {
	out, err := runCLI(args, workingDir{}, realUpdater{})
	if out.Stdout != "" {
		fmt.Print(out.Stdout)
	}
	if out.Stderr != "" {
		fmt.Fprint(os.Stderr, out.Stderr)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Message)
		return err.Code
	}
	return 0
}

func runCLI(args []string, fsys fsRoot, updater updater) (commandOutput, *hyperError) {
	command := ""
	if len(args) > 0 {
		command = args[0]
	}
	rest := []string{}
	if len(args) > 1 {
		rest = args[1:]
	}

	switch command {
	case "", "help", "--help", "-h":
		return stdout(usage()), nil
	case "init":
		if len(rest) > 0 {
			return commandOutput{}, newError("hyper init does not take an objective.\n\nRun `hyper init`, fill in plan.md, then use `hyper run [focus]`.", 2)
		}
		return initHyper(fsys)
	case "run":
		opts, err := parseRunOptions(rest)
		if err != nil {
			return commandOutput{}, err
		}
		return runHyper(fsys, opts)
	case "status":
		return statusHyper(fsys, rest)
	case "doctor":
		return doctorHyper(fsys)
	case "repair":
		return repairHyper(fsys)
	case "migrate":
		return migrateHyper(fsys)
	case "resume":
		return resumeHyper(fsys)
	case "complete":
		return completeHyper(fsys)
	case "advance":
		return advanceHyper(fsys)
	case "version":
		return versionHyper()
	case "update":
		source := ""
		if len(rest) > 0 {
			source = rest[0]
		}
		return updateHyper(source, updater)
	case "internal":
		return runInternal(rest, fsys)
	default:
		return commandOutput{}, newError(fmt.Sprintf("Unknown command: %s\n\n%s", command, usage()), 2)
	}
}

func runInternal(args []string, fsys fsRoot) (commandOutput, *hyperError) {
	if len(args) > 0 && args[0] == "learn" {
		return learnCurrentGoal(fsys)
	}
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	return commandOutput{}, newError(fmt.Sprintf("Unknown internal command: %s\n\n%s", sub, usage()), 2)
}

func usage() string {
	return strings.Join([]string{
		"Hyper Run",
		"",
		"Usage:",
		"  hyper init",
		"  hyper run [--auto] [--until stage] [focus]",
		"  hyper complete",
		"  hyper advance",
		"  hyper status",
		"  hyper status --short",
		"  hyper doctor",
		"  hyper repair",
		"  hyper resume",
		"  hyper migrate",
		"  hyper version",
		"  hyper update [source]",
		"",
		"Primary flow:",
		"  Run `hyper init` once in a project to install Hyper Run settings.",
		"  Edit plan.md, then use `hyper run [focus]` to create the next runtime packet.",
		"  Use `hyper run --auto --until service-quality [focus]` or `--until sustained-service-quality` when Codex should keep planning packets toward a target stage.",
		"  After updating evidence.md and next.md, use `hyper complete` to turn evidence into pressure, candidates, and readiness.",
		"  `hyper complete` runs the finish gate first; fix review.md findings in the same packet before continuing.",
		"  When `hyper status` says the stage gate is ready, use `hyper advance` to apply the accepted stage change.",
		"",
		"Method:",
		"  " + growthRuntimeDefinition,
		"  " + growthLoopDefinition,
		"  " + runtimeProtocolDefinition,
		"",
		"Principles:",
		"  " + growthPrinciplesLine(),
		"",
		"Codex Desktop convention:",
		"  `$hyper run` means: run the CLI, read the generated runtime packet, and execute it.",
		"",
	}, "\n")
}

type fsRoot interface {
	root() string
}

type workingDir struct{}

func (workingDir) root() string {
	root, _ := os.Getwd()
	return root
}
