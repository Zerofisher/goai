package types

import "time"

// ProblemRequest represents a user's request for assistance
type ProblemRequest struct {
	Description     string            `json:"description"`
	Context         *ProjectContext   `json:"context"`
	Requirements    []string          `json:"requirements"`
	Constraints     []string          `json:"constraints"`
	PreferredStyle  *CodingStyle      `json:"preferred_style"`
}

// Analysis represents the result of problem analysis
type Analysis struct {
	ProblemDomain       string            `json:"problem_domain"`
	TechnicalStack      []string          `json:"technical_stack"`
	ArchitecturePattern string            `json:"architecture_pattern"`
	RiskFactors         []RiskFactor      `json:"risk_factors"`
	Recommendations     []Recommendation  `json:"recommendations"`
	Complexity          ComplexityLevel   `json:"complexity"`
}

// ExecutionPlan represents a detailed plan for implementation
type ExecutionPlan struct {
	Steps           []PlanStep        `json:"steps"`
	Dependencies    []Dependency      `json:"dependencies"`
	Timeline        *Timeline         `json:"timeline"`
	TestStrategy    *TestStrategy     `json:"test_strategy"`
	ValidationRules []ValidationRule  `json:"validation_rules"`
}

// ProjectContext holds information about the current project
type ProjectContext struct {
	WorkingDirectory string           `json:"working_directory"`
	ProjectConfig    *GOAIConfig      `json:"project_config"`
	ProjectStructure *ProjectStructure `json:"project_structure"`
	RecentChanges    []*GitChange     `json:"recent_changes"`
	Dependencies     []*Dependency    `json:"dependencies"`
	OpenFiles        []*FileInfo      `json:"open_files"`
	CodingStandards  *CodingStandards `json:"coding_standards"`
	LoadedAt         time.Time        `json:"loaded_at"`
	GitInfo          *GitInfo         `json:"git_info"`
}

// CodeResult represents the output of code generation
type CodeResult struct {
	GeneratedFiles []GeneratedFile   `json:"generated_files"`
	Tests          *TestSuite       `json:"tests"`
	Documentation  *Documentation   `json:"documentation"`
	Metadata       *GenerationMeta  `json:"metadata"`
}

// ValidationReport contains validation results
type ValidationReport struct {
	StaticReport     *StaticReport     `json:"static_report"`
	TestResults      *TestResults      `json:"test_results"`
	ComplianceReport *ComplianceReport `json:"compliance_report"`
	OverallStatus    ValidationStatus  `json:"overall_status"`
}

// Supporting types
type ComplexityLevel string

const (
	ComplexityLow    ComplexityLevel = "low"
	ComplexityMedium ComplexityLevel = "medium"
	ComplexityHigh   ComplexityLevel = "high"
)

type ValidationStatus string

const (
	ValidationPassed  ValidationStatus = "passed"
	ValidationFailed  ValidationStatus = "failed"
	ValidationWarning ValidationStatus = "warning"
)

type RiskFactor struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Mitigation  string `json:"mitigation"`
}

type Recommendation struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Impact      string `json:"impact"`
}

type PlanStep struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Dependencies []string      `json:"dependencies"`
	EstimatedTime time.Duration `json:"estimated_time"`
	Priority     int           `json:"priority"`
}

