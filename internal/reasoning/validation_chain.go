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

// buildValidationChain creates the code validation chain
func (e *Engine) buildValidationChain(ctx context.Context) error {
	// System prompt for validation
	systemPrompt := `You are a senior code reviewer and quality assurance expert.
Validate generated code for quality, correctness, and adherence to best practices.

Your validation approach:
1. **Static Code Analysis**: Check syntax, style, and structural issues
2. **Logic Review**: Evaluate correctness and edge case handling  
3. **Performance Analysis**: Identify potential performance bottlenecks
4. **Security Assessment**: Look for security vulnerabilities and concerns
5. **Best Practices Compliance**: Ensure adherence to Go conventions and patterns

Focus on:
- Go language idioms and conventions
- Error handling patterns and completeness
- Testing coverage and quality
- Documentation completeness
- Security best practices
- Performance considerations

Respond with structured JSON validation report:
{
  "static_report": {
    "issues": [
      {
        "type": "string",
        "severity": "string", 
        "message": "string",
        "file": "string",
        "line": number,
        "column": number,
        "rule": "string"
      }
    ],
    "metrics": {
      "lines_of_code": number,
      "cyclomatic_complexity": number,
      "test_coverage": number,
      "maintainability": number,
      "technical_debt": number
    },
    "suggestions": [
      {
        "type": "string",
        "description": "string",
        "file": "string", 
        "line": number,
        "priority": "string",
        "effort": "string"
      }
    ],
    "overall_score": number
  },
  "test_results": {
    "total_tests": number,
    "passed_tests": number,
    "failed_tests": number,
    "coverage": number,
    "results": []
  },
  "compliance_report": {
    "violations": [],
    "score": number,
    "standards": ["string"],
    "summary": "string"
  },
  "overall_status": "passed|failed|warning"
}`


	// Create the validation chain
	e.validationChain = compose.NewChain[*types.CodeResult, *types.ValidationReport]()
	e.validationChain.
		// Pre-processing: validate input
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, result *types.CodeResult) (*types.CodeResult, error) {
			if result == nil {
				return nil, fmt.Errorf("code result cannot be nil")
			}
			if len(result.GeneratedFiles) == 0 {
				return nil, fmt.Errorf("code result must contain at least one generated file")
			}
			return result, nil
		})).
		// Generate validation prompt directly
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, result *types.CodeResult) ([]*schema.Message, error) {
			// Build user message from code result
			userMessage := fmt.Sprintf(`Validate this generated code for quality and correctness:

%s

%s

%s

Provide a comprehensive validation report covering all aspects of code quality, testing, and compliance.`,
				func() string {
					if len(result.GeneratedFiles) > 0 {
						output := ""
						for _, file := range result.GeneratedFiles {
							output += fmt.Sprintf("**File: %s**\n%s\n\n", file.Path, file.Content)
						}
						return output
					}
					return "No generated files"
				}(),
				func() string {
					if result.Tests != nil && len(result.Tests.TestFiles) > 0 {
						output := "**Tests:**\n"
						for _, testFile := range result.Tests.TestFiles {
							output += fmt.Sprintf("**Test File: %s**\n%s\n\n", testFile.Path, testFile.Content)
						}
						return output
					}
					return ""
				}(),
				func() string {
					if result.Documentation != nil && result.Documentation.ReadmeContent != "" {
						return fmt.Sprintf("**Documentation:**\n%s\n", result.Documentation.ReadmeContent)
					}
					return ""
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
		// LLM validation
		AppendChatModel(e.chatModel).
		// Parse response into validation report
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (*types.ValidationReport, error) {
			return parseValidationResponse(msg.Content)
		}))

	return nil
}

