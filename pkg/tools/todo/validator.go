package todo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Zerofisher/goai/pkg/todo"
)

// Validator provides input validation for todo operations.
type Validator struct {
	maxItems int
}

// NewValidator creates a new todo validator.
func NewValidator() *Validator {
	return &Validator{
		maxItems: todo.MaxTodoItems,
	}
}

// ValidateTodoList validates a complete todo list structure.
func (v *Validator) ValidateTodoList(input interface{}) error {
	// Check if input is a slice
	slice, ok := input.([]interface{})
	if !ok {
		return fmt.Errorf("todo list must be an array, got %T", input)
	}

	return v.ValidateTodoItems(slice)
}

// ValidateTodoItems validates an array of todo items.
func (v *Validator) ValidateTodoItems(items []interface{}) error {
	if len(items) > v.maxItems {
		return fmt.Errorf("todo list cannot exceed %d items, got %d", v.maxItems, len(items))
	}

	seenIDs := make(map[string]bool)
	inProgressCount := 0

	for i, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("todo item at index %d must be an object, got %T", i, item)
		}

		if err := v.ValidateTodoItem(itemMap, i); err != nil {
			return fmt.Errorf("invalid todo at index %d: %w", i, err)
		}

		// Check for duplicate IDs
		if id := v.extractID(itemMap, i); id != "" {
			if seenIDs[id] {
				return fmt.Errorf("duplicate todo ID: %s", id)
			}
			seenIDs[id] = true
		}

		// Count in-progress items
		if status, _ := itemMap["status"].(string); status == string(todo.StatusInProgress) {
			inProgressCount++
		}
	}

	// Validate business rule: only one in-progress item
	if inProgressCount > 1 {
		return fmt.Errorf("only one task can be in progress at a time, found %d", inProgressCount)
	}

	return nil
}

// ValidateTodoItem validates a single todo item.
func (v *Validator) ValidateTodoItem(item map[string]interface{}, index int) error {
	// Validate content
	content, hasContent := item["content"].(string)
	if !hasContent {
		return fmt.Errorf("content is required")
	}
	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content cannot be empty")
	}
	if len(content) > 500 {
		return fmt.Errorf("content is too long (max 500 characters)")
	}

	// Validate activeForm
	activeForm, hasActiveForm := item["activeForm"].(string)
	if !hasActiveForm {
		return fmt.Errorf("activeForm is required")
	}
	if strings.TrimSpace(activeForm) == "" {
		return fmt.Errorf("activeForm cannot be empty")
	}
	if len(activeForm) > 500 {
		return fmt.Errorf("activeForm is too long (max 500 characters)")
	}

	// Validate status
	statusStr, hasStatus := item["status"].(string)
	if !hasStatus {
		return fmt.Errorf("status is required")
	}

	status := todo.Status(statusStr)
	if !status.IsValid() {
		validStatuses := make([]string, len(todo.ValidStatuses))
		for i, s := range todo.ValidStatuses {
			validStatuses[i] = string(s)
		}
		return fmt.Errorf("status must be one of: %s", strings.Join(validStatuses, ", "))
	}

	// Validate optional ID if present
	if idRaw, hasID := item["id"]; hasID {
		id := fmt.Sprintf("%v", idRaw)
		if strings.TrimSpace(id) == "" {
			return fmt.Errorf("id cannot be empty if provided")
		}
		if len(id) > 100 {
			return fmt.Errorf("id is too long (max 100 characters)")
		}
	}

	return nil
}

// ValidateStatusTransition validates if a status transition is allowed.
func (v *Validator) ValidateStatusTransition(oldStatus, newStatus todo.Status) error {
	if !oldStatus.IsValid() {
		return fmt.Errorf("invalid old status: %s", oldStatus)
	}

	if !newStatus.IsValid() {
		return fmt.Errorf("invalid new status: %s", newStatus)
	}

	// All transitions are currently allowed, including reopening completed tasks
	// This can be customized later if needed

	return nil
}

