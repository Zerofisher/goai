package types

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ValidationError represents a validation error with field context
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", ve.Field, ve.Message)
}

// ValidationResult holds the result of validation
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

// AddError adds a validation error
func (vr *ValidationResult) AddError(field, message string, value interface{}) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// FirstError returns the first validation error or nil
func (vr *ValidationResult) FirstError() *ValidationError {
	if len(vr.Errors) > 0 {
		return &vr.Errors[0]
	}
	return nil
}

// DataValidator provides validation methods for all data models
type DataValidator struct {
	// Configuration for validation rules
	MaxDescriptionLength int
	MaxRequirements      int
	MaxConstraints       int
	MaxSteps             int
	MinEstimatedTime     time.Duration
	MaxEstimatedTime     time.Duration
}

// NewDataValidator creates a new validator with default settings
func NewDataValidator() *DataValidator {
	return &DataValidator{
		MaxDescriptionLength: 10000,
		MaxRequirements:      50,
		MaxConstraints:       20,
		MaxSteps:             100,
		MinEstimatedTime:     time.Minute,
		MaxEstimatedTime:     time.Hour * 24 * 30, // 30 days
	}
}

// ValidateProblemRequest validates a ProblemRequest
func (v *DataValidator) ValidateProblemRequest(req *ProblemRequest) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if req == nil {
		result.AddError("request", "request cannot be nil", nil)
		return result
	}
	
	// Validate description
	if strings.TrimSpace(req.Description) == "" {
		result.AddError("description", "description cannot be empty", req.Description)
	}
	
	if len(req.Description) > v.MaxDescriptionLength {
		result.AddError("description", 
			fmt.Sprintf("description too long (max %d characters)", v.MaxDescriptionLength), 
			len(req.Description))
	}
	
	// Validate requirements
	if len(req.Requirements) > v.MaxRequirements {
		result.AddError("requirements", 
			fmt.Sprintf("too many requirements (max %d)", v.MaxRequirements), 
			len(req.Requirements))
	}
	
	for i, req_item := range req.Requirements {
		if strings.TrimSpace(req_item) == "" {
			result.AddError(fmt.Sprintf("requirements[%d]", i), 
				"requirement cannot be empty", req_item)
		}
	}
	
	// Validate constraints
	if len(req.Constraints) > v.MaxConstraints {
		result.AddError("constraints", 
			fmt.Sprintf("too many constraints (max %d)", v.MaxConstraints), 
			len(req.Constraints))
	}
	
	for i, constraint := range req.Constraints {
		if strings.TrimSpace(constraint) == "" {
			result.AddError(fmt.Sprintf("constraints[%d]", i), 
				"constraint cannot be empty", constraint)
		}
	}
	
	// Validate context if present
	if req.Context != nil {
		contextResult := v.ValidateProjectContext(req.Context)
		for _, err := range contextResult.Errors {
			result.AddError("context."+err.Field, err.Message, err.Value)
		}
	}
	
	// Validate coding style if present
	if req.PreferredStyle != nil {
		styleResult := v.ValidateCodingStyle(req.PreferredStyle)
		for _, err := range styleResult.Errors {
			result.AddError("preferred_style."+err.Field, err.Message, err.Value)
		}
	}
	
	return result
}

// ValidateAnalysis validates an Analysis
func (v *DataValidator) ValidateAnalysis(analysis *Analysis) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if analysis == nil {
		result.AddError("analysis", "analysis cannot be nil", nil)
		return result
	}
	
	// Validate problem domain
	if strings.TrimSpace(analysis.ProblemDomain) == "" {
		result.AddError("problem_domain", "problem domain cannot be empty", analysis.ProblemDomain)
	}
	
	// Validate complexity level
	validComplexity := map[ComplexityLevel]bool{
		ComplexityLow:    true,
		ComplexityMedium: true,
		ComplexityHigh:   true,
	}
	
	if !validComplexity[analysis.Complexity] {
		result.AddError("complexity", "invalid complexity level", analysis.Complexity)
	}
	
	// Validate technical stack
	if len(analysis.TechnicalStack) == 0 {
		result.AddError("technical_stack", "technical stack cannot be empty", nil)
	}
	
	for i, tech := range analysis.TechnicalStack {
		if strings.TrimSpace(tech) == "" {
			result.AddError(fmt.Sprintf("technical_stack[%d]", i), 
				"technology cannot be empty", tech)
		}
	}
	
	// Validate risk factors
	for i, risk := range analysis.RiskFactors {
		if strings.TrimSpace(risk.Type) == "" {
			result.AddError(fmt.Sprintf("risk_factors[%d].type", i), 
				"risk type cannot be empty", risk.Type)
		}
		if strings.TrimSpace(risk.Description) == "" {
			result.AddError(fmt.Sprintf("risk_factors[%d].description", i), 
				"risk description cannot be empty", risk.Description)
		}
	}
	
	// Validate recommendations
	for i, rec := range analysis.Recommendations {
		if strings.TrimSpace(rec.Category) == "" {
			result.AddError(fmt.Sprintf("recommendations[%d].category", i), 
				"recommendation category cannot be empty", rec.Category)
		}
		if strings.TrimSpace(rec.Description) == "" {
			result.AddError(fmt.Sprintf("recommendations[%d].description", i), 
				"recommendation description cannot be empty", rec.Description)
		}
	}
	
	return result
}