type Dependency struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	Source      string `json:"source"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type Timeline struct {
	StartTime     time.Time     `json:"start_time"`
	EstimatedEnd  time.Time     `json:"estimated_end"`
	Milestones    []Milestone   `json:"milestones"`
	TotalDuration time.Duration `json:"total_duration"`
}

type Milestone struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	Completed   bool      `json:"completed"`
}

type TestStrategy struct {
	UnitTests        bool     `json:"unit_tests"`
	IntegrationTests bool     `json:"integration_tests"`
	EndToEndTests    bool     `json:"end_to_end_tests"`
	TestFrameworks   []string `json:"test_frameworks"`
	CoverageTarget   float64  `json:"coverage_target"`
}

type ValidationRule struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Parameters  map[string]interface{} `json:"parameters"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type GOAIConfig struct {
	ProjectName     string            `json:"project_name"`
	Language        string            `json:"language"`
	Framework       string            `json:"framework"`
	Description     string            `json:"description"`
	Version         string            `json:"version"`
	Author          string            `json:"author"`
	Repository      string            `json:"repository"`
	License         string            `json:"license"`
	Dependencies    []string          `json:"dependencies"`
	Exclusions      []string          `json:"exclusions"`
	CodingStyle     *CodingStyle      `json:"coding_style"`
	CodingStandards *CodingStandards  `json:"coding_standards"`
	TestConfig      *TestConfig       `json:"test_config"`
	Templates       map[string]string `json:"templates"`
	Preferences     map[string]interface{} `json:"preferences"`
}

type ProjectStructure struct {
	RootPath    string      `json:"root_path"`
	Directories []Directory `json:"directories"`
	Files       []FileInfo  `json:"files"`
	Patterns    []Pattern   `json:"patterns"`
}

type Directory struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type FileInfo struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Extension    string    `json:"extension"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
	IsOpen       bool      `json:"is_open"`
}

type Pattern struct {
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type GitChange struct {
	FilePath    string    `json:"file_path"`
	ChangeType  string    `json:"change_type"`
	CommitHash  string    `json:"commit_hash"`
	Timestamp   time.Time `json:"timestamp"`
	Author      string    `json:"author"`
	Message     string    `json:"message"`
	Diff        string    `json:"diff"`
	Additions   int       `json:"additions"`
	Deletions   int       `json:"deletions"`
}

type FileChangeEvent struct {
	FilePath  string `json:"file_path"`
	EventType string `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

type GitInfo struct {
	CurrentBranch    string    `json:"current_branch"`
	CurrentCommit    string    `json:"current_commit"`
	RemoteURLs       []string  `json:"remote_urls"`
	LastCommitTime   time.Time `json:"last_commit_time"`
	LastCommitMsg    string    `json:"last_commit_msg"`
	LastCommitAuthor string    `json:"last_commit_author"`
}

type GitStatus struct {
	IsClean bool              `json:"is_clean"`
	Files   map[string]string `json:"files"`
}

type CodingStandards struct {
	Language         string            `json:"language"`
	StyleGuide       string            `json:"style_guide"`
	LintingRules     []string          `json:"linting_rules"`
	FormattingRules  map[string]string `json:"formatting_rules"`
	NamingConventions map[string]string `json:"naming_conventions"`
}

type CodingStyle struct {
	IndentSize      int               `json:"indent_size"`
	UseSpaces       bool              `json:"use_spaces"`
	MaxLineLength   int               `json:"max_line_length"`
	NamingStyle     string            `json:"naming_style"`
	CommentStyle    string            `json:"comment_style"`
	FormatOnSave    bool              `json:"format_on_save"`
	CustomRules     map[string]string `json:"custom_rules"`
}

type TestConfig struct {
	Framework     string   `json:"framework"`
	TestPatterns  []string `json:"test_patterns"`
	CoverageGoal  float64  `json:"coverage_goal"`
	RequireTests  bool     `json:"require_tests"`
	TestTimeout   string   `json:"test_timeout"`
}

type CodeSpec struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Language     string            `json:"language"`
	Package      string            `json:"package"`
	Functions    []FunctionSpec    `json:"functions"`
	Types        []TypeSpec        `json:"types"`
	Dependencies []string          `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type FunctionSpec struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters"`
	ReturnType  string      `json:"return_type"`
	Visibility  string      `json:"visibility"`
	Documentation string    `json:"documentation"`
}

type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Optional    bool   `json:"optional"`
	DefaultValue string `json:"default_value"`
}

type TypeSpec struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Fields      []FieldSpec `json:"fields"`
	Methods     []FunctionSpec `json:"methods"`
}

type FieldSpec struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Visibility  string `json:"visibility"`
}

