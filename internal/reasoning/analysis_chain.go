package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/Zerofisher/goai/pkg/types"
)

// buildAnalysisChain creates the problem analysis chain
func (e *Engine) buildAnalysisChain(ctx context.Context) error {
	// System prompt for problem analysis
	systemPrompt := `You are a senior software architect specializing in Go development and software design. 
Analyze programming problems using this structured approach:

1. **Problem Domain Analysis**: Identify the core domain, business context, and technical scope
2. **Technical Challenge Assessment**: Evaluate complexity, technical hurdles, and architectural requirements  
3. **Architecture Pattern Recommendation**: Suggest appropriate design patterns and architectural approaches
4. **Implementation Strategy**: Outline step-by-step technical approach with Go best practices
5. **Risk Assessment**: Identify potential issues, edge cases, and mitigation strategies

Consider Go language idioms, performance characteristics, concurrency patterns, and ecosystem best practices.
Focus on maintainable, testable, and scalable solutions.

Respond with a structured JSON analysis following this format:
{
  "problem_domain": "string",
  "technical_stack": ["string"],
  "architecture_pattern": "string", 
  "risk_factors": [{"type": "string", "description": "string", "severity": "string", "mitigation": "string"}],
  "recommendations": [{"category": "string", "description": "string", "priority": "string", "impact": "string"}],
  "complexity": "low|medium|high"
}`


	// Create the analysis chain
	e.analysisChain = compose.NewChain[*types.ProblemRequest, *types.Analysis]()
	e.analysisChain.
		// Pre-processing: enhance request with context
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, req *types.ProblemRequest) (*types.ProblemRequest, error) {
			// Add any pre-processing logic here
			// For example, sanitize inputs, add default values, etc.
			if req.Context == nil {
				req.Context = &types.ProjectContext{}
			}
			return req, nil
		})).
		// Generate structured prompt directly without template
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, req *types.ProblemRequest) ([]*schema.Message, error) {
			// Build the user message directly
			userMessage := fmt.Sprintf(`Analyze this programming requirement:

**Problem Description**: %s

**Project Context**:
- Working Directory: %s
- Project Language: Go

**Requirements**:
%s

**Constraints**:
%s

Provide a comprehensive technical analysis following the structured approach above.`, 
				req.Description,
				func() string {
					if req.Context != nil {
						return req.Context.WorkingDirectory
					}
					return "/tmp"
				}(),
				func() string {
					if len(req.Requirements) > 0 {
						var result strings.Builder
						for _, r := range req.Requirements {
							result.WriteString("- " + r + "\n")
						}
						return result.String()
					}
					return "- No specific requirements"
				}(),
				func() string {
					if len(req.Constraints) > 0 {
						var result strings.Builder
						for _, c := range req.Constraints {
							result.WriteString("- " + c + "\n")
						}
						return result.String()
					}
					return "- No specific constraints"
				}(),
			)
			
			return []*schema.Message{
				{
					Role: schema.System,
					Content: systemPrompt,
				},
				{
					Role: schema.User,
					Content: userMessage,
				},
			}, nil
		})).
		// LLM reasoning
		AppendChatModel(e.chatModel).
		// Parse response into structured analysis
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (*types.Analysis, error) {
			return parseAnalysisResponse(msg.Content)
		}))

	return nil
}

// parseAnalysisResponse parses LLM response into structured Analysis
func parseAnalysisResponse(content string) (*types.Analysis, error) {
	// Clean the content to extract JSON
	content = strings.TrimSpace(content)
	
	// Find JSON block if wrapped in markdown
	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json") + 7
		end := strings.LastIndex(content, "```")
		if start < end && end > start {
			content = content[start:end]
		}
	} else if strings.Contains(content, "```") {
		start := strings.Index(content, "```") + 3
		end := strings.LastIndex(content, "```")
		if start < end && end > start {
			content = content[start:end]
		}
	}

	// Try to parse as JSON
	var rawAnalysis struct {
		ProblemDomain       string                  `json:"problem_domain"`
		TechnicalStack      []string                `json:"technical_stack"`
		ArchitecturePattern string                  `json:"architecture_pattern"`
		RiskFactors         []map[string]string     `json:"risk_factors"`
		Recommendations     []map[string]string     `json:"recommendations"`
		Complexity          string                  `json:"complexity"`
	}

	if err := json.Unmarshal([]byte(content), &rawAnalysis); err != nil {
		// Fallback: create a basic analysis from the text response
		return createFallbackAnalysis(content), nil
	}

	// Convert to typed structures
	analysis := &types.Analysis{
		ProblemDomain:       rawAnalysis.ProblemDomain,
		TechnicalStack:      rawAnalysis.TechnicalStack,
		ArchitecturePattern: rawAnalysis.ArchitecturePattern,
		Complexity:          types.ComplexityLevel(rawAnalysis.Complexity),
	}

	// Convert risk factors
	for _, rf := range rawAnalysis.RiskFactors {
		analysis.RiskFactors = append(analysis.RiskFactors, types.RiskFactor{
			Type:        rf["type"],
			Description: rf["description"],
			Severity:    rf["severity"],
			Mitigation:  rf["mitigation"],
		})
	}

	// Convert recommendations
	for _, rec := range rawAnalysis.Recommendations {
		analysis.Recommendations = append(analysis.Recommendations, types.Recommendation{
			Category:    rec["category"],
			Description: rec["description"],
			Priority:    rec["priority"],
			Impact:      rec["impact"],
		})
	}

	return analysis, nil
}

// createFallbackAnalysis creates a basic analysis when JSON parsing fails
func createFallbackAnalysis(content string) *types.Analysis {
	complexity := types.ComplexityMedium
	if strings.Contains(strings.ToLower(content), "complex") || 
	   strings.Contains(strings.ToLower(content), "difficult") {
		complexity = types.ComplexityHigh
	} else if strings.Contains(strings.ToLower(content), "simple") || 
	          strings.Contains(strings.ToLower(content), "straightforward") {
		complexity = types.ComplexityLow
	}

	return &types.Analysis{
		ProblemDomain:       "General Programming",
		TechnicalStack:      []string{"Go"},
		ArchitecturePattern: "Standard Go Application",
		RiskFactors: []types.RiskFactor{
			{
				Type:        "parsing",
				Description: "LLM response could not be parsed as structured JSON",
				Severity:    "medium",
				Mitigation:  "Using fallback analysis, consider retrying with refined prompts",
			},
		},
		Recommendations: []types.Recommendation{
			{
				Category:    "implementation",
				Description: "Review the unstructured analysis and implement according to Go best practices",
				Priority:    "high",
				Impact:      "medium",
			},
		},
		Complexity: complexity,
	}
}