package edit

import (
	"fmt"
	"os"
	"strings"
)

// Strategy defines the interface for edit strategies.
type Strategy interface {
	// Name returns the strategy name
	Name() string
	// Validate validates the input parameters for this strategy
	Validate(input map[string]interface{}) error
	// Execute performs the edit operation
	Execute(path string, input map[string]interface{}) (*EditResult, error)
}

// ReplaceStrategy implements text replacement with line number or text matching.
type ReplaceStrategy struct {
	diff *DiffGenerator
}

// NewReplaceStrategy creates a new replace strategy.
func NewReplaceStrategy() *ReplaceStrategy {
	return &ReplaceStrategy{
		diff: NewDiffGenerator(),
	}
}

// Name returns the strategy name.
func (s *ReplaceStrategy) Name() string {
	return "replace"
}

// Validate validates replace strategy parameters.
func (s *ReplaceStrategy) Validate(input map[string]interface{}) error {
	// Must have old_text
	if _, ok := input["old_text"]; !ok {
		return fmt.Errorf("replace strategy requires 'old_text' parameter")
	}

	// Must have new_text
	if _, ok := input["new_text"]; !ok {
		return fmt.Errorf("replace strategy requires 'new_text' parameter")
	}

	// Validate line_start/line_end if provided
	if startRaw, ok := input["line_start"]; ok {
		if startFloat, ok := startRaw.(float64); ok {
			if startFloat < 1 {
				return fmt.Errorf("line_start must be >= 1")
			}
		} else {
			return fmt.Errorf("line_start must be a number")
		}
	}

	if endRaw, ok := input["line_end"]; ok {
		if endFloat, ok := endRaw.(float64); ok {
			if endFloat < 1 {
				return fmt.Errorf("line_end must be >= 1")
			}
		} else {
			return fmt.Errorf("line_end must be a number")
		}
	}

	return nil
}

// Execute performs the replace operation.
func (s *ReplaceStrategy) Execute(path string, input map[string]interface{}) (*EditResult, error) {
	oldText, _ := input["old_text"].(string)
	newText, _ := input["new_text"].(string)

	// Read original content
	originalContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	originalStr := string(originalContent)
	var newContent string
	var lineStart, lineEnd int
	var replaceAll bool

	// Check for line range constraint
	if startRaw, ok := input["line_start"]; ok {
		lineStart = int(startRaw.(float64))
		if endRaw, ok := input["line_end"]; ok {
			lineEnd = int(endRaw.(float64))
		} else {
			lineEnd = lineStart
		}

		// Replace within line range
		newContent, err = s.replaceInLineRange(originalStr, oldText, newText, lineStart, lineEnd)
		if err != nil {
			return nil, err
		}
	} else {
		// Check if replace_all is set
		if replaceAllRaw, ok := input["replace_all"]; ok {
			if replaceAllBool, ok := replaceAllRaw.(bool); ok {
				replaceAll = replaceAllBool
			}
		}

		// Replace in entire file
		if replaceAll {
			newContent = strings.ReplaceAll(originalStr, oldText, newText)
		} else {
			newContent = strings.Replace(originalStr, oldText, newText, 1)
		}

		// Check if any changes were made
		if originalStr == newContent {
			return nil, fmt.Errorf("text not found: %s", oldText)
		}
	}

	// Write updated content
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Count lines modified
	linesModified := countLinesModified(originalStr, newContent)

	// Generate diff
	diff := s.diff.GenerateDiff(originalStr, newContent, path)

	return &EditResult{
		Path:          path,
		Strategy:      "replace",
		LinesModified: linesModified,
		Diff:          diff,
	}, nil
}

// replaceInLineRange replaces text only within the specified line range.
func (s *ReplaceStrategy) replaceInLineRange(content, oldText, newText string, lineStart, lineEnd int) (string, error) {
	lines := strings.Split(content, "\n")

	if lineStart < 1 || lineStart > len(lines) {
		return "", fmt.Errorf("line_start %d out of range (1-%d)", lineStart, len(lines))
	}

	if lineEnd < lineStart || lineEnd > len(lines) {
		lineEnd = len(lines)
	}

	// Replace only in the specified range (convert to 0-based)
	for i := lineStart - 1; i < lineEnd; i++ {
		lines[i] = strings.Replace(lines[i], oldText, newText, -1)
	}

	return strings.Join(lines, "\n"), nil
}

// InsertStrategy implements text insertion at line number or after anchor.
type InsertStrategy struct{}

// NewInsertStrategy creates a new insert strategy.
func NewInsertStrategy() *InsertStrategy {
	return &InsertStrategy{}
}

// Name returns the strategy name.
func (s *InsertStrategy) Name() string {
	return "insert"
}

