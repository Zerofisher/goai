package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Zerofisher/goai/internal/reasoning"
	contextpkg "github.com/Zerofisher/goai/pkg/context"
	"github.com/Zerofisher/goai/pkg/indexing"
	"github.com/Zerofisher/goai/pkg/tools"
	"github.com/Zerofisher/goai/pkg/types"
)

var rootCmd = &cobra.Command{
	Use:   "goai",
	Short: "GoAI Coder - Reasoning-based programming assistant",
	Long: `GoAI Coder is a reasoning-based programming assistant that helps developers
with intelligent code generation, analysis, and problem-solving.

Built on the Eino framework, it provides context-aware assistance with a focus
on Go language development while maintaining extensibility for other languages.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("GoAI Coder - Reasoning-based programming assistant")
		fmt.Println("Use 'goai --help' to see available commands")
	},
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug mode")

	// Add index subcommands
	indexCmd.AddCommand(indexBuildCmd)
	indexCmd.AddCommand(indexStatusCmd)
	indexCmd.AddCommand(indexRefreshCmd)
	indexCmd.AddCommand(indexClearCmd)

	// Add search flags
	searchCmd.Flags().IntP("limit", "l", 10, "maximum number of results to return")
	searchCmd.Flags().StringP("type", "t", "hybrid", "search type (full_text, semantic, symbol, hybrid)")

	// Add tool subcommands
	toolCmd.AddCommand(toolListCmd)
	toolCmd.AddCommand(toolExecuteCmd)

	// Add commands
	rootCmd.AddCommand(indexCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(thinkCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(fixCmd)
	rootCmd.AddCommand(toolCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Think command for problem analysis
var thinkCmd = &cobra.Command{
	Use:   "think [description]",
	Short: "Analyze programming problems through reasoning chains",
	Long: `The think command analyzes programming problems using structured reasoning.
It identifies the problem domain, technical challenges, architecture patterns,
and provides implementation guidance.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		description := strings.Join(args, " ")
		
		fmt.Printf("🤔 Analyzing problem: %s\n", description)
		fmt.Println("⏳ Starting reasoning chain analysis...")

		// Create context manager first
		workdir, _ := os.Getwd()
		contextManager, err := contextpkg.NewContextManager(workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating context manager: %v\n", err)
			os.Exit(1)
		}

		// Create reasoning engine
		reasoningEngine, err := reasoning.NewEngine(context.Background(), contextManager)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating reasoning engine: %v\n", err)
			fmt.Println("💡 Make sure OPENAI_API_KEY environment variable is set")
			os.Exit(1)
		}

		// Build project context
		projectContext, err := contextManager.BuildProjectContext(workdir)
		if err != nil {
			// Continue without project context
			projectContext = &types.ProjectContext{
				WorkingDirectory: workdir,
			}
		}

		// Create problem request
		req := &types.ProblemRequest{
			Description: description,
			Context:     projectContext,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// Perform analysis
		analysis, err := reasoningEngine.AnalyzeProblem(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing problem: %v\n", err)
			os.Exit(1)
		}

		// Display results
		fmt.Println("\n✅ Analysis Complete!")
		fmt.Printf("🎯 Problem Domain: %s\n", analysis.ProblemDomain)
		fmt.Printf("🔍 Complexity: %s\n", analysis.Complexity)
		fmt.Printf("🏗️ Architecture Pattern: %s\n", analysis.ArchitecturePattern)
		
		if len(analysis.TechnicalStack) > 0 {
			fmt.Println("\n⚙️ Technical Stack:")
			for i, tech := range analysis.TechnicalStack {
				fmt.Printf("  %d. %s\n", i+1, tech)
			}
		}

		if len(analysis.RiskFactors) > 0 {
			fmt.Println("\n⚠️ Risk Factors:")
			for i, risk := range analysis.RiskFactors {
				fmt.Printf("  %d. %s (Severity: %s)\n", i+1, risk.Description, risk.Severity)
			}
		}

		if len(analysis.Recommendations) > 0 {
			fmt.Println("\n💡 Recommendations:")
			for i, rec := range analysis.Recommendations {
				fmt.Printf("  %d. %s (Priority: %s)\n", i+1, rec.Description, rec.Priority)
			}
		}
	},
}

