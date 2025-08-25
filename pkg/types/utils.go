package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ToJSON converts any data structure to JSON string
func ToJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(bytes), nil
}

// FromJSON converts JSON string to a data structure
func FromJSON(jsonStr string, target interface{}) error {
	if err := json.Unmarshal([]byte(jsonStr), target); err != nil {
		return fmt.Errorf("failed to unmarshal from JSON: %w", err)
	}
	return nil
}

// ProblemRequestBuilder provides a fluent interface for building ProblemRequest
type ProblemRequestBuilder struct {
	request *ProblemRequest
}

// NewProblemRequestBuilder creates a new builder
func NewProblemRequestBuilder() *ProblemRequestBuilder {
	return &ProblemRequestBuilder{
		request: &ProblemRequest{
			Requirements: make([]string, 0),
			Constraints:  make([]string, 0),
		},
	}
}

// WithDescription sets the description
func (b *ProblemRequestBuilder) WithDescription(description string) *ProblemRequestBuilder {
	b.request.Description = strings.TrimSpace(description)
	return b
}

// AddRequirement adds a requirement
func (b *ProblemRequestBuilder) AddRequirement(requirement string) *ProblemRequestBuilder {
	if req := strings.TrimSpace(requirement); req != "" {
		b.request.Requirements = append(b.request.Requirements, req)
	}
	return b
}

// AddRequirements adds multiple requirements
func (b *ProblemRequestBuilder) AddRequirements(requirements ...string) *ProblemRequestBuilder {
	for _, req := range requirements {
		b.AddRequirement(req)
	}
	return b
}

// AddConstraint adds a constraint
func (b *ProblemRequestBuilder) AddConstraint(constraint string) *ProblemRequestBuilder {
	if cons := strings.TrimSpace(constraint); cons != "" {
		b.request.Constraints = append(b.request.Constraints, cons)
	}
	return b
}

// AddConstraints adds multiple constraints
func (b *ProblemRequestBuilder) AddConstraints(constraints ...string) *ProblemRequestBuilder {
	for _, cons := range constraints {
		b.AddConstraint(cons)
	}
	return b
}

// WithContext sets the project context
func (b *ProblemRequestBuilder) WithContext(context *ProjectContext) *ProblemRequestBuilder {
	b.request.Context = context
	return b
}

// WithPreferredStyle sets the preferred coding style
func (b *ProblemRequestBuilder) WithPreferredStyle(style *CodingStyle) *ProblemRequestBuilder {
	b.request.PreferredStyle = style
	return b
}

// Build returns the constructed ProblemRequest
func (b *ProblemRequestBuilder) Build() *ProblemRequest {
	return b.request
}

// AnalysisBuilder provides a fluent interface for building Analysis
type AnalysisBuilder struct {
	analysis *Analysis
}

// NewAnalysisBuilder creates a new builder
func NewAnalysisBuilder() *AnalysisBuilder {
	return &AnalysisBuilder{
		analysis: &Analysis{
			TechnicalStack:  make([]string, 0),
			RiskFactors:     make([]RiskFactor, 0),
			Recommendations: make([]Recommendation, 0),
		},
	}
}

// WithProblemDomain sets the problem domain
func (b *AnalysisBuilder) WithProblemDomain(domain string) *AnalysisBuilder {
	b.analysis.ProblemDomain = strings.TrimSpace(domain)
	return b
}

// WithComplexity sets the complexity level
func (b *AnalysisBuilder) WithComplexity(complexity ComplexityLevel) *AnalysisBuilder {
	b.analysis.Complexity = complexity
	return b
}

// WithArchitecturePattern sets the architecture pattern
func (b *AnalysisBuilder) WithArchitecturePattern(pattern string) *AnalysisBuilder {
	b.analysis.ArchitecturePattern = strings.TrimSpace(pattern)
	return b
}

// AddTechnology adds a technology to the stack
func (b *AnalysisBuilder) AddTechnology(tech string) *AnalysisBuilder {
	if technology := strings.TrimSpace(tech); technology != "" {
		b.analysis.TechnicalStack = append(b.analysis.TechnicalStack, technology)
	}
	return b
}

