package indexing

import (
	"context"
	"testing"
)

func TestGoTreeSitterParser(t *testing.T) {
	parser := NewGoTreeSitterParser()
	
	if parser == nil {
		t.Fatal("NewGoTreeSitterParser returned nil")
	}
	
	// Test supported languages
	languages := parser.GetSupportedLanguages()
	if len(languages) != 1 || languages[0] != "go" {
		t.Errorf("Expected supported languages to be [\"go\"], got %v", languages)
	}
}

func TestParseGoFile(t *testing.T) {
	parser := NewGoTreeSitterParser()
	ctx := context.Background()
	
	goCode := `package main

import (
	"fmt"
	"net/http"
)

// Constants for the application
const (
	DefaultPort = 8080
	AppName     = "test-app"
)

// Global variable
var globalCounter int

// User represents a user in the system
type User struct {
	ID   int    ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}

// UserService provides user-related operations
type UserService interface {
	GetUser(id int) (*User, error)
	CreateUser(name string) (*User, error)
}

// createUser creates a new user
func createUser(name string) *User {
	return &User{
		ID:   1,
		Name: name,
	}
}

// GetUser retrieves a user by ID
func (u *User) GetUser(id int) (*User, error) {
	return createUser("test"), nil
}

// main function
func main() {
	fmt.Println("Hello, World!")
	http.ListenAndServe(":8080", nil)
}`

	result, err := parser.ParseFile(ctx, "test.go", []byte(goCode))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	// Verify basic structure
	if result.FilePath != "test.go" {
		t.Errorf("Expected FilePath to be 'test.go', got '%s'", result.FilePath)
	}
	
	if result.Language != "go" {
		t.Errorf("Expected Language to be 'go', got '%s'", result.Language)
	}
	
	// Verify imports
	expectedImports := []string{"fmt", "net/http"}
	if len(result.Imports) != len(expectedImports) {
		t.Errorf("Expected %d imports, got %d", len(expectedImports), len(result.Imports))
	}
	for i, imp := range expectedImports {
		if i >= len(result.Imports) || result.Imports[i] != imp {
			t.Errorf("Expected import[%d] to be '%s', got '%s'", i, imp, result.Imports[i])
		}
	}
	
	// Verify symbols
	if len(result.Symbols) == 0 {
		t.Fatal("Expected symbols to be parsed, but got none")
	}
	
	// Create a map for easier symbol lookup
	symbolMap := make(map[string]*SymbolInfo)
	for _, symbol := range result.Symbols {
		symbolMap[symbol.Name] = symbol
	}
	
	// Test function parsing
	if createUserFunc, exists := symbolMap["createUser"]; exists {
		if createUserFunc.Kind != SymbolKindFunction {
			t.Errorf("Expected createUser to be a function, got %s", createUserFunc.Kind)
		}
		if !contains(createUserFunc.Signature, "createUser") {
			t.Errorf("Expected signature to contain 'createUser', got: %s", createUserFunc.Signature)
		}
	} else {
		t.Error("Expected to find 'createUser' function")
	}
	
	// Test method parsing
	if getUserMethod, exists := symbolMap["GetUser"]; exists {
		if getUserMethod.Kind != SymbolKindMethod {
			t.Errorf("Expected GetUser to be a method, got %s", getUserMethod.Kind)
		}
		if getUserMethod.Parent != "User" {
			t.Errorf("Expected GetUser parent to be 'User', got '%s'", getUserMethod.Parent)
		}
	} else {
		t.Error("Expected to find 'GetUser' method")
	}
	
	// Test struct parsing
	if userStruct, exists := symbolMap["User"]; exists {
		if userStruct.Kind != SymbolKindStruct {
			t.Errorf("Expected User to be a struct, got %s", userStruct.Kind)
		}
		if len(userStruct.Children) != 2 {
			t.Errorf("Expected User struct to have 2 fields, got %d", len(userStruct.Children))
		}
	} else {
		t.Error("Expected to find 'User' struct")
	}
	
	// Test interface parsing
	if userServiceInterface, exists := symbolMap["UserService"]; exists {
		if userServiceInterface.Kind != SymbolKindInterface {
			t.Errorf("Expected UserService to be an interface, got %s", userServiceInterface.Kind)
		}
		if len(userServiceInterface.Children) != 2 {
			t.Errorf("Expected UserService interface to have 2 methods, got %d", len(userServiceInterface.Children))
		}
	} else {
		t.Error("Expected to find 'UserService' interface")
	}
	
	// Test constants
	expectedConstants := []string{"DefaultPort", "AppName"}
	foundConstants := 0
	for _, constName := range expectedConstants {
		if constSymbol, exists := symbolMap[constName]; exists {
			if constSymbol.Kind != SymbolKindConstant {
				t.Errorf("Expected %s to be a constant, got %s", constName, constSymbol.Kind)
			}
			foundConstants++
		}
	}
	if foundConstants != len(expectedConstants) {
		t.Errorf("Expected to find %d constants, found %d", len(expectedConstants), foundConstants)
	}
	
	// Test variables
	if globalVar, exists := symbolMap["globalCounter"]; exists {
		if globalVar.Kind != SymbolKindVariable {
			t.Errorf("Expected globalCounter to be a variable, got %s", globalVar.Kind)
		}
	} else {
		t.Error("Expected to find 'globalCounter' variable")
	}
}

