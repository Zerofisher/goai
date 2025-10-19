package file

import (
	"encoding/json"
	"fmt"
)

// ToolResponse represents the standardized JSON response format for file tools.
// All file tools return responses in the format: {"ok":true,"summary":"...","data":{...}}
type ToolResponse struct {
	Ok      bool        `json:"ok"`
	Summary string      `json:"summary"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ReadFileData contains the data returned by read_file tool.
type ReadFileData struct {
	Path    string    `json:"path"`
	Content string    `json:"content"`
	Range   *LineRange `json:"range,omitempty"`
	Bytes   int       `json:"bytes"`
}

// LineRange represents the line range that was read.
type LineRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// WriteFileData contains the data returned by write_file tool.
type WriteFileData struct {
	Path    string `json:"path"`
	Bytes   int    `json:"bytes"`
	Created bool   `json:"created"`
	Mode    string `json:"mode,omitempty"`
}

// ListFilesData contains the data returned by list_files tool.
type ListFilesData struct {
	Files []FileEntry `json:"files"`
	Count int         `json:"count"`
	Path  string      `json:"path"`
}

// FileEntry represents a file or directory entry.
type FileEntry struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	IsDir    bool   `json:"is_dir"`
	Modified string `json:"modified,omitempty"`
	Mode     string `json:"mode,omitempty"`
}

// Success creates a successful response with data.
func Success(summary string, data interface{}) string {
	resp := ToolResponse{
		Ok:      true,
		Summary: summary,
		Data:    data,
	}
	return marshalResponse(resp)
}

// Error creates an error response.
func Error(summary string, err error) string {
	resp := ToolResponse{
		Ok:      false,
		Summary: summary,
		Error:   err.Error(),
	}
	return marshalResponse(resp)
}

// marshalResponse converts the response to JSON string.
func marshalResponse(resp ToolResponse) string {
	data, err := json.Marshal(resp)
	if err != nil {
		// Fallback to plain text error if JSON marshaling fails
		return fmt.Sprintf(`{"ok":false,"summary":"JSON marshaling error","error":"%s"}`, err.Error())
	}
	return string(data)
}
