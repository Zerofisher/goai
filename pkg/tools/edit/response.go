package edit

import (
	"encoding/json"
	"fmt"
)

// ToolResponse represents the standardized JSON response format for edit operations.
type ToolResponse struct {
	Ok      bool        `json:"ok"`
	Summary string      `json:"summary"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// EditResult contains the data returned by edit operations.
type EditResult struct {
	Path          string   `json:"path"`
	Strategy      string   `json:"strategy"`
	LinesModified int      `json:"lines_modified"`
	BackupCreated bool     `json:"backup_created,omitempty"`
	BackupPath    string   `json:"backup_path,omitempty"`
	Diff          string   `json:"diff,omitempty"`
	Conflicts     []string `json:"conflicts,omitempty"`
}

// Success creates a successful response with data.
func Success(summary string, data *EditResult) string {
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