// extractID extracts the ID from a todo item map.
func (v *Validator) extractID(item map[string]interface{}, defaultIndex int) string {
	if idRaw, hasID := item["id"]; hasID {
		return fmt.Sprintf("%v", idRaw)
	}
	return fmt.Sprintf("%d", defaultIndex+1)
}

// ValidateInput validates the complete input structure for the todo tool.
func (v *Validator) ValidateInput(input map[string]interface{}) error {
	if input == nil {
		return fmt.Errorf("input cannot be nil")
	}

	// Check for required fields
	todos, hasTodos := input["todos"]
	if !hasTodos {
		return fmt.Errorf("'todos' field is required")
	}

	// Check for unexpected fields
	for key := range input {
		if key != "todos" {
			return fmt.Errorf("unexpected field: %s", key)
		}
	}

	// Validate the todos array
	todosArray, ok := todos.([]interface{})
	if !ok {
		return fmt.Errorf("'todos' must be an array, got %T", todos)
	}

	return v.ValidateTodoItems(todosArray)
}

// ValidatePartialUpdate validates a partial update to existing todos.
func (v *Validator) ValidatePartialUpdate(updates map[string]interface{}, existing []todo.TodoItem) error {
	// Extract the ID to update
	id, hasID := updates["id"].(string)
	if !hasID || id == "" {
		return fmt.Errorf("id is required for partial update")
	}

	// Find the existing item
	var existingItem *todo.TodoItem
	for i := range existing {
		if existing[i].ID == id {
			existingItem = &existing[i]
			break
		}
	}

	if existingItem == nil {
		return fmt.Errorf("todo with ID %s not found", id)
	}

	// Validate status transition if status is being updated
	if newStatusStr, hasStatus := updates["status"].(string); hasStatus {
		newStatus := todo.Status(newStatusStr)
		if !newStatus.IsValid() {
			return fmt.Errorf("invalid status: %s", newStatusStr)
		}

		if err := v.ValidateStatusTransition(existingItem.Status, newStatus); err != nil {
			return err
		}

		// Check in-progress constraint
		if newStatus == todo.StatusInProgress {
			for i := range existing {
				if existing[i].ID != id && existing[i].Status == todo.StatusInProgress {
					return fmt.Errorf("cannot set to in-progress: %s is already in progress", existing[i].ID)
				}
			}
		}
	}

	// Validate content if being updated
	if content, hasContent := updates["content"].(string); hasContent {
		if strings.TrimSpace(content) == "" {
			return fmt.Errorf("content cannot be empty")
		}
		if len(content) > 500 {
			return fmt.Errorf("content is too long (max 500 characters)")
		}
	}

	// Validate activeForm if being updated
	if activeForm, hasActiveForm := updates["activeForm"].(string); hasActiveForm {
		if strings.TrimSpace(activeForm) == "" {
			return fmt.Errorf("activeForm cannot be empty")
		}
		if len(activeForm) > 500 {
			return fmt.Errorf("activeForm is too long (max 500 characters)")
		}
	}

	return nil
}

// SanitizeInput cleans and normalizes input data.
func (v *Validator) SanitizeInput(input map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	if todos, hasTodos := input["todos"]; hasTodos {
		if todosArray, ok := todos.([]interface{}); ok {
			sanitizedTodos := make([]interface{}, 0, len(todosArray))
			for _, item := range todosArray {
				if itemMap, ok := item.(map[string]interface{}); ok {
					sanitizedTodos = append(sanitizedTodos, v.sanitizeTodoItem(itemMap))
				}
			}
			sanitized["todos"] = sanitizedTodos
		}
	}

	return sanitized
}

