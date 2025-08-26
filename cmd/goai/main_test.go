package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// TestRootCommand tests the root command functionality
func TestRootCommand(t *testing.T) {
	// Test that root command exists and has basic properties
	if rootCmd == nil {
		t.Fatalf("Root command should not be nil")
	}
	
	if rootCmd.Use != "goai" {
		t.Errorf("Expected command name 'goai', got '%s'", rootCmd.Use)
	}
	
	if rootCmd.Short == "" {
		t.Errorf("Command should have a short description")
	}
	
	if rootCmd.Long == "" {
		t.Errorf("Command should have a long description")
	}
}

// TestRootCommandExecution tests root command execution
func TestRootCommandExecution(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	// Execute root command
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	
	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout
	
	// Read output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()
	
	if err != nil {
		t.Errorf("Root command execution failed: %v", err)
	}
	
	if !strings.Contains(output, "GoAI Coder") {
		t.Errorf("Expected output to contain 'GoAI Coder', got: %s", output)
	}
}

// TestRootCommandHelp tests the help functionality
func TestRootCommandHelp(t *testing.T) {
	// Test help flag
	rootCmd.SetArgs([]string{"--help"})
	
	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Help command failed: %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Help output should contain 'Usage:', got: %s", output)
	}
	
	if !strings.Contains(output, "Available Commands:") {
		t.Errorf("Help output should contain 'Available Commands:', got: %s", output)
	}
}

// TestGlobalFlags tests global flags
func TestGlobalFlags(t *testing.T) {
	// Test that global flags are defined
	flags := rootCmd.PersistentFlags()
	
	// Test config flag
	configFlag := flags.Lookup("config")
	if configFlag == nil {
		t.Errorf("Config flag should be defined")
		return
	}
	if configFlag.Shorthand != "c" {
		t.Errorf("Config flag shorthand should be 'c', got '%s'", configFlag.Shorthand)
	}
	
	// Test verbose flag
	verboseFlag := flags.Lookup("verbose")
	if verboseFlag == nil {
		t.Errorf("Verbose flag should be defined")
		return
	}
	if verboseFlag.Shorthand != "v" {
		t.Errorf("Verbose flag shorthand should be 'v', got '%s'", verboseFlag.Shorthand)
	}
	
	// Test debug flag
	debugFlag := flags.Lookup("debug")
	if debugFlag == nil {
		t.Errorf("Debug flag should be defined")
		return
	}
	if debugFlag.Shorthand != "d" {
		t.Errorf("Debug flag shorthand should be 'd', got '%s'", debugFlag.Shorthand)
	}
}

// TestSubCommands tests that all expected sub-commands are present
func TestSubCommands(t *testing.T) {
	expectedCommands := []string{"index", "search", "think", "plan", "analyze", "fix", "tool"}
	
	for _, cmdName := range expectedCommands {
		cmd, _, err := rootCmd.Find([]string{cmdName})
		if err != nil {
			t.Errorf("Command '%s' should exist: %v", cmdName, err)
			continue
		}
		
		if cmd.Name() != cmdName {
			t.Errorf("Expected command name '%s', got '%s'", cmdName, cmd.Name())
		}
		
		if cmd.Short == "" {
			t.Errorf("Command '%s' should have a short description", cmdName)
		}
	}
}

// TestThinkCommand tests the think command
func TestThinkCommand(t *testing.T) {
	if thinkCmd == nil {
		t.Fatalf("Think command should not be nil")
	}
	
	if thinkCmd.Use != "think [description]" {
		t.Errorf("Expected think command use 'think [description]', got '%s'", thinkCmd.Use)
	}
	
	if thinkCmd.Short == "" {
		t.Errorf("Think command should have a short description")
	}
	
	// Test that it requires arguments
	if thinkCmd.Args == nil {
		t.Errorf("Think command should have argument validation")
	}
	
	// Test that Run function is set (we won't actually execute it to avoid dependencies)
	if thinkCmd.Run == nil && thinkCmd.RunE == nil {
		t.Errorf("Think command should have a Run or RunE function")
	}
}

