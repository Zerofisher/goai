package indexing

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"
)

// DefaultChunker implements Chunker with configurable chunking strategies
type DefaultChunker struct {
	maxChunkSize    int
	overlapSize     int
	minChunkSize    int
	respectBoundaries bool // Respect function/class boundaries when chunking code
}

// NewDefaultChunker creates a new chunker with default settings
func NewDefaultChunker() *DefaultChunker {
	return &DefaultChunker{
		maxChunkSize:      1000,  // Maximum characters per chunk
		overlapSize:       100,   // Overlap between chunks for context
		minChunkSize:      50,    // Minimum chunk size to avoid tiny chunks
		respectBoundaries: true,  // Try to keep code blocks together
	}
}

// ChunkFile chunks a file into indexable segments
func (c *DefaultChunker) ChunkFile(ctx context.Context, file *FileContent) ([]*Chunk, error) {
	content := string(file.Content)
	language := file.Info.Language
	
	// Choose chunking strategy based on file type
	switch language {
	case "go", "javascript", "typescript", "python", "java", "cpp", "rust":
		return c.chunkCodeFile(content, file.Info.Path, language), nil
	case "markdown", "text":
		return c.chunkTextFile(content, file.Info.Path, language), nil
	default:
		return c.chunkTextFile(content, file.Info.Path, language), nil
	}
}

// ChunkText chunks raw text with language-specific logic
func (c *DefaultChunker) ChunkText(text string, language string) ([]*Chunk, error) {
	switch language {
	case "go", "javascript", "typescript", "python", "java", "cpp", "rust":
		return c.chunkCode(text, "", language), nil
	case "markdown":
		return c.chunkMarkdown(text, ""), nil
	default:
		return c.chunkPlainText(text, "", language), nil
	}
}

// chunkCodeFile chunks code files with function/class awareness
func (c *DefaultChunker) chunkCodeFile(content, filePath, language string) []*Chunk {
	if !c.respectBoundaries {
		return c.chunkPlainText(content, filePath, language)
	}
	
	return c.chunkCode(content, filePath, language)
}

// chunkCode performs intelligent code chunking
func (c *DefaultChunker) chunkCode(content, filePath, language string) []*Chunk {
	lines := strings.Split(content, "\n")
	var chunks []*Chunk
	var currentChunk strings.Builder
	var currentStart int
	var currentByteOffset int
	
	chunkID := 0
	
	for i, line := range lines {
		// Check if this line starts a new logical block
		if c.isBlockStart(line, language) && currentChunk.Len() > c.minChunkSize {
			// Create chunk from accumulated content
			if currentChunk.Len() > 0 {
				chunk := c.createChunk(
					fmt.Sprintf("%s_chunk_%d", filepath.Base(filePath), chunkID),
					filePath,
					strings.TrimSpace(currentChunk.String()),
					currentStart+1, // Convert to 1-based line numbers
					i,
					currentByteOffset,
					currentByteOffset+currentChunk.Len(),
					language,
					c.determineChunkType(currentChunk.String(), language),
				)
				chunks = append(chunks, chunk)
				chunkID++
			}
			
			// Start new chunk
			currentChunk.Reset()
			currentChunk.WriteString(line)
			if i < len(lines)-1 {
				currentChunk.WriteString("\n")
			}
			currentStart = i
			currentByteOffset += len(strings.Join(lines[:i], "\n"))
			if i > 0 {
				currentByteOffset++ // Account for newline
			}
		} else {
			// Add line to current chunk
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n")
			}
			currentChunk.WriteString(line)
			
			// Check if chunk is getting too large
			if currentChunk.Len() > c.maxChunkSize {
				// Create chunk and start new one with overlap
				chunk := c.createChunk(
					fmt.Sprintf("%s_chunk_%d", filepath.Base(filePath), chunkID),
					filePath,
					strings.TrimSpace(currentChunk.String()),
					currentStart+1,
					i+1,
					currentByteOffset,
					currentByteOffset+currentChunk.Len(),
					language,
					c.determineChunkType(currentChunk.String(), language),
				)
				chunks = append(chunks, chunk)
				chunkID++
				
				// Create overlap for context continuity
				overlapContent := c.getOverlap(currentChunk.String())
				currentChunk.Reset()
				currentChunk.WriteString(overlapContent)
				currentStart = i - strings.Count(overlapContent, "\n")
				currentByteOffset = chunk.EndByte - len(overlapContent)
			}
		}
	}
	
	// Handle remaining content
	if currentChunk.Len() > c.minChunkSize {
		chunk := c.createChunk(
			fmt.Sprintf("%s_chunk_%d", filepath.Base(filePath), chunkID),
			filePath,
			strings.TrimSpace(currentChunk.String()),
			currentStart+1,
			len(lines),
			currentByteOffset,
			currentByteOffset+currentChunk.Len(),
			language,
			c.determineChunkType(currentChunk.String(), language),
		)
		chunks = append(chunks, chunk)
	}
	
	return chunks
}

