package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type planImportCandidate struct {
	Path    string
	Reason  string
	Excerpt string
}

func maybeWritePlanImportCandidates(root, planBody string) (string, *hyperError) {
	if !planNeedsImport(planBody) {
		return "", nil
	}
	candidates := findPlanImportCandidates(root)
	if len(candidates) == 0 {
		return "", nil
	}
	relPath := filepath.ToSlash(filepath.Join(hyperDir, "plan-candidates.md"))
	if err := writeText(filepath.Join(root, filepath.FromSlash(relPath)), formatPlanImportCandidates(candidates)); err != nil {
		return "", err
	}
	return relPath, nil
}

func planNeedsImport(body string) bool {
	plan := parsePlan(body)
	return firstRuntimeValue(plan["Product"]) == "" ||
		firstRuntimeValue(plan["MVP"]) == "" ||
		firstRuntimeValue(plan["Success Criteria"]) == ""
}

func findPlanImportCandidates(root string) []planImportCandidate {
	paths := []string{}
	for _, rel := range []string{"README.md", "README_ko.md"} {
		if exists(filepath.Join(root, rel)) {
			paths = append(paths, rel)
		}
	}
	docsRoot := filepath.Join(root, "docs")
	if exists(docsRoot) {
		_ = filepath.WalkDir(docsRoot, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}
			if strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
				if rel, relErr := filepath.Rel(root, path); relErr == nil {
					paths = append(paths, filepath.ToSlash(rel))
				}
			}
			return nil
		})
	}
	sort.Strings(paths)

	candidates := []planImportCandidate{}
	for _, rel := range paths {
		body := readIfExists(filepath.Join(root, filepath.FromSlash(rel)))
		score, reason := planCandidateScore(body)
		if score < 2 {
			continue
		}
		candidates = append(candidates, planImportCandidate{
			Path:    rel,
			Reason:  reason,
			Excerpt: markdownExcerpt(body, 10),
		})
	}
	return candidates
}

func planCandidateScore(body string) (int, string) {
	normalized := strings.ToLower(body)
	score := 0
	reasons := []string{}
	for _, group := range []struct {
		label    string
		keywords []string
	}{
		{label: "product context", keywords: []string{"product", "제품", "서비스", "한 줄 정의"}},
		{label: "MVP context", keywords: []string{"mvp", "minimum viable", "첫 버전", "우선순위"}},
		{label: "target users", keywords: []string{"target users", "users", "사용자", "타겟"}},
		{label: "success criteria", keywords: []string{"success criteria", "완료 기준", "검증", "validation"}},
		{label: "technical context", keywords: []string{"stack", "기술 스택", "api", "database", "frontend", "backend"}},
	} {
		if hasAny(normalized, group.keywords...) {
			score++
			reasons = append(reasons, group.label)
		}
	}
	return score, strings.Join(reasons, ", ")
}

func markdownExcerpt(body string, limit int) string {
	lines := []string{}
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "![") || strings.HasPrefix(trimmed, "<") {
			continue
		}
		lines = append(lines, trimmed)
		if len(lines) >= limit {
			break
		}
	}
	return strings.Join(lines, "\n")
}

func formatPlanImportCandidates(candidates []planImportCandidate) string {
	lines := []string{
		"# Plan Import Candidates",
		"",
		"Hyper Run found existing project documents that may help fill `plan.md`.",
		"Keep `plan.md` human-owned; copy only the parts that define product, MVP, constraints, and success criteria.",
		"",
	}
	for _, candidate := range candidates {
		lines = append(lines,
			"## "+candidate.Path,
			"",
			"Reason: "+candidate.Reason,
			"",
			"```md",
			candidate.Excerpt,
			"```",
			"",
		)
	}
	return strings.Join(lines, "\n")
}
