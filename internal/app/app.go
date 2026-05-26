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
		if helpRequested(rest) {
			return stdout(commandUsage("init")), nil
		}
		if len(rest) > 0 {
			return commandOutput{}, newError("hyper init does not take an objective.\n\nRun `hyper init`, fill in plan.md, then use `hyper run [focus]`.", 2)
		}
		return initHyper(fsys)
	case "run":
		if helpRequested(rest) {
			return stdout(commandUsage("run")), nil
		}
		opts, err := parseRunOptions(rest)
		if err != nil {
			return commandOutput{}, err
		}
		return runHyper(fsys, opts)
	case "status":
		if helpRequested(rest) {
			return stdout(commandUsage("status")), nil
		}
		return statusHyper(fsys, rest)
	case "doctor":
		if helpRequested(rest) {
			return stdout(commandUsage("doctor")), nil
		}
		return doctorHyper(fsys)
	case "repair":
		if helpRequested(rest) {
			return stdout(commandUsage("repair")), nil
		}
		return repairHyper(fsys)
	case "migrate":
		if helpRequested(rest) {
			return stdout(commandUsage("migrate")), nil
		}
		return migrateHyper(fsys)
	case "resume":
		if helpRequested(rest) {
			return stdout(commandUsage("resume")), nil
		}
		return resumeHyper(fsys)
	case "complete":
		if helpRequested(rest) {
			return stdout(commandUsage("complete")), nil
		}
		return completeHyper(fsys)
	case "advance":
		if helpRequested(rest) {
			return stdout(commandUsage("advance")), nil
		}
		return advanceHyper(fsys)
	case "version":
		if helpRequested(rest) {
			return stdout(commandUsage("version")), nil
		}
		return versionHyper()
	case "update":
		if helpRequested(rest) {
			return stdout(commandUsage("update")), nil
		}
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

func helpRequested(args []string) bool {
	for i, arg := range args {
		switch strings.TrimSpace(arg) {
		case "--help", "-h":
			return true
		case "help":
			return i == 0 && len(args) == 1
		}
	}
	return false
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

func commandUsage(command string) string {
	lines := map[string][]string{
		"init": {
			"Usage:",
			"  hyper init",
			"",
			"Creates `plan.md`, `.hyper/`, and Codex Desktop routing files in the current project.",
			"Use `hyper run [focus]` for the current work objective; do not pass the objective to `hyper init`.",
		},
		"run": {
			"Usage:",
			"  hyper run [--auto] [--until stage] [focus]",
			"",
			"Creates the next runtime packet from `plan.md`, prior evidence, pressure, and readiness.",
			"Options:",
			"  --auto              Continue packet-by-packet through the generated next-packet plan.",
			"  --until <stage>     Plan auto continuation toward tiny-mvp, usable-mvp, beta, service-quality, or sustained-service-quality.",
		},
		"status": {
			"Usage:",
			"  hyper status",
			"  hyper status --short",
			"",
			"Shows current stage, gate, proof, pressure, next action, and blocking gaps.",
		},
		"doctor": {
			"Usage:",
			"  hyper doctor",
			"",
			"Checks install path, version, project state, SQLite, migration freshness, and Codex Desktop routing.",
		},
		"repair": {
			"Usage:",
			"  hyper repair",
			"",
			"Refreshes generated project state when files are missing or stale.",
		},
		"migrate": {
			"Usage:",
			"  hyper migrate",
			"",
			"Refreshes project state, growth rules, readiness, and next-packet planning after a CLI update.",
		},
		"resume": {
			"Usage:",
			"  hyper resume",
			"",
			"Prints the current runtime packet handoff if an active packet exists.",
		},
		"complete": {
			"Usage:",
			"  hyper complete",
			"",
			"Runs the finish gate, learns from `evidence.md` and `next.md`, then refreshes growth and readiness.",
		},
		"advance": {
			"Usage:",
			"  hyper advance",
			"",
			"Updates `plan.md` to the next stage only when the readiness gate is ready and the user accepts the change.",
		},
		"version": {
			"Usage:",
			"  hyper version",
			"",
			"Shows build version, commit, build date, platform, executable path, and update source.",
		},
		"update": {
			"Usage:",
			"  hyper update [source]",
			"",
			"Installs the latest Hyper Run binary from the configured GitHub release or provided source.",
		},
	}
	body := lines[command]
	if len(body) == 0 {
		return usage()
	}
	out := append([]string{"Hyper Run " + command, ""}, body...)
	return strings.Join(out, "\n")
}

type fsRoot interface {
	root() string
}

type workingDir struct{}

func (workingDir) root() string {
	root, _ := os.Getwd()
	return root
}
