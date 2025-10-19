package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Zerofisher/goai/pkg/agent"
	"github.com/Zerofisher/goai/pkg/config"
	"github.com/Zerofisher/goai/pkg/reminder"
	"github.com/Zerofisher/goai/pkg/todo"
	"github.com/Zerofisher/goai/pkg/tools/bash"
	"github.com/Zerofisher/goai/pkg/tools/edit"
	"github.com/Zerofisher/goai/pkg/tools/file"
	"github.com/Zerofisher/goai/pkg/tools/search"
	todotool "github.com/Zerofisher/goai/pkg/tools/todo"

	// Import LLM providers to register their factories
	_ "github.com/Zerofisher/goai/pkg/llm/anthropic"
	_ "github.com/Zerofisher/goai/pkg/llm/openai"
)

const (
	// Version information
	Version = "0.2.0"
	AppName = "GoAI Coder"
)

func main() {
	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nGracefully shutting down...")
		cancel()
		os.Exit(0)
	}()

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Create agent
	agent, err := createAgent(cfg)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		os.Exit(1)
	}

	// Warn if API key is not set
	if cfg.Model.APIKey == "" || cfg.Model.APIKey == "${OPENAI_API_KEY}" {
		fmt.Println("\n⚠️  WARNING: OPENAI_API_KEY environment variable is not set!")
		fmt.Println("   The agent will not be able to make LLM requests.")
		fmt.Println("   Please set your API key: export OPENAI_API_KEY='your-key-here'")
		fmt.Println()
	}

	// Print welcome message
	printWelcome()

	// Run interactive loop
	runInteractiveLoop(ctx, agent)
}

// loadConfig loads the configuration from file or environment
func loadConfig() (*config.Config, error) {
	// Check for config file
	configPaths := []string{
		"goai.yaml",
		"goai.yml",
		".goai.yaml",
		".goai.yml",
		filepath.Join(os.Getenv("HOME"), ".config", "goai", "config.yaml"),
	}

	var cfg *config.Config
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			loadedCfg, err := config.LoadFromFile(path)
			if err == nil {
				cfg = loadedCfg
				fmt.Printf("Loaded configuration from: %s\n", path)
				break
			}
		}
	}

	// Use default config if no file found
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Override with environment variables
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		cfg.Model.APIKey = apiKey
	}
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		cfg.Model.Name = model
	}

	// Set work directory to current directory if not set
	if cfg.WorkDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		cfg.WorkDir = cwd
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// createAgent creates and configures the agent
func createAgent(cfg *config.Config) (*agent.Agent, error) {
	// LLM client factories are auto-registered via init() functions
	// (e.g., OpenAI client in pkg/llm/openai.go)

	// Create agent
	a, err := agent.NewAgent(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Register tools
	if err := registerTools(a, cfg); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	return a, nil
}

// registerTools registers all available tools with the agent
func registerTools(a *agent.Agent, cfg *config.Config) error {
	dispatcher := a.GetDispatcher()
	enabledTools := []string{}

	// Register bash tool
	if isToolEnabled(cfg, "bash") {
		bashTool := bash.NewBashTool(cfg.WorkDir, 30*time.Second)
		if err := dispatcher.Register(bashTool); err != nil {
			return fmt.Errorf("failed to register bash tool: %w", err)
		}
		enabledTools = append(enabledTools, "bash")
	}

	// Register file tools (read, write, list) - enabled with "file" config
	if isToolEnabled(cfg, "file") {
		readTool := file.NewReadTool(cfg.WorkDir, 10*1024*1024) // 10MB max
		if err := dispatcher.Register(readTool); err != nil {
			return fmt.Errorf("failed to register read_file tool: %w", err)
		}

		writeTool := file.NewWriteTool(cfg.WorkDir, 10*1024*1024) // 10MB max
		if err := dispatcher.Register(writeTool); err != nil {
			return fmt.Errorf("failed to register write_file tool: %w", err)
		}

		listTool := file.NewListTool(cfg.WorkDir, 1000) // max 1000 items
		if err := dispatcher.Register(listTool); err != nil {
			return fmt.Errorf("failed to register list_files tool: %w", err)
		}
		enabledTools = append(enabledTools, "read_file", "write_file", "list_files")
	}

	// Register edit tool
	if isToolEnabled(cfg, "edit", "edit_file") {
		editTool := edit.NewEditTool(cfg.WorkDir)
		if err := dispatcher.Register(editTool); err != nil {
			return fmt.Errorf("failed to register edit_file tool: %w", err)
		}
		enabledTools = append(enabledTools, "edit_file")
	}

	// Register search tool
	if isToolEnabled(cfg, "search") {
		searchTool := search.NewSearchTool(cfg.WorkDir, nil)
		if err := dispatcher.Register(searchTool); err != nil {
			return fmt.Errorf("failed to register search tool: %w", err)
		}
		enabledTools = append(enabledTools, "search")
	}

	// Register todo tool
	if isToolEnabled(cfg, "todo") {
		todoMgr := todo.NewManager()
		reminderSys := reminder.NewSystem(3, 5) // Remind after 3 rounds, every 5 rounds
		todoToolInstance := todotool.NewTodoTool(todoMgr, reminderSys)
		if err := dispatcher.Register(todoToolInstance); err != nil {
			return fmt.Errorf("failed to register todo tool: %w", err)
		}
		enabledTools = append(enabledTools, "todo")
	}

	if len(enabledTools) > 0 {
		fmt.Printf("Tools enabled: %s\n", strings.Join(enabledTools, ", "))
	}

	return nil
}

// isToolEnabled checks if a tool is enabled in the configuration
func isToolEnabled(cfg *config.Config, names ...string) bool {
	for _, enabledTool := range cfg.Tools.Enabled {
		for _, name := range names {
			if enabledTool == name {
				return true
			}
		}
	}
	return false
}

// printWelcome prints the welcome message
func printWelcome() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("%s %s\n", AppName, Version)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("Welcome to GoAI Coder - Your intelligent programming assistant")
	fmt.Println()
	fmt.Println("Available commands:")
	fmt.Println("  /help    - Show this help message")
	fmt.Println("  /clear   - Clear the conversation")
	fmt.Println("  /stats   - Show agent statistics")
	fmt.Println("  /reset   - Reset the agent state")
	fmt.Println("  /exit    - Exit the application")
	fmt.Println()
	fmt.Println("Type your query or command and press Enter.")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println()
}

// printHelp prints the help message
func printHelp() {
	fmt.Println("\nAvailable Commands:")
	fmt.Println("==================")
	fmt.Println()
	fmt.Println("Special Commands:")
	fmt.Println("  /help, /h     - Show this help message")
	fmt.Println("  /clear, /c    - Clear the conversation history")
	fmt.Println("  /stats, /s    - Display agent statistics")
	fmt.Println("  /reset, /r    - Reset the agent state")
	fmt.Println("  /exit, /quit  - Exit the application")
	fmt.Println()
	fmt.Println("Usage Tips:")
	fmt.Println("  - Type your programming questions or requests naturally")
	fmt.Println("  - The agent can help with code generation, debugging, and explanations")
	fmt.Println("  - Use available tools to read/write files and execute commands")
	fmt.Println()
}