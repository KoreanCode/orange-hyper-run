package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func explicitStatus(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if value, ok := strings.CutPrefix(strings.TrimSpace(line), "Status:"); ok {
			status := strings.ToLower(strings.TrimSpace(value))
			for _, allowed := range []string{"active", "blocked", "waiting_user", "completed", "stale"} {
				if status == allowed {
					return status
				}
			}
		}
	}
	return ""
}

func hasNonPendingSection(text, heading string) bool {
	return len(usefulSectionLines(text, heading)) > 0
}

func sectionBody(text, heading string) string {
	lines := strings.Split(text, "\n")
	start := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## "+heading {
			start = i + 1
			break
		}
	}
	if start == -1 {
		return ""
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}

func firstSectionLine(text, heading string) string {
	for _, line := range strings.Split(sectionBody(text, heading), "\n") {
		trimmed := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
		if trimmed != "" && !isPlaceholder(trimmed) {
			return trimmed
		}
	}
	return ""
}

func firstLabelValue(text, label string) string {
	prefix := label + ":"
	for _, line := range strings.Split(text, "\n") {
		if value, ok := strings.CutPrefix(strings.TrimSpace(line), prefix); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstUsefulLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
		if trimmed != "" && !isPlaceholder(trimmed) {
			return trimmed
		}
	}
	return ""
}

func normalizeLabel(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeSentence(value string) string {
	normalized := strings.ToLower(oneLine(value))
	normalized = strings.Trim(normalized, " .,:;!?\t")
	return normalized
}

func hasAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func oneLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

func compactText(value string, limit int) string {
	value = oneLine(value)
	if limit <= 0 || len([]rune(value)) <= limit {
		return value
	}
	runes := []rune(value)
	return strings.TrimSpace(string(runes[:limit])) + "..."
}

func compactMultiline(value string, maxLines, lineLimit int) string {
	lines := []string{}
	for _, line := range strings.Split(value, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		lines = append(lines, compactText(trimmed, lineLimit))
		if maxLines > 0 && len(lines) >= maxLines {
			break
		}
	}
	total := 0
	for _, line := range strings.Split(value, "\n") {
		if strings.TrimSpace(line) != "" {
			total++
		}
	}
	if maxLines > 0 && total > maxLines {
		lines = append(lines, "- ...")
	}
	return strings.Join(lines, "\n")
}

func joinNonEmpty(values []string, sep string) string {
	nonEmpty := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			nonEmpty = append(nonEmpty, value)
		}
	}
	return strings.Join(nonEmpty, sep)
}

func hashText(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func nowISO() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

func readIfExists(path string) string {
	body, _ := os.ReadFile(path)
	return string(body)
}

func writeIfMissing(path, body string) *hyperError {
	if exists(path) {
		return nil
	}
	return writeText(path, body)
}

func writeText(path, body string) *hyperError {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ioError(err)
	}
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		return ioError(err)
	}
	return nil
}

func writeJSON(path string, value any) *hyperError {
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return newError(err.Error(), 1)
	}
	return writeText(path, body.String())
}

func appendJSONL(path string, value map[string]any) *hyperError {
	body, err := json.Marshal(value)
	if err != nil {
		return newError(err.Error(), 1)
	}
	return appendText(path, string(body)+"\n")
}

func appendText(path, value string) *hyperError {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return ioError(err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return ioError(err)
	}
	defer file.Close()
	if _, err := file.WriteString(value); err != nil {
		return ioError(err)
	}
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}
