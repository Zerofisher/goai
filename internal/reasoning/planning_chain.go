package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/Zerofisher/goai/pkg/types"
)

// buildPlanningChain creates the execution planning chain
func (e *Engine) buildPlanningChain(ctx context.Context) error {
	// System prompt for execution planning
	systemPrompt := `You are a senior technical lead responsible for creating detailed implementation plans.
Transform technical analysis into actionable execution plans with specific steps, dependencies, and timelines.

Your planning approach:
1. **Break Down Complexity**: Decompose the solution into manageable, concrete steps
2. **Identify Dependencies**: Map out step dependencies and prerequisites
3. **Estimate Effort**: Provide realistic time estimates for each step
4. **Plan Testing Strategy**: Include comprehensive testing approach
5. **Define Validation Rules**: Specify quality checks and acceptance criteria

Focus on:
- Go language best practices and conventions
- Clear, atomic steps that can be implemented independently
- Proper error handling and edge case considerations
- Performance and scalability considerations
- Testing strategy at unit, integration, and system levels

Respond with structured JSON following this format:
{
  "steps": [
    {
      "id": "string",
      "name": "string", 
      "description": "string",
      "dependencies": ["string"],
      "estimated_time": "duration_string",
      "priority": number
    }
  ],
  "dependencies": [
    {
      "name": "string",
      "version": "string",
      "type": "string", 
      "description": "string"
    }
  ],
  "timeline": {
    "total_estimate": "duration_string",
    "phases": [
      {
        "name": "string",
        "duration": "duration_string",
        "steps": ["string"]
      }
    ]
  },
  "test_strategy": {
    "approach": "string",
    "levels": ["string"],
    "frameworks": ["string"],
    "coverage_target": number
  },
  "validation_rules": [
    {
      "type": "string",
      "description": "string",
      "criteria": "string"
    }
  ]
}`


	// Create the planning chain
	e.planningChain = compose.NewChain[*types.Analysis, *types.ExecutionPlan]()
	e.planningChain.
		// Pre-processing: validate analysis input
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, analysis *types.Analysis) (*types.Analysis, error) {
			if analysis == nil {
				return nil, fmt.Errorf("analysis cannot be nil")
			}
			// Set defaults if missing
			if analysis.ProblemDomain == "" {
				analysis.ProblemDomain = "General Programming"
			}
			if len(analysis.TechnicalStack) == 0 {
				analysis.TechnicalStack = []string{"Go"}
			}
			return analysis, nil
		})).
		// Generate planning prompt directly
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, analysis *types.Analysis) ([]*schema.Message, error) {
			// Build user message from analysis
			userMessage := fmt.Sprintf(`Create a detailed implementation plan based on this technical analysis:

**Problem Domain**: %s
**Architecture Pattern**: %s
**Technical Stack**: %s
**Complexity Level**: %s

**Risk Factors**:
%s

**Recommendations**:
%s

Create a comprehensive implementation plan that addresses all aspects of the analysis, includes proper Go project structure, testing strategy, and quality validation steps.`,
				analysis.ProblemDomain,
				analysis.ArchitecturePattern,
				strings.Join(analysis.TechnicalStack, ", "),
				string(analysis.Complexity),
				func() string {
					if len(analysis.RiskFactors) > 0 {
						result := ""
						for _, rf := range analysis.RiskFactors {
							result += fmt.Sprintf("- %s: %s (%s)\n  Mitigation: %s\n", rf.Type, rf.Description, rf.Severity, rf.Mitigation)
						}
						return result
					}
					return "- No specific risk factors identified"
				}(),
				func() string {
					if len(analysis.Recommendations) > 0 {
						result := ""
						for _, rec := range analysis.Recommendations {
							result += fmt.Sprintf("- %s: %s (Priority: %s)\n", rec.Category, rec.Description, rec.Priority)
						}
						return result
					}
					return "- No specific recommendations"
				}(),
			)

			return []*schema.Message{
				{
					Role:    schema.System,
					Content: systemPrompt,
				},
				{
					Role:    schema.User,
					Content: userMessage,
				},
			}, nil
		})).
		// LLM planning
		AppendChatModel(e.chatModel).
		// Parse response into execution plan
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (*types.ExecutionPlan, error) {
			return parsePlanningResponse(msg.Content)
		}))

	return nil
}

