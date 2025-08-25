package types

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// TestToJSON tests JSON serialization utility
func TestToJSON(t *testing.T) {
	// Test simple struct
	data := struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}{
		Name:  "test",
		Value: 42,
	}
	
	jsonStr, err := ToJSON(data)
	if err != nil {
		t.Errorf("ToJSON failed: %v", err)
	}
	
	if jsonStr == "" {
		t.Errorf("JSON string should not be empty")
	}
	
	// Verify it contains expected data
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Errorf("Failed to parse generated JSON: %v", err)
	}
	
	if result["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", result["name"])
	}
	
	if result["value"] != float64(42) {
		t.Errorf("Expected value 42, got %v", result["value"])
	}
}

// TestFromJSON tests JSON deserialization utility
func TestFromJSON(t *testing.T) {
	jsonStr := `{"name": "test", "value": 42}`
	
	var result struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	
	err := FromJSON(jsonStr, &result)
	if err != nil {
		t.Errorf("FromJSON failed: %v", err)
	}
	
	if result.Name != "test" {
		t.Errorf("Expected name 'test', got %s", result.Name)
	}
	
	if result.Value != 42 {
		t.Errorf("Expected value 42, got %d", result.Value)
	}
	
	// Test invalid JSON
	err = FromJSON("invalid json", &result)
	if err == nil {
		t.Errorf("Should have error for invalid JSON")
	}
}

// TestProblemRequestBuilder tests the fluent builder for ProblemRequest
func TestProblemRequestBuilder(t *testing.T) {
	builder := NewProblemRequestBuilder()
	
	if builder == nil {
		t.Fatalf("Builder should not be nil")
	}
	
	context := &ProjectContext{WorkingDirectory: "/test"}
	style := &CodingStyle{IndentSize: 4}
	
	request := builder.
		WithDescription("Build a REST API").
		AddRequirement("Authentication").
		AddRequirement("Rate limiting").
		AddRequirements("Logging", "Monitoring").
		AddConstraint("Use Go").
		AddConstraints("PostgreSQL", "Docker").
		WithContext(context).
		WithPreferredStyle(style).
		Build()
	
	if request.Description != "Build a REST API" {
		t.Errorf("Expected description 'Build a REST API', got %s", request.Description)
	}
	
	if len(request.Requirements) != 4 {
		t.Errorf("Expected 4 requirements, got %d", len(request.Requirements))
	}
	
	expectedReqs := []string{"Authentication", "Rate limiting", "Logging", "Monitoring"}
	for i, expected := range expectedReqs {
		if request.Requirements[i] != expected {
			t.Errorf("Expected requirement[%d] '%s', got '%s'", i, expected, request.Requirements[i])
		}
	}
	
	if len(request.Constraints) != 3 {
		t.Errorf("Expected 3 constraints, got %d", len(request.Constraints))
	}
	
	if request.Context != context {
		t.Errorf("Context should be set")
	}
	
	if request.PreferredStyle != style {
		t.Errorf("PreferredStyle should be set")
	}
	
	// Test empty strings are filtered out
	builder2 := NewProblemRequestBuilder()
	request2 := builder2.
		AddRequirement("").
		AddRequirement("valid").
		AddConstraint("  ").
		AddConstraint("valid constraint").
		Build()
	
	if len(request2.Requirements) != 1 {
		t.Errorf("Empty requirements should be filtered out, got %d", len(request2.Requirements))
	}
	
	if len(request2.Constraints) != 1 {
		t.Errorf("Empty constraints should be filtered out, got %d", len(request2.Constraints))
	}
}