// Plan command for execution planning
var planCmd = &cobra.Command{
	Use:   "plan [description]",
	Short: "Generate detailed execution plans from problem analysis",
	Long: `The plan command creates detailed execution plans based on problem analysis.
It includes project structure design, component division, and implementation steps.
If no description is provided, it will use the result from the most recent 'think' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		var analysis *types.Analysis
		
		if len(args) > 0 {
			// First analyze the problem if description is provided
			description := strings.Join(args, " ")
			fmt.Printf("🤔 Analyzing problem: %s\n", description)
			
			// Create context manager first
			workdir, _ := os.Getwd()
			contextManager, err := contextpkg.NewContextManager(workdir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating context manager: %v\n", err)
				os.Exit(1)
			}

			// Create reasoning engine
			reasoningEngine, err := reasoning.NewEngine(context.Background(), contextManager)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating reasoning engine: %v\n", err)
				fmt.Println("💡 Make sure OPENAI_API_KEY environment variable is set")
				os.Exit(1)
			}

			// Build project context
			projectContext, err := contextManager.BuildProjectContext(workdir)
			if err != nil {
				projectContext = &types.ProjectContext{
					WorkingDirectory: workdir,
				}
			}

			// Create problem request
			req := &types.ProblemRequest{
				Description: description,
				Context:     projectContext,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			// Perform analysis
			analysis, err = reasoningEngine.AnalyzeProblem(ctx, req)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error analyzing problem: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("No description provided. Please provide a problem description.")
			os.Exit(1)
		}

		fmt.Println("📋 Generating execution plan...")

		// Create context manager if not already created
		workdir, _ := os.Getwd()
		contextManager, err := contextpkg.NewContextManager(workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating context manager: %v\n", err)
			os.Exit(1)
		}

		// Create reasoning engine
		reasoningEngine, err := reasoning.NewEngine(context.Background(), contextManager)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating reasoning engine: %v\n", err)
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		// Generate execution plan
		plan, err := reasoningEngine.GeneratePlan(ctx, analysis)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating plan: %v\n", err)
			os.Exit(1)
		}

		// Display execution plan
		fmt.Println("\n✅ Execution Plan Generated!")
		
		if len(plan.Steps) > 0 {
			fmt.Println("\n📝 Implementation Steps:")
			for i, step := range plan.Steps {
				fmt.Printf("  %d. %s\n", i+1, step.Description)
				if step.EstimatedTime > 0 {
					fmt.Printf("     ⏱️ Estimated time: %s\n", step.EstimatedTime)
				}
				if len(step.Dependencies) > 0 {
					fmt.Printf("     🔗 Dependencies: %v\n", step.Dependencies)
				}
			}
		}

		if len(plan.Dependencies) > 0 {
			fmt.Println("\n🔗 Project Dependencies:")
			for i, dep := range plan.Dependencies {
				fmt.Printf("  %d. %s (%s)\n", i+1, dep.Name, dep.Type)
			}
		}

		if plan.TestStrategy != nil {
			fmt.Println("\n🧪 Test Strategy:")
			if plan.TestStrategy.UnitTests {
				fmt.Println("     ✅ Unit Tests")
			}
			if plan.TestStrategy.IntegrationTests {
				fmt.Println("     ✅ Integration Tests")
			}
			if plan.TestStrategy.EndToEndTests {
				fmt.Println("     ✅ End-to-End Tests")
			}
			if len(plan.TestStrategy.TestFrameworks) > 0 {
				fmt.Printf("     📦 Frameworks: %v\n", plan.TestStrategy.TestFrameworks)
			}
			if plan.TestStrategy.CoverageTarget > 0 {
				fmt.Printf("     🎯 Coverage Target: %.1f%%\n", plan.TestStrategy.CoverageTarget*100)
			}
		}

		if plan.Timeline != nil {
			fmt.Printf("\n⏰ Timeline: %s\n", plan.Timeline.TotalDuration)
			if !plan.Timeline.StartTime.IsZero() {
				fmt.Printf("     Start: %s\n", plan.Timeline.StartTime.Format("2006-01-02"))
			}
			if !plan.Timeline.EstimatedEnd.IsZero() {
				fmt.Printf("     End: %s\n", plan.Timeline.EstimatedEnd.Format("2006-01-02"))
			}
			if len(plan.Timeline.Milestones) > 0 {
				fmt.Println("     📍 Milestones:")
				for _, milestone := range plan.Timeline.Milestones {
					fmt.Printf("       - %s: %s\n", milestone.Name, milestone.DueDate.Format("2006-01-02"))
				}
			}
		}
	},
}

// Analyze command for project analysis
var analyzeCmd = &cobra.Command{
	Use:   "analyze [path]",
	Short: "Analyze project structure and provide recommendations",
	Long: `The analyze command evaluates project architecture, code quality,
and identifies potential improvements and issues.`,
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		
		absPath, err := filepath.Abs(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("🔍 Analyzing project at: %s\n", absPath)

		// Create context manager
		contextManager, err := contextpkg.NewContextManager(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating context manager: %v\n", err)
			os.Exit(1)
		}

		// Build project context
		projectContext, err := contextManager.BuildProjectContext(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building project context: %v\n", err)
			os.Exit(1)
		}

		// Display project information
		fmt.Println("\n📊 Project Analysis Results:")
		fmt.Printf("📁 Working Directory: %s\n", projectContext.WorkingDirectory)
		
		if projectContext.GitInfo != nil {
			fmt.Printf("🌿 Git Branch: %s\n", projectContext.GitInfo.CurrentBranch)
			if projectContext.GitInfo.CurrentCommit != "" {
				fmt.Printf("📝 Last Commit: %s\n", projectContext.GitInfo.CurrentCommit[:8])
				fmt.Printf("👤 Author: %s\n", projectContext.GitInfo.LastCommitAuthor)
				if projectContext.GitInfo.LastCommitMsg != "" {
					fmt.Printf("💬 Message: %s\n", projectContext.GitInfo.LastCommitMsg)
				}
			}
		}

		if projectContext.ProjectStructure != nil {
			fmt.Printf("📦 Root Path: %s\n", projectContext.ProjectStructure.RootPath)
			fmt.Printf("📁 Total Files: %d\n", len(projectContext.ProjectStructure.Files))
			
			if len(projectContext.ProjectStructure.Directories) > 0 {
				fmt.Println("\n📂 Directory Structure:")
				for _, dir := range projectContext.ProjectStructure.Directories {
					fmt.Printf("  📁 %s (%s)\n", dir.Name, dir.Type)
				}
			}
		}

		if len(projectContext.Dependencies) > 0 {
			fmt.Println("\n🔗 Dependencies:")
			for _, dep := range projectContext.Dependencies {
				fmt.Printf("  📦 %s (%s)\n", dep.Name, dep.Type)
			}
		}

		if len(projectContext.RecentChanges) > 0 {
			fmt.Println("\n📈 Recent Changes:")
			for i, change := range projectContext.RecentChanges {
				if i >= 5 { // Show only last 5 changes
					break
				}
				fmt.Printf("  %s: %s (%s)\n", change.ChangeType, change.FilePath, change.Timestamp.Format("2006-01-02 15:04"))
			}
		}

		// Check if index exists for additional analysis
		indexManager, err := indexing.NewEnhancedIndexManager(absPath)
		if err == nil {
			defer func() { _ = indexManager.Close() }()
			
			if indexManager.IsIndexReady(absPath) {
				fmt.Println("\n🔍 Index Analysis:")
				status, err := indexManager.GetIndexStatus(absPath)
				if err == nil {
					fmt.Printf("  📊 Indexed Files: %d / %d\n", status.IndexedFiles, status.TotalFiles)
					fmt.Printf("  💾 Index Size: %s\n", formatBytes(status.IndexSize))
				}
			} else {
				fmt.Println("\n💡 Recommendation: Run 'goai index build' for enhanced analysis capabilities")
			}
		}

		// Provide general recommendations
		fmt.Println("\n🎯 Recommendations:")
		if projectContext.GitInfo == nil {
			fmt.Println("  📝 Consider initializing Git version control")
		}
		if projectContext.ProjectConfig == nil {
			fmt.Println("  ⚙️ Consider creating a GOAI.md configuration file")
		}
		fmt.Println("  🔍 Use 'goai search' to explore your codebase")
		fmt.Println("  🤔 Use 'goai think' for problem analysis")
		fmt.Println("  📋 Use 'goai plan' for implementation planning")
	},
}

// Fix command for bug analysis and fixing
var fixCmd = &cobra.Command{
	Use:   "fix [description]",
	Short: "Analyze and fix bugs with reasoning explanations",
	Long: `The fix command analyzes buggy code, identifies root causes,
and provides multiple solution approaches with explanations.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		description := strings.Join(args, " ")
		
		fmt.Printf("🐛 Analyzing bug: %s\n", description)
		fmt.Println("⏳ Starting bug analysis...")

		// Create context manager first
		workdir, _ := os.Getwd()
		contextManager, err := contextpkg.NewContextManager(workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating context manager: %v\n", err)
			os.Exit(1)
		}

		// Build project context
		projectContext, err := contextManager.BuildProjectContext(workdir)
		if err != nil {
			projectContext = &types.ProjectContext{
				WorkingDirectory: workdir,
			}
		}

		// Create problem request for bug analysis
		req := &types.ProblemRequest{
			Description: fmt.Sprintf("Bug analysis: %s", description),
			Context:     projectContext,
		}

		// Create reasoning engine
		reasoningEngine, err := reasoning.NewEngine(context.Background(), contextManager)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating reasoning engine: %v\n", err)
			fmt.Println("💡 Make sure OPENAI_API_KEY environment variable is set")
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()

		// Perform analysis
		analysis, err := reasoningEngine.AnalyzeProblem(ctx, req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing bug: %v\n", err)
			os.Exit(1)
		}

		// Display bug analysis
		fmt.Println("\n🔍 Bug Analysis Results:")
		fmt.Printf("🎯 Problem Domain: %s\n", analysis.ProblemDomain)
		fmt.Printf("🔍 Complexity: %s\n", analysis.Complexity)
		
		if len(analysis.RiskFactors) > 0 {
			fmt.Println("\n🚨 Root Causes & Risk Factors:")
			for i, risk := range analysis.RiskFactors {
				fmt.Printf("  %d. %s (Severity: %s)\n", i+1, risk.Description, risk.Severity)
				if risk.Mitigation != "" {
					fmt.Printf("     💡 Mitigation: %s\n", risk.Mitigation)
				}
			}
		}

		if len(analysis.Recommendations) > 0 {
			fmt.Println("\n🔧 Fix Recommendations:")
			for i, rec := range analysis.Recommendations {
				fmt.Printf("  %d. %s (Priority: %s)\n", i+1, rec.Description, rec.Priority)
			}
		}

		if len(analysis.TechnicalStack) > 0 {
			fmt.Println("\n🛠️ Technical Context:")
			for _, tech := range analysis.TechnicalStack {
				fmt.Printf("  • %s\n", tech)
			}
		}

		fmt.Println("\n💡 Next Steps:")
		fmt.Println("  1. Review the recommended fixes above")
		fmt.Println("  2. Use 'goai plan' to create an implementation plan")
		fmt.Println("  3. Use 'goai tool' commands to apply changes")
		fmt.Println("  4. Test the fixes thoroughly")
	},
}

