package main

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
	if hasNonPendingSection(evidenceText, "Blocker") {
		return goalState{State: "blocked", Reason: firstNonBlank(firstSectionLine(evidenceText, "Blocker"), "Blocker section is populated.")}
	}
	if hasNonPendingSection(nextText, "Recommended Next Goal") && hasNonPendingSection(evidenceText, "Validation") {
		return goalState{State: "completed", Reason: "Evidence and next recommendation are populated."}
	}
	return goalState{State: "active", Reason: "Runtime packet evidence is still pending."}
}

func memoriesForDerivedState(state goalState, goalID, evidenceText, nextText string) []memory {
	memories := []memory{}
	memories = appendSectionMemories(memories, goalID, evidenceText, "Readiness Evidence", "pattern", 0.7)
	memories = appendSectionMemories(memories, goalID, evidenceText, "Decisions", "decision", 0.75)
	memories = appendSectionMemories(memories, goalID, evidenceText, "Reusable Patterns", "pattern", 0.75)
	memories = appendLearnNoteMemories(memories, goalID, nextText)

	switch state.State {
	case "blocked":
		memories = appendMemoryIfUseful(memories, "failure", fmt.Sprintf("%s blocked: %s", goalID, state.Reason), 0.8)
	case "completed":
		memories = appendMemoryIfUseful(memories, "pattern", fmt.Sprintf("%s validated: %s", goalID, firstNonBlank(firstUsefulLine(sectionBody(evidenceText, "Validation")), state.Reason)), 0.65)
		memories = appendMemoryIfUseful(memories, "decision", fmt.Sprintf("%s next runtime episode: %s", goalID, firstNonBlank(firstUsefulLine(sectionBody(nextText, "Recommended Next Goal")), "See next.md")), 0.6)
	case "waiting_user":
		memories = appendMemoryIfUseful(memories, "constraint", fmt.Sprintf("%s waiting for user: %s", goalID, state.Reason), 0.8)
	}
	return dedupeMemories(memories)
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

func appendMemoryIfUseful(memories []memory, kind, text string, confidence float64) []memory {
	text = oneLine(text)
	if text == "" || isPlaceholder(text) {
		return memories
	}
	return append(memories, memory{Kind: kind, Text: text, Confidence: confidence})
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