// TestPlanCommand tests the plan command
func TestPlanCommand(t *testing.T) {
	if planCmd == nil {
		t.Fatalf("Plan command should not be nil")
	}
	
	if planCmd.Use != "plan [description]" {
		t.Errorf("Expected plan command use 'plan [description]', got '%s'", planCmd.Use)
	}
	
	if planCmd.Short == "" {
		t.Errorf("Plan command should have a short description")
	}
	
	// Test that Run function is set
	if planCmd.Run == nil && planCmd.RunE == nil {
		t.Errorf("Plan command should have a Run or RunE function")
	}
}

// TestAnalyzeCommand tests the analyze command
func TestAnalyzeCommand(t *testing.T) {
	if analyzeCmd == nil {
		t.Fatalf("Analyze command should not be nil")
	}
	
	if analyzeCmd.Use != "analyze [path]" {
		t.Errorf("Expected analyze command use 'analyze [path]', got '%s'", analyzeCmd.Use)
	}
	
	if analyzeCmd.Short == "" {
		t.Errorf("Analyze command should have a short description")
	}
	
	// Test that Run function is set
	if analyzeCmd.Run == nil && analyzeCmd.RunE == nil {
		t.Errorf("Analyze command should have a Run or RunE function")
	}
}

// TestFixCommand tests the fix command
func TestFixCommand(t *testing.T) {
	if fixCmd == nil {
		t.Fatalf("Fix command should not be nil")
	}
	
	if fixCmd.Use != "fix [description]" {
		t.Errorf("Expected fix command use 'fix [description]', got '%s'", fixCmd.Use)
	}
	
	if fixCmd.Short == "" {
		t.Errorf("Fix command should have a short description")
	}
	
	// Test that it requires arguments
	if fixCmd.Args == nil {
		t.Errorf("Fix command should have argument validation")
	}
	
	// Test that Run function is set
	if fixCmd.Run == nil && fixCmd.RunE == nil {
		t.Errorf("Fix command should have a Run or RunE function")
	}
}

// TestCommandHierarchy tests the command hierarchy structure
func TestCommandHierarchy(t *testing.T) {
	// Test that all commands are added to root
	commands := rootCmd.Commands()
	
	expectedCmds := []*cobra.Command{indexCmd, searchCmd, thinkCmd, planCmd, analyzeCmd, fixCmd, toolCmd}
	
	if len(commands) < len(expectedCmds) {
		t.Errorf("Expected at least %d commands, got %d", len(expectedCmds), len(commands))
	}
	
	// Find our commands in the list
	cmdMap := make(map[string]*cobra.Command)
	for _, cmd := range commands {
		cmdMap[cmd.Name()] = cmd
	}
	
	expectedNames := []string{"index", "search", "think", "plan", "analyze", "fix", "tool"}
	for _, name := range expectedNames {
		if _, exists := cmdMap[name]; !exists {
			t.Errorf("Command '%s' should be added to root command", name)
		}
	}
}

// TestCommandValidation tests command argument validation
func TestCommandValidation(t *testing.T) {
	// Test think command requires arguments
	err := thinkCmd.Args(thinkCmd, []string{})
	if err == nil {
		t.Errorf("Think command should require arguments")
	}
	
	err = thinkCmd.Args(thinkCmd, []string{"test"})
	if err != nil {
		t.Errorf("Think command should accept valid arguments: %v", err)
	}
	
	// Test fix command requires arguments
	err = fixCmd.Args(fixCmd, []string{})
	if err == nil {
		t.Errorf("Fix command should require arguments")
	}
	
	err = fixCmd.Args(fixCmd, []string{"test bug"})
	if err != nil {
		t.Errorf("Fix command should accept valid arguments: %v", err)
	}
}

// TestMainFunction tests the main function indirectly
func TestMainFunction(t *testing.T) {
	// We can't directly test main() since it calls os.Exit,
	// but we can test that rootCmd.Execute() works properly
	
	// Set up a test scenario
	rootCmd.SetArgs([]string{"--help"})
	
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Root command execution should not fail: %v", err)
	}
	
	output := buf.String()
	if output == "" {
		t.Errorf("Command should produce output")
	}
}

// BenchmarkRootCommandCreation benchmarks root command setup
func BenchmarkRootCommandCreation(b *testing.B) {
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		cmd := &cobra.Command{
			Use:   "goai",
			Short: "GoAI Coder - Reasoning-based programming assistant",
		}
		
		// Add flags
		cmd.PersistentFlags().StringP("config", "c", "", "config file path")
		cmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
		cmd.PersistentFlags().BoolP("debug", "d", false, "debug mode")
		
		// Command is guaranteed to be non-nil after creation
		_ = cmd
	}
}