// AddTechnologies adds multiple technologies
func (b *AnalysisBuilder) AddTechnologies(techs ...string) *AnalysisBuilder {
	for _, tech := range techs {
		b.AddTechnology(tech)
	}
	return b
}

// AddRiskFactor adds a risk factor
func (b *AnalysisBuilder) AddRiskFactor(riskType, description, severity, mitigation string) *AnalysisBuilder {
	b.analysis.RiskFactors = append(b.analysis.RiskFactors, RiskFactor{
		Type:        strings.TrimSpace(riskType),
		Description: strings.TrimSpace(description),
		Severity:    strings.TrimSpace(severity),
		Mitigation:  strings.TrimSpace(mitigation),
	})
	return b
}

// AddRecommendation adds a recommendation
func (b *AnalysisBuilder) AddRecommendation(category, description, priority, impact string) *AnalysisBuilder {
	b.analysis.Recommendations = append(b.analysis.Recommendations, Recommendation{
		Category:    strings.TrimSpace(category),
		Description: strings.TrimSpace(description),
		Priority:    strings.TrimSpace(priority),
		Impact:      strings.TrimSpace(impact),
	})
	return b
}

// Build returns the constructed Analysis
func (b *AnalysisBuilder) Build() *Analysis {
	return b.analysis
}

// ExecutionPlanBuilder provides a fluent interface for building ExecutionPlan
type ExecutionPlanBuilder struct {
	plan *ExecutionPlan
}

// NewExecutionPlanBuilder creates a new builder
func NewExecutionPlanBuilder() *ExecutionPlanBuilder {
	return &ExecutionPlanBuilder{
		plan: &ExecutionPlan{
			Steps:        make([]PlanStep, 0),
			Dependencies: make([]Dependency, 0),
		},
	}
}

// AddStep adds a step to the plan
func (b *ExecutionPlanBuilder) AddStep(id, name, description string, estimatedTime time.Duration, priority int) *ExecutionPlanBuilder {
	b.plan.Steps = append(b.plan.Steps, PlanStep{
		ID:            strings.TrimSpace(id),
		Name:          strings.TrimSpace(name),
		Description:   strings.TrimSpace(description),
		Dependencies:  make([]string, 0),
		EstimatedTime: estimatedTime,
		Priority:      priority,
	})
	return b
}

// AddStepWithDependencies adds a step with dependencies
func (b *ExecutionPlanBuilder) AddStepWithDependencies(id, name, description string, estimatedTime time.Duration, priority int, dependencies []string) *ExecutionPlanBuilder {
	cleanDeps := make([]string, 0, len(dependencies))
	for _, dep := range dependencies {
		if cleanDep := strings.TrimSpace(dep); cleanDep != "" {
			cleanDeps = append(cleanDeps, cleanDep)
		}
	}
	
	b.plan.Steps = append(b.plan.Steps, PlanStep{
		ID:            strings.TrimSpace(id),
		Name:          strings.TrimSpace(name),
		Description:   strings.TrimSpace(description),
		Dependencies:  cleanDeps,
		EstimatedTime: estimatedTime,
		Priority:      priority,
	})
	return b
}

// AddDependency adds a dependency to the plan
func (b *ExecutionPlanBuilder) AddDependency(name, version, depType string, required bool, description string) *ExecutionPlanBuilder {
	b.plan.Dependencies = append(b.plan.Dependencies, Dependency{
		Name:        strings.TrimSpace(name),
		Version:     strings.TrimSpace(version),
		Type:        strings.TrimSpace(depType),
		Required:    required,
		Description: strings.TrimSpace(description),
	})
	return b
}

// WithTimeline sets the timeline
func (b *ExecutionPlanBuilder) WithTimeline(startTime, estimatedEnd time.Time, totalDuration time.Duration) *ExecutionPlanBuilder {
	b.plan.Timeline = &Timeline{
		StartTime:     startTime,
		EstimatedEnd:  estimatedEnd,
		TotalDuration: totalDuration,
		Milestones:    make([]Milestone, 0),
	}
	return b
}

