package app

import (
	"fmt"
	"path/filepath"
	"strings"
)

func deriveCurrentGoalState(root, goalID string) goalState {
	if strings.TrimSpace(goalID) == "" {
		return goalState{State: "initialized", Reason: "No current runtime packet recorded."}
	}
	goalDir := filepath.Join(root, hyperDir, "goals", goalID)
	return deriveGoalState(readIfExists(filepath.Join(goalDir, "evidence.md")), readIfExists(filepath.Join(goalDir, "next.md")))
}

func deriveGoalState(evidenceText, nextText string) goalState {
	if status := firstNonBlank(explicitStatus(evidenceText), explicitStatus(nextText)); status != "" {
		reason := firstNonBlank(firstLabelValue(evidenceText, "Reason"), firstLabelValue(nextText, "Reason"), "Explicit status marker: "+status)
		return goalState{State: status, Reason: reason}
	}
	blockers, waiting := blockerSectionSignals(evidenceText)
	if len(blockers) > 0 {
		return goalState{State: "blocked", Reason: firstNonBlank(blockers[0], "Blocker section is populated.")}
	}
	if hasNonPendingSection(nextText, "Recommended Next Goal") && hasNonPendingSection(evidenceText, "Validation") {
		if surfaceProofFollowupRequiredFromEvidence(evidenceText) {
			return goalState{State: "completed", Reason: "Evidence and next recommendation are populated; surface proof follow-up is needed."}
		}
		return goalState{State: "completed", Reason: "Evidence and next recommendation are populated."}
	}
	if len(waiting) > 0 {
		return goalState{State: "waiting_user", Reason: firstNonBlank(waiting[0], "Waiting for user input.")}
	}
	return goalState{State: "active", Reason: "Runtime packet evidence is still pending."}
}

func memoriesForDerivedState(state goalState, goalID, evidenceText, nextText string) []memory {
	memories := []memory{}
	memories = appendSectionMemories(memories, goalID, evidenceText, "Readiness Evidence", "pattern", 0.7)
	memories = appendSurfaceProofMemories(memories, goalID, evidenceText)
	memories = appendSectionMemories(memories, goalID, evidenceText, "Pressure Signals", "pattern", 0.7)
	memories = appendSectionMemories(memories, goalID, evidenceText, "Decisions", "decision", 0.75)
	memories = appendSectionMemories(memories, goalID, evidenceText, "Reusable Patterns", "pattern", 0.75)
	memories = appendLearnNoteMemories(memories, goalID, nextText)

	switch state.State {
	case "blocked":
		memories = appendMemoryIfUseful(memories, "failure", fmt.Sprintf("%s blocked: %s", goalID, state.Reason), 0.8)
	case "completed":
		for _, validation := range usefulValidationMemories(sectionBody(evidenceText, "Validation")) {
			memories = appendMemoryIfUseful(memories, "pattern", fmt.Sprintf("%s validation pattern: %s", goalID, validation), 0.65)
		}
	case "waiting_user":
		memories = appendMemoryIfUseful(memories, "constraint", fmt.Sprintf("%s waiting for user: %s", goalID, state.Reason), 0.8)
	}
	return dedupeMemories(memories)
}

func rejectedMemoryQualityCounts(goalID, evidenceText, nextText string) map[string]int {
	rejected := map[string]int{}
	for _, line := range usefulSectionLines(evidenceText, "Readiness Evidence") {
		countRejectedMemoryQuality(rejected, "pattern", fmt.Sprintf("%s readiness evidence: %s", goalID, line), 0.7)
	}
	for _, line := range usefulSectionLines(evidenceText, "Pressure Signals") {
		countRejectedMemoryQuality(rejected, "pattern", fmt.Sprintf("%s pressure signals: %s", goalID, line), 0.7)
	}
	for _, line := range usefulSectionLines(evidenceText, "Decisions") {
		countRejectedMemoryQuality(rejected, "decision", fmt.Sprintf("%s decisions: %s", goalID, line), 0.75)
	}
	for _, line := range usefulSectionLines(evidenceText, "Reusable Patterns") {
		countRejectedMemoryQuality(rejected, "pattern", fmt.Sprintf("%s reusable patterns: %s", goalID, line), 0.75)
	}
	for _, line := range usefulSectionLines(nextText, "Learn Notes") {
		if learnNoteInstructionLine(line) {
			continue
		}
		kind, value := parseLearnNote(line)
		if kind == "" {
			countRejectedMemoryQuality(rejected, "invalid", line, 0)
			continue
		}
		countRejectedMemoryQuality(rejected, kind, fmt.Sprintf("%s learn %s: %s", goalID, kind, value), 0.7)
	}
	for _, validation := range usefulValidationMemories(sectionBody(evidenceText, "Validation")) {
		countRejectedMemoryQuality(rejected, "pattern", fmt.Sprintf("%s validation pattern: %s", goalID, validation), 0.65)
	}
	return rejected
}