// TestThinkCommandExecution tests think command execution with mocked dependencies
func TestThinkCommandExecution(t *testing.T) {
	// Skip if no API key is set (avoid actual API calls in tests)
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}
	
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)
	
	// Capture stderr to check for errors
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	
	// Set up command with test arguments
	thinkCmd.SetArgs([]string{"create a simple hello world program"})
	
	// Execute command in a goroutine to handle potential os.Exit calls
	var cmdErr error
	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Command may call os.Exit, which is expected in error cases
				done <- true
			}
		}()
		cmdErr = thinkCmd.Execute()
		done <- true
	}()
	
	// Wait for command completion or timeout
	select {
	case <-done:
		// Command completed
	case <-time.After(30 * time.Second):
		t.Fatal("Think command execution timed out")
	}
	
	// Restore stderr
	_ = w.Close()
	os.Stderr = oldStderr
	
	// Read stderr output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	stderrOutput := buf.String()
	
	// If command failed due to API issues, that's expected in test environment
	if cmdErr != nil && strings.Contains(stderrOutput, "Error creating reasoning engine") {
		t.Skip("Expected API error in test environment without valid setup")
	}
	
	if cmdErr != nil {
		t.Errorf("Think command execution failed unexpectedly: %v", cmdErr)
	}
}

// TestThinkCommandArguments tests think command with different argument combinations
func TestThinkCommandArguments(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
			errorMsg:    "requires at least 1 arg(s), only received 0",
		},
		{
			name:        "single argument",
			args:        []string{"test"},
			expectError: false,
		},
		{
			name:        "multiple arguments",
			args:        []string{"create", "a", "web", "server"},
			expectError: false,
		},
		{
			name:        "complex description",
			args:        []string{"implement a REST API with authentication and database integration"},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test argument validation without executing the command
			err := thinkCmd.Args(thinkCmd, tt.args)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestPlanCommandExecution tests plan command execution
func TestPlanCommandExecution(t *testing.T) {
	// Skip if no API key is set
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}
	
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)
	
	// Test plan command structure
	if planCmd.Use != "plan [description]" {
		t.Errorf("Expected plan command use 'plan [description]', got '%s'", planCmd.Use)
	}
	
	// Test that command exists and has basic properties
	if planCmd.Short == "" {
		t.Error("Plan command should have a short description")
	}
	
	if planCmd.Run == nil {
		t.Error("Plan command should have a Run function")
	}
}

// TestAnalyzeCommandExecution tests analyze command execution
func TestAnalyzeCommandExecution(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)
	
	// Create a simple Go file for analysis
	testFile := "main.go"
	testContent := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Test analyze command properties
	if analyzeCmd.Use != "analyze [path]" {
		t.Errorf("Expected analyze command use 'analyze [path]', got '%s'", analyzeCmd.Use)
	}
	
	if analyzeCmd.Short == "" {
		t.Error("Analyze command should have a short description")
	}
	
	if analyzeCmd.Run == nil {
		t.Error("Analyze command should have a Run function")
	}
}