// parsePlanningResponse parses LLM response into structured ExecutionPlan
func parsePlanningResponse(content string) (*types.ExecutionPlan, error) {
	// Clean the content to extract JSON
	content = strings.TrimSpace(content)
	
	// Find JSON block if wrapped in markdown
	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json") + 7
		end := strings.LastIndex(content, "```")
		if start < end && end > start {
			content = content[start:end]
		}
	}

	// Try to parse as JSON
	var rawPlan struct {
		Steps           []map[string]interface{} `json:"steps"`
		Dependencies    []map[string]string      `json:"dependencies"`
		Timeline        map[string]interface{}   `json:"timeline"`
		TestStrategy    map[string]interface{}   `json:"test_strategy"`
		ValidationRules []map[string]string      `json:"validation_rules"`
	}

	if err := json.Unmarshal([]byte(content), &rawPlan); err != nil {
		return createFallbackPlan(content), nil
	}

	// Convert to typed structures
	plan := &types.ExecutionPlan{}

	// Convert steps
	for i, stepData := range rawPlan.Steps {
		step := types.PlanStep{
			ID:          getStringFromMap(stepData, "id", fmt.Sprintf("step_%d", i+1)),
			Name:        getStringFromMap(stepData, "name", fmt.Sprintf("Step %d", i+1)),
			Description: getStringFromMap(stepData, "description", ""),
			Priority:    getIntFromMap(stepData, "priority", i+1),
		}
		
		// Parse duration
		if durationStr := getStringFromMap(stepData, "estimated_time", "1h"); durationStr != "" {
			if duration, err := time.ParseDuration(durationStr); err == nil {
				step.EstimatedTime = duration
			} else {
				step.EstimatedTime = time.Hour // default 1 hour
			}
		}

		// Parse dependencies
		if deps, ok := stepData["dependencies"].([]interface{}); ok {
			for _, dep := range deps {
				if depStr, ok := dep.(string); ok {
					step.Dependencies = append(step.Dependencies, depStr)
				}
			}
		}

		plan.Steps = append(plan.Steps, step)
	}

	// Convert dependencies using adapter
	plan.Dependencies = adaptDependencies(rawPlan.Dependencies)

	// Convert timeline using adapter
	if rawPlan.Timeline != nil {
		if totalEst := getStringFromMap(rawPlan.Timeline, "total_estimate", ""); totalEst != "" {
			if duration, err := time.ParseDuration(totalEst); err == nil {
				plan.Timeline = adaptTimeline(duration)
			}
		}
	}

	// Convert test strategy using adapter
	if rawPlan.TestStrategy != nil {
		approach := getStringFromMap(rawPlan.TestStrategy, "approach", "unit_and_integration")
		coverageTarget := float64(getIntFromMap(rawPlan.TestStrategy, "coverage_target", 80))
		
		var levels []string
		if levelsData, ok := rawPlan.TestStrategy["levels"].([]interface{}); ok {
			for _, level := range levelsData {
				if levelStr, ok := level.(string); ok {
					levels = append(levels, levelStr)
				}
			}
		}
		
		var frameworks []string
		if frameworksData, ok := rawPlan.TestStrategy["frameworks"].([]interface{}); ok {
			for _, framework := range frameworksData {
				if frameworkStr, ok := framework.(string); ok {
					frameworks = append(frameworks, frameworkStr)
				}
			}
		}
		
		plan.TestStrategy = adaptTestStrategy(approach, levels, frameworks, coverageTarget)
	}

	// Convert validation rules using adapter
	plan.ValidationRules = adaptValidationRules(rawPlan.ValidationRules)

	return plan, nil
}

// Helper functions for safe map access
func getStringFromMap(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func getIntFromMap(m map[string]interface{}, key string, defaultValue int) int {
	if val, ok := m[key]; ok {
		if num, ok := val.(float64); ok {
			return int(num)
		}
		if num, ok := val.(int); ok {
			return num
		}
	}
	return defaultValue
}

// createFallbackPlan creates a basic plan when JSON parsing fails
func createFallbackPlan(content string) *types.ExecutionPlan {
	return &types.ExecutionPlan{
		Steps: []types.PlanStep{
			{
				ID:            "step_1",
				Name:          "Analyze Requirements",
				Description:   "Review and understand the requirements based on analysis",
				EstimatedTime: time.Hour,
				Priority:      1,
			},
			{
				ID:            "step_2", 
				Name:          "Design Solution",
				Description:   "Create detailed design based on the analysis",
				Dependencies:  []string{"step_1"},
				EstimatedTime: 2 * time.Hour,
				Priority:      2,
			},
			{
				ID:            "step_3",
				Name:          "Implement Solution",
				Description:   "Code the solution following Go best practices",
				Dependencies:  []string{"step_2"},
				EstimatedTime: 4 * time.Hour,
				Priority:      3,
			},
			{
				ID:            "step_4",
				Name:          "Write Tests",
				Description:   "Create comprehensive test suite",
				Dependencies:  []string{"step_3"},
				EstimatedTime: 2 * time.Hour,
				Priority:      4,
			},
			{
				ID:            "step_5",
				Name:          "Validate Solution",
				Description:   "Review and validate the implementation",
				Dependencies:  []string{"step_4"},
				EstimatedTime: time.Hour,
				Priority:      5,
			},
		},
		Dependencies: []types.Dependency{}, // Fixed: removed pointer
		Timeline: adaptTimeline(10 * time.Hour), // Use adapter
		TestStrategy: adaptTestStrategy("comprehensive", []string{"unit", "integration"}, []string{"testing", "testify"}, 80.0), // Use adapter
		ValidationRules: adaptValidationRules([]map[string]string{ // Use adapter
			{
				"type":        "code_quality",
				"description": "Ensure code follows Go best practices",
				"criteria":    "gofmt, golint, go vet pass",
			},
			{
				"type":        "test_coverage",
				"description": "Maintain adequate test coverage",
				"criteria":    "coverage >= 80%",
			},
		}),
	}
}