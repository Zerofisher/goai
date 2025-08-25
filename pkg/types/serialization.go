package types

import (
	"encoding/json"
	"fmt"
)

// ValidatedJSON provides JSON serialization with validation
type ValidatedJSON struct {
	validator *DataValidator
}

// NewValidatedJSON creates a new ValidatedJSON instance
func NewValidatedJSON() *ValidatedJSON {
	return &ValidatedJSON{
		validator: NewDataValidator(),
	}
}

// MarshalProblemRequest marshals ProblemRequest to JSON with validation
func (vj *ValidatedJSON) MarshalProblemRequest(req *ProblemRequest) ([]byte, error) {
	// Validate before marshaling
	result := vj.validator.ValidateProblemRequest(req)
	if result.HasErrors() {
		return nil, fmt.Errorf("validation failed: %s", result.Errors[0].Error())
	}
	
	return json.MarshalIndent(req, "", "  ")
}

// UnmarshalProblemRequest unmarshals JSON to ProblemRequest with validation
func (vj *ValidatedJSON) UnmarshalProblemRequest(data []byte) (*ProblemRequest, error) {
	var req ProblemRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	// Validate after unmarshaling
	result := vj.validator.ValidateProblemRequest(&req)
	if result.HasErrors() {
		return nil, fmt.Errorf("validation failed: %s", result.Errors[0].Error())
	}
	
	return &req, nil
}

// MarshalAnalysis marshals Analysis to JSON with validation
func (vj *ValidatedJSON) MarshalAnalysis(analysis *Analysis) ([]byte, error) {
	result := vj.validator.ValidateAnalysis(analysis)
	if result.HasErrors() {
		return nil, fmt.Errorf("validation failed: %s", result.Errors[0].Error())
	}
	
	return json.MarshalIndent(analysis, "", "  ")
}

