package reasoning

import (
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// Adapter functions to work with existing data structures

// adaptTestStrategy converts reasoning chain test strategy to existing format
func adaptTestStrategy(approach string, levels []string, frameworks []string, coverageTarget float64) *types.TestStrategy {
	strategy := &types.TestStrategy{
		CoverageTarget:   coverageTarget,
		TestFrameworks:   frameworks,
		UnitTests:        containsLevel(levels, "unit"),
		IntegrationTests: containsLevel(levels, "integration"),
		EndToEndTests:    containsLevel(levels, "e2e"),
	}
	return strategy
}

// adaptTimeline converts reasoning chain timeline to existing format
func adaptTimeline(totalEstimate time.Duration) *types.Timeline {
	return &types.Timeline{
		TotalDuration: totalEstimate,
		StartTime:     time.Now(),
		EstimatedEnd:  time.Now().Add(totalEstimate),
		Milestones:    []types.Milestone{},
	}
}

// adaptValidationRules converts reasoning chain validation rules to existing format
func adaptValidationRules(rules []map[string]string) []types.ValidationRule {
	var adapted []types.ValidationRule
	for _, rule := range rules {
		adapted = append(adapted, types.ValidationRule{
			Name:        rule["type"],
			Type:        rule["type"],
			Description: rule["description"],
			Required:    true, // Default to required
			Parameters:  map[string]interface{}{"criteria": rule["criteria"]}, // Store criteria in parameters
		})
	}
	return adapted
}

// adaptDependencies converts dependency maps to existing format
func adaptDependencies(deps []map[string]string) []types.Dependency {
	var adapted []types.Dependency
	for _, dep := range deps {
		adapted = append(adapted, types.Dependency{
			Name:        dep["name"],
			Version:     dep["version"],
			Type:        dep["type"],
			Description: dep["description"],
			Required:    true, // Default to required
		})
	}
	return adapted
}

// Helper functions

func containsLevel(levels []string, level string) bool {
	for _, l := range levels {
		if l == level {
			return true
		}
	}
	return false
}