// AddMilestone adds a milestone to the timeline
func (b *ExecutionPlanBuilder) AddMilestone(name, description string, dueDate time.Time) *ExecutionPlanBuilder {
	if b.plan.Timeline == nil {
		b.plan.Timeline = &Timeline{Milestones: make([]Milestone, 0)}
	}
	
	b.plan.Timeline.Milestones = append(b.plan.Timeline.Milestones, Milestone{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		DueDate:     dueDate,
		Completed:   false,
	})
	return b
}

// WithTestStrategy sets the test strategy
func (b *ExecutionPlanBuilder) WithTestStrategy(unitTests, integrationTests, e2eTests bool, frameworks []string, coverageTarget float64) *ExecutionPlanBuilder {
	cleanFrameworks := make([]string, 0, len(frameworks))
	for _, framework := range frameworks {
		if f := strings.TrimSpace(framework); f != "" {
			cleanFrameworks = append(cleanFrameworks, f)
		}
	}
	
	b.plan.TestStrategy = &TestStrategy{
		UnitTests:        unitTests,
		IntegrationTests: integrationTests,
		EndToEndTests:    e2eTests,
		TestFrameworks:   cleanFrameworks,
		CoverageTarget:   coverageTarget,
	}
	return b
}

// Build returns the constructed ExecutionPlan
func (b *ExecutionPlanBuilder) Build() *ExecutionPlan {
	return b.plan
}

// ProjectContextBuilder provides a fluent interface for building ProjectContext
type ProjectContextBuilder struct {
	context *ProjectContext
}

// NewProjectContextBuilder creates a new builder
func NewProjectContextBuilder() *ProjectContextBuilder {
	return &ProjectContextBuilder{
		context: &ProjectContext{
			RecentChanges: make([]*GitChange, 0),
			Dependencies:  make([]*Dependency, 0),
			OpenFiles:     make([]*FileInfo, 0),
		},
	}
}

// WithWorkingDirectory sets the working directory
func (b *ProjectContextBuilder) WithWorkingDirectory(workdir string) *ProjectContextBuilder {
	b.context.WorkingDirectory = strings.TrimSpace(workdir)
	return b
}

// WithProjectConfig sets the project configuration
func (b *ProjectContextBuilder) WithProjectConfig(config *GOAIConfig) *ProjectContextBuilder {
	b.context.ProjectConfig = config
	return b
}

// WithProjectStructure sets the project structure
func (b *ProjectContextBuilder) WithProjectStructure(structure *ProjectStructure) *ProjectContextBuilder {
	b.context.ProjectStructure = structure
	return b
}

// AddGitChange adds a git change
func (b *ProjectContextBuilder) AddGitChange(filePath, changeType, author, message, diff string, timestamp time.Time) *ProjectContextBuilder {
	b.context.RecentChanges = append(b.context.RecentChanges, &GitChange{
		FilePath:   strings.TrimSpace(filePath),
		ChangeType: strings.TrimSpace(changeType),
		Timestamp:  timestamp,
		Author:     strings.TrimSpace(author),
		Message:    strings.TrimSpace(message),
		Diff:       diff,
	})
	return b
}

// AddOpenFile adds an open file
func (b *ProjectContextBuilder) AddOpenFile(path, name, extension string, size int64, modTime time.Time, isOpen bool) *ProjectContextBuilder {
	b.context.OpenFiles = append(b.context.OpenFiles, &FileInfo{
		Path:         strings.TrimSpace(path),
		Name:         strings.TrimSpace(name),
		Extension:    strings.TrimSpace(extension),
		Size:         size,
		ModifiedTime: modTime,
		IsOpen:       isOpen,
	})
	return b
}

// Build returns the constructed ProjectContext
func (b *ProjectContextBuilder) Build() *ProjectContext {
	return b.context
}

// Data transformation utilities