// TestAnalysisBuilder tests the fluent builder for Analysis
func TestAnalysisBuilder(t *testing.T) {
	builder := NewAnalysisBuilder()
	
	analysis := builder.
		WithProblemDomain("web-development").
		WithComplexity(ComplexityHigh).
		WithArchitecturePattern("Microservices").
		AddTechnology("Go").
		AddTechnology("PostgreSQL").
		AddTechnologies("Docker", "Kubernetes", "Redis").
		AddRiskFactor("performance", "High load", "high", "Use caching").
		AddRecommendation("architecture", "Use event sourcing", "medium", "high").
		Build()
	
	if analysis.ProblemDomain != "web-development" {
		t.Errorf("Expected problem domain 'web-development', got %s", analysis.ProblemDomain)
	}
	
	if analysis.Complexity != ComplexityHigh {
		t.Errorf("Expected complexity high, got %s", analysis.Complexity)
	}
	
	if analysis.ArchitecturePattern != "Microservices" {
		t.Errorf("Expected pattern 'Microservices', got %s", analysis.ArchitecturePattern)
	}
	
	expectedTech := []string{"Go", "PostgreSQL", "Docker", "Kubernetes", "Redis"}
	if len(analysis.TechnicalStack) != len(expectedTech) {
		t.Errorf("Expected %d technologies, got %d", len(expectedTech), len(analysis.TechnicalStack))
	}
	
	for i, expected := range expectedTech {
		if analysis.TechnicalStack[i] != expected {
			t.Errorf("Expected tech[%d] '%s', got '%s'", i, expected, analysis.TechnicalStack[i])
		}
	}
	
	if len(analysis.RiskFactors) != 1 {
		t.Errorf("Expected 1 risk factor, got %d", len(analysis.RiskFactors))
	}
	
	if analysis.RiskFactors[0].Type != "performance" {
		t.Errorf("Expected risk type 'performance', got %s", analysis.RiskFactors[0].Type)
	}
	
	if len(analysis.Recommendations) != 1 {
		t.Errorf("Expected 1 recommendation, got %d", len(analysis.Recommendations))
	}
	
	if analysis.Recommendations[0].Category != "architecture" {
		t.Errorf("Expected rec category 'architecture', got %s", analysis.Recommendations[0].Category)
	}
}

// TestExecutionPlanBuilder tests the fluent builder for ExecutionPlan
func TestExecutionPlanBuilder(t *testing.T) {
	builder := NewExecutionPlanBuilder()
	now := time.Now()
	
	plan := builder.
		AddStep("step-1", "Setup", "Initialize project", time.Hour*2, 1).
		AddStepWithDependencies("step-2", "Implementation", "Build features", time.Hour*8, 2, []string{"step-1"}).
		AddDependency("gin", "v1.9.1", "library", true, "Web framework").
		WithTimeline(now, now.Add(time.Hour*24), time.Hour*24).
		AddMilestone("Alpha", "First release", now.Add(time.Hour*12)).
		WithTestStrategy(true, true, false, []string{"testing", "testify"}, 85.0).
		Build()
	
	if len(plan.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(plan.Steps))
	}
	
	step1 := plan.Steps[0]
	if step1.ID != "step-1" {
		t.Errorf("Expected step ID 'step-1', got %s", step1.ID)
	}
	
	if step1.Name != "Setup" {
		t.Errorf("Expected step name 'Setup', got %s", step1.Name)
	}
	
	if step1.EstimatedTime != time.Hour*2 {
		t.Errorf("Expected estimated time 2h, got %v", step1.EstimatedTime)
	}
	
	step2 := plan.Steps[1]
	if len(step2.Dependencies) != 1 || step2.Dependencies[0] != "step-1" {
		t.Errorf("Expected step-2 to depend on step-1, got %v", step2.Dependencies)
	}
	
	if len(plan.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(plan.Dependencies))
	}
	
	dep := plan.Dependencies[0]
	if dep.Name != "gin" || dep.Version != "v1.9.1" {
		t.Errorf("Expected gin v1.9.1, got %s %s", dep.Name, dep.Version)
	}
	
	if plan.Timeline == nil {
		t.Errorf("Timeline should be set")
	}
	
	if len(plan.Timeline.Milestones) != 1 {
		t.Errorf("Expected 1 milestone, got %d", len(plan.Timeline.Milestones))
	}
	
	if plan.Timeline.Milestones[0].Name != "Alpha" {
		t.Errorf("Expected milestone 'Alpha', got %s", plan.Timeline.Milestones[0].Name)
	}
	
	if plan.TestStrategy == nil {
		t.Errorf("Test strategy should be set")
	}
	
	if !plan.TestStrategy.UnitTests {
		t.Errorf("Unit tests should be enabled")
	}
	
	if plan.TestStrategy.CoverageTarget != 85.0 {
		t.Errorf("Expected coverage 85.0, got %f", plan.TestStrategy.CoverageTarget)
	}
}