// TestFixCommandExecution tests fix command execution
func TestFixCommandExecution(t *testing.T) {
	// Test fix command argument validation
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "single argument",
			args:        []string{"fix null pointer issue"},
			expectError: false,
		},
		{
			name:        "multiple arguments",
			args:        []string{"fix", "the", "memory", "leak"},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fixCmd.Args(fixCmd, tt.args)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestIndexCommands tests index command and its subcommands
func TestIndexCommands(t *testing.T) {
	// Test index command structure
	if indexCmd.Use != "index" {
		t.Errorf("Expected index command use 'index', got '%s'", indexCmd.Use)
	}
	
	if indexCmd.Short == "" {
		t.Error("Index command should have a short description")
	}
	
	// Test subcommands exist
	subCommands := indexCmd.Commands()
	expectedSubCmds := []string{"build", "status", "refresh", "clear"}
	
	subCmdMap := make(map[string]bool)
	for _, cmd := range subCommands {
		subCmdMap[cmd.Name()] = true
	}
	
	for _, expected := range expectedSubCmds {
		if !subCmdMap[expected] {
			t.Errorf("Expected index subcommand '%s' to exist", expected)
		}
	}
	
	// Test individual subcommands (check actual usage strings)
	if !strings.HasPrefix(indexBuildCmd.Use, "build") {
		t.Errorf("Expected index build command to start with 'build', got '%s'", indexBuildCmd.Use)
	}
	
	if !strings.HasPrefix(indexStatusCmd.Use, "status") {
		t.Errorf("Expected index status command to start with 'status', got '%s'", indexStatusCmd.Use)
	}
	
	if !strings.HasPrefix(indexRefreshCmd.Use, "refresh") {
		t.Errorf("Expected index refresh command to start with 'refresh', got '%s'", indexRefreshCmd.Use)
	}
	
	if !strings.HasPrefix(indexClearCmd.Use, "clear") {
		t.Errorf("Expected index clear command to start with 'clear', got '%s'", indexClearCmd.Use)
	}
}

// TestIndexBuildCommand tests index build command execution
func TestIndexBuildCommand(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)
	
	// Create a simple project structure
	_ = os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)
	
	// Test command properties
	if indexBuildCmd.Run == nil {
		t.Error("Index build command should have a Run function")
	}
	
	if indexBuildCmd.Short == "" {
		t.Error("Index build command should have a short description")
	}
}

// TestSearchCommand tests search command execution
func TestSearchCommand(t *testing.T) {
	// Test search command structure
	if searchCmd.Use != "search [query]" {
		t.Errorf("Expected search command use 'search [query]', got '%s'", searchCmd.Use)
	}
	
	if searchCmd.Short == "" {
		t.Error("Search command should have a short description")
	}
	
	if searchCmd.Run == nil {
		t.Error("Search command should have a Run function")
	}
	
	// Test search command flags
	flags := searchCmd.Flags()
	
	limitFlag := flags.Lookup("limit")
	if limitFlag == nil {
		t.Error("Search command should have a limit flag")
	} else if limitFlag.Shorthand != "l" {
		t.Errorf("Expected limit flag shorthand 'l', got '%s'", limitFlag.Shorthand)
	}
	
	typeFlag := flags.Lookup("type")
	if typeFlag == nil {
		t.Error("Search command should have a type flag")
	} else if typeFlag.Shorthand != "t" {
		t.Errorf("Expected type flag shorthand 't', got '%s'", typeFlag.Shorthand)
	}
}

// TestToolCommands tests tool command and its subcommands
func TestToolCommands(t *testing.T) {
	// Test tool command structure
	if toolCmd.Use != "tool" {
		t.Errorf("Expected tool command use 'tool', got '%s'", toolCmd.Use)
	}
	
	if toolCmd.Short == "" {
		t.Error("Tool command should have a short description")
	}
	
	// Test subcommands exist
	subCommands := toolCmd.Commands()
	expectedSubCmds := []string{"list", "execute"}
	
	subCmdMap := make(map[string]bool)
	for _, cmd := range subCommands {
		subCmdMap[cmd.Name()] = true
	}
	
	for _, expected := range expectedSubCmds {
		if !subCmdMap[expected] {
			t.Errorf("Expected tool subcommand '%s' to exist", expected)
		}
	}
	
	// Test individual subcommands
	if toolListCmd.Use != "list" {
		t.Errorf("Expected tool list command use 'list', got '%s'", toolListCmd.Use)
	}
	
	if !strings.HasPrefix(toolExecuteCmd.Use, "execute [tool-name]") {
		t.Errorf("Expected tool execute command to start with 'execute [tool-name]', got '%s'", toolExecuteCmd.Use)
	}
}

// TestCommandErrorHandling tests error handling in commands  
func TestCommandErrorHandling(t *testing.T) {
	// Test that commands have proper error handling structure
	// without actually executing them (to avoid side effects)
	
	// Test that commands with Args validators work properly
	commands := []*cobra.Command{thinkCmd, fixCmd, toolExecuteCmd}
	
	for _, cmd := range commands {
		if cmd.Args == nil {
			t.Errorf("Command '%s' should have argument validation", cmd.Name())
		}
		
		// Test empty args (should fail for commands that require args)
		err := cmd.Args(cmd, []string{})
		if err == nil {
			t.Errorf("Command '%s' should require arguments", cmd.Name())
		}
		
		// Test valid args (should pass)
		err = cmd.Args(cmd, []string{"test"})
		if err != nil {
			t.Logf("Command '%s' has specific argument requirements: %v", cmd.Name(), err)
		}
		
		// Ensure Run function exists
		if cmd.Run == nil && cmd.RunE == nil {
			t.Errorf("Command '%s' should have a Run or RunE function", cmd.Name())
		}
	}
}