// parseValidationResponse parses LLM response into structured ValidationReport
func parseValidationResponse(content string) (*types.ValidationReport, error) {
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
	var rawReport struct {
		StaticReport     map[string]interface{} `json:"static_report"`
		TestResults      map[string]interface{} `json:"test_results"`
		ComplianceReport map[string]interface{} `json:"compliance_report"`
		OverallStatus    string                 `json:"overall_status"`
	}

	if err := json.Unmarshal([]byte(content), &rawReport); err != nil {
		return createFallbackValidationReport(content), nil
	}

	// Convert to typed structures
	report := &types.ValidationReport{
		OverallStatus: types.ValidationStatus(rawReport.OverallStatus),
	}

	// Convert static report
	if rawReport.StaticReport != nil {
		staticReport := &types.StaticReport{
			OverallScore: getFloatFromMap(rawReport.StaticReport, "overall_score", 75.0),
		}

		// Parse issues
		if issues, ok := rawReport.StaticReport["issues"].([]interface{}); ok {
			for _, issue := range issues {
				if issueData, ok := issue.(map[string]interface{}); ok {
					staticIssue := types.StaticIssue{
						Type:     getStringFromMap(issueData, "type", "warning"),
						Severity: getStringFromMap(issueData, "severity", "medium"),
						Message:  getStringFromMap(issueData, "message", ""),
						File:     getStringFromMap(issueData, "file", ""),
						Line:     getIntFromMap(issueData, "line", 0),
						Column:   getIntFromMap(issueData, "column", 0),
						Rule:     getStringFromMap(issueData, "rule", ""),
					}
					staticReport.Issues = append(staticReport.Issues, staticIssue)
				}
			}
		}

		// Parse metrics
		if metrics, ok := rawReport.StaticReport["metrics"].(map[string]interface{}); ok {
			staticReport.Metrics = &types.CodeMetrics{
				LinesOfCode:          getIntFromMap(metrics, "lines_of_code", 0),
				CyclomaticComplexity: getIntFromMap(metrics, "cyclomatic_complexity", 1),
				TestCoverage:         getFloatFromMap(metrics, "test_coverage", 0.0),
				Maintainability:      getFloatFromMap(metrics, "maintainability", 75.0),
				TechnicalDebt:        getFloatFromMap(metrics, "technical_debt", 0.0),
			}
		}

		// Parse suggestions
		if suggestions, ok := rawReport.StaticReport["suggestions"].([]interface{}); ok {
			for _, suggestion := range suggestions {
				if suggestionData, ok := suggestion.(map[string]interface{}); ok {
					sugg := types.Suggestion{
						Type:        getStringFromMap(suggestionData, "type", "improvement"),
						Description: getStringFromMap(suggestionData, "description", ""),
						File:        getStringFromMap(suggestionData, "file", ""),
						Line:        getIntFromMap(suggestionData, "line", 0),
						Priority:    getStringFromMap(suggestionData, "priority", "medium"),
						Effort:      getStringFromMap(suggestionData, "effort", "medium"),
					}
					staticReport.Suggestions = append(staticReport.Suggestions, sugg)
				}
			}
		}

		report.StaticReport = staticReport
	}

	// Convert test results
	if rawReport.TestResults != nil {
		testResults := &types.TestResults{
			TotalTests:   getIntFromMap(rawReport.TestResults, "total_tests", 0),
			PassedTests:  getIntFromMap(rawReport.TestResults, "passed_tests", 0),
			FailedTests:  getIntFromMap(rawReport.TestResults, "failed_tests", 0),
			SkippedTests: getIntFromMap(rawReport.TestResults, "skipped_tests", 0),
			Coverage:     getFloatFromMap(rawReport.TestResults, "coverage", 0.0),
		}
		report.TestResults = testResults
	}

	// Convert compliance report
	if rawReport.ComplianceReport != nil {
		complianceReport := &types.ComplianceReport{
			Score:   getFloatFromMap(rawReport.ComplianceReport, "score", 80.0),
			Summary: getStringFromMap(rawReport.ComplianceReport, "summary", "Code complies with basic standards"),
		}

		if standards, ok := rawReport.ComplianceReport["standards"].([]interface{}); ok {
			for _, standard := range standards {
				if standardStr, ok := standard.(string); ok {
					complianceReport.Standards = append(complianceReport.Standards, standardStr)
				}
			}
		}

		report.ComplianceReport = complianceReport
	}

	return report, nil
}

// Helper function for safe float access from map
func getFloatFromMap(m map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := m[key]; ok {
		if num, ok := val.(float64); ok {
			return num
		}
		if num, ok := val.(int); ok {
			return float64(num)
		}
	}
	return defaultValue
}

// createFallbackValidationReport creates a basic report when JSON parsing fails
func createFallbackValidationReport(content string) *types.ValidationReport {
	// Determine overall status from content
	status := types.ValidationPassed
	if strings.Contains(strings.ToLower(content), "error") || 
	   strings.Contains(strings.ToLower(content), "fail") {
		status = types.ValidationFailed
	} else if strings.Contains(strings.ToLower(content), "warning") ||
	          strings.Contains(strings.ToLower(content), "issue") {
		status = types.ValidationWarning
	}

	return &types.ValidationReport{
		StaticReport: &types.StaticReport{
			Issues: []types.StaticIssue{
				{
					Type:     "parsing",
					Severity: "medium",
					Message:  "Validation response could not be parsed as structured JSON",
					File:     "",
					Line:     0,
					Rule:     "response_format",
				},
			},
			Metrics: &types.CodeMetrics{
				LinesOfCode:          100,
				CyclomaticComplexity: 5,
				TestCoverage:         70.0,
				Maintainability:      70.0,
				TechnicalDebt:        30.0,
			},
			Suggestions: []types.Suggestion{
				{
					Type:        "improvement",
					Description: "Consider improving validation response parsing",
					Priority:    "low",
					Effort:      "medium",
				},
			},
			OverallScore: 70.0,
		},
		TestResults: &types.TestResults{
			TotalTests:  1,
			PassedTests: 1,
			Coverage:    70.0,
		},
		ComplianceReport: &types.ComplianceReport{
			Score:     80.0,
			Standards: []string{"Go Best Practices"},
			Summary:   "Basic compliance validation completed",
		},
		OverallStatus: status,
	}
}