// Validate validates insert strategy parameters.
func (s *InsertStrategy) Validate(input map[string]interface{}) error {
	// Must have text to insert
	if _, ok := input["text"]; !ok {
		return fmt.Errorf("insert strategy requires 'text' parameter")
	}

	// Must have either line or after_anchor
	if _, hasLine := input["line"]; !hasLine {
		if _, hasAnchor := input["after_anchor"]; !hasAnchor {
			return fmt.Errorf("insert strategy requires either 'line' or 'after_anchor' parameter")
		}
	}

	// Validate line if provided
	if lineRaw, ok := input["line"]; ok {
		if lineFloat, ok := lineRaw.(float64); ok {
			if lineFloat < 0 {
				return fmt.Errorf("line must be >= 0")
			}
		} else {
			return fmt.Errorf("line must be a number")
		}
	}

	return nil
}

// Execute performs the insert operation.
func (s *InsertStrategy) Execute(path string, input map[string]interface{}) (*EditResult, error) {
	text, _ := input["text"].(string)

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var insertLine int
	var found bool

	// Determine insertion point
	if lineRaw, ok := input["line"]; ok {
		insertLine = int(lineRaw.(float64))
		found = true
	} else if anchor, ok := input["after_anchor"].(string); ok {
		// Find the anchor line
		for i, line := range lines {
			if strings.Contains(line, anchor) {
				insertLine = i + 2 // Insert after this line (1-based + 1)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("anchor not found: %s", anchor)
		}
	}

	if !found {
		return nil, fmt.Errorf("no insertion point specified")
	}

	// Perform insertion
	if insertLine < 0 {
		insertLine = 0
	} else if insertLine > len(lines) {
		insertLine = len(lines)
	}

	// Convert 1-based to 0-based for array indexing (unless it's 0, meaning prepend)
	insertIdx := insertLine
	if insertLine > 0 {
		insertIdx = insertLine - 1
	}

	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:insertIdx]...)
	newLines = append(newLines, text)
	newLines = append(newLines, lines[insertIdx:]...)

	// Write back
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &EditResult{
		Path:          path,
		Strategy:      "insert",
		LinesModified: 1,
	}, nil
}

// AnchoredStrategy implements context-aware editing with before/after anchors.
type AnchoredStrategy struct {
	diff *DiffGenerator
}

// NewAnchoredStrategy creates a new anchored strategy.
func NewAnchoredStrategy() *AnchoredStrategy {
	return &AnchoredStrategy{
		diff: NewDiffGenerator(),
	}
}

// Name returns the strategy name.
func (s *AnchoredStrategy) Name() string {
	return "anchored"
}

// Validate validates anchored strategy parameters.
func (s *AnchoredStrategy) Validate(input map[string]interface{}) error {
	// Must have old_text and new_text
	if _, ok := input["old_text"]; !ok {
		return fmt.Errorf("anchored strategy requires 'old_text' parameter")
	}

	if _, ok := input["new_text"]; !ok {
		return fmt.Errorf("anchored strategy requires 'new_text' parameter")
	}

	// Must have at least one anchor
	_, hasBeforeAnchor := input["before_anchor"]
	_, hasAfterAnchor := input["after_anchor"]

	if !hasBeforeAnchor && !hasAfterAnchor {
		return fmt.Errorf("anchored strategy requires at least one of 'before_anchor' or 'after_anchor'")
	}

	return nil
}

// Execute performs the anchored edit operation.
func (s *AnchoredStrategy) Execute(path string, input map[string]interface{}) (*EditResult, error) {
	oldText, _ := input["old_text"].(string)
	newText, _ := input["new_text"].(string)
	beforeAnchor, hasBeforeAnchor := input["before_anchor"].(string)
	afterAnchor, hasAfterAnchor := input["after_anchor"].(string)

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)

	// Find the region constrained by anchors
	region, err := s.findAnchoredRegion(originalContent, beforeAnchor, hasBeforeAnchor, afterAnchor, hasAfterAnchor)
	if err != nil {
		return nil, err
	}

	// Replace within the region
	if !strings.Contains(region, oldText) {
		return nil, fmt.Errorf("old_text not found within anchored region")
	}

	newRegion := strings.Replace(region, oldText, newText, 1)

	// Reconstruct the full content
	var newContent string
	if hasBeforeAnchor && hasAfterAnchor {
		beforeIdx := strings.Index(originalContent, beforeAnchor)
		afterIdx := strings.Index(originalContent, afterAnchor)
		prefix := originalContent[:beforeIdx+len(beforeAnchor)]
		suffix := originalContent[afterIdx:]
		newContent = prefix + "\n" + newRegion + "\n" + suffix
	} else if hasBeforeAnchor {
		beforeIdx := strings.Index(originalContent, beforeAnchor)
		prefix := originalContent[:beforeIdx+len(beforeAnchor)]
		newContent = prefix + "\n" + newRegion
	} else if hasAfterAnchor {
		afterIdx := strings.Index(originalContent, afterAnchor)
		suffix := originalContent[afterIdx:]
		newContent = newRegion + "\n" + suffix
	}

	// Write back
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	// Count lines modified
	linesModified := countLinesModified(originalContent, newContent)

	// Generate diff
	diff := s.diff.GenerateDiff(originalContent, newContent, path)

	return &EditResult{
		Path:          path,
		Strategy:      "anchored",
		LinesModified: linesModified,
		Diff:          diff,
	}, nil
}