// sanitizeTodoItem cleans a single todo item.
func (v *Validator) sanitizeTodoItem(item map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	// Sanitize content
	if content, ok := item["content"].(string); ok {
		sanitized["content"] = strings.TrimSpace(content)
	}

	// Sanitize activeForm
	if activeForm, ok := item["activeForm"].(string); ok {
		sanitized["activeForm"] = strings.TrimSpace(activeForm)
	}

	// Sanitize status (lowercase)
	if status, ok := item["status"].(string); ok {
		sanitized["status"] = strings.ToLower(strings.TrimSpace(status))
	}

	// Sanitize ID if present
	if id, ok := item["id"]; ok {
		sanitized["id"] = fmt.Sprintf("%v", id)
	}

	return sanitized
}

// ValidateSchema validates that input matches the expected schema structure.
func (v *Validator) ValidateSchema(input interface{}, schema map[string]interface{}) error {
	schemaType, hasType := schema["type"].(string)
	if !hasType {
		return fmt.Errorf("schema must have a type")
	}

	inputType := reflect.TypeOf(input)
	if inputType == nil {
		return fmt.Errorf("input is nil")
	}

	switch schemaType {
	case "object":
		if inputType.Kind() != reflect.Map {
			return fmt.Errorf("expected object, got %s", inputType.Kind())
		}
		return v.validateObjectSchema(input, schema)

	case "array":
		if inputType.Kind() != reflect.Slice && inputType.Kind() != reflect.Array {
			return fmt.Errorf("expected array, got %s", inputType.Kind())
		}
		return v.validateArraySchema(input, schema)

	case "string":
		if inputType.Kind() != reflect.String {
			return fmt.Errorf("expected string, got %s", inputType.Kind())
		}

	case "number", "integer":
		if !isNumeric(inputType.Kind()) {
			return fmt.Errorf("expected number, got %s", inputType.Kind())
		}

	case "boolean":
		if inputType.Kind() != reflect.Bool {
			return fmt.Errorf("expected boolean, got %s", inputType.Kind())
		}

	default:
		return fmt.Errorf("unknown schema type: %s", schemaType)
	}

	return nil
}

// validateObjectSchema validates an object against its schema.
func (v *Validator) validateObjectSchema(input interface{}, schema map[string]interface{}) error {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return fmt.Errorf("cannot validate object: input is not a map")
	}

	// Check required fields
	if required, hasRequired := schema["required"].([]string); hasRequired {
		for _, field := range required {
			if _, exists := inputMap[field]; !exists {
				return fmt.Errorf("required field '%s' is missing", field)
			}
		}
	}

	// Validate properties
	if properties, hasProperties := schema["properties"].(map[string]interface{}); hasProperties {
		for key, value := range inputMap {
			if propSchema, hasProp := properties[key].(map[string]interface{}); hasProp {
				if err := v.ValidateSchema(value, propSchema); err != nil {
					return fmt.Errorf("field '%s': %w", key, err)
				}
			} else if additionalProps, hasAdditional := schema["additionalProperties"].(bool); hasAdditional && !additionalProps {
				return fmt.Errorf("unexpected field: %s", key)
			}
		}
	}

	return nil
}

// validateArraySchema validates an array against its schema.
func (v *Validator) validateArraySchema(input interface{}, schema map[string]interface{}) error {
	inputSlice, ok := input.([]interface{})
	if !ok {
		return fmt.Errorf("cannot validate array: input is not a slice")
	}

	// Check max items
	if maxItems, hasMax := schema["maxItems"].(int); hasMax {
		if len(inputSlice) > maxItems {
			return fmt.Errorf("array exceeds maximum length of %d", maxItems)
		}
	}

	// Check min items
	if minItems, hasMin := schema["minItems"].(int); hasMin {
		if len(inputSlice) < minItems {
			return fmt.Errorf("array is below minimum length of %d", minItems)
		}
	}

	// Validate items
	if itemSchema, hasItemSchema := schema["items"].(map[string]interface{}); hasItemSchema {
		for i, item := range inputSlice {
			if err := v.ValidateSchema(item, itemSchema); err != nil {
				return fmt.Errorf("item at index %d: %w", i, err)
			}
		}
	}

	return nil
}

// isNumeric checks if a reflect.Kind represents a numeric type.
func isNumeric(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}