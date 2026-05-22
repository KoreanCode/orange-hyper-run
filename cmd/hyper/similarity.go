package main

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	similarContextDisplayLimit = 3
	similarContextTextLimit    = 220
)

func findSimilarContext(db *sql.DB, query string, limit int) ([]similarContext, *hyperError) {
	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return nil, nil
	}
	var candidates []similarContext
	runRows, err := db.Query(`select id, objective, stage, coalesce(summary, '') from runs order by started_at desc limit 100`)
	if err != nil {
		return nil, dbError(err)
	}
	for runRows.Next() {
		var id, objective, stage, summary string
		if err := runRows.Scan(&id, &objective, &stage, &summary); err != nil {
			runRows.Close()
			return nil, dbError(err)
		}
		candidates = append(candidates, similarContext{Source: "run", ID: id, Kind: stage, Text: joinNonEmpty([]string{objective, stage, summary}, " - ")})
	}
	runRows.Close()

	goalRows, err := db.Query(`select id, objective, coalesce(scope, ''), coalesce(non_goals, ''), coalesce(validation, ''), coalesce(stop_condition, '') from goals order by created_at desc limit 100`)
	if err != nil {
		return nil, dbError(err)
	}
	for goalRows.Next() {
		var id, objective, scope, nonGoals, validation, stopCondition string
		if err := goalRows.Scan(&id, &objective, &scope, &nonGoals, &validation, &stopCondition); err != nil {
			goalRows.Close()
			return nil, dbError(err)
		}
		if nonGoals != "" {
			nonGoals = "Non-goals: " + nonGoals
		}
		candidates = append(candidates, similarContext{Source: "goal", ID: id, Kind: "goal", Text: joinNonEmpty([]string{objective, scope, nonGoals, validation, stopCondition}, " - ")})
	}
	goalRows.Close()

	memRows, err := db.Query(`select id, kind, text, coalesce(quality, '') from memories where stale_at is null order by created_at desc limit 200`)
	if err != nil {
		return nil, dbError(err)
	}
	for memRows.Next() {
		var id int64
		var kind, text, quality string
		if err := memRows.Scan(&id, &kind, &text, &quality); err != nil {
			memRows.Close()
			return nil, dbError(err)
		}
		if memoryQualityIsIgnored(quality) {
			continue
		}
		candidates = append(candidates, similarContext{Source: "memory", ID: strconv.FormatInt(id, 10), Kind: firstNonBlank(quality, kind), Text: text})
	}
	memRows.Close()

	for i := range candidates {
		candidates[i].Score = scoreText(candidates[i].Text, query, queryTokens)
	}
	filtered := candidates[:0]
	for _, candidate := range candidates {
		if candidate.Score > 0 {
			filtered = append(filtered, candidate)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Score == filtered[j].Score {
			if filtered[i].Source == filtered[j].Source {
				return filtered[i].ID < filtered[j].ID
			}
			return filtered[i].Source < filtered[j].Source
		}
		return filtered[i].Score > filtered[j].Score
	})
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

func formatSimilarContext(items []similarContext) string {
	if len(items) == 0 {
		return "None yet."
	}
	if len(items) > similarContextDisplayLimit {
		items = items[:similarContextDisplayLimit]
	}
	lines := make([]string, 0, len(items))
	for _, item := range items {
		kind := ""
		if item.Kind != "" {
			kind = " (" + item.Kind + ")"
		}
		lines = append(lines, fmt.Sprintf("- %s:%s%s score %.2f - %s", item.Source, item.ID, kind, item.Score, compactSimilarText(item.Text)))
	}
	return strings.Join(lines, "\n")
}

func compactSimilarText(text string) string {
	return compactText(text, similarContextTextLimit)
}

func scoreText(text, rawQuery string, queryTokens map[string]struct{}) float64 {
	textTokens := tokenize(text)
	if len(textTokens) == 0 {
		return 0
	}
	overlap := 0
	for token := range queryTokens {
		if _, ok := textTokens[token]; ok {
			overlap++
		}
	}
	coverage := float64(overlap) / float64(len(queryTokens))
	density := float64(overlap) / float64(len(textTokens))
	bonus := 0.0
	if strings.Contains(strings.ToLower(oneLine(text)), strings.ToLower(oneLine(rawQuery))) {
		bonus = 0.25
	}
	return coverage*0.7 + density*0.3 + bonus
}

func tokenize(value string) map[string]struct{} {
	stops := map[string]struct{}{"and": {}, "are": {}, "for": {}, "from": {}, "into": {}, "the": {}, "this": {}, "that": {}, "with": {}, "without": {}, "tbd": {}, "goal": {}, "mvp": {}, "run": {}}
	result := map[string]struct{}{}
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !(r == '_' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r >= '가' && r <= '힣')
	})
	for _, field := range fields {
		if len([]rune(field)) > 1 {
			if _, stop := stops[field]; !stop {
				result[field] = struct{}{}
			}
		}
	}
	return result
}