func countRejectedMemoryQuality(counts map[string]int, kind, text string, confidence float64) {
	text = oneLine(text)
	if text == "" || isPlaceholder(text) || noisyMemoryText(text) {
		counts["noisy"]++
		return
	}
	quality := memoryQuality(kind, text, confidence)
	if memoryQualityIsIgnored(quality) {
		counts[quality]++
	}
}

func appendSurfaceProofMemories(memories []memory, goalID, text string) []memory {
	for _, line := range usefulSectionLines(text, "Surface Proof Evidence") {
		signal := surfaceProofValue(line)
		if !usefulSurfaceProofMemory(signal) {
			continue
		}
		kind := "pattern"
		confidence := 0.72
		if surfaceProofGapSignal(signal) {
			kind = "failure"
			confidence = 0.8
		}
		memories = appendMemoryIfUseful(memories, kind, fmt.Sprintf("%s surface proof evidence: %s", goalID, signal), confidence)
	}
	return memories
}

func usefulSurfaceProofMemory(text string) bool {
	normalized := normalizeSentence(text)
	if normalized == "" || isPlaceholder(normalized) || noisyMemoryText(text) {
		return false
	}
	if looksLikeSurfaceProof(text) {
		return true
	}
	return hasAny(normalized,
		"overflow", "overlap", "text clipping", "clipped", "responsive", "breakpoint", "mobile", "desktop",
		"missing state", "empty state", "loading state", "error state", "accessibility", "focus", "keyboard",
		"figma", "design token", "component contract", "console error", "network error",
	)
}

func surfaceProofGapSignal(text string) bool {
	normalized := normalizeSentence(text)
	return hasAny(normalized,
		"gap", "risk", "failed", "failure", "missing", "blocked", "overflow", "overlap", "clipped",
		"not checked", "not verified", "could not", "cannot", "console error", "network error",
	)
}

func appendSectionMemories(memories []memory, goalID, text, heading, kind string, confidence float64) []memory {
	for _, line := range usefulSectionLines(text, heading) {
		memories = appendMemoryIfUseful(memories, kind, fmt.Sprintf("%s %s: %s", goalID, strings.ToLower(heading), line), confidence)
	}
	return memories
}

func appendLearnNoteMemories(memories []memory, goalID, nextText string) []memory {
	for _, line := range usefulSectionLines(nextText, "Learn Notes") {
		kind, value := parseLearnNote(line)
		if kind == "" {
			continue
		}
		memories = appendMemoryIfUseful(memories, kind, fmt.Sprintf("%s learn %s: %s", goalID, kind, value), 0.7)
	}
	return memories
}

func parseLearnNote(line string) (string, string) {
	trimmed := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
	if learnNoteInstructionLine(trimmed) {
		return "", ""
	}
	key, value, ok := strings.Cut(trimmed, ":")
	if !ok {
		return "", ""
	}
	kind := strings.ToLower(strings.TrimSpace(key))
	switch kind {
	case "decision", "pattern", "failure", "constraint":
	default:
		return "", ""
	}
	value = strings.TrimSpace(value)
	if isPlaceholder(value) {
		return "", ""
	}
	return kind, value
}

func learnNoteInstructionLine(line string) bool {
	normalized := normalizeSentence(line)
	return hasAny(normalized, "write only durable signals", "leave a line as pending")
}