// findAnchoredRegion finds the text region between anchors.
func (s *AnchoredStrategy) findAnchoredRegion(content, beforeAnchor string, hasBeforeAnchor bool, afterAnchor string, hasAfterAnchor bool) (string, error) {
	if hasBeforeAnchor && hasAfterAnchor {
		// Find region between both anchors
		beforeIdx := strings.Index(content, beforeAnchor)
		if beforeIdx == -1 {
			return "", fmt.Errorf("before_anchor not found: %s", beforeAnchor)
		}

		afterIdx := strings.Index(content[beforeIdx:], afterAnchor)
		if afterIdx == -1 {
			return "", fmt.Errorf("after_anchor not found: %s", afterAnchor)
		}

		startIdx := beforeIdx + len(beforeAnchor)
		endIdx := beforeIdx + afterIdx

		return strings.TrimSpace(content[startIdx:endIdx]), nil
	} else if hasBeforeAnchor {
		// Find region after before_anchor to end
		beforeIdx := strings.Index(content, beforeAnchor)
		if beforeIdx == -1 {
			return "", fmt.Errorf("before_anchor not found: %s", beforeAnchor)
		}

		return strings.TrimSpace(content[beforeIdx+len(beforeAnchor):]), nil
	} else if hasAfterAnchor {
		// Find region from beginning to after_anchor
		afterIdx := strings.Index(content, afterAnchor)
		if afterIdx == -1 {
			return "", fmt.Errorf("after_anchor not found: %s", afterAnchor)
		}

		return strings.TrimSpace(content[:afterIdx]), nil
	}

	return "", fmt.Errorf("no anchors specified")
}

// ApplyPatchStrategy implements unified diff patch application.
type ApplyPatchStrategy struct{}

// NewApplyPatchStrategy creates a new apply_patch strategy.
func NewApplyPatchStrategy() *ApplyPatchStrategy {
	return &ApplyPatchStrategy{}
}

// Name returns the strategy name.
func (s *ApplyPatchStrategy) Name() string {
	return "apply_patch"
}

// Validate validates apply_patch strategy parameters.
func (s *ApplyPatchStrategy) Validate(input map[string]interface{}) error {
	if _, ok := input["patch"]; !ok {
		return fmt.Errorf("apply_patch strategy requires 'patch' parameter")
	}

	return nil
}

// Execute applies a unified diff patch.
func (s *ApplyPatchStrategy) Execute(path string, input map[string]interface{}) (*EditResult, error) {
	patch, _ := input["patch"].(string)

	// Read original file
	originalContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Apply the patch
	newContent, linesModified, err := s.applyUnifiedDiff(string(originalContent), patch)
	if err != nil {
		return nil, fmt.Errorf("failed to apply patch: %w", err)
	}

	// Write back
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &EditResult{
		Path:          path,
		Strategy:      "apply_patch",
		LinesModified: linesModified,
	}, nil
}

// applyUnifiedDiff applies a unified diff patch to content.
func (s *ApplyPatchStrategy) applyUnifiedDiff(content, patch string) (string, int, error) {
	lines := strings.Split(content, "\n")
	patchLines := strings.Split(patch, "\n")

	// Parse the patch
	linesModified := 0

	// Skip header lines until we find @@ line
	i := 0
	for i < len(patchLines) {
		if strings.HasPrefix(patchLines[i], "@@") {
			i++
			break
		}
		i++
	}

	// Apply hunks
	newLines := make([]string, 0, len(lines))
	lineIdx := 0

	for i < len(patchLines) {
		line := patchLines[i]

		if strings.HasPrefix(line, "@@") {
			// New hunk
			i++
			continue
		}

		if strings.HasPrefix(line, "+") {
			// Add line
			newLines = append(newLines, strings.TrimPrefix(line, "+"))
			linesModified++
		} else if strings.HasPrefix(line, "-") {
			// Remove line (skip)
			lineIdx++
			linesModified++
		} else if strings.HasPrefix(line, " ") {
			// Context line
			if lineIdx < len(lines) {
				newLines = append(newLines, lines[lineIdx])
				lineIdx++
			}
		}

		i++
	}

	// Add remaining lines
	for lineIdx < len(lines) {
		newLines = append(newLines, lines[lineIdx])
		lineIdx++
	}

	return strings.Join(newLines, "\n"), linesModified, nil
}

// countLinesModified counts how many lines were changed between two strings.
func countLinesModified(original, new string) int {
	originalLines := strings.Split(original, "\n")
	newLines := strings.Split(new, "\n")

	modified := 0
	maxLen := len(originalLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen; i++ {
		if i >= len(originalLines) || i >= len(newLines) {
			modified++
		} else if originalLines[i] != newLines[i] {
			modified++
		}
	}

	return modified
}
