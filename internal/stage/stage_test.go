package stage

import (
	"errors"
	"testing"
)

func TestNormalizeReturnsCanonicalStage(t *testing.T) {
	tests := map[string]string{
		"Tiny MVP":                                  TinyMVP,
		"current stage -> usable mvp":               UsableMVP,
		"ship toward sustained service quality":     SustainedServiceQuality,
		"production quality handoff":                ServiceQuality,
		"unknown launch stage":                      "unknown launch stage",
		"Beta roadmap before Service Quality proof": Beta,
	}
	for input, expected := range tests {
		if got := Normalize(input); got != expected {
			t.Fatalf("Normalize(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestKnownAndRankUseCanonicalStages(t *testing.T) {
	if !Known("service-quality") {
		t.Fatal("service-quality should be a known stage")
	}
	if got := Rank("sustained-service-quality"); got != 5 {
		t.Fatalf("Rank(sustained-service-quality) = %d, want 5", got)
	}
	if Known("enterprise launch") {
		t.Fatal("enterprise launch should not be a known stage")
	}
}

func TestParseTargetAliases(t *testing.T) {
	tests := map[string]string{
		"tiny":                      TinyMVP,
		"usable-mvp":                UsableMVP,
		"beta":                      Beta,
		"production_quality":        ServiceQuality,
		"sustained-service-quality": SustainedServiceQuality,
	}
	for input, expected := range tests {
		got, err := ParseTarget(input)
		if err != nil {
			t.Fatalf("ParseTarget(%q) returned error: %v", input, err)
		}
		if got != expected {
			t.Fatalf("ParseTarget(%q) = %q, want %q", input, got, expected)
		}
	}
}

func TestParseTargetErrors(t *testing.T) {
	if _, err := ParseTarget(""); !errors.Is(err, ErrMissingTarget) {
		t.Fatalf("empty target error = %v, want ErrMissingTarget", err)
	}
	if _, err := ParseTarget("enterprise launch"); !errors.Is(err, ErrUnknownTarget) {
		t.Fatalf("unknown target error = %v, want ErrUnknownTarget", err)
	}
}
