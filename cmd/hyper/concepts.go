package main

import (
	"fmt"
	"strings"
)

const (
	growthRuntimeDefinition   = "Evidence-first project growth protocol: execution logs create pressure, pressure creates candidates, and repeated proof promotes project-specific structure."
	growthLoopDefinition      = "Execution -> Evidence -> Pressure Ledger -> Candidate -> Structure when proven."
	runtimeProtocolDefinition = "Agent-agnostic runtime packet protocol for Codex, CLI agents, and other coding assistants."
)

func growthPrinciples() []string {
	return []string{
		"No structure before pressure.",
		"No stage advancement without evidence.",
		"No harness before repeated need.",
		"No memory without reusable signal.",
	}
}

func growthPrinciplesLine() string {
	return strings.Join(growthPrinciples(), " ")
}

func stageGrowthContract(stage string) string {
	normalized := normalizeLabel(stage)
	switch {
	case strings.Contains(normalized, "tiny") && strings.Contains(normalized, "mvp"):
		return "Existence proof: prove one useful flow exists with the smallest reversible product slice."
	case strings.Contains(normalized, "usable") && strings.Contains(normalized, "mvp"):
		return "Usability proof: make the primary flow usable end-to-end for a real user."
	case strings.Contains(normalized, "beta"):
		return "Repeatability proof: prove reliability around realistic data, failures, validation, docs, and release readiness."
	case strings.Contains(normalized, "service") || strings.Contains(normalized, "production"):
		return "Operability proof: treat security, deployment, operations, rollback, and repeatable validation as required product behavior."
	default:
		return "Advance the current stage with evidence that can change the next runtime packet."
	}
}

func pressureLedgerFor(pressures []growthPressure, candidates []growthCandidate) pressureLedger {
	return pressureLedger{
		Method:              growthRuntimeDefinition,
		Protocol:            runtimeProtocolDefinition,
		Principles:          growthPrinciples(),
		OpenPressures:       visibleGrowthPressureCount(pressures),
		CandidateStructures: visibleGrowthCandidateCount(candidates),
		ActiveStructures:    activeStructureCount(candidates),
	}
}

func visibleGrowthPressureCount(pressures []growthPressure) int {
	count := 0
	for _, pressure := range pressures {
		if visibleGrowthPressure(pressure) {
			count++
		}
	}
	return count
}

func visibleGrowthPressure(pressure growthPressure) bool {
	return !isNoisyGrowthSignal(pressure.Signal)
}

func visibleGrowthCandidateCount(candidates []growthCandidate) int {
	count := 0
	for _, candidate := range candidates {
		if visibleGrowthCandidate(candidate) {
			count++
		}
	}
	return count
}

func visibleGrowthCandidate(candidate growthCandidate) bool {
	signal := strings.TrimSpace(candidate.Signal)
	return candidate.Status != "retired" && (signal == "" || !isNoisyGrowthSignal(signal))
}

func activeStructureCount(candidates []growthCandidate) int {
	activeCount := 0
	for _, candidate := range candidates {
		if candidate.Status == "active" && visibleGrowthCandidate(candidate) {
			activeCount++
		}
	}
	return activeCount
}

func growthLoopStateSummary(growth growthState) string {
	pressureCount := visibleGrowthPressureCount(growth.Pressures)
	candidateCount := visibleGrowthCandidateCount(growth.Candidates)
	if pressureCount == 0 {
		return "Pressure Ledger is empty; the next run starts from plan.md and repository state."
	}
	activeCount := activeStructureCount(growth.Candidates)
	if activeCount > 0 {
		return fmt.Sprintf("%d pressure(s), %d candidate(s), %d active structure(s).", pressureCount, candidateCount, activeCount)
	}
	if candidateCount > 0 {
		return fmt.Sprintf("%d pressure(s), %d candidate(s); structure stays candidate until repeated evidence proves it.", pressureCount, candidateCount)
	}
	return fmt.Sprintf("%d pressure(s) observed; waiting for repeated evidence before creating structure.", pressureCount)
}
