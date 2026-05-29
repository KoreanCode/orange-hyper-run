package app

import "strings"

func selfReviewRequired(stage string, readiness readinessState) bool {
	stage = firstNonBlank(readiness.Stage, stage)
	normalized := normalizeLabel(stage)
	return strings.Contains(normalized, "service") || strings.Contains(normalized, "production")
}

func selfReviewEvidenceTemplate(stage string, readiness readinessState) string {
	if !selfReviewRequired(stage, readiness) {
		return ""
	}
	return strings.Join([]string{
		"## Self Review",
		"",
		"Plan alignment: Pending. Check whether the result still matches plan.md North Star, target user, and current stage.",
		"Core loop quality: Pending. Check whether the core loop feels coherent, not merely functional.",
		"Product satisfaction: Pending. Check visual polish, copy, flow feel, and target-user fit.",
		"No drift: Pending. Check that the packet did not add broad features or move outside plan.md non-goals.",
		"Validation match: Pending. Check that test, browser, screenshot, docs, or release evidence matches the actual result.",
		"Verdict: Pending. Use `pass` only if this is service-quality enough; use `fail` and list fixes when it is not.",
		"",
	}, "\n")
}

func selfReviewTaskLine(stage string, readiness readinessState) string {
	if !selfReviewRequired(stage, readiness) {
		return ""
	}
	return "- [ ] Fill Self Review with plan alignment, core loop quality, product satisfaction, no drift, validation match, and a pass/fail verdict\n"
}

func selfReviewChecklistLine(stage string, readiness readinessState) string {
	if !selfReviewRequired(stage, readiness) {
		return ""
	}
	return "- Self Review records plan alignment, core loop quality, product satisfaction, no drift, validation match, and `Verdict: pass`."
}

func selfReviewProofLine(stage string, readiness readinessState) string {
	if !selfReviewRequired(stage, readiness) {
		return ""
	}
	return "- Self Review Proof: before completion, judge the actual result against plan.md direction, core loop quality, product satisfaction, no drift, and validation match; fail the packet when it is not service-quality enough."
}

func selfReviewFinishGateFinding(stage string, readiness readinessState, evidenceText string) string {
	if !selfReviewRequired(stage, readiness) {
		return ""
	}
	body := sectionBody(evidenceText, "Self Review")
	if strings.TrimSpace(body) == "" {
		return "Add `## Self Review` with plan alignment, core loop quality, product satisfaction, no drift, validation match, and `Verdict: pass` or `Verdict: fail`."
	}
	verdict := selfReviewLabelValue(body, "Verdict")
	if selfReviewVerdict(verdict) == "fail" {
		return "Self Review verdict is fail; fix the listed quality gaps in this same packet before completing."
	}
	for _, field := range []string{"Plan alignment", "Core loop quality", "Product satisfaction", "No drift", "Validation match"} {
		value := selfReviewLabelValue(body, field)
		if selfReviewValuePending(value) {
			return "Complete Self Review `" + field + "` with a concrete judgment, not a pending placeholder."
		}
		if selfReviewValueNegative(value) {
			return "Self Review `" + field + "` records an unresolved quality gap; keep this packet open and fix it before completing."
		}
	}
	switch selfReviewVerdict(verdict) {
	case "pass":
		return ""
	case "fail":
		return "Self Review verdict is fail; fix the listed quality gaps in this same packet before completing."
	default:
		return "Set Self Review `Verdict: pass` or `Verdict: fail`; do not complete Service Quality packets without an explicit self judgment."
	}
}

func selfReviewLabelValue(body, label string) string {
	prefix := strings.ToLower(label) + ":"
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(strings.TrimLeft(line, "-*0123456789. "))
		if value, ok := strings.CutPrefix(strings.ToLower(trimmed), prefix); ok {
			index := len(trimmed) - len(value)
			return strings.TrimSpace(trimmed[index:])
		}
	}
	return ""
}

func selfReviewValuePending(value string) bool {
	normalized := normalizeSentence(value)
	return normalized == "" ||
		strings.Contains(normalized, "pending") ||
		strings.Contains(normalized, "todo") ||
		strings.Contains(normalized, "tbd") ||
		strings.Contains(normalized, "not checked") ||
		strings.Contains(normalized, "not reviewed")
}

func selfReviewValueNegative(value string) bool {
	normalized := normalizeSentence(value)
	if hasAny(normalized, "no failure", "no failures", "without failure", "no failed check", "no blocking gap") {
		return false
	}
	return hasAny(normalized,
		"failed",
		"failure remains",
		"fails",
		"not acceptable",
		"not accepted",
		"not satisfying",
		"unsatisfied",
		"needs work",
		"need work",
		"unfinished",
		"awkward",
		"blocking gap",
		"below quality",
	)
}

func selfReviewVerdict(value string) string {
	if selfReviewValuePending(value) {
		return ""
	}
	normalized := normalizeSentence(value)
	if normalized == "fail" || strings.HasPrefix(normalized, "fail ") || strings.HasPrefix(normalized, "fail;") || selfReviewValueNegative(normalized) || hasAny(normalized, "not pass", "do not pass") {
		return "fail"
	}
	if hasAny(normalized, "pass", "passed", "acceptable", "accepted", "ready") {
		return "pass"
	}
	return ""
}
