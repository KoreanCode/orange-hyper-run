package stage

import (
	"errors"
	"strings"
)

const (
	TinyMVP                 = "Tiny MVP"
	UsableMVP               = "Usable MVP"
	Beta                    = "Beta"
	ServiceQuality          = "Service Quality"
	SustainedServiceQuality = "Sustained Service Quality"
)

var (
	ErrMissingTarget = errors.New("missing target stage")
	ErrUnknownTarget = errors.New("unknown target stage")
)

const AllowedTargets = "tiny-mvp, usable-mvp, beta, service-quality, sustained-service-quality"

type patternSet struct {
	name     string
	patterns []string
}

var normalizePatterns = []patternSet{
	{name: TinyMVP, patterns: []string{"tiny mvp"}},
	{name: UsableMVP, patterns: []string{"usable mvp"}},
	{name: Beta, patterns: []string{"beta"}},
	{name: SustainedServiceQuality, patterns: []string{"sustained service quality", "sustained quality"}},
	{name: ServiceQuality, patterns: []string{"service quality", "production"}},
}

// Normalize returns the canonical stage name when value clearly contains a known stage.
func Normalize(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	normalized := strings.ToLower(value)
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")
	normalized = strings.Join(strings.Fields(normalized), " ")
	bestName := ""
	bestIndex := len(normalized) + 1
	for _, candidate := range normalizePatterns {
		for _, pattern := range candidate.patterns {
			index := strings.Index(normalized, pattern)
			if index >= 0 && index < bestIndex {
				bestIndex = index
				bestName = candidate.name
			}
		}
	}
	if bestName != "" {
		return bestName
	}
	return value
}

func Known(value string) bool {
	switch Normalize(value) {
	case TinyMVP, UsableMVP, Beta, ServiceQuality, SustainedServiceQuality:
		return true
	default:
		return false
	}
}

func Rank(value string) int {
	switch Normalize(value) {
	case TinyMVP:
		return 1
	case UsableMVP:
		return 2
	case Beta:
		return 3
	case ServiceQuality:
		return 4
	case SustainedServiceQuality:
		return 5
	default:
		return 0
	}
}

func ParseTarget(value string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.ReplaceAll(normalized, "_", " ")
	normalized = strings.Join(strings.Fields(normalized), " ")
	switch normalized {
	case "", "none":
		return "", ErrMissingTarget
	case "tiny", "tiny mvp":
		return TinyMVP, nil
	case "usable", "usable mvp":
		return UsableMVP, nil
	case "beta":
		return Beta, nil
	case "service", "service quality", "production", "production quality":
		return ServiceQuality, nil
	case "sustained", "sustained quality", "sustained service", "sustained service quality":
		return SustainedServiceQuality, nil
	default:
		return "", ErrUnknownTarget
	}
}
