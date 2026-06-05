package app

import (
	"errors"
	"strings"

	runtimeStage "github.com/KoreanCode/orange-hyper-run/internal/stage"
)

const planTargetStageSource = "plan.md Target Stage"

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
			opts.RunTargetSource = "--until"
		case strings.HasPrefix(arg, "--until="):
			target, err := normalizeRunUntilTarget(strings.TrimPrefix(arg, "--until="))
			if err != nil {
				return runOptions{}, err
			}
			opts.RunUntil = target
			opts.AutoContinue = true
			opts.RunTargetSource = "--until"
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
		opts.RunTargetSource = planTargetStageSource
		return opts, nil
	}
	if opts.AutoContinue && previous.AutoContinue && strings.TrimSpace(previous.RunUntil) != "" {
		opts.AutoContinue = true
		opts.RunUntil = previous.RunUntil
		opts.RunTargetSource = previousRunTargetSource(previous)
		return opts, nil
	}
	return opts, nil
}

func missingTargetStageAdvisory(opts runOptions, plan map[string]string) []string {
	if opts.AutoContinue || strings.TrimSpace(opts.RunUntil) != "" {
		return nil
	}
	if _, ok, err := planRunTarget(plan); ok || err != nil {
		return nil
	}
	if !longRunningFocus(opts.Focus) {
		return nil
	}
	return []string{
		"Run target notice: this is a single packet because plan.md has no Target Stage.",
		"To continue packet by packet, add `Target Stage: Service Quality` to plan.md or run `hyper run --auto --until service-quality [focus]`.",
	}
}

func longRunningFocus(focus string) bool {
	normalized := normalizeSentence(focus)
	if normalized == "" {
		return false
	}
	return hasAny(normalized,
		"service quality",
		"service-quality",
		"sustained service",
		"sustained-service",
		"service readiness",
		"production quality",
		"production-ready",
		"until launch",
		"keep going",
		"continue until",
		"finish the service",
		"서비스 수준",
		"서비스화",
		"서비스 품질",
		"완성형",
		"끝까지",
		"지속 개발",
		"계속 개발",
	)
}

func previousRunTargetSource(previous projectState) string {
	source := strings.TrimSpace(previous.RunTargetSource)
	if source != "" && source != planTargetStageSource {
		return source
	}
	return "previous auto target"
}

func validatePlanStageFields(plan map[string]string) *hyperError {
	if err := planCurrentStageError(plan); err != nil {
		return err
	}
	_, _, err := planRunTarget(plan)
	return err
}

func planCurrentStageError(plan map[string]string) *hyperError {
	value := firstRuntimeValue(plan["Current Stage"])
	if value == "" {
		return nil
	}
	stage := normalizeRuntimeStage(value)
	if knownRuntimeStage(stage) {
		return nil
	}
	return newError("Invalid plan.md Current Stage: "+value+"\n\nUse one of: "+runtimeStage.AllowedTargets+".", 2)
}

func applyPlanTargetToState(state projectState, plan map[string]string) projectState {
	target, ok, err := planRunTarget(plan)
	if err != nil || !ok {
		if !ok && state.RunTargetSource == planTargetStageSource {
			state.AutoContinue = false
			state.RunUntil = ""
			state.RunTargetSource = ""
		}
		return state
	}
	if strings.TrimSpace(state.RunUntil) != "" && state.RunTargetSource != planTargetStageSource {
		return state
	}
	state.AutoContinue = true
	state.RunUntil = target
	state.RunTargetSource = planTargetStageSource
	return state
}

func planRunTarget(plan map[string]string) (string, bool, *hyperError) {
	value := firstRuntimeValue(plan["Target Stage"])
	if value == "" {
		return "", false, nil
	}
	target, err := normalizeRunUntilTarget(value)
	if err != nil {
		return "", true, newError("Invalid plan.md Target Stage: "+value+"\n\nUse one of: "+runtimeStage.AllowedTargets+".", 2)
	}
	return target, true, nil
}

func normalizeRunUntilTarget(value string) (string, *hyperError) {
	target, err := runtimeStage.ParseTarget(value)
	switch {
	case err == nil:
		return target, nil
	case errors.Is(err, runtimeStage.ErrMissingTarget):
		return "", newError("Missing value for --until.\n\nUse one of: "+runtimeStage.AllowedTargets+".", 2)
	default:
		return "", newError("Unknown --until stage: "+value+"\n\nUse one of: "+runtimeStage.AllowedTargets+".", 2)
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