func usefulSectionLines(text, heading string) []string {
	lines := []string{}
	for _, line := range strings.Split(sectionBody(text, heading), "\n") {
		trimmed := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
		if trimmed == "" || isPlaceholder(trimmed) {
			continue
		}
		lines = append(lines, trimmed)
	}
	return lines
}

func blockerSectionSignals(text string) ([]string, []string) {
	lines := []string{}
	waiting := []string{}
	for _, line := range usefulSectionLines(text, "Blocker") {
		switch blockerLineDisposition(line) {
		case "non_blocking":
			continue
		case "surface_proof_followup":
			continue
		case "waiting_user":
			waiting = append(waiting, line)
		default:
			lines = append(lines, line)
		}
	}
	return lines, waiting
}

func blockerLineDisposition(line string) string {
	normalized := normalizeSentence(line)
	if waitingUserBlockerLine(normalized) {
		return "waiting_user"
	}
	if surfaceProofFollowupBlockerLine(normalized) {
		return "surface_proof_followup"
	}
	if normalized == "" || isNoIssueText(normalized) || isPassiveNoChangeText(normalized) {
		return "non_blocking"
	}
	if strings.HasPrefix(normalized, "non-blocking note") {
		return "non_blocking"
	}
	if strings.HasPrefix(normalized, "operational note") || strings.HasPrefix(normalized, "note") {
		recovered := hasAny(normalized,
			"succeeded",
			"worked around",
			"workaround succeeded",
			"fallback worked",
			"resolved",
			"recovered",
			"copied into",
		)
		stillBlocked := hasAny(normalized,
			"still blocking",
			"still blocked",
			"cannot proceed",
			"can't proceed",
			"unable to proceed",
			"no workaround",
		)
		if recovered && !stillBlocked {
			return "non_blocking"
		}
	}
	return "blocking"
}

func surfaceProofFollowupRequiredFromEvidence(text string) bool {
	for _, line := range usefulSectionLines(text, "Blocker") {
		if blockerLineDisposition(line) == "surface_proof_followup" {
			return true
		}
	}
	return false
}

func surfaceProofFollowupBlockerLine(normalized string) bool {
	if normalized == "" {
		return false
	}
	hasSurfaceBlock := hasAny(normalized,
		"surface proof",
		"browser proof",
		"browser surface",
		"screenshot proof",
		"browser url policy",
		"browser use url policy",
		"localhost browser access",
	)
	hasBlocked := hasAny(normalized,
		"blocked",
		"could not",
		"cannot",
		"unable",
		"policy",
	)
	hasImplementationClear := hasAny(normalized,
		"no implementation blocker",
		"command validation passed",
		"active command validation passed",
		"validation passed",
		"recorded as a surface-proof gap",
		"surface-proof gap",
		"surface proof gap",
	)
	return hasSurfaceBlock && hasBlocked && hasImplementationClear
}

func waitingUserBlockerLine(normalized string) bool {
	if normalized == "" {
		return false
	}
	if hasAny(normalized,
		"waiting for user",
		"awaiting user",
		"pending user",
		"user decision",
		"user approval",
		"user confirmation",
		"user accepts",
		"user explicitly accepts",
		"until the user accepts",
		"until user accepts",
		"until the user confirms",
		"until user confirms",
	) {
		return true
	}
	return hasAny(normalized, "stage advancement", "stage change", "plan.md current stage") &&
		hasAny(normalized, "accept", "approval", "confirm", "decide")
}

func appendMemoryIfUseful(memories []memory, kind, text string, confidence float64) []memory {
	text = oneLine(text)
	if text == "" || isPlaceholder(text) || noisyMemoryText(text) {
		return memories
	}
	quality := memoryQuality(kind, text, confidence)
	if memoryQualityIsIgnored(quality) {
		return memories
	}
	return append(memories, memory{Kind: kind, Text: text, Confidence: confidence, Quality: quality})
}