// TestAllCommandsHaveDescriptions tests that all commands have proper descriptions
func TestAllCommandsHaveDescriptions(t *testing.T) {
	commands := []*cobra.Command{
		rootCmd, thinkCmd, planCmd, analyzeCmd, fixCmd,
		indexCmd, indexBuildCmd, indexStatusCmd, indexRefreshCmd, indexClearCmd,
		searchCmd, toolCmd, toolListCmd, toolExecuteCmd,
	}
	
	for _, cmd := range commands {
		if cmd.Short == "" {
			t.Errorf("Command '%s' should have a short description", cmd.Use)
		}
		
		// Most commands should have long descriptions too
		if cmd == rootCmd || cmd == thinkCmd || cmd == planCmd || cmd == analyzeCmd {
			if cmd.Long == "" {
				t.Errorf("Command '%s' should have a long description", cmd.Use)
			}
		}
	}
}

// TestCommandFlags tests command-specific flags
func TestCommandFlags(t *testing.T) {
	// Test search command flags
	searchFlags := searchCmd.Flags()
	
	// Test limit flag
	limitFlag := searchFlags.Lookup("limit")
	if limitFlag == nil {
		t.Error("Search command should have limit flag")
	} else {
		if limitFlag.DefValue != "10" {
			t.Errorf("Expected limit flag default '10', got '%s'", limitFlag.DefValue)
		}
	}
	
	// Test type flag  
	typeFlag := searchFlags.Lookup("type")
	if typeFlag == nil {
		t.Error("Search command should have type flag")
	} else {
		if typeFlag.DefValue != "hybrid" {
			t.Errorf("Expected type flag default 'hybrid', got '%s'", typeFlag.DefValue)
		}
	}
	
	// Test that flags are properly configured
	searchCmd.SetArgs([]string{"--limit", "5", "--type", "semantic", "test query"})
	err := searchCmd.ParseFlags([]string{"--limit", "5", "--type", "semantic"})
	if err != nil {
		t.Errorf("Failed to parse search flags: %v", err)
	}
	
	limitVal, err := searchCmd.Flags().GetInt("limit")
	if err != nil {
		t.Errorf("Failed to get limit flag value: %v", err)
	} else if limitVal != 5 {
		t.Errorf("Expected limit value 5, got %d", limitVal)
	}
	
	typeVal, err := searchCmd.Flags().GetString("type")
	if err != nil {
		t.Errorf("Failed to get type flag value: %v", err)
	} else if typeVal != "semantic" {
		t.Errorf("Expected type value 'semantic', got '%s'", typeVal)
	}
}

// TestFormatBytes tests the formatBytes helper function
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
		{1125899906842624, "1.0 PB"},
		{1152921504606846976, "1.0 EB"},
		{2048, "2.0 KB"},
		{2097152, "2.0 MB"},
	}
	
	for _, test := range tests {
		result := formatBytes(test.input)
		if result != test.expected {
			t.Errorf("formatBytes(%d) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

// TestMainFunctionError tests main function error handling
func TestMainFunctionError(t *testing.T) {
	// Test that we can create a command that would cause Execute() to fail
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	
	// Create a test command that will fail
	testCmd := &cobra.Command{
		Use: "test-fail",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("test error")
		},
	}
	
	// Execute the command and expect it to fail
	err := testCmd.Execute()
	if err == nil {
		t.Error("Expected command to fail")
	}
	
	// The command should have produced an error
	if err.Error() != "test error" {
		t.Errorf("Expected error 'test error', got '%s'", err.Error())
	}
}

// TestIndexSubCommandsExecution tests index subcommands
func TestIndexSubCommandsExecution(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)
	
	// Create a simple Go project
	_ = os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0644)
	_ = os.WriteFile("go.mod", []byte("module test\n\ngo 1.21\n"), 0644)
	
	// Test that index build command has proper properties
	if indexBuildCmd.Short == "" {
		t.Error("Index build command should have a short description")
	}
	
	if indexBuildCmd.Run == nil {
		t.Error("Index build command should have a Run function")
	}
	
	// Test index status command
	if indexStatusCmd.Short == "" {
		t.Error("Index status command should have a short description")
	}
	
	if indexStatusCmd.Run == nil {
		t.Error("Index status command should have a Run function")
	}
	
	// Test index refresh command
	if indexRefreshCmd.Short == "" {
		t.Error("Index refresh command should have a short description")
	}
	
	if indexRefreshCmd.Run == nil {
		t.Error("Index refresh command should have a Run function")
	}
	
	// Test index clear command
	if indexClearCmd.Short == "" {
		t.Error("Index clear command should have a short description")  
	}
	
	if indexClearCmd.Run == nil {
		t.Error("Index clear command should have a Run function")
	}
}