func TestParseNonGoFile(t *testing.T) {
	parser := NewGoTreeSitterParser()
	ctx := context.Background()
	
	result, err := parser.ParseFile(ctx, "test.txt", []byte("not go code"))
	if err != nil {
		t.Fatalf("ParseFile should handle non-Go files gracefully: %v", err)
	}
	
	if result.Language != "unknown" {
		t.Errorf("Expected Language to be 'unknown' for non-Go files, got '%s'", result.Language)
	}
	
	if len(result.Symbols) != 0 {
		t.Errorf("Expected no symbols for non-Go files, got %d", len(result.Symbols))
	}
}

func TestParseInvalidGoCode(t *testing.T) {
	parser := NewGoTreeSitterParser()
	ctx := context.Background()
	
	invalidCode := `package main
	func invalid syntax {
		// missing parentheses and return type
	`
	
	// Should not crash on invalid code
	result, err := parser.ParseFile(ctx, "invalid.go", []byte(invalidCode))
	if err != nil {
		// Parser may legitimately fail on invalid syntax
		t.Logf("ParseFile failed on invalid code (expected): %v", err)
		return
	}
	
	if result.Language != "go" {
		t.Errorf("Expected Language to be 'go' even for invalid code, got '%s'", result.Language)
	}
	
	// May or may not parse symbols from invalid code, but shouldn't crash
	t.Logf("Parsed %d symbols from invalid code", len(result.Symbols))
}

func TestFunctionSignatureBuilding(t *testing.T) {
	parser := NewGoTreeSitterParser()
	ctx := context.Background()
	
	testCases := []struct {
		name     string
		code     string
		funcName string
		expected string
	}{
		{
			name: "simple function",
			code: `package main
func add(a, b int) int {
	return a + b
}`,
			funcName: "add",
			expected: "add",
		},
		{
			name: "method with receiver",
			code: `package main
type User struct{}
func (u *User) GetName() string {
	return "test"
}`,
			funcName: "GetName",
			expected: "GetName",
		},
		{
			name: "function with multiple return values",
			code: `package main
func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}
	return a / b, nil
}`,
			funcName: "divide",
			expected: "divide",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.ParseFile(ctx, "test.go", []byte(tc.code))
			if err != nil {
				t.Fatalf("ParseFile failed: %v", err)
			}
			
			found := false
			for _, symbol := range result.Symbols {
				if symbol.Name == tc.funcName {
					if !contains(symbol.Signature, tc.expected) {
						t.Errorf("Expected signature to contain '%s', got: %s", tc.expected, symbol.Signature)
					}
					found = true
					break
				}
			}
			
			if !found {
				t.Errorf("Function '%s' not found in parsed symbols", tc.funcName)
			}
		})
	}
}

func TestTypeSpecParsing(t *testing.T) {
	parser := NewGoTreeSitterParser()
	ctx := context.Background()
	
	code := `package main

type StringAlias string
type IntSlice []int
type StringMap map[string]int

type Config struct {
	Host string
	Port int
}

type Reader interface {
	Read() ([]byte, error)
}
`
	
	result, err := parser.ParseFile(ctx, "test.go", []byte(code))
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	
	symbolMap := make(map[string]*SymbolInfo)
	for _, symbol := range result.Symbols {
		symbolMap[symbol.Name] = symbol
	}
	
	// Test type alias
	if alias, exists := symbolMap["StringAlias"]; exists {
		if alias.Kind != SymbolKindType {
			t.Errorf("Expected StringAlias to be a type, got %s", alias.Kind)
		}
	}
	
	// Test struct
	if config, exists := symbolMap["Config"]; exists {
		if config.Kind != SymbolKindStruct {
			t.Errorf("Expected Config to be a struct, got %s", config.Kind)
		}
		if len(config.Children) != 2 {
			t.Errorf("Expected Config to have 2 fields, got %d", len(config.Children))
		}
		expectedFields := []string{"Host", "Port"}
		for i, field := range expectedFields {
			if i >= len(config.Children) || config.Children[i] != field {
				t.Errorf("Expected field[%d] to be '%s', got '%s'", i, field, config.Children[i])
			}
		}
	}
	
	// Test interface
	if reader, exists := symbolMap["Reader"]; exists {
		if reader.Kind != SymbolKindInterface {
			t.Errorf("Expected Reader to be an interface, got %s", reader.Kind)
		}
		if len(reader.Children) != 1 || reader.Children[0] != "Read" {
			t.Errorf("Expected Reader to have method 'Read', got %v", reader.Children)
		}
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) <= len(s) && (substr == "" || findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}