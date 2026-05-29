package app

import (
	"strings"
)

func parseRunOptions(args []string) (runOptions, *hyperError) {
	opts := runOptions{}
	focus := []string{}
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		switch {
		case arg == "":
			continue
		case arg == "--auto":
			opts.AutoContinue = true
		case arg == "--until":
			i++
			if i >= len(args) {
				return runOptions{}, newError("Missing value for --until.\n\nUsage:\n  hyper run [--auto] [--until stage] [focus]", 2)
			}
			target, err := normalizeRunUntilTarget(args[i])
			if err != nil {
				return runOptions{}, err
			}
			opts.RunUntil = target
			opts.AutoContinue = true
		case strings.HasPrefix(arg, "--until="):
			target, err := normalizeRunUntilTarget(strings.TrimPrefix(arg, "--until="))
			if err != nil {
				return runOptions{}, err
			}
			opts.RunUntil = target
			opts.AutoContinue = true
		case strings.HasPrefix(arg, "--"):
			return runOptions{}, newError("Unknown run option: "+arg+"\n\nUsage:\n  hyper run [--auto] [--until stage] [focus]", 2)
		default:
			focus = append(focus, arg)
		}
	}
	opts.Focus = strings.Join(focus, " ")
	return opts, nil
}

func applyDefaultRunTarget(opts runOptions, plan map[string]string, previous projectState) (runOptions, *hyperError) {
	if strings.TrimSpace(opts.RunUntil) != "" {
		return opts, nil
	}
	if target, ok, err := planRunTarget(plan); ok || err != nil {
		if err != nil {
			return runOptions{}, err
		}
		opts.AutoContinue = true
		opts.RunUntil = target
		opts.RunTargetSource = "plan.md Target Stage"
		return opts, nil
	}
	if opts.AutoContinue && previous.AutoContinue && strings.TrimSpace(previous.RunUntil) != "" {
		opts.AutoContinue = true
		opts.RunUntil = previous.RunUntil
		opts.RunTargetSource = "previous auto target"
		return opts, nil
	}
	return opts, nil
}

func planRunTarget(plan map[string]string) (string, bool, *hyperError) {
	value := firstRuntimeValue(plan["Target Stage"])
	if value == "" {
		return "", false, nil
	}
	target, err := normalizeRunUntilTarget(value)
	if err != nil {
		return "", true, newError("Invalid plan.md Target Stage: "+value+"\n\nUse one of: tiny-mvp, usable-mvp, beta, service-quality, sustained-service-quality.", 2)
	}
	return target, true, nil
}

func normalizeRunUntilTarget(value string) (string, *hyperError) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")
	normalized = strings.Join(strings.Fields(normalized), " ")
	switch normalized {
	case "", "none":
		return "", newError("Missing value for --until.\n\nUse one of: tiny-mvp, usable-mvp, beta, service-quality, sustained-service-quality.", 2)
	case "tiny", "tiny mvp":
		return "Tiny MVP", nil
	case "usable", "usable mvp":
		return "Usable MVP", nil
	case "beta":
		return "Beta", nil
	case "service", "service quality", "production", "production quality":
		return "Service Quality", nil
	case "sustained", "sustained quality", "sustained service", "sustained service quality":
		return "Sustained Service Quality", nil
	default:
		return "", newError("Unknown --until stage: "+value+"\n\nUse one of: tiny-mvp, usable-mvp, beta, service-quality, sustained-service-quality.", 2)
	}
}

func formatRunMode(opts runOptions) string {
	if !opts.AutoContinue {
		return "single packet"
	}
	if opts.RunUntil != "" {
		return "auto until " + opts.RunUntil
	}
	return "auto"
}

func runTargetSourceLine(opts runOptions) string {
	if strings.TrimSpace(opts.RunTargetSource) == "" || strings.TrimSpace(opts.RunUntil) == "" {
		return ""
	}
	return "Run target source: " + opts.RunTargetSource
}