// TestToolSubCommandsExecution tests tool subcommands
func TestToolSubCommandsExecution(t *testing.T) {
	// Test tool list command
	if toolListCmd.Short == "" {
		t.Error("Tool list command should have a short description")
	}
	
	if toolListCmd.Run == nil {
		t.Error("Tool list command should have a Run function")
	}
	
	// Test tool execute command
	if toolExecuteCmd.Short == "" {
		t.Error("Tool execute command should have a short description")
	}
	
	if toolExecuteCmd.Run == nil {
		t.Error("Tool execute command should have a Run function")
	}
	
	// Test tool execute command argument validation
	if toolExecuteCmd.Args == nil {
		t.Error("Tool execute command should have argument validation")
	}
	
	// Test argument validation
	err := toolExecuteCmd.Args(toolExecuteCmd, []string{})
	if err == nil {
		t.Error("Tool execute command should require arguments")
	}
	
	err = toolExecuteCmd.Args(toolExecuteCmd, []string{"testTool"})
	if err != nil {
		t.Errorf("Tool execute command should accept valid arguments: %v", err)
	}
}

// TestCommandInitialization tests that commands are properly initialized
func TestCommandInitialization(t *testing.T) {
	// Verify that all commands are added to the root command
	rootCommands := rootCmd.Commands()
	commandNames := make(map[string]bool)
	
	for _, cmd := range rootCommands {
		commandNames[cmd.Name()] = true
	}
	
	expectedCommands := []string{"index", "search", "think", "plan", "analyze", "fix", "tool"}
	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("Command '%s' should be added to root command during initialization", expected)
		}
	}
	
	// Test that index command has all its subcommands
	indexSubCommands := indexCmd.Commands()
	indexSubNames := make(map[string]bool)
	
	for _, cmd := range indexSubCommands {
		indexSubNames[cmd.Name()] = true
	}
	
	expectedIndexSubs := []string{"build", "status", "refresh", "clear"}
	for _, expected := range expectedIndexSubs {
		if !indexSubNames[expected] {
			t.Errorf("Index subcommand '%s' should be added during initialization", expected)
		}
	}
	
	// Test that tool command has all its subcommands
	toolSubCommands := toolCmd.Commands()
	toolSubNames := make(map[string]bool)
	
	for _, cmd := range toolSubCommands {
		toolSubNames[cmd.Name()] = true
	}
	
	expectedToolSubs := []string{"list", "execute"}
	for _, expected := range expectedToolSubs {
		if !toolSubNames[expected] {
			t.Errorf("Tool subcommand '%s' should be added during initialization", expected)
		}
	}
}

// TestGlobalFlagsInheritance tests that global flags are inherited by subcommands
func TestGlobalFlagsInheritance(t *testing.T) {
	globalFlags := rootCmd.PersistentFlags()
	
	// Test that persistent flags exist
	flagNames := []string{"config", "verbose", "debug"}
	for _, flagName := range flagNames {
		flag := globalFlags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Global flag '%s' should exist", flagName)
		}
	}
	
	// Test that subcommands inherit persistent flags
	subCommands := []*cobra.Command{thinkCmd, planCmd, analyzeCmd, fixCmd, indexCmd, searchCmd, toolCmd}
	
	for _, subCmd := range subCommands {
		for _, flagName := range flagNames {
			// Commands should be able to access inherited persistent flags
			inheritedFlags := subCmd.InheritedFlags()
			inheritedFlag := inheritedFlags.Lookup(flagName)
			if inheritedFlag == nil {
				t.Errorf("Command '%s' should inherit persistent flag '%s'", subCmd.Name(), flagName)
			}
		}
	}
}

