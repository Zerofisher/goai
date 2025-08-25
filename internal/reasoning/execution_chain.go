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

// buildExecutionChain creates the code generation and execution chain
func (e *Engine) buildExecutionChain(ctx context.Context) error {
	// System prompt for code execution
	systemPrompt := `You are an expert Go developer responsible for implementing detailed execution plans.
Transform execution plans into working, production-ready code following Go best practices.

Your implementation approach:
1. **Follow Go Conventions**: Use standard Go project layout, naming conventions, and idioms
2. **Write Clean Code**: Clear, readable, maintainable code with proper documentation
3. **Error Handling**: Comprehensive error handling with meaningful error messages
4. **Testing**: Include unit tests for all significant functions
5. **Documentation**: Add appropriate comments and documentation strings

Implementation Guidelines:
- Use standard Go project structure (cmd/, pkg/, internal/)
- Follow effective Go style guide and best practices
- Include proper package documentation
- Implement graceful error handling and recovery
- Use context for cancellation and timeouts where appropriate
- Include comprehensive logging for debugging and monitoring

Respond with structured JSON containing the generated code:
{
  "generated_files": [
    {
      "path": "string",
      "content": "string",
      "file_type": "string",
      "description": "string"
    }
  ],
  "tests": {
    "test_files": [
      {
        "path": "string", 
        "content": "string",
        "description": "string"
      }
    ],
    "coverage_estimate": number,
    "test_approach": "string"
  },
  "documentation": {
    "readme_content": "string",
    "api_documentation": "string",
    "usage_examples": ["string"]
  },
  "metadata": {
    "total_files": number,
    "estimated_lines": number,
    "complexity_score": number,
    "go_version": "string"
  }
}`


	// Create the execution chain
	e.executionChain = compose.NewChain[*types.ExecutionPlan, *types.CodeResult]()
	e.executionChain.
		// Pre-processing: validate execution plan
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, plan *types.ExecutionPlan) (*types.ExecutionPlan, error) {
			if plan == nil {
				return nil, fmt.Errorf("execution plan cannot be nil")
			}
			if len(plan.Steps) == 0 {
				return nil, fmt.Errorf("execution plan must contain at least one step")
			}
			return plan, nil
		})).
		// Generate execution prompt directly
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, plan *types.ExecutionPlan) ([]*schema.Message, error) {
			// Build user message from execution plan
			userMessage := fmt.Sprintf(`Implement the following execution plan with high-quality Go code:

%s

%s

%s

%s

Generate complete, working Go code that implements all steps in the plan. Include comprehensive tests and documentation.`,
				func() string {
					if len(plan.Steps) > 0 {
						result := ""
						for _, step := range plan.Steps {
							deps := strings.Join(step.Dependencies, ", ")
							if deps == "" {
								deps = "None"
							}
							result += fmt.Sprintf("**Step %d: %s**\n- Description: %s\n- Dependencies: %s\n- Estimated Time: %s\n\n",
								step.Priority, step.Name, step.Description, deps, step.EstimatedTime.String())
						}
						return result
					}
					return "No steps defined"
				}(),
				func() string {
					if len(plan.Dependencies) > 0 {
						result := "**Project Dependencies**:\n"
						for _, dep := range plan.Dependencies {
							result += fmt.Sprintf("- %s %s (%s): %s\n", dep.Name, dep.Version, dep.Type, dep.Description)
						}
						return result
					}
					return ""
				}(),
				func() string {
					if plan.TestStrategy != nil {
						frameworks := strings.Join(plan.TestStrategy.TestFrameworks, ", ")
						return fmt.Sprintf("**Test Strategy**:\n- Unit Tests: %t\n- Integration Tests: %t\n- Frameworks: %s\n- Coverage Target: %.0f%%\n",
							plan.TestStrategy.UnitTests, plan.TestStrategy.IntegrationTests, frameworks, plan.TestStrategy.CoverageTarget)
					}
					return ""
				}(),
				func() string {
					if len(plan.ValidationRules) > 0 {
						result := "**Validation Requirements**:\n"
						for _, rule := range plan.ValidationRules {
							result += fmt.Sprintf("- %s: %s\n", rule.Type, rule.Description)
						}
						return result
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
		// LLM code generation
		AppendChatModel(e.chatModel).
		// Parse response into code result
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, msg *schema.Message) (*types.CodeResult, error) {
			return parseExecutionResponse(msg.Content)
		}))

	return nil
}