// ValidateExecutionPlan validates an ExecutionPlan
func (v *DataValidator) ValidateExecutionPlan(plan *ExecutionPlan) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if plan == nil {
		result.AddError("plan", "execution plan cannot be nil", nil)
		return result
	}
	
	// Validate steps
	if len(plan.Steps) == 0 {
		result.AddError("steps", "execution plan must have at least one step", nil)
	}
	
	if len(plan.Steps) > v.MaxSteps {
		result.AddError("steps", 
			fmt.Sprintf("too many steps (max %d)", v.MaxSteps), 
			len(plan.Steps))
	}
	
	stepIDs := make(map[string]bool)
	for i, step := range plan.Steps {
		// Validate step ID uniqueness
		if step.ID == "" {
			result.AddError(fmt.Sprintf("steps[%d].id", i), 
				"step ID cannot be empty", step.ID)
		} else if stepIDs[step.ID] {
			result.AddError(fmt.Sprintf("steps[%d].id", i), 
				"duplicate step ID", step.ID)
		}
		stepIDs[step.ID] = true
		
		// Validate step name
		if strings.TrimSpace(step.Name) == "" {
			result.AddError(fmt.Sprintf("steps[%d].name", i), 
				"step name cannot be empty", step.Name)
		}
		
		// Validate estimated time
		if step.EstimatedTime < v.MinEstimatedTime {
			result.AddError(fmt.Sprintf("steps[%d].estimated_time", i), 
				fmt.Sprintf("estimated time too short (min %v)", v.MinEstimatedTime), 
				step.EstimatedTime)
		}
		
		if step.EstimatedTime > v.MaxEstimatedTime {
			result.AddError(fmt.Sprintf("steps[%d].estimated_time", i), 
				fmt.Sprintf("estimated time too long (max %v)", v.MaxEstimatedTime), 
				step.EstimatedTime)
		}
		
		// Validate priority
		if step.Priority < 0 {
			result.AddError(fmt.Sprintf("steps[%d].priority", i), 
				"priority cannot be negative", step.Priority)
		}
	}
	
	// Validate dependencies exist
	for i, step := range plan.Steps {
		for j, depID := range step.Dependencies {
			if !stepIDs[depID] && depID != "" {
				result.AddError(fmt.Sprintf("steps[%d].dependencies[%d]", i, j), 
					"dependency step ID not found", depID)
			}
		}
	}
	
	// Validate timeline if present
	if plan.Timeline != nil {
		timelineResult := v.ValidateTimeline(plan.Timeline)
		for _, err := range timelineResult.Errors {
			result.AddError("timeline."+err.Field, err.Message, err.Value)
		}
	}
	
	// Validate test strategy if present
	if plan.TestStrategy != nil {
		testResult := v.ValidateTestStrategy(plan.TestStrategy)
		for _, err := range testResult.Errors {
			result.AddError("test_strategy."+err.Field, err.Message, err.Value)
		}
	}
	
	return result
}

// ValidateProjectContext validates a ProjectContext
func (v *DataValidator) ValidateProjectContext(context *ProjectContext) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if context == nil {
		result.AddError("context", "project context cannot be nil", nil)
		return result
	}
	
	// Validate working directory
	if strings.TrimSpace(context.WorkingDirectory) == "" {
		result.AddError("working_directory", "working directory cannot be empty", context.WorkingDirectory)
	}
	
	// Validate project config if present
	if context.ProjectConfig != nil {
		configResult := v.ValidateGOAIConfig(context.ProjectConfig)
		for _, err := range configResult.Errors {
			result.AddError("project_config."+err.Field, err.Message, err.Value)
		}
	}
	
	return result
}