// MergeAnalyses combines multiple analyses into one
func MergeAnalyses(analyses ...*Analysis) *Analysis {
	if len(analyses) == 0 {
		return &Analysis{}
	}
	
	merged := &Analysis{
		TechnicalStack:  make([]string, 0),
		RiskFactors:     make([]RiskFactor, 0),
		Recommendations: make([]Recommendation, 0),
	}
	
	// Use first analysis as base
	if len(analyses) > 0 {
		first := analyses[0]
		merged.ProblemDomain = first.ProblemDomain
		merged.ArchitecturePattern = first.ArchitecturePattern
		merged.Complexity = first.Complexity
	}
	
	// Merge technical stacks (deduplicated)
	techSet := make(map[string]bool)
	for _, analysis := range analyses {
		if analysis == nil {
			continue
		}
		for _, tech := range analysis.TechnicalStack {
			if !techSet[tech] {
				merged.TechnicalStack = append(merged.TechnicalStack, tech)
				techSet[tech] = true
			}
		}
		
		// Merge risk factors and recommendations
		merged.RiskFactors = append(merged.RiskFactors, analysis.RiskFactors...)
		merged.Recommendations = append(merged.Recommendations, analysis.Recommendations...)
	}
	
	return merged
}

// SplitExecutionPlan splits a large execution plan into phases
func SplitExecutionPlan(plan *ExecutionPlan, maxStepsPerPhase int) []*ExecutionPlan {
	if plan == nil || len(plan.Steps) <= maxStepsPerPhase {
		return []*ExecutionPlan{plan}
	}
	
	phases := make([]*ExecutionPlan, 0)
	
	for i := 0; i < len(plan.Steps); i += maxStepsPerPhase {
		end := i + maxStepsPerPhase
		if end > len(plan.Steps) {
			end = len(plan.Steps)
		}
		
		phase := &ExecutionPlan{
			Steps:        plan.Steps[i:end],
			Dependencies: plan.Dependencies, // Share dependencies across phases
			Timeline:     plan.Timeline,     // Share timeline
			TestStrategy: plan.TestStrategy, // Share test strategy
		}
		
		phases = append(phases, phase)
	}
	
	return phases
}

// ExtractComplexityMetrics calculates complexity metrics from various data
func ExtractComplexityMetrics(request *ProblemRequest, analysis *Analysis, plan *ExecutionPlan) map[string]interface{} {
	metrics := make(map[string]interface{})
	
	if request != nil {
		metrics["requirements_count"] = len(request.Requirements)
		metrics["constraints_count"] = len(request.Constraints)
		metrics["description_length"] = len(request.Description)
	}
	
	if analysis != nil {
		metrics["technical_stack_size"] = len(analysis.TechnicalStack)
		metrics["risk_factors_count"] = len(analysis.RiskFactors)
		metrics["recommendations_count"] = len(analysis.Recommendations)
		metrics["complexity_level"] = string(analysis.Complexity)
	}
	
	if plan != nil {
		metrics["steps_count"] = len(plan.Steps)
		metrics["dependencies_count"] = len(plan.Dependencies)
		
		// Calculate total estimated time
		var totalTime time.Duration
		for _, step := range plan.Steps {
			totalTime += step.EstimatedTime
		}
		metrics["total_estimated_hours"] = totalTime.Hours()
		
		// Calculate average step complexity
		if len(plan.Steps) > 0 {
			avgPriority := 0
			for _, step := range plan.Steps {
				avgPriority += step.Priority
			}
			metrics["average_step_priority"] = float64(avgPriority) / float64(len(plan.Steps))
		}
	}
	
	return metrics
}

// ConvertToMap converts any struct to map[string]interface{}
func ConvertToMap(data interface{}) (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}
	
	return result, nil
}

// DeepCopy creates a deep copy of any serializable data structure
func DeepCopy(src, dst interface{}) error {
	jsonBytes, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("failed to marshal source: %w", err)
	}
	
	if err := json.Unmarshal(jsonBytes, dst); err != nil {
		return fmt.Errorf("failed to unmarshal to destination: %w", err)
	}
	
	return nil
}