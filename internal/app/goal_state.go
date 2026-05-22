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
		if validation := firstUsefulValidationMemory(sectionBody(evidenceText, "Validation")); validation != "" {
			memories = appendMemoryIfUseful(memories, "pattern", fmt.Sprintf("%s validation pattern: %s", goalID, validation), 0.65)
		}
	case "waiting_user":
		memories = appendMemoryIfUseful(memories, "constraint", fmt.Sprintf("%s waiting for user: %s", goalID, state.Reason), 0.8)
	}
	return dedupeMemories(memories)
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

func blockerSectionLines(text string) []string {
	blockers, _ := blockerSectionSignals(text)
	return blockers
}

func blockerSectionSignals(text string) ([]string, []string) {
	lines := []string{}
	waiting := []string{}
	for _, line := range usefulSectionLines(text, "Blocker") {
		switch blockerLineDisposition(line) {
		case "non_blocking":
			continue
		case "waiting_user":
			waiting = append(waiting, line)
		default:
			lines = append(lines, line)
		}
	}
	return lines, waiting
}

func nonBlockingBlockerLine(line string) bool {
	return blockerLineDisposition(line) == "non_blocking"
}

func blockerLineDisposition(line string) string {
	normalized := normalizeSentence(line)
	if waitingUserBlockerLine(normalized) {
		return "waiting_user"
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
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
		if trimmed != "" && !isPlaceholder(trimmed) && !noisyMemoryText(trimmed) {
			return trimmed
		}
	}
	return ""
}

func noisyMemoryText(text string) bool {
	normalized := strings.ToLower(oneLine(text))
	if normalized == "" {
		return true
	}
	return hasAny(normalized,
		"hyper run created", "`hyper run` created", "created goal-", "created `goal-", "runtime packet created",
		"created runtime packet", "screenshot saved", "screenshot path", "pending.", "no learnable signal",
	) || isNoIssueText(normalized) || isPassiveNoChangeText(normalized)
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
	seen := map[string]bool{}
	deduped := []memory{}
	for _, mem := range memories {
		key := mem.Kind + "\x00" + mem.Text
		if seen[key] {
			continue
		}
		seen[key] = true
		deduped = append(deduped, mem)
	}
	return deduped
}