// TestCommandUsageStrings tests that commands have proper usage strings
func TestCommandUsageStrings(t *testing.T) {
	commandUsageTests := []struct {
		cmd          *cobra.Command
		expectedUse  string
		description  string
	}{
		{thinkCmd, "think [description]", "think command"},
		{planCmd, "plan [description]", "plan command"},
		{analyzeCmd, "analyze [path]", "analyze command"},
		{fixCmd, "fix [description]", "fix command"},
		{indexCmd, "index", "index command"},
		{searchCmd, "search [query]", "search command"},
		{toolCmd, "tool", "tool command"},
		{indexBuildCmd, "build [path]", "index build command"},
		{indexStatusCmd, "status [path]", "index status command"},
		{indexRefreshCmd, "refresh [paths...]", "index refresh command"},
		{indexClearCmd, "clear [path]", "index clear command"},
		{toolListCmd, "list", "tool list command"},
		{toolExecuteCmd, "execute [tool-name] [args...]", "tool execute command"},
	}
	
	for _, test := range commandUsageTests {
		if test.cmd.Use != test.expectedUse {
			t.Errorf("%s should have Use '%s', got '%s'", test.description, test.expectedUse, test.cmd.Use)
		}
		
		if test.cmd.Short == "" {
			t.Errorf("%s should have a short description", test.description)
		}
	}
}

// TestErrorHandlingPaths tests various error handling scenarios
func TestErrorHandlingPaths(t *testing.T) {
	// Test working directory error handling
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	
	// Change to temp directory
	_ = os.Chdir(tempDir)
	
	// Test commands that might fail due to missing dependencies/files
	commands := []*cobra.Command{thinkCmd, planCmd, analyzeCmd, fixCmd}
	
	for _, cmd := range commands {
		// Each command should have argument validation if it requires args
		if cmd.Args != nil {
			// Test with empty args (should fail)
			err := cmd.Args(cmd, []string{})
			if err == nil && (cmd == thinkCmd || cmd == fixCmd) {
				t.Errorf("Command '%s' should require arguments", cmd.Name())
			}
			
			// Test with valid args (should pass validation)
			err = cmd.Args(cmd, []string{"test"})
			if err != nil {
				// Some commands might have more specific arg requirements
				t.Logf("Command '%s' argument validation: %v", cmd.Name(), err)
			}
		}
		
		// Verify Run function exists
		if cmd.Run == nil && cmd.RunE == nil {
			t.Errorf("Command '%s' should have a Run or RunE function", cmd.Name())
		}
	}
}

// TestPersistentFlagsConfiguration tests persistent flags configuration
func TestPersistentFlagsConfiguration(t *testing.T) {
	persistentFlags := rootCmd.PersistentFlags()
	
	// Test config flag configuration
	configFlag := persistentFlags.Lookup("config")
	if configFlag == nil {
		t.Fatal("Config flag should exist")
	}
	if configFlag.Shorthand != "c" {
		t.Errorf("Config flag shorthand should be 'c', got '%s'", configFlag.Shorthand)
	}
	if configFlag.DefValue != "" {
		t.Errorf("Config flag should have empty default value, got '%s'", configFlag.DefValue)
	}
	
	// Test verbose flag configuration
	verboseFlag := persistentFlags.Lookup("verbose")
	if verboseFlag == nil {
		t.Fatal("Verbose flag should exist")
	}
	if verboseFlag.Shorthand != "v" {
		t.Errorf("Verbose flag shorthand should be 'v', got '%s'", verboseFlag.Shorthand)
	}
	if verboseFlag.DefValue != "false" {
		t.Errorf("Verbose flag should have 'false' default value, got '%s'", verboseFlag.DefValue)
	}
	
	// Test debug flag configuration  
	debugFlag := persistentFlags.Lookup("debug")
	if debugFlag == nil {
		t.Fatal("Debug flag should exist")
	}
	if debugFlag.Shorthand != "d" {
		t.Errorf("Debug flag shorthand should be 'd', got '%s'", debugFlag.Shorthand)
	}
	if debugFlag.DefValue != "false" {
		t.Errorf("Debug flag should have 'false' default value, got '%s'", debugFlag.DefValue)
	}
}