// chunkTextFile chunks text files (markdown, documentation, etc.)
func (c *DefaultChunker) chunkTextFile(content, filePath, language string) []*Chunk {
	if language == "markdown" {
		return c.chunkMarkdown(content, filePath)
	}
	return c.chunkPlainText(content, filePath, language)
}

// chunkMarkdown chunks markdown with section awareness
func (c *DefaultChunker) chunkMarkdown(content, filePath string) []*Chunk {
	lines := strings.Split(content, "\n")
	var chunks []*Chunk
	var currentChunk strings.Builder
	var currentStart int
	var currentByteOffset int
	
	chunkID := 0
	
	for i, line := range lines {
		// Check for markdown headers
		if strings.HasPrefix(strings.TrimSpace(line), "#") && currentChunk.Len() > c.minChunkSize {
			// Create chunk from accumulated content
			if currentChunk.Len() > 0 {
				chunk := c.createChunk(
					fmt.Sprintf("%s_section_%d", filepath.Base(filePath), chunkID),
					filePath,
					strings.TrimSpace(currentChunk.String()),
					currentStart+1,
					i,
					currentByteOffset,
					currentByteOffset+currentChunk.Len(),
					"markdown",
					ChunkTypeDocumentation,
				)
				chunks = append(chunks, chunk)
				chunkID++
			}
			
			// Start new chunk
			currentChunk.Reset()
			currentChunk.WriteString(line)
			if i < len(lines)-1 {
				currentChunk.WriteString("\n")
			}
			currentStart = i
			currentByteOffset += len(strings.Join(lines[:i], "\n"))
			if i > 0 {
				currentByteOffset++
			}
		} else {
			// Add line to current chunk
			if currentChunk.Len() > 0 {
				currentChunk.WriteString("\n")
			}
			currentChunk.WriteString(line)
			
			// Check size limits
			if currentChunk.Len() > c.maxChunkSize {
				chunk := c.createChunk(
					fmt.Sprintf("%s_section_%d", filepath.Base(filePath), chunkID),
					filePath,
					strings.TrimSpace(currentChunk.String()),
					currentStart+1,
					i+1,
					currentByteOffset,
					currentByteOffset+currentChunk.Len(),
					"markdown",
					ChunkTypeDocumentation,
				)
				chunks = append(chunks, chunk)
				chunkID++
				
				// Start new chunk with overlap
				overlapContent := c.getOverlap(currentChunk.String())
				currentChunk.Reset()
				currentChunk.WriteString(overlapContent)
				currentStart = i - strings.Count(overlapContent, "\n")
				currentByteOffset = chunk.EndByte - len(overlapContent)
			}
		}
	}
	
	// Handle remaining content
	if currentChunk.Len() > c.minChunkSize {
		chunk := c.createChunk(
			fmt.Sprintf("%s_section_%d", filepath.Base(filePath), chunkID),
			filePath,
			strings.TrimSpace(currentChunk.String()),
			currentStart+1,
			len(lines),
			currentByteOffset,
			currentByteOffset+currentChunk.Len(),
			"markdown",
			ChunkTypeDocumentation,
		)
		chunks = append(chunks, chunk)
	}
	
	return chunks
}

// chunkPlainText chunks plain text with paragraph awareness
func (c *DefaultChunker) chunkPlainText(content, filePath, language string) []*Chunk {
	// Split by paragraphs first
	paragraphs := strings.Split(content, "\n\n")
	var chunks []*Chunk
	var currentChunk strings.Builder
	var currentStart int
	var currentByteOffset int
	
	chunkID := 0
	lineCount := 0
	
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		
		// Add paragraph to current chunk
		if currentChunk.Len() > 0 {
			currentChunk.WriteString("\n\n")
			currentByteOffset += 2
		}
		currentChunk.WriteString(paragraph)
		
		paragraphLines := strings.Count(paragraph, "\n") + 1
		
		// Check if chunk is getting too large
		if currentChunk.Len() > c.maxChunkSize {
			chunk := c.createChunk(
				fmt.Sprintf("%s_chunk_%d", filepath.Base(filePath), chunkID),
				filePath,
				strings.TrimSpace(currentChunk.String()),
				currentStart+1,
				lineCount+paragraphLines,
				currentByteOffset-currentChunk.Len(),
				currentByteOffset,
				language,
				ChunkTypeCode,
			)
			chunks = append(chunks, chunk)
			chunkID++
			
			// Start new chunk
			currentChunk.Reset()
			currentStart = lineCount
			// Don't include overlap for plain text to avoid confusion
		}
		
		lineCount += paragraphLines
	}
	
	// Handle remaining content
	if currentChunk.Len() > c.minChunkSize {
		chunk := c.createChunk(
			fmt.Sprintf("%s_chunk_%d", filepath.Base(filePath), chunkID),
			filePath,
			strings.TrimSpace(currentChunk.String()),
			currentStart+1,
			lineCount,
			currentByteOffset-currentChunk.Len(),
			currentByteOffset,
			language,
			ChunkTypeCode,
		)
		chunks = append(chunks, chunk)
	}
	
	return chunks
}