// Index command for codebase indexing operations
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Manage codebase indexing for enhanced search and analysis",
	Long: `The index command manages codebase indexing operations including building,
refreshing, checking status, and clearing indexes for full-text search,
symbol indexing, and vector embeddings.`,
}

var indexBuildCmd = &cobra.Command{
	Use:   "build [path]",
	Short: "Build comprehensive index of the codebase",
	Long: `Build a comprehensive index including full-text search, symbol index,
and vector embeddings for the specified directory (default: current directory).`,
	Run: func(cmd *cobra.Command, args []string) {
		workdir := "."
		if len(args) > 0 {
			workdir = args[0]
		}
		
		absPath, err := filepath.Abs(workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Building index for: %s\n", absPath)
		
		indexManager, err := indexing.NewEnhancedIndexManager(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating index manager: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = indexManager.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if err := indexManager.BuildIndex(ctx, absPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error building index: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ Index built successfully")
	},
}

var indexStatusCmd = &cobra.Command{
	Use:   "status [path]",
	Short: "Show indexing status and statistics",
	Long: `Display current indexing status, statistics, and health information
for the specified directory (default: current directory).`,
	Run: func(cmd *cobra.Command, args []string) {
		workdir := "."
		if len(args) > 0 {
			workdir = args[0]
		}

		absPath, err := filepath.Abs(workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
			os.Exit(1)
		}

		indexManager, err := indexing.NewEnhancedIndexManager(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating index manager: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = indexManager.Close() }()

		status, err := indexManager.GetIndexStatus(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting index status: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Index Status for: %s\n", absPath)
		fmt.Printf("  Ready: %t\n", indexManager.IsIndexReady(absPath))
		fmt.Printf("  Last Updated: %s\n", status.LastUpdated.Format(time.RFC3339))
		fmt.Printf("  Total Files: %d\n", status.TotalFiles)
		fmt.Printf("  Indexed Files: %d\n", status.IndexedFiles)
		fmt.Printf("  Index Size: %s\n", formatBytes(status.IndexSize))
	},
}

var indexRefreshCmd = &cobra.Command{
	Use:   "refresh [paths...]",
	Short: "Refresh index for specific paths or entire codebase",
	Long: `Refresh the index for specific file paths or the entire codebase.
If no paths are specified, refreshes the entire index.`,
	Run: func(cmd *cobra.Command, args []string) {
		workdir := "."
		var paths []string

		if len(args) == 0 {
			paths = []string{workdir}
		} else {
			paths = args
		}

		absPath, err := filepath.Abs(workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
			os.Exit(1)
		}

		indexManager, err := indexing.NewEnhancedIndexManager(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating index manager: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = indexManager.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if err := indexManager.RefreshIndex(ctx, paths); err != nil {
			fmt.Fprintf(os.Stderr, "Error refreshing index: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Index refreshed for %d path(s)\n", len(paths))
	},
}

var indexClearCmd = &cobra.Command{
	Use:   "clear [path]",
	Short: "Clear all indexes for the specified path",
	Long: `Clear all indexes (full-text, symbol, and embedding) for the specified 
directory (default: current directory). This will remove all indexed data.`,
	Run: func(cmd *cobra.Command, args []string) {
		workdir := "."
		if len(args) > 0 {
			workdir = args[0]
		}

		absPath, err := filepath.Abs(workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
			os.Exit(1)
		}

		indexManager, err := indexing.NewEnhancedIndexManager(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating index manager: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = indexManager.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := indexManager.ClearIndex(ctx, absPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error clearing index: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Index cleared for: %s\n", absPath)
	},
}

// Search command for code search operations
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search codebase using multiple retrieval methods",
	Long: `Search the indexed codebase using full-text search, semantic search,
symbol search, and hybrid retrieval with intelligent reranking.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := strings.Join(args, " ")
		workdir := "."

		absPath, err := filepath.Abs(workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
			os.Exit(1)
		}

		indexManager, err := indexing.NewEnhancedIndexManager(absPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating index manager: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = indexManager.Close() }()

		// Check if index is ready
		if !indexManager.IsIndexReady(absPath) {
			fmt.Println("Index not ready. Run 'goai index build' first.")
			os.Exit(1)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Get search options from flags
		limit, _ := cmd.Flags().GetInt("limit")
		if limit == 0 {
			limit = 10 // default limit
		}

		searchReq := &indexing.SearchRequest{
			Query:      query,
			WorkingDir: absPath,
			MaxResults: limit,
			SearchTypes: []indexing.SearchType{indexing.SearchTypeHybrid},
			IncludeContent: true,
		}

		results, err := indexManager.SearchWithHybridRetrieval(ctx, searchReq)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error searching: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Search Results for: %s\n", query)
		fmt.Printf("Found %d results\n\n", len(results.Results))

		for i, item := range results.Results {
			if i >= limit {
				break
			}
			fmt.Printf("%d. %s (score: %.3f)\n", i+1, item.FilePath, item.Score)
			if len(item.Content) > 0 {
				// Show first 2 lines of content
				lines := strings.Split(item.Content, "\n")
				for j, line := range lines {
					if j >= 2 {
						break
					}
					fmt.Printf("   %s\n", line)
				}
			}
			fmt.Println()
		}
	},
}

// Tool command for tool system operations
var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "Manage and execute development tools",
	Long: `The tool command provides access to the tool system for file operations,
code analysis, and system interactions.`,
}

var toolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available tools",
	Long:  `Display all available tools with their descriptions and capabilities.`,
	Run: func(cmd *cobra.Command, args []string) {
		factory := tools.NewToolFactory()
		availableTools := factory.GetAvailableToolsInfo()
		
		fmt.Println("Available Tools:")
		fmt.Println()
		
		categories := make(map[string][]tools.ToolInfo)
		for _, tool := range availableTools {
			categories[tool.Category] = append(categories[tool.Category], tool)
		}
		
		for category, categoryTools := range categories {
			// Capitalize first letter of category
			categoryTitle := strings.ToUpper(category[:1]) + category[1:]
			fmt.Printf("%s:\n", categoryTitle)
			for _, tool := range categoryTools {
				confirmStr := ""
				if tool.RequiresConfirmation {
					confirmStr = " (requires confirmation)"
				}
				fmt.Printf("  %s - %s%s\n", tool.Name, tool.Description, confirmStr)
			}
			fmt.Println()
		}
	},
}

var toolExecuteCmd = &cobra.Command{
	Use:   "execute [tool-name] [args...]",
	Short: "Execute a specific tool",
	Long: `Execute a specific tool with the provided arguments.
Use 'goai tool list' to see available tools.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		toolName := args[0]
		toolArgs := args[1:]

		// Create index manager for tools that need it
		workdir, _ := os.Getwd()
		indexManager, err := indexing.NewEnhancedIndexManager(workdir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating index manager: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = indexManager.Close() }()

		// Create tool manager using factory
		factory := tools.NewToolFactory()
		toolManager, err := factory.CreateDefaultToolManager(indexManager)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating tool manager: %v\n", err)
			os.Exit(1)
		}
		
		// Convert string args to tool parameters
		params := make(map[string]any)
		for i, arg := range toolArgs {
			params[fmt.Sprintf("arg%d", i)] = arg
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := toolManager.ExecuteTool(ctx, toolName, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing tool '%s': %v\n", toolName, err)
			os.Exit(1)
		}

		fmt.Printf("Tool '%s' executed successfully:\n", toolName)
		if result.Success {
			fmt.Printf("Output: %s\n", result.Output)
			if result.Data != nil {
				fmt.Printf("Data: %v\n", result.Data)
			}
		} else {
			fmt.Printf("Error: %s\n", result.Error)
		}
		
		if len(result.ModifiedFiles) > 0 {
			fmt.Printf("Modified files: %v\n", result.ModifiedFiles)
		}
	},
}

// Helper function to format bytes
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}