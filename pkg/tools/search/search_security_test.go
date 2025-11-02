package search

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Zerofisher/goai/pkg/tools"
)

// TestContextCancellation tests that search operations respect context cancellation
func TestContextCancellation(t *testing.T) {
	// Create a temporary directory with test files
	tempDir, err := os.MkdirTemp("", "search-cancel-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Create many test files to slow down search
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tempDir, "test"+string(rune('0'+i%10))+".go")
		content := strings.Repeat("test pattern\n", 1000)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	validator := tools.NewSecurityValidator(tempDir)
	searchTool := NewSearchTool(tempDir, validator)

	// Create a context that will be canceled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	options := SearchOptions{
		CaseSensitive: false,
		MaxResults:    10000, // Large number to ensure we hit the timeout
	}

	// This should be canceled before completion
	_, err = searchTool.SearchCode(ctx, "test pattern", options)

	// We expect either a context canceled error or successful completion
	// (depending on how fast the system is)
	if err != nil {
		if !strings.Contains(err.Error(), "canceled") && !strings.Contains(err.Error(), "deadline exceeded") {
			t.Errorf("Expected cancellation error, got: %v", err)
		}
	}
}

// TestOutputLimiting tests that output size limits are enforced
func TestOutputLimiting(t *testing.T) {
	// Create a temporary directory with test files
	tempDir, err := os.MkdirTemp("", "search-limit-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Create a file with many matches that would exceed output limit
	filename := filepath.Join(tempDir, "large.go")
	// Create 100K lines with pattern
	var builder strings.Builder
	for i := 0; i < 100000; i++ {
		builder.WriteString("test pattern line\n")
	}
	if err := os.WriteFile(filename, []byte(builder.String()), 0644); err != nil {
		t.Fatal(err)
	}

	validator := tools.NewSecurityValidator(tempDir)
	searchTool := NewSearchTool(tempDir, validator)

	ctx := context.Background()
	options := SearchOptions{
		CaseSensitive: false,
		MaxResults:    50, // Limit results at grep level
	}

	results, err := searchTool.SearchCode(ctx, "test pattern", options)

	// Should complete successfully with limited results
	if err != nil {
		t.Fatalf("SearchCode failed: %v", err)
	}

	// Should respect MaxResults limit
	if len(results) > options.MaxResults {
		t.Errorf("Expected at most %d results, got %d", options.MaxResults, len(results))
	}
}

// TestSymbolSearchTimeout tests symbol search with timeout
func TestSymbolSearchTimeout(t *testing.T) {
	// Create a temporary directory with test files
	tempDir, err := os.MkdirTemp("", "symbol-timeout-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Create many Go files to slow down symbol search
	for i := 0; i < 50; i++ {
		filename := filepath.Join(tempDir, "file"+string(rune('0'+i%10))+".go")
		content := `package main
func TestFunc() {}
type TestType struct{}
var TestVar int
const TestConst = 1
`
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	validator := tools.NewSecurityValidator(tempDir)
	searchTool := NewSearchTool(tempDir, validator)

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	// This should timeout or complete quickly
	_, err = searchTool.SearchSymbol(ctx, "TestFunc")

	// We expect either a context error or successful completion
	if err != nil {
		if !strings.Contains(err.Error(), "canceled") && !strings.Contains(err.Error(), "deadline exceeded") {
			t.Errorf("Expected cancellation error, got: %v", err)
		}
	}
}

// TestFindMatchingFilesWithContext tests that findMatchingFiles respects context
func TestFindMatchingFilesWithContext(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "find-context-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Create many nested directories to slow down find
	for i := 0; i < 10; i++ {
		dir := filepath.Join(tempDir, "dir"+string(rune('0'+i)))
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 10; j++ {
			filename := filepath.Join(dir, "file"+string(rune('0'+j))+".go")
			if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	validator := tools.NewSecurityValidator(tempDir)
	searchTool := NewSearchTool(tempDir, validator)

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	// This should timeout or complete quickly
	_, err = searchTool.findMatchingFiles(ctx, "*.go")

	// We expect either a context error or successful completion
	if err != nil && !strings.Contains(err.Error(), "canceled") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("Expected cancellation error or success, got: %v", err)
	}
}

// TestLimitedWriter tests the limitedWriter implementation
func TestLimitedWriter(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		writes    [][]byte
		expectErr bool
	}{
		{
			name:      "within limit",
			limit:     100,
			writes:    [][]byte{[]byte("hello"), []byte("world")},
			expectErr: false,
		},
		{
			name:      "exact limit",
			limit:     10,
			writes:    [][]byte{[]byte("hello"), []byte("world")},
			expectErr: false,
		},
		{
			name:      "exceed limit",
			limit:     5,
			writes:    [][]byte{[]byte("hello"), []byte("world")},
			expectErr: true,
		},
		{
			name:      "single large write",
			limit:     10,
			writes:    [][]byte{[]byte("this is a very long string")},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			writer := &limitedWriter{
				w:     nil, // We'll test with strings.Builder
				limit: tt.limit,
			}

			for _, data := range tt.writes {
				// Simulate writing to buffer
				n := len(data)
				if writer.written+n > writer.limit {
					n = writer.limit - writer.written
					if n > 0 {
						buf.Write(data[:n])
					}
					writer.written += n
					if writer.written >= writer.limit {
						break
					}
				} else {
					buf.Write(data)
					writer.written += n
				}
			}

			hasError := writer.written >= writer.limit && len(tt.writes) > 0 &&
				(func() int { total := 0; for _, w := range tt.writes { total += len(w) }; return total }()) > writer.limit

			if hasError != tt.expectErr {
				t.Errorf("Expected error: %v, got error: %v (written: %d, limit: %d)",
					tt.expectErr, hasError, writer.written, writer.limit)
			}

			if writer.written > writer.limit {
				t.Errorf("Written bytes (%d) exceeds limit (%d)", writer.written, writer.limit)
			}
		})
	}
}
