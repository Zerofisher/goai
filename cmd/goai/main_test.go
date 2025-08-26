package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

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