// isBlockStart determines if a line starts a new logical block
func (c *DefaultChunker) isBlockStart(line, language string) bool {
	trimmed := strings.TrimSpace(line)
	
	switch language {
	case "go":
		return strings.HasPrefix(trimmed, "func ") ||
			strings.HasPrefix(trimmed, "type ") ||
			strings.HasPrefix(trimmed, "var ") ||
			strings.HasPrefix(trimmed, "const ") ||
			strings.HasPrefix(trimmed, "package ")
			
	case "javascript", "typescript":
		return strings.HasPrefix(trimmed, "function ") ||
			strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "interface ") ||
			strings.HasPrefix(trimmed, "const ") ||
			strings.HasPrefix(trimmed, "let ") ||
			strings.HasPrefix(trimmed, "var ") ||
			strings.HasPrefix(trimmed, "export ")
			
	case "python":
		return strings.HasPrefix(trimmed, "def ") ||
			strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "import ") ||
			strings.HasPrefix(trimmed, "from ")
			
	case "java":
		return strings.HasPrefix(trimmed, "public ") ||
			strings.HasPrefix(trimmed, "private ") ||
			strings.HasPrefix(trimmed, "protected ") ||
			strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "interface ") ||
			strings.HasPrefix(trimmed, "enum ")
			
	case "cpp", "c":
		return strings.HasPrefix(trimmed, "#include") ||
			strings.HasPrefix(trimmed, "#define") ||
			strings.HasPrefix(trimmed, "class ") ||
			strings.HasPrefix(trimmed, "struct ") ||
			strings.Contains(trimmed, "{") && !strings.HasPrefix(trimmed, "//")
	}
	
	return false
}

// determineChunkType determines the type of content in a chunk
func (c *DefaultChunker) determineChunkType(content, language string) ChunkType {
	trimmed := strings.TrimSpace(content)
	
	// Check for markdown first
	if language == "markdown" {
		return ChunkTypeDocumentation
	}
	
	// Check for comments
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "#") {
		return ChunkTypeComment
	}
	
	// Check for specific patterns based on language
	switch language {
	case "go", "javascript", "typescript", "python", "java", "cpp", "rust":
		if strings.Contains(trimmed, "func ") || strings.Contains(trimmed, "function ") || strings.Contains(trimmed, "def ") {
			return ChunkTypeFunction
		}
		if strings.Contains(trimmed, "class ") || strings.Contains(trimmed, "struct ") {
			return ChunkTypeClass
		}
		if strings.Contains(trimmed, "test") || strings.Contains(trimmed, "Test") {
			return ChunkTypeTest
		}
		return ChunkTypeCode
		
	default:
		return ChunkTypeCode
	}
}

// getOverlap extracts overlapping content for context continuity
func (c *DefaultChunker) getOverlap(content string) string {
	if len(content) <= c.overlapSize {
		return content
	}
	
	// Try to find a good break point (end of line, sentence, etc.)
	overlap := content[len(content)-c.overlapSize:]
	
	// Find the first newline to start overlap from a complete line
	if idx := strings.Index(overlap, "\n"); idx != -1 {
		overlap = overlap[idx+1:]
	}
	
	return overlap
}

// createChunk creates a new chunk with computed hash
func (c *DefaultChunker) createChunk(id, filePath, content string, startLine, endLine, startByte, endByte int, language string, chunkType ChunkType) *Chunk {
	// Generate content hash for change detection
	hash := sha256.Sum256([]byte(content))
	
	return &Chunk{
		ID:        id,
		FilePath:  filePath,
		Content:   content,
		StartLine: startLine,
		EndLine:   endLine,
		StartByte: startByte,
		EndByte:   endByte,
		Language:  language,
		ChunkType: chunkType,
		Metadata: map[string]interface{}{
			"hash":        fmt.Sprintf("%x", hash),
			"char_count":  len(content),
			"word_count":  len(strings.Fields(content)),
			"line_count":  endLine - startLine + 1,
		},
	}
}

// SetMaxChunkSize sets the maximum chunk size
func (c *DefaultChunker) SetMaxChunkSize(size int) {
	c.maxChunkSize = size
}

// SetOverlapSize sets the overlap size between chunks
func (c *DefaultChunker) SetOverlapSize(size int) {
	c.overlapSize = size
}

// SetMinChunkSize sets the minimum chunk size
func (c *DefaultChunker) SetMinChunkSize(size int) {
	c.minChunkSize = size
}

// SetRespectBoundaries sets whether to respect code block boundaries
func (c *DefaultChunker) SetRespectBoundaries(respect bool) {
	c.respectBoundaries = respect
}