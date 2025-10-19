package edit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReplaceStrategy tests the replace strategy.
func TestReplaceStrategy(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	content := "Hello World\nHello Universe\nGoodbye World"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	// Test simple replace
	input := map[string]interface{}{
		"path":     "test.txt",
		"strategy": "replace",
		"old_text": "World",
		"new_text": "Earth",
	}

	result, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	var resp ToolResponse
	if err := json.Unmarshal([]byte(result), &resp); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if !resp.Ok {
		t.Errorf("Expected ok=true, got error: %s", resp.Error)
	}

	// Verify file was modified
	newContent, _ := os.ReadFile(testFile)
	if !strings.Contains(string(newContent), "Hello Earth") {
		t.Error("Replace did not work correctly")
	}
}

// TestReplaceStrategy_ReplaceAll tests replacing all occurrences.
func TestReplaceStrategy_ReplaceAll(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	content := "foo bar foo baz foo"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	input := map[string]interface{}{
		"path":        "test.txt",
		"strategy":    "replace",
		"old_text":    "foo",
		"new_text":    "FOO",
		"replace_all": true,
	}

	result, _ := tool.Execute(ctx, input)

	var resp ToolResponse
	json.Unmarshal([]byte(result), &resp)

	if !resp.Ok {
		t.Errorf("Replace all failed: %s", resp.Error)
	}

	newContent, _ := os.ReadFile(testFile)
	expected := "FOO bar FOO baz FOO"
	if string(newContent) != expected {
		t.Errorf("Content = %q, want %q", string(newContent), expected)
	}
}

// TestReplaceStrategy_LineRange tests replace within line range.
func TestReplaceStrategy_LineRange(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	content := "line 1: foo\nline 2: foo\nline 3: foo"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	input := map[string]interface{}{
		"path":       "test.txt",
		"strategy":   "replace",
		"old_text":   "foo",
		"new_text":   "bar",
		"line_start": float64(2),
		"line_end":   float64(2),
	}

	result, _ := tool.Execute(ctx, input)
	var resp ToolResponse
	json.Unmarshal([]byte(result), &resp)

	if !resp.Ok {
		t.Errorf("Line range replace failed: %s", resp.Error)
	}

	newContent, _ := os.ReadFile(testFile)
	expected := "line 1: foo\nline 2: bar\nline 3: foo"
	if string(newContent) != expected {
		t.Errorf("Content = %q, want %q", string(newContent), expected)
	}
}

// TestInsertStrategy tests the insert strategy.
func TestInsertStrategy(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	content := "Line 1\nLine 3"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	input := map[string]interface{}{
		"path":     "test.txt",
		"strategy": "insert",
		"text":     "Line 2",
		"line":     float64(2),
	}

	result, _ := tool.Execute(ctx, input)
	var resp ToolResponse
	json.Unmarshal([]byte(result), &resp)

	if !resp.Ok {
		t.Errorf("Insert failed: %s", resp.Error)
	}

	newContent, _ := os.ReadFile(testFile)
	if !strings.Contains(string(newContent), "Line 2") {
		t.Error("Insert did not work")
	}
}

// TestInsertStrategy_AfterAnchor tests inserting after an anchor.
func TestInsertStrategy_AfterAnchor(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	content := "# Header\n\n# Footer"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	input := map[string]interface{}{
		"path":         "test.txt",
		"strategy":     "insert",
		"text":         "New content",
		"after_anchor": "# Header",
	}

	result, _ := tool.Execute(ctx, input)
	var resp ToolResponse
	json.Unmarshal([]byte(result), &resp)

	if !resp.Ok {
		t.Errorf("Insert after anchor failed: %s", resp.Error)
	}

	newContent, _ := os.ReadFile(testFile)
	lines := strings.Split(string(newContent), "\n")
	if len(lines) < 3 || lines[1] != "New content" {
		t.Error("Insert after anchor did not work correctly")
	}
}

// TestAnchoredStrategy tests the anchored strategy.
func TestAnchoredStrategy(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	content := "Start\n# BEGIN\nOld content\n# END\nFinish"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	input := map[string]interface{}{
		"path":          "test.txt",
		"strategy":      "anchored",
		"old_text":      "Old content",
		"new_text":      "New content",
		"before_anchor": "# BEGIN",
		"after_anchor":  "# END",
	}

	result, _ := tool.Execute(ctx, input)
	var resp ToolResponse
	json.Unmarshal([]byte(result), &resp)

	if !resp.Ok {
		t.Errorf("Anchored edit failed: %s", resp.Error)
	}

	newContent, _ := os.ReadFile(testFile)
	if !strings.Contains(string(newContent), "New content") {
		t.Error("Anchored edit did not work")
	}
	if strings.Contains(string(newContent), "Old content") {
		t.Error("Old content still exists after anchored edit")
	}
}