// TestProjectContextBuilder tests the fluent builder for ProjectContext
func TestProjectContextBuilder(t *testing.T) {
	builder := NewProjectContextBuilder()
	now := time.Now()
	
	config := &GOAIConfig{ProjectName: "test"}
	structure := &ProjectStructure{RootPath: "/root"}
	
	context := builder.
		WithWorkingDirectory("/project/root").
		WithProjectConfig(config).
		WithProjectStructure(structure).
		AddGitChange("main.go", "modified", "author", "commit msg", "diff", now).
		AddOpenFile("/project/main.go", "main.go", ".go", 1024, now, true).
		Build()
	
	if context.WorkingDirectory != "/project/root" {
		t.Errorf("Expected working dir '/project/root', got %s", context.WorkingDirectory)
	}
	
	if context.ProjectConfig != config {
		t.Errorf("Project config should be set")
	}
	
	if context.ProjectStructure != structure {
		t.Errorf("Project structure should be set")
	}
	
	if len(context.RecentChanges) != 1 {
		t.Errorf("Expected 1 git change, got %d", len(context.RecentChanges))
	}
	
	change := context.RecentChanges[0]
	if change.FilePath != "main.go" {
		t.Errorf("Expected file path 'main.go', got %s", change.FilePath)
	}
	
	if len(context.OpenFiles) != 1 {
		t.Errorf("Expected 1 open file, got %d", len(context.OpenFiles))
	}
	
	file := context.OpenFiles[0]
	if file.Name != "main.go" {
		t.Errorf("Expected file name 'main.go', got %s", file.Name)
	}
	
	if !file.IsOpen {
		t.Errorf("File should be marked as open")
	}
}

// TestMergeAnalyses tests analysis merging utility
func TestMergeAnalyses(t *testing.T) {
	// Test empty analyses
	merged := MergeAnalyses()
	if merged == nil {
		t.Errorf("Merged analysis should not be nil")
	}
	
	// Test single analysis
	analysis1 := &Analysis{
		ProblemDomain:       "web",
		TechnicalStack:      []string{"Go", "PostgreSQL"},
		ArchitecturePattern: "REST",
		Complexity:          ComplexityMedium,
		RiskFactors: []RiskFactor{
			{Type: "performance", Description: "High load"},
		},
	}
	
	merged = MergeAnalyses(analysis1)
	if merged.ProblemDomain != analysis1.ProblemDomain {
		t.Errorf("Expected domain %s, got %s", analysis1.ProblemDomain, merged.ProblemDomain)
	}
	
	// Test multiple analyses
	analysis2 := &Analysis{
		TechnicalStack: []string{"Docker", "Go"}, // Go is duplicate
		RiskFactors: []RiskFactor{
			{Type: "security", Description: "Auth issues"},
		},
	}
	
	merged = MergeAnalyses(analysis1, analysis2)
	
	// Should have unique technologies
	expectedTech := []string{"Go", "PostgreSQL", "Docker"}
	if len(merged.TechnicalStack) != len(expectedTech) {
		t.Errorf("Expected %d unique technologies, got %d", len(expectedTech), len(merged.TechnicalStack))
	}
	
	// Should have combined risk factors
	if len(merged.RiskFactors) != 2 {
		t.Errorf("Expected 2 risk factors, got %d", len(merged.RiskFactors))
	}
	
	// Test with nil analysis
	merged = MergeAnalyses(analysis1, nil, analysis2)
	if len(merged.TechnicalStack) != 3 {
		t.Errorf("Nil analysis should be ignored, expected 3 tech items, got %d", len(merged.TechnicalStack))
	}
}

// TestSplitExecutionPlan tests execution plan splitting utility
func TestSplitExecutionPlan(t *testing.T) {
	// Test nil plan
	phases := SplitExecutionPlan(nil, 2)
	if len(phases) != 1 || phases[0] != nil {
		t.Errorf("Nil plan should return [nil]")
	}
	
	// Test plan with few steps
	plan := &ExecutionPlan{
		Steps: []PlanStep{
			{ID: "1", Name: "Step 1"},
			{ID: "2", Name: "Step 2"},
		},
		Dependencies: []Dependency{{Name: "dep1"}},
	}
	
	phases = SplitExecutionPlan(plan, 5)
	if len(phases) != 1 {
		t.Errorf("Small plan should not be split, got %d phases", len(phases))
	}
	
	if phases[0] != plan {
		t.Errorf("Small plan should return original plan")
	}
	
	// Test plan with many steps
	largeSteps := make([]PlanStep, 7)
	for i := range largeSteps {
		largeSteps[i] = PlanStep{ID: string(rune('1' + i)), Name: "Step"}
	}
	
	largePlan := &ExecutionPlan{
		Steps:        largeSteps,
		Dependencies: []Dependency{{Name: "dep1"}},
	}
	
	phases = SplitExecutionPlan(largePlan, 3)
	if len(phases) != 3 {
		t.Errorf("Expected 3 phases for 7 steps with max 3, got %d", len(phases))
	}
	
	// Check phase sizes
	if len(phases[0].Steps) != 3 {
		t.Errorf("Expected first phase to have 3 steps, got %d", len(phases[0].Steps))
	}
	
	if len(phases[1].Steps) != 3 {
		t.Errorf("Expected second phase to have 3 steps, got %d", len(phases[1].Steps))
	}
	
	if len(phases[2].Steps) != 1 {
		t.Errorf("Expected third phase to have 1 step, got %d", len(phases[2].Steps))
	}
	
	// Check dependencies are shared
	for i, phase := range phases {
		if len(phase.Dependencies) != 1 || phase.Dependencies[0].Name != "dep1" {
			t.Errorf("Phase %d should have shared dependencies", i)
		}
	}
}

