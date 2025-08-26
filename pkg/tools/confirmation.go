package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

// ConsoleConfirmationHandler implements ConfirmationHandler using console input
type ConsoleConfirmationHandler struct{}

func NewConsoleConfirmationHandler() *ConsoleConfirmationHandler {
	return &ConsoleConfirmationHandler{}
}

func (c *ConsoleConfirmationHandler) RequestConfirmation(ctx context.Context, preview *ToolPreview) (bool, error) {
	// Print the preview information
	fmt.Printf("\n=== Tool Execution Preview ===\n")
	fmt.Printf("Tool: %s\n", preview.ToolName)
	fmt.Printf("Description: %s\n", preview.Description)
	
	if len(preview.ExpectedChanges) > 0 {
		fmt.Printf("\nExpected Changes:\n")
		for i, change := range preview.ExpectedChanges {
			fmt.Printf("  %d. [%s] %s\n", i+1, change.Type, change.Description)
			if change.Target != "" {
				fmt.Printf("     Target: %s\n", change.Target)
			}
			if change.Preview != "" && len(change.Preview) < 200 {
				fmt.Printf("     Preview: %s\n", change.Preview)
			}
		}
	}
	
	if preview.EstimatedTime != "" {
		fmt.Printf("Estimated Time: %s\n", preview.EstimatedTime)
	}
	
	fmt.Printf("\n")
	
	// Ask for confirmation
	for {
		fmt.Print("Do you want to proceed? [y/N]: ")
		
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("failed to read user input: %w", err)
		}
		
		input = strings.TrimSpace(strings.ToLower(input))
		
		switch input {
		case "y", "yes":
			return true, nil
		case "n", "no", "":
			return false, nil
		default:
			fmt.Println("Please enter 'y' for yes or 'n' for no.")
		}
	}
}

// MockConfirmationHandler for testing that always returns the configured response
type MockConfirmationHandler struct {
	AlwaysConfirm bool
}

func NewMockConfirmationHandler(alwaysConfirm bool) *MockConfirmationHandler {
	return &MockConfirmationHandler{
		AlwaysConfirm: alwaysConfirm,
	}
}

func (m *MockConfirmationHandler) RequestConfirmation(ctx context.Context, preview *ToolPreview) (bool, error) {
	return m.AlwaysConfirm, nil
}