// TestApplyPatchStrategy tests the apply_patch strategy.
func TestApplyPatchStrategy(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	content := "Line 1\nLine 2\nLine 3\nLine 4"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	// Create a simple unified diff patch
	patch := `@@ -1,4 +1,4 @@
 Line 1
-Line 2
+Line 2 modified
 Line 3
 Line 4`

	input := map[string]interface{}{
		"path":     "test.txt",
		"strategy": "apply_patch",
		"patch":    patch,
	}

	result, _ := tool.Execute(ctx, input)
	var resp ToolResponse
	json.Unmarshal([]byte(result), &resp)

	if !resp.Ok {
		t.Errorf("Apply patch failed: %s", resp.Error)
	}

	newContent, _ := os.ReadFile(testFile)
	if !strings.Contains(string(newContent), "Line 2 modified") {
		t.Error("Patch was not applied correctly")
	}
}

// TestBackupAndRestore tests backup creation and restoration on failure.
func TestBackupAndRestore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	originalContent := "Original content"
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	// Try to replace non-existent text (should fail)
	input := map[string]interface{}{
		"path":          "test.txt",
		"strategy":      "replace",
		"old_text":      "NonExistent",
		"new_text":      "Something",
		"create_backup": true,
	}

	result, _ := tool.Execute(ctx, input)
	var resp ToolResponse
	json.Unmarshal([]byte(result), &resp)

	// Should fail
	if resp.Ok {
		t.Error("Expected operation to fail, but it succeeded")
	}

	// Verify original content is still intact
	restoredContent, _ := os.ReadFile(testFile)
	if string(restoredContent) != originalContent {
		t.Error("Content was not restored after failed operation")
	}
}

// TestConflictDetection tests conflict detection.
func TestConflictDetection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	content := "foo bar foo baz foo"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	// Replace without replace_all should detect conflict
	input := map[string]interface{}{
		"path":             "test.txt",
		"strategy":         "replace",
		"old_text":         "foo",
		"new_text":         "FOO",
		"detect_conflicts": true,
	}

	result, _ := tool.Execute(ctx, input)
	var resp ToolResponse
	json.Unmarshal([]byte(result), &resp)

	if !resp.Ok {
		t.Fatalf("Operation failed: %s", resp.Error)
	}

	// Check if conflicts were detected
	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data is not a map")
	}

	conflictsRaw, ok := dataMap["conflicts"]
	if !ok {
		t.Error("No conflicts field in response")
	}

	conflicts, ok := conflictsRaw.([]interface{})
	if !ok || len(conflicts) == 0 {
		t.Error("Expected conflicts to be detected")
	}
}

// TestPathSecurity tests path validation.
func TestPathSecurity(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	dangerousPaths := []string{
		"../../../etc/passwd",
		"../../.ssh/id_rsa",
		"/etc/passwd",
	}

	for _, path := range dangerousPaths {
		input := map[string]interface{}{
			"path":     path,
			"strategy": "replace",
			"old_text": "test",
			"new_text": "test",
		}

		result, _ := tool.Execute(ctx, input)
		var resp ToolResponse
		json.Unmarshal([]byte(result), &resp)

		if resp.Ok {
			t.Errorf("Expected security error for path %s, but operation succeeded", path)
		}
	}
}

// TestJSONOutputFormat tests that all operations return proper JSON.
func TestJSONOutputFormat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "edit_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(tempDir)
	ctx := context.Background()

	strategies := []string{"replace", "insert", "anchored"}

	for _, strategy := range strategies {
		var input map[string]interface{}

		switch strategy {
		case "replace":
			input = map[string]interface{}{
				"path":     "test.txt",
				"strategy": strategy,
				"old_text": "content",
				"new_text": "new content",
			}
		case "insert":
			input = map[string]interface{}{
				"path":     "test.txt",
				"strategy": strategy,
				"text":     "inserted",
				"line":     float64(1),
			}
		case "anchored":
			input = map[string]interface{}{
				"path":          "test.txt",
				"strategy":      strategy,
				"old_text":      "content",
				"new_text":      "modified",
				"before_anchor": "content",
			}
		}

		result, _ := tool.Execute(ctx, input)

		var resp ToolResponse
		if err := json.Unmarshal([]byte(result), &resp); err != nil {
			t.Errorf("Strategy %s did not return valid JSON: %v", strategy, err)
			continue
		}

		// Validate response structure
		if resp.Ok {
			if resp.Data == nil {
				t.Errorf("Strategy %s: ok=true but data is nil", strategy)
			}
			if resp.Summary == "" {
				t.Errorf("Strategy %s: summary is empty", strategy)
			}
		} else {
			if resp.Error == "" {
				t.Errorf("Strategy %s: ok=false but error is empty", strategy)
			}
		}
	}
}
