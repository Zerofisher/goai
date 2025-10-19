package edit

import (
	"fmt"
	"strings"
)

// DiffGenerator generates unified diff output for file changes.
type DiffGenerator struct {
	contextLines int
}

// NewDiffGenerator creates a new diff generator.
func NewDiffGenerator() *DiffGenerator {
	return &DiffGenerator{
		contextLines: 3, // Default context lines
	}
}

// GenerateDiff generates a unified diff between two strings.
func (g *DiffGenerator) GenerateDiff(original, modified, filename string) string {
	originalLines := strings.Split(original, "\n")
	modifiedLines := strings.Split(modified, "\n")

	// Use simple line-by-line comparison for now
	// In production, you might want to use a proper diff algorithm like Myers' algorithm
	changes := g.findChanges(originalLines, modifiedLines)

	if len(changes) == 0 {
		return "No changes detected"
	}

	return g.formatUnifiedDiff(changes, filename, originalLines, modifiedLines)
}

// Change represents a change in the file.
type Change struct {
	Type      string // "add", "delete", "modify"
	StartLine int
	EndLine   int
	Lines     []string
	NewLines  []string // For modifications
}

// findChanges finds the differences between two sets of lines.
func (g *DiffGenerator) findChanges(original, modified []string) []Change {
	var changes []Change

	// Simple algorithm: find longest common subsequence
	// For simplicity, we'll use a basic approach

	i, j := 0, 0
	for i < len(original) || j < len(modified) {
		if i >= len(original) {
			// All remaining modified lines are additions
			changes = append(changes, Change{
				Type:      "add",
				StartLine: i,
				Lines:     modified[j:],
			})
			break
		} else if j >= len(modified) {
			// All remaining original lines are deletions
			changes = append(changes, Change{
				Type:      "delete",
				StartLine: i,
				EndLine:   len(original) - 1,
				Lines:     original[i:],
			})
			break
		} else if original[i] == modified[j] {
			// Lines match, move forward
			i++
			j++
		} else {
			// Find the extent of the change
			changeStart := i
			modStart := j

			// Look ahead to find where lines sync up again
			syncFound := false
			for di := 0; di <= 5 && changeStart+di < len(original); di++ {
				for dj := 0; dj <= 5 && modStart+dj < len(modified); dj++ {
					if original[changeStart+di] == modified[modStart+dj] {
						// Found sync point
						if di > 0 {
							changes = append(changes, Change{
								Type:      "delete",
								StartLine: changeStart,
								EndLine:   changeStart + di - 1,
								Lines:     original[changeStart : changeStart+di],
							})
						}
						if dj > 0 {
							changes = append(changes, Change{
								Type:      "add",
								StartLine: changeStart,
								Lines:     modified[modStart : modStart+dj],
							})
						}
						i = changeStart + di
						j = modStart + dj
						syncFound = true
						break
					}
				}
				if syncFound {
					break
				}
			}

			if !syncFound {
				// Treat as modification
				changes = append(changes, Change{
					Type:      "modify",
					StartLine: i,
					EndLine:   i,
					Lines:     []string{original[i]},
					NewLines:  []string{modified[j]},
				})
				i++
				j++
			}
		}
	}

	return changes
}

// formatUnifiedDiff formats changes as a unified diff.
func (g *DiffGenerator) formatUnifiedDiff(changes []Change, filename string, original, modified []string) string {
	var result strings.Builder

	// Write header
	result.WriteString(fmt.Sprintf("--- a/%s\n", filename))
	result.WriteString(fmt.Sprintf("+++ b/%s\n", filename))

	// Group changes into hunks
	hunks := g.groupIntoHunks(changes, original, modified)

	for _, hunk := range hunks {
		result.WriteString(g.formatHunk(hunk, original, modified))
	}

	return result.String()
}

// Hunk represents a group of related changes.
type Hunk struct {
	StartOriginal int
	LenOriginal   int
	StartModified int
	LenModified   int
	Changes       []Change
}