func memoryQuality(kind, text string, confidence float64) string {
	normalized := normalizeSentence(text)
	if normalized == "" || noisyMemoryText(text) || isPassiveNoChangeText(normalized) {
		return "passive"
	}
	if hasAny(normalized, "changed files", "notes:", "screenshot path", "screenshot saved") {
		return "one_off"
	}
	if weakLearnSignal(kind, text, confidence) {
		return "weak"
	}
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "decision", "constraint", "failure":
		return "durable"
	case "pattern":
		if confidence >= 0.75 || hasAny(normalized, "reusable pattern", "learn pattern", "before every", "before each", "repeatable") {
			return "durable"
		}
		return "weak"
	default:
		if confidence >= 0.8 {
			return "durable"
		}
		return "weak"
	}
}

func memoryQualityIsIgnored(quality string) bool {
	quality = strings.ToLower(strings.TrimSpace(quality))
	return quality == "passive" || quality == "one_off"
}

func firstUsefulValidationMemory(text string) string {
	memories := usefulValidationMemories(text)
	if len(memories) == 0 {
		return ""
	}
	return memories[0]
}

func usefulValidationMemories(text string) []string {
	command := ""
	commandEmitted := false
	seen := map[string]bool{}
	memories := []string{}
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
		if validationCommandBoundary(trimmed) {
			if nextCommand := firstBacktickCommand(trimmed); nextCommand != "" {
				command = nextCommand
				commandEmitted = false
			}
		}
		if usefulValidationSignal(trimmed) {
			memory := trimmed
			if firstBacktickCommand(trimmed) == "" && command != "" {
				if commandEmitted {
					continue
				}
				memory = commandValidationMemory(command, trimmed)
				commandEmitted = true
			} else if firstBacktickCommand(trimmed) != "" {
				command = firstBacktickCommand(trimmed)
				commandEmitted = true
			}
			if !seen[memory] {
				seen[memory] = true
				memories = append(memories, memory)
			}
		} else if command == "" {
			command = firstBacktickCommand(trimmed)
		}
	}
	return memories
}

func commandValidationMemory(command, outcome string) string {
	command = strings.TrimSpace(command)
	if command == "" {
		return strings.TrimSpace(outcome)
	}
	normalized := normalizeSentence(outcome)
	switch {
	case hasAny(normalized, "failed", "failure", "error"):
		return "`" + command + "` failed."
	case hasAny(normalized, "blocked"):
		return "`" + command + "` blocked."
	case hasAny(normalized, "warning", "warn"):
		return "`" + command + "` completed with warning."
	default:
		return "`" + command + "` passed."
	}
}

func weakLearnSignal(kind, text string, confidence float64) bool {
	normalized := normalizeSentence(text)
	if normalized == "" {
		return true
	}
	if strings.ToLower(strings.TrimSpace(kind)) == "pattern" && strings.Contains(normalized, "validation pattern:") {
		return !usefulValidationSignal(memorySignal(text))
	}
	tokens := pressureTokens(memorySignal(text))
	if len(tokens) < 3 && confidence < 0.8 {
		return true
	}
	return false
}

func usefulValidationSignal(text string) bool {
	trimmed := strings.TrimSpace(strings.TrimLeft(text, "-*0123456789. "))
	if trimmed == "" || isPlaceholder(trimmed) || noisyMemoryText(trimmed) {
		return false
	}
	normalized := strings.ToLower(trimmed)
	if firstBacktickCommand(trimmed) != "" {
		return hasAny(normalized, "pass", "passed", "success", "succeeded", "fail", "failed", "error", "warning", "warn")
	}
	hasTool := hasAny(normalized,
		"go test", "npm run", "pnpm", "yarn", "pytest", "cargo test", "go vet", "staticcheck",
		"govulncheck", "playwright", "vitest", "jest", "build", "test", "lint", "smoke",
		"browser", "screenshot", "deploy", "url",
	)
	hasOutcome := hasAny(normalized,
		"passed", "pass", "succeeded", "success", "verified", "checked", "captured", "built",
		"created", "failed", "blocked", "warning", "warn",
	)
	return hasTool && hasOutcome
}