// TestExtractComplexityMetrics tests complexity metrics extraction
func TestExtractComplexityMetrics(t *testing.T) {
	request := &ProblemRequest{
		Description:  "Build API",
		Requirements: []string{"auth", "logging"},
		Constraints:  []string{"Go"},
	}
	
	analysis := &Analysis{
		TechnicalStack:  []string{"Go", "PostgreSQL"},
		RiskFactors:     []RiskFactor{{Type: "perf"}},
		Recommendations: []Recommendation{{Category: "arch"}},
		Complexity:      ComplexityHigh,
	}
	
	plan := &ExecutionPlan{
		Steps: []PlanStep{
			{EstimatedTime: time.Hour * 2, Priority: 1},
			{EstimatedTime: time.Hour * 4, Priority: 2},
		},
		Dependencies: []Dependency{{Name: "gin"}},
	}
	
	metrics := ExtractComplexityMetrics(request, analysis, plan)
	
	if metrics["requirements_count"] != 2 {
		t.Errorf("Expected 2 requirements, got %v", metrics["requirements_count"])
	}
	
	if metrics["technical_stack_size"] != 2 {
		t.Errorf("Expected 2 tech items, got %v", metrics["technical_stack_size"])
	}
	
	if metrics["complexity_level"] != "high" {
		t.Errorf("Expected complexity 'high', got %v", metrics["complexity_level"])
	}
	
	if metrics["steps_count"] != 2 {
		t.Errorf("Expected 2 steps, got %v", metrics["steps_count"])
	}
	
	if metrics["total_estimated_hours"] != 6.0 {
		t.Errorf("Expected 6.0 hours, got %v", metrics["total_estimated_hours"])
	}
	
	if metrics["average_step_priority"] != 1.5 {
		t.Errorf("Expected avg priority 1.5, got %v", metrics["average_step_priority"])
	}
	
	// Test with nil inputs
	metrics = ExtractComplexityMetrics(nil, nil, nil)
	if len(metrics) != 0 {
		t.Errorf("Empty metrics expected for nil inputs, got %d items", len(metrics))
	}
}

// TestConvertToMap tests struct to map conversion
func TestConvertToMap(t *testing.T) {
	data := struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
		Flag  bool   `json:"flag"`
	}{
		Name:  "test",
		Value: 42,
		Flag:  true,
	}
	
	result, err := ConvertToMap(data)
	if err != nil {
		t.Errorf("ConvertToMap failed: %v", err)
	}
	
	if result["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", result["name"])
	}
	
	if result["value"] != float64(42) {
		t.Errorf("Expected value 42, got %v", result["value"])
	}
	
	if result["flag"] != true {
		t.Errorf("Expected flag true, got %v", result["flag"])
	}
}

// TestDeepCopy tests deep copy utility
func TestDeepCopy(t *testing.T) {
	original := struct {
		Name   string   `json:"name"`
		Values []int    `json:"values"`
		Nested struct { 
			Field string `json:"field"`
		} `json:"nested"`
	}{
		Name:   "original",
		Values: []int{1, 2, 3},
		Nested: struct {
			Field string `json:"field"`
		}{Field: "nested_value"},
	}
	
	var copy struct {
		Name   string   `json:"name"`
		Values []int    `json:"values"`
		Nested struct { 
			Field string `json:"field"`
		} `json:"nested"`
	}
	
	err := DeepCopy(original, &copy)
	if err != nil {
		t.Errorf("DeepCopy failed: %v", err)
	}
	
	if copy.Name != original.Name {
		t.Errorf("Expected name %s, got %s", original.Name, copy.Name)
	}
	
	if !reflect.DeepEqual(copy.Values, original.Values) {
		t.Errorf("Values should be equal: %v vs %v", copy.Values, original.Values)
	}
	
	if copy.Nested.Field != original.Nested.Field {
		t.Errorf("Expected nested field %s, got %s", original.Nested.Field, copy.Nested.Field)
	}
	
	// Verify it's a deep copy (modifying copy shouldn't affect original)
	copy.Name = "modified"
	copy.Values[0] = 999
	
	if original.Name == "modified" {
		t.Errorf("Original should not be affected by copy modification")
	}
	
	if original.Values[0] == 999 {
		t.Errorf("Original values should not be affected by copy modification")
	}
}