// groupIntoHunks groups changes into hunks with context.
func (g *DiffGenerator) groupIntoHunks(changes []Change, original, modified []string) []Hunk {
	if len(changes) == 0 {
		return nil
	}

	var hunks []Hunk
	var currentHunk *Hunk

	for _, change := range changes {
		if currentHunk == nil {
			// Start new hunk
			startLine := change.StartLine - g.contextLines
			if startLine < 0 {
				startLine = 0
			}

			currentHunk = &Hunk{
				StartOriginal: startLine,
				StartModified: startLine,
				Changes:       []Change{change},
			}
		} else {
			// Check if this change should be in the same hunk
			lastChange := currentHunk.Changes[len(currentHunk.Changes)-1]
			if change.StartLine <= lastChange.EndLine+2*g.contextLines {
				// Add to current hunk
				currentHunk.Changes = append(currentHunk.Changes, change)
			} else {
				// Finish current hunk and start new one
				g.finalizeHunk(currentHunk, original, modified)
				hunks = append(hunks, *currentHunk)

				startLine := change.StartLine - g.contextLines
				if startLine < 0 {
					startLine = 0
				}

				currentHunk = &Hunk{
					StartOriginal: startLine,
					StartModified: startLine,
					Changes:       []Change{change},
				}
			}
		}
	}

	// Add the last hunk
	if currentHunk != nil {
		g.finalizeHunk(currentHunk, original, modified)
		hunks = append(hunks, *currentHunk)
	}

	return hunks
}

// finalizeHunk calculates the final size of a hunk.
func (g *DiffGenerator) finalizeHunk(hunk *Hunk, original, modified []string) {
	if len(hunk.Changes) == 0 {
		return
	}

	lastChange := hunk.Changes[len(hunk.Changes)-1]
	endLine := lastChange.EndLine + g.contextLines
	if endLine >= len(original) {
		endLine = len(original) - 1
	}

	hunk.LenOriginal = endLine - hunk.StartOriginal + 1
	hunk.LenModified = hunk.LenOriginal // Adjust based on actual changes
}

// formatHunk formats a single hunk.
func (g *DiffGenerator) formatHunk(hunk Hunk, original, modified []string) string {
	var result strings.Builder

	// Write hunk header
	result.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
		hunk.StartOriginal+1, hunk.LenOriginal,
		hunk.StartModified+1, hunk.LenModified))

	// Write hunk content
	currentLine := hunk.StartOriginal
	for _, change := range hunk.Changes {
		// Write context before change
		for currentLine < change.StartLine {
			if currentLine < len(original) {
				result.WriteString(" " + original[currentLine] + "\n")
			}
			currentLine++
		}

		// Write the change
		switch change.Type {
		case "delete":
			for _, line := range change.Lines {
				result.WriteString("-" + line + "\n")
			}
			currentLine = change.EndLine + 1
		case "add":
			for _, line := range change.Lines {
				result.WriteString("+" + line + "\n")
			}
		case "modify":
			for _, line := range change.Lines {
				result.WriteString("-" + line + "\n")
			}
			for _, line := range change.NewLines {
				result.WriteString("+" + line + "\n")
			}
			currentLine = change.EndLine + 1
		}
	}

	// Write context after last change
	endLine := hunk.StartOriginal + hunk.LenOriginal
	for currentLine < endLine && currentLine < len(original) {
		result.WriteString(" " + original[currentLine] + "\n")
		currentLine++
	}

	return result.String()
}

// PreviewChanges generates a human-readable preview of changes.
func (g *DiffGenerator) PreviewChanges(original, modified string) string {
	originalLines := strings.Split(original, "\n")
	modifiedLines := strings.Split(modified, "\n")

	changes := g.findChanges(originalLines, modifiedLines)

	if len(changes) == 0 {
		return "No changes will be made"
	}

	var result strings.Builder
	result.WriteString("📝 Preview of changes:\n")
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	additions := 0
	deletions := 0
	modifications := 0

	for _, change := range changes {
		switch change.Type {
		case "add":
			additions += len(change.Lines)
		case "delete":
			deletions += len(change.Lines)
		case "modify":
			modifications += len(change.Lines)
		}
	}

	result.WriteString(fmt.Sprintf("  ➕ %d addition(s)\n", additions))
	result.WriteString(fmt.Sprintf("  ➖ %d deletion(s)\n", deletions))
	result.WriteString(fmt.Sprintf("  ✏️  %d modification(s)\n", modifications))
	result.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return result.String()
}

// GeneratePatch generates a patch file that can be applied with the patch command.
func (g *DiffGenerator) GeneratePatch(original, modified, filename string) string {
	var result strings.Builder

	// Write patch header
	result.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filename, filename))
	result.WriteString("index 0000000..1111111 100644\n")
	result.WriteString(fmt.Sprintf("--- a/%s\n", filename))
	result.WriteString(fmt.Sprintf("+++ b/%s\n", filename))

	// Generate the unified diff
	diff := g.GenerateDiff(original, modified, filename)

	// Remove the header lines we already added
	lines := strings.Split(diff, "\n")
	if len(lines) > 2 {
		for i := 2; i < len(lines); i++ {
			result.WriteString(lines[i])
			if i < len(lines)-1 {
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}