// UnmarshalAnalysis unmarshals JSON to Analysis with validation
func (vj *ValidatedJSON) UnmarshalAnalysis(data []byte) (*Analysis, error) {
	var analysis Analysis
	if err := json.Unmarshal(data, &analysis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	result := vj.validator.ValidateAnalysis(&analysis)
	if result.HasErrors() {
		return nil, fmt.Errorf("validation failed: %s", result.Errors[0].Error())
	}
	
	return &analysis, nil
}

// MarshalExecutionPlan marshals ExecutionPlan to JSON with validation
func (vj *ValidatedJSON) MarshalExecutionPlan(plan *ExecutionPlan) ([]byte, error) {
	result := vj.validator.ValidateExecutionPlan(plan)
	if result.HasErrors() {
		return nil, fmt.Errorf("validation failed: %s", result.Errors[0].Error())
	}
	
	return json.MarshalIndent(plan, "", "  ")
}

// UnmarshalExecutionPlan unmarshals JSON to ExecutionPlan with validation
func (vj *ValidatedJSON) UnmarshalExecutionPlan(data []byte) (*ExecutionPlan, error) {
	var plan ExecutionPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	result := vj.validator.ValidateExecutionPlan(&plan)
	if result.HasErrors() {
		return nil, fmt.Errorf("validation failed: %s", result.Errors[0].Error())
	}
	
	return &plan, nil
}

// MarshalProjectContext marshals ProjectContext to JSON with validation
func (vj *ValidatedJSON) MarshalProjectContext(context *ProjectContext) ([]byte, error) {
	result := vj.validator.ValidateProjectContext(context)
	if result.HasErrors() {
		return nil, fmt.Errorf("validation failed: %s", result.Errors[0].Error())
	}
	
	return json.MarshalIndent(context, "", "  ")
}

// UnmarshalProjectContext unmarshals JSON to ProjectContext with validation
func (vj *ValidatedJSON) UnmarshalProjectContext(data []byte) (*ProjectContext, error) {
	var context ProjectContext
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	result := vj.validator.ValidateProjectContext(&context)
	if result.HasErrors() {
		return nil, fmt.Errorf("validation failed: %s", result.Errors[0].Error())
	}
	
	return &context, nil
}

// Convenience methods for common operations

// ProblemRequestToJSON converts ProblemRequest to validated JSON string
func ProblemRequestToJSON(req *ProblemRequest) (string, error) {
	vj := NewValidatedJSON()
	data, err := vj.MarshalProblemRequest(req)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ProblemRequestFromJSON converts JSON string to validated ProblemRequest
func ProblemRequestFromJSON(jsonStr string) (*ProblemRequest, error) {
	vj := NewValidatedJSON()
	return vj.UnmarshalProblemRequest([]byte(jsonStr))
}

// AnalysisToJSON converts Analysis to validated JSON string
func AnalysisToJSON(analysis *Analysis) (string, error) {
	vj := NewValidatedJSON()
	data, err := vj.MarshalAnalysis(analysis)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// AnalysisFromJSON converts JSON string to validated Analysis
func AnalysisFromJSON(jsonStr string) (*Analysis, error) {
	vj := NewValidatedJSON()
	return vj.UnmarshalAnalysis([]byte(jsonStr))
}

// ExecutionPlanToJSON converts ExecutionPlan to validated JSON string
func ExecutionPlanToJSON(plan *ExecutionPlan) (string, error) {
	vj := NewValidatedJSON()
	data, err := vj.MarshalExecutionPlan(plan)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ExecutionPlanFromJSON converts JSON string to validated ExecutionPlan
func ExecutionPlanFromJSON(jsonStr string) (*ExecutionPlan, error) {
	vj := NewValidatedJSON()
	return vj.UnmarshalExecutionPlan([]byte(jsonStr))
}

// ProjectContextToJSON converts ProjectContext to validated JSON string
func ProjectContextToJSON(context *ProjectContext) (string, error) {
	vj := NewValidatedJSON()
	data, err := vj.MarshalProjectContext(context)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ProjectContextFromJSON converts JSON string to validated ProjectContext
func ProjectContextFromJSON(jsonStr string) (*ProjectContext, error) {
	vj := NewValidatedJSON()
	return vj.UnmarshalProjectContext([]byte(jsonStr))
}

// ValidateAndFormat validates and formats any supported data structure
func ValidateAndFormat(data interface{}) (string, error) {
	switch v := data.(type) {
	case *ProblemRequest:
		return ProblemRequestToJSON(v)
	case ProblemRequest:
		return ProblemRequestToJSON(&v)
	case *Analysis:
		return AnalysisToJSON(v)
	case Analysis:
		return AnalysisToJSON(&v)
	case *ExecutionPlan:
		return ExecutionPlanToJSON(v)
	case ExecutionPlan:
		return ExecutionPlanToJSON(&v)
	case *ProjectContext:
		return ProjectContextToJSON(v)
	case ProjectContext:
		return ProjectContextToJSON(&v)
	default:
		// Fall back to regular JSON for unsupported types
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal data: %w", err)
		}
		return string(bytes), nil
	}
}

// BatchValidateJSON validates multiple data structures
type BatchValidateResult struct {
	Valid  bool              `json:"valid"`
	Errors map[string]string `json:"errors,omitempty"`
}

// ValidateBatch validates multiple data structures at once
func ValidateBatch(items map[string]interface{}) *BatchValidateResult {
	result := &BatchValidateResult{
		Valid:  true,
		Errors: make(map[string]string),
	}
	
	vj := NewValidatedJSON()
	
	for name, item := range items {
		var err error
		
		switch v := item.(type) {
		case *ProblemRequest:
			_, err = vj.MarshalProblemRequest(v)
		case ProblemRequest:
			_, err = vj.MarshalProblemRequest(&v)
		case *Analysis:
			_, err = vj.MarshalAnalysis(v)
		case Analysis:
			_, err = vj.MarshalAnalysis(&v)
		case *ExecutionPlan:
			_, err = vj.MarshalExecutionPlan(v)
		case ExecutionPlan:
			_, err = vj.MarshalExecutionPlan(&v)
		case *ProjectContext:
			_, err = vj.MarshalProjectContext(v)
		case ProjectContext:
			_, err = vj.MarshalProjectContext(&v)
		default:
			// Skip validation for unknown types
			continue
		}
		
		if err != nil {
			result.Valid = false
			result.Errors[name] = err.Error()
		}
	}
	
	return result
}

// SafeUnmarshal attempts to unmarshal JSON to a specific type with error handling
func SafeUnmarshal(data []byte, targetType string) (interface{}, error) {
	vj := NewValidatedJSON()
	
	switch targetType {
	case "ProblemRequest":
		return vj.UnmarshalProblemRequest(data)
	case "Analysis":
		return vj.UnmarshalAnalysis(data)
	case "ExecutionPlan":
		return vj.UnmarshalExecutionPlan(data)
	case "ProjectContext":
		return vj.UnmarshalProjectContext(data)
	default:
		return nil, fmt.Errorf("unsupported target type: %s", targetType)
	}
}

// JSONPatch represents a JSON patch operation (RFC 6902 style)
type JSONPatch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// ApplyJSONPatch applies a simple JSON patch to update fields
// Note: This is a simplified implementation, not full RFC 6902
func ApplyJSONPatch(original interface{}, patches []JSONPatch) (interface{}, error) {
	// Convert to map for easy manipulation
	dataMap, err := ConvertToMap(original)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to map: %w", err)
	}
	
	// Apply patches
	for _, patch := range patches {
		switch patch.Op {
		case "replace":
			// Simple path replacement (only supports top-level paths like "/field_name")
			if len(patch.Path) > 1 && patch.Path[0] == '/' {
				fieldName := patch.Path[1:]
				dataMap[fieldName] = patch.Value
			}
		case "add":
			if len(patch.Path) > 1 && patch.Path[0] == '/' {
				fieldName := patch.Path[1:]
				dataMap[fieldName] = patch.Value
			}
		case "remove":
			if len(patch.Path) > 1 && patch.Path[0] == '/' {
				fieldName := patch.Path[1:]
				delete(dataMap, fieldName)
			}
		default:
			return nil, fmt.Errorf("unsupported patch operation: %s", patch.Op)
		}
	}
	
	return dataMap, nil
}