func noisyMemoryText(text string) bool {
	normalized := strings.ToLower(oneLine(text))
	if normalized == "" {
		return true
	}
	if isHyperProtocolNoiseText(normalized) {
		return true
	}
	return hasAny(normalized,
		"hyper run created", "`hyper run` created", "created goal-", "created `goal-", "runtime packet created",
		"created runtime packet", "screenshot saved", "screenshot path", "pending.", "no learnable signal",
		"only documentation changed", "no code changed", "status only", "summary only",
	) || isNoIssueText(normalized) || isPassiveNoChangeText(normalized)
}

func isHyperProtocolNoiseText(normalized string) bool {
	return hasAny(normalized,
		"stage advancement remains a recommendation pending user acceptance",
		"stage advancement is a recommendation pending user acceptance",
		"stage advancement recommendation pending user acceptance",
		"do not edit `plan.md current stage` until the user accepts stage advancement",
		"do not edit plan.md current stage until the user accepts stage advancement",
		"do not run `hyper advance` unless the user accepts the stage advancement",
		"recommend updating plan.md current stage",
		"review readiness evidence, then run `hyper advance`",
		"stage advancement is acceptable",
		"stage advancement is allowed",
		"allow stage advancement",
		"service quality advancement is acceptable",
		"service quality advancement is allowed",
		"allow service quality advancement",
	)
}

func isPassiveNoChangeText(normalized string) bool {
	return hasAny(normalized,
		"not changed in this episode",
		"was not changed",
		"were not changed",
		"remains unchanged",
		"remain unchanged",
		"unchanged in ",
		"configuration was not changed",
		"configuration were not changed",
		"no auth, secrets",
		"no auth, secret",
		"no secrets, privileged flows",
		"no privileged flows",
		"no third-party write surfaces were added",
	)
}

func dedupeMemories(memories []memory) []memory {
	deduped := []memory{}
	for _, mem := range memories {
		duplicateIndex := -1
		for i, existing := range deduped {
			if memoriesOverlap(existing, mem) {
				duplicateIndex = i
				break
			}
		}
		if duplicateIndex >= 0 {
			if memoryPreferred(mem, deduped[duplicateIndex]) {
				deduped[duplicateIndex] = mem
			}
			continue
		}
		deduped = append(deduped, mem)
	}
	return deduped
}

func memoriesOverlap(left, right memory) bool {
	if !strings.EqualFold(strings.TrimSpace(left.Kind), strings.TrimSpace(right.Kind)) {
		return false
	}
	leftSignal := memorySignal(left.Text)
	rightSignal := memorySignal(right.Text)
	if leftSignal == "" || rightSignal == "" {
		return normalizeSentence(left.Text) == normalizeSentence(right.Text)
	}
	leftTokens := tokenSet(pressureTokens(leftSignal))
	rightTokens := tokenSet(pressureTokens(rightSignal))
	if len(leftTokens) == 0 || len(rightTokens) == 0 {
		return normalizeSentence(leftSignal) == normalizeSentence(rightSignal)
	}
	if tokenJaccard(leftTokens, rightTokens) >= 0.82 {
		return true
	}
	intersection := 0
	for token := range leftTokens {
		if rightTokens[token] {
			intersection++
		}
	}
	smaller := len(leftTokens)
	if len(rightTokens) < smaller {
		smaller = len(rightTokens)
	}
	return smaller > 0 && float64(intersection)/float64(smaller) >= 0.86
}

func memoryPreferred(candidate, existing memory) bool {
	candidateRank := memoryQualityRank(candidate.Quality)
	existingRank := memoryQualityRank(existing.Quality)
	if candidateRank != existingRank {
		return candidateRank > existingRank
	}
	if candidate.Confidence > existing.Confidence+0.01 {
		return true
	}
	if existing.Confidence > candidate.Confidence+0.01 {
		return false
	}
	return len(memorySignal(candidate.Text)) < len(memorySignal(existing.Text))
}

func memoryQualityRank(quality string) int {
	switch strings.ToLower(strings.TrimSpace(quality)) {
	case "durable":
		return 3
	case "weak":
		return 2
	case "one_off":
		return 1
	case "passive":
		return 0
	default:
		return 1
	}
}