// parseExecutionResponse parses LLM response into structured CodeResult
func parseExecutionResponse(content string) (*types.CodeResult, error) {
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
	var rawResult struct {
		GeneratedFiles []map[string]interface{} `json:"generated_files"`
		Tests          map[string]interface{}   `json:"tests"`
		Documentation  map[string]interface{}   `json:"documentation"`
		Metadata       map[string]interface{}   `json:"metadata"`
	}

	if err := json.Unmarshal([]byte(content), &rawResult); err != nil {
		return createFallbackCodeResult(content), nil
	}

	// Convert to typed structures
	result := &types.CodeResult{}

	// Convert generated files
	for _, fileData := range rawResult.GeneratedFiles {
		file := types.GeneratedFile{
			Path:        getStringFromMap(fileData, "path", "main.go"),
			Content:     getStringFromMap(fileData, "content", "package main\n\nfunc main() {\n\t// TODO: Implement\n}"),
			FileType:    getStringFromMap(fileData, "file_type", "source"),
			Description: getStringFromMap(fileData, "description", "Generated Go file"),
		}
		result.GeneratedFiles = append(result.GeneratedFiles, file)
	}

	// Convert tests
	if rawResult.Tests != nil {
		testSuite := &types.TestSuite{
			TestApproach:     getStringFromMap(rawResult.Tests, "test_approach", "unit_testing"),
			CoverageEstimate: float64(getIntFromMap(rawResult.Tests, "coverage_estimate", 80)),
		}

		if testFiles, ok := rawResult.Tests["test_files"].([]interface{}); ok {
			for _, testFile := range testFiles {
				if testData, ok := testFile.(map[string]interface{}); ok {
					test := types.TestFile{
						Path:        getStringFromMap(testData, "path", "main_test.go"),
						Content:     getStringFromMap(testData, "content", "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {\n\t// TODO: Implement test\n}"),
						Description: getStringFromMap(testData, "description", "Generated test file"),
					}
					testSuite.TestFiles = append(testSuite.TestFiles, test)
				}
			}
		}
		result.Tests = testSuite
	}

	// Convert documentation
	if rawResult.Documentation != nil {
		doc := &types.Documentation{
			ReadmeContent:    getStringFromMap(rawResult.Documentation, "readme_content", "# Generated Project\n\nTODO: Add description"),
			APIDocumentation: getStringFromMap(rawResult.Documentation, "api_documentation", ""),
		}

		if examples, ok := rawResult.Documentation["usage_examples"].([]interface{}); ok {
			for _, example := range examples {
				if exampleStr, ok := example.(string); ok {
					doc.UsageExamples = append(doc.UsageExamples, exampleStr)
				}
			}
		}
		result.Documentation = doc
	}

	// Convert metadata
	if rawResult.Metadata != nil {
		metadata := &types.GenerationMeta{
			TotalFiles:      getIntFromMap(rawResult.Metadata, "total_files", len(result.GeneratedFiles)),
			EstimatedLines:  getIntFromMap(rawResult.Metadata, "estimated_lines", 100),
			ComplexityScore: float64(getIntFromMap(rawResult.Metadata, "complexity_score", 50)),
			GoVersion:       getStringFromMap(rawResult.Metadata, "go_version", "1.21"),
		}
		result.Metadata = metadata
	}

	return result, nil
}

// createFallbackCodeResult creates a basic code result when JSON parsing fails
func createFallbackCodeResult(content string) *types.CodeResult {
	// Extract any code blocks from the content
	codeBlocks := extractCodeBlocks(content)
	
	result := &types.CodeResult{
		GeneratedFiles: []types.GeneratedFile{
			{
				Path:        "main.go",
				Content:     getMainCodeBlock(codeBlocks),
				FileType:    "source",
				Description: "Main implementation file",
			},
		},
		Tests: &types.TestSuite{
			TestFiles: []types.TestFile{
				{
					Path:        "main_test.go",
					Content:     getTestCodeBlock(codeBlocks),
					Description: "Test file for main implementation",
				},
			},
			TestApproach:     "basic_unit_testing",
			CoverageEstimate: 70.0,
		},
		Documentation: &types.Documentation{
			ReadmeContent:    "# Generated Go Project\n\nThis project was generated by GoAI.\n\n## Usage\n\n```bash\ngo run main.go\n```",
			APIDocumentation: "",
			UsageExamples:    []string{"go run main.go"},
		},
		Metadata: &types.GenerationMeta{
			TotalFiles:      2,
			EstimatedLines:  50,
			ComplexityScore: 25.0,
			GoVersion:       "1.21",
		},
	}

	return result
}

// extractCodeBlocks extracts Go code blocks from markdown-formatted content
func extractCodeBlocks(content string) []string {
	var codeBlocks []string
	
	lines := strings.Split(content, "\n")
	var currentBlock strings.Builder
	inCodeBlock := false
	
	for _, line := range lines {
		if strings.HasPrefix(line, "```go") || strings.HasPrefix(line, "```golang") {
			inCodeBlock = true
			continue
		} else if strings.HasPrefix(line, "```") && inCodeBlock {
			if currentBlock.Len() > 0 {
				codeBlocks = append(codeBlocks, currentBlock.String())
				currentBlock.Reset()
			}
			inCodeBlock = false
		} else if inCodeBlock {
			currentBlock.WriteString(line)
			currentBlock.WriteString("\n")
		}
	}
	
	return codeBlocks
}

// getMainCodeBlock returns the main implementation code or a default
func getMainCodeBlock(codeBlocks []string) string {
	if len(codeBlocks) > 0 {
		return codeBlocks[0]
	}
	
	return `package main

import "fmt"

func main() {
	fmt.Println("Hello from GoAI generated code!")
	// TODO: Implement the actual functionality
}
`
}

// getTestCodeBlock returns test code or generates a basic test
func getTestCodeBlock(codeBlocks []string) string {
	// Look for test-related code blocks
	for _, block := range codeBlocks {
		if strings.Contains(block, "func Test") || strings.Contains(block, "import \"testing\"") {
			return block
		}
	}
	
	return `package main

import "testing"

func TestMain(t *testing.T) {
	// TODO: Implement meaningful tests
	t.Log("Generated test - implement actual test logic")
}
`
}