// ValidateCodingStyle validates a CodingStyle
func (v *DataValidator) ValidateCodingStyle(style *CodingStyle) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if style == nil {
		result.AddError("style", "coding style cannot be nil", nil)
		return result
	}
	
	// Validate indent size
	if style.IndentSize < 1 || style.IndentSize > 8 {
		result.AddError("indent_size", "indent size must be between 1 and 8", style.IndentSize)
	}
	
	// Validate max line length
	if style.MaxLineLength < 80 || style.MaxLineLength > 200 {
		result.AddError("max_line_length", "max line length must be between 80 and 200", style.MaxLineLength)
	}
	
	return result
}

// ValidateTimeline validates a Timeline
func (v *DataValidator) ValidateTimeline(timeline *Timeline) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if timeline == nil {
		result.AddError("timeline", "timeline cannot be nil", nil)
		return result
	}
	
	// Validate start and end times
	if timeline.EstimatedEnd.Before(timeline.StartTime) {
		result.AddError("estimated_end", "end time cannot be before start time", timeline.EstimatedEnd)
	}
	
	// Validate milestones
	for i, milestone := range timeline.Milestones {
		if strings.TrimSpace(milestone.Name) == "" {
			result.AddError(fmt.Sprintf("milestones[%d].name", i), 
				"milestone name cannot be empty", milestone.Name)
		}
		
		if milestone.DueDate.Before(timeline.StartTime) {
			result.AddError(fmt.Sprintf("milestones[%d].due_date", i), 
				"milestone due date cannot be before timeline start", milestone.DueDate)
		}
	}
	
	return result
}

// ValidateTestStrategy validates a TestStrategy
func (v *DataValidator) ValidateTestStrategy(strategy *TestStrategy) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if strategy == nil {
		result.AddError("strategy", "test strategy cannot be nil", nil)
		return result
	}
	
	// Validate coverage target
	if strategy.CoverageTarget < 0 || strategy.CoverageTarget > 100 {
		result.AddError("coverage_target", "coverage target must be between 0 and 100", strategy.CoverageTarget)
	}
	
	// Validate at least one test type is enabled
	if !strategy.UnitTests && !strategy.IntegrationTests && !strategy.EndToEndTests {
		result.AddError("test_types", "at least one test type must be enabled", nil)
	}
	
	return result
}

// ValidateGOAIConfig validates a GOAIConfig
func (v *DataValidator) ValidateGOAIConfig(config *GOAIConfig) *ValidationResult {
	result := &ValidationResult{Valid: true}
	
	if config == nil {
		result.AddError("config", "GOAI config cannot be nil", nil)
		return result
	}
	
	// Validate project name
	if strings.TrimSpace(config.ProjectName) == "" {
		result.AddError("project_name", "project name cannot be empty", config.ProjectName)
	}
	
	// Validate project name format (alphanumeric, dashes, underscores)
	projectNameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !projectNameRegex.MatchString(config.ProjectName) {
		result.AddError("project_name", 
			"project name can only contain letters, numbers, dashes, and underscores", 
			config.ProjectName)
	}
	
	// Validate language
	supportedLanguages := map[string]bool{
		"go":         true,
		"javascript": true,
		"typescript": true,
		"python":     true,
		"java":       true,
		"rust":       true,
		"cpp":        true,
		"c":          true,
	}
	
	if config.Language != "" && !supportedLanguages[strings.ToLower(config.Language)] {
		result.AddError("language", "unsupported language", config.Language)
	}
	
	return result
}

// ValidateString validates a string field with various constraints
func ValidateString(value, fieldName string, required bool, minLen, maxLen int) *ValidationError {
	if required && strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: "field is required",
			Value:   value,
		}
	}
	
	if len(value) < minLen {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("minimum length is %d", minLen),
			Value:   len(value),
		}
	}
	
	if maxLen > 0 && len(value) > maxLen {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("maximum length is %d", maxLen),
			Value:   len(value),
		}
	}
	
	return nil
}

// ValidateEmail validates an email address format
func ValidateEmail(email, fieldName string) *ValidationError {
	if email == "" {
		return nil // Allow empty emails
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return &ValidationError{
			Field:   fieldName,
			Message: "invalid email format",
			Value:   email,
		}
	}
	
	return nil
}

// ValidateURL validates a URL format
func ValidateURL(url, fieldName string) *ValidationError {
	if url == "" {
		return nil // Allow empty URLs
	}
	
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return &ValidationError{
			Field:   fieldName,
			Message: "invalid URL format",
			Value:   url,
		}
	}
	
	return nil
}