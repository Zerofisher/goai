package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Zerofisher/goai/pkg/tools"
)

func main() {
	fmt.Println("GoAI Tools System Example")
	fmt.Println("=========================")

	// Create tool factory
	factory := tools.NewToolFactory()

	// Create tool manager (without indexing for simplicity)
	manager, err := factory.CreateTestToolManager(nil, true) // Auto-confirm for demo
	if err != nil {
		log.Fatalf("Failed to create tool manager: %v", err)
	}

	ctx := context.Background()

	// Demonstrate listing available tools
	fmt.Println("\n1. Available Tools:")
	allTools := manager.ListTools("")
	for _, tool := range allTools {
		fmt.Printf("   - %s (%s): %s\n", tool.Name(), tool.Category(), tool.Description())
	}

	// Demonstrate file operations
	fmt.Println("\n2. File Operations Demo:")
	
	// Create a temporary file
	tempFile := "/tmp/goai-tools-demo.txt"
	writeParams := map[string]any{
		"path":    tempFile,
		"content": "Hello from GoAI Tools System!\nThis is a demonstration of file operations.",
	}

	fmt.Printf("Writing to %s...\n", tempFile)
	result, err := manager.ExecuteTool(ctx, "writeFile", writeParams)
	if err != nil {
		log.Printf("Write failed: %v", err)
	} else if result.Success {
		fmt.Printf("✓ Write successful: %s\n", result.Output)
	} else {
		fmt.Printf("✗ Write failed: %s\n", result.Error)
	}

	// Read the file back
	readParams := map[string]any{
		"path": tempFile,
	}

	fmt.Printf("Reading from %s...\n", tempFile)
	result, err = manager.ExecuteTool(ctx, "readFile", readParams)
	if err != nil {
		log.Printf("Read failed: %v", err)
	} else if result.Success {
		fmt.Printf("✓ Read successful. Content:\n%s\n", result.Data)
	} else {
		fmt.Printf("✗ Read failed: %s\n", result.Error)
	}

	// Edit the file
	editParams := map[string]any{
		"path":       tempFile,
		"oldContent": "Hello from GoAI Tools System!",
		"newContent": "Greetings from the GoAI Tools System!",
	}

	fmt.Printf("Editing %s...\n", tempFile)
	result, err = manager.ExecuteTool(ctx, "editFile", editParams)
	if err != nil {
		log.Printf("Edit failed: %v", err)
	} else if result.Success {
		fmt.Printf("✓ Edit successful: %s\n", result.Output)
	} else {
		fmt.Printf("✗ Edit failed: %s\n", result.Error)
	}

	// Demonstrate listing files
	fmt.Println("\n3. File Listing Demo:")
	listParams := map[string]any{
		"path":      "/tmp",
		"pattern":   "goai-*",
		"maxResults": 5,
	}

	result, err = manager.ExecuteTool(ctx, "listFiles", listParams)
	if err != nil {
		log.Printf("List failed: %v", err)
	} else if result.Success {
		if fileList, ok := result.Data.([]map[string]any); ok {
			fmt.Printf("✓ Found %d matching files:\n", len(fileList))
			for _, file := range fileList {
				fmt.Printf("   - %s (%d bytes)\n", file["name"], int64(file["size"].(int64)))
			}
		}
	} else {
		fmt.Printf("✗ List failed: %s\n", result.Error)
	}

	// Demonstrate system command (safe example)
	fmt.Println("\n4. System Command Demo:")
	cmdParams := map[string]any{
		"command": "echo",
		"args":    []string{"Hello from system command!"},
		"timeout": 5,
	}

	result, err = manager.ExecuteTool(ctx, "runCommand", cmdParams)
	if err != nil {
		log.Printf("Command failed: %v", err)
	} else if result.Success {
		if cmdResult, ok := result.Data.(map[string]any); ok {
			fmt.Printf("✓ Command executed successfully:\n")
			fmt.Printf("   stdout: %s", cmdResult["stdout"])
		}
	} else {
		fmt.Printf("✗ Command failed: %s\n", result.Error)
	}

	// Demonstrate preview functionality
	fmt.Println("\n5. Preview Demo:")
	previewParams := map[string]any{
		"path":    "/tmp/goai-preview-demo.txt",
		"content": "This is a preview demonstration",
	}

	preview, err := manager.ExecuteWithPreview(ctx, "writeFile", previewParams)
	if err != nil {
		log.Printf("Preview failed: %v", err)
	} else {
		fmt.Printf("✓ Preview for writeFile operation:\n")
		fmt.Printf("   Description: %s\n", preview.Description)
		fmt.Printf("   Requires Confirmation: %v\n", preview.RequiresConfirmation)
		if len(preview.ExpectedChanges) > 0 {
			fmt.Printf("   Expected Changes:\n")
			for i, change := range preview.ExpectedChanges {
				fmt.Printf("     %d. [%s] %s -> %s\n", i+1, change.Type, change.Description, change.Target)
			}
		}
	}

	// Clean up
	fmt.Println("\n6. Cleanup:")
	if err := os.Remove(tempFile); err != nil {
		log.Printf("Warning: Failed to clean up %s: %v", tempFile, err)
	} else {
		fmt.Printf("✓ Cleaned up temporary file: %s\n", tempFile)
	}

	fmt.Println("\nGoAI Tools System demonstration completed!")
}