type GeneratedCode struct {
	Files       []GeneratedFile `json:"files"`
	Language    string          `json:"language"`
	Framework   string          `json:"framework"`
	Metadata    *GenerationMeta `json:"metadata"`
	Dependencies []string       `json:"dependencies"`
}

type GeneratedFile struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Content      string `json:"content"`
	Type         string `json:"type"`
	FileType     string `json:"file_type"`     // For backward compatibility and new chains
	Description  string `json:"description"`
	Dependencies []string `json:"dependencies"`
}

type TestSuite struct {
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	TestFiles    []TestFile `json:"test_files"`
	Framework    string     `json:"framework"`
	CoverageGoal float64    `json:"coverage_goal"`
	TestApproach     string `json:"test_approach"`        // New field for reasoning chains
	CoverageEstimate float64 `json:"coverage_estimate"`   // New field for reasoning chains
}

type TestFile struct {
	Path        string     `json:"path"`
	Name        string     `json:"name"`
	Content     string     `json:"content"`
	TestCases   []TestCase `json:"test_cases"`
	Dependencies []string  `json:"dependencies"`
	Description  string    `json:"description"`  // New field for reasoning chains
}

type TestCase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Input       string `json:"input"`
	Expected    string `json:"expected"`
	Type        string `json:"type"`
}

type Documentation struct {
	Files       []DocumentationFile `json:"files"`
	Format      string             `json:"format"`
	Language    string             `json:"language"`
	Metadata    map[string]interface{} `json:"metadata"`
	ReadmeContent    string   `json:"readme_content"`      // New field for reasoning chains
	APIDocumentation string   `json:"api_documentation"`   // New field for reasoning chains
	UsageExamples    []string `json:"usage_examples"`      // New field for reasoning chains
}

type DocumentationFile struct {
	Path        string `json:"path"`
	Name        string `json:"name"`
	Content     string `json:"content"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type GenerationMeta struct {
	Timestamp   time.Time         `json:"timestamp"`
	Generator   string            `json:"generator"`
	Version     string            `json:"version"`
	Parameters  map[string]interface{} `json:"parameters"`
	ExecutionTime time.Duration   `json:"execution_time"`
	TotalFiles      int           `json:"total_files"`        // New field for reasoning chains
	EstimatedLines  int           `json:"estimated_lines"`    // New field for reasoning chains 
	ComplexityScore float64       `json:"complexity_score"`   // New field for reasoning chains
	GoVersion       string        `json:"go_version"`         // New field for reasoning chains
}

type StaticReport struct {
	Issues      []StaticIssue     `json:"issues"`
	Metrics     *CodeMetrics      `json:"metrics"`
	Suggestions []Suggestion      `json:"suggestions"`
	OverallScore float64          `json:"overall_score"`
}

type StaticIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	Rule        string `json:"rule"`
}

type CodeMetrics struct {
	LinesOfCode         int     `json:"lines_of_code"`
	CyclomaticComplexity int    `json:"cyclomatic_complexity"`
	TestCoverage        float64 `json:"test_coverage"`
	Maintainability     float64 `json:"maintainability"`
	TechnicalDebt       float64 `json:"technical_debt"`
}

type Suggestion struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Priority    string `json:"priority"`
	Effort      string `json:"effort"`
}

type TestResults struct {
	TotalTests   int           `json:"total_tests"`
	PassedTests  int           `json:"passed_tests"`
	FailedTests  int           `json:"failed_tests"`
	SkippedTests int           `json:"skipped_tests"`
	Coverage     float64       `json:"coverage"`
	Duration     time.Duration `json:"duration"`
	Results      []TestResult  `json:"results"`
}

type TestResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Error    string        `json:"error"`
	Output   string        `json:"output"`
}

type ComplianceReport struct {
	Violations  []ComplianceViolation `json:"violations"`
	Score       float64               `json:"score"`
	Standards   []string              `json:"standards"`
	Summary     string                `json:"summary"`
}

type ComplianceViolation struct {
	Rule        string `json:"rule"`
	Description string `json:"description"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Severity    string `json:"severity"`
	Suggestion  string `json:"suggestion"`
}

