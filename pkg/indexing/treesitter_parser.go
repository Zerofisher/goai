package indexing

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// GoTreeSitterParser implements TreeSitterParser for Go files using go/parser
type GoTreeSitterParser struct {
	fset *token.FileSet
}

// NewGoTreeSitterParser creates a new Go parser
func NewGoTreeSitterParser() *GoTreeSitterParser {
	return &GoTreeSitterParser{
		fset: token.NewFileSet(),
	}
}

// ParseFile parses a Go file and extracts symbols
func (p *GoTreeSitterParser) ParseFile(ctx context.Context, filePath string, content []byte) (*ParsedFile, error) {
	// Only handle Go files for now
	if !strings.HasSuffix(filePath, ".go") {
		return &ParsedFile{
			FilePath: filePath,
			Language: "unknown",
			Symbols:  []*SymbolInfo{},
		}, nil
	}

	// Parse Go file
	file, err := parser.ParseFile(p.fset, filePath, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go file: %w", err)
	}

	parsedFile := &ParsedFile{
		FilePath: filePath,
		Language: "go",
		Symbols:  []*SymbolInfo{},
		Imports:  []string{},
	}

	// Extract imports
	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		parsedFile.Imports = append(parsedFile.Imports, importPath)
	}

	// Walk AST and extract symbols
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			symbol := p.extractFunction(node)
			if symbol != nil {
				parsedFile.Symbols = append(parsedFile.Symbols, symbol)
			}

		case *ast.GenDecl:
			symbols := p.extractGeneralDeclaration(node)
			parsedFile.Symbols = append(parsedFile.Symbols, symbols...)

		case *ast.TypeSpec:
			symbol := p.extractTypeSpec(node)
			if symbol != nil {
				parsedFile.Symbols = append(parsedFile.Symbols, symbol)
			}
		}
		return true
	})

	// Set file path for all symbols
	for _, symbol := range parsedFile.Symbols {
		symbol.FilePath = filePath
		symbol.Language = "go"
	}

	return parsedFile, nil
}

// GetSupportedLanguages returns supported languages
func (p *GoTreeSitterParser) GetSupportedLanguages() []string {
	return []string{"go"}
}

// extractFunction extracts function/method symbols
func (p *GoTreeSitterParser) extractFunction(fn *ast.FuncDecl) *SymbolInfo {
	if fn.Name == nil {
		return nil
	}

	pos := p.fset.Position(fn.Pos())
	end := p.fset.Position(fn.End())

	symbol := &SymbolInfo{
		Name:        fn.Name.Name,
		Kind:        SymbolKindFunction,
		StartLine:   pos.Line,
		EndLine:     end.Line,
		StartColumn: pos.Column,
		EndColumn:   end.Column,
		Signature:   p.buildFunctionSignature(fn),
	}

	// Check if it's a method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		symbol.Kind = SymbolKindMethod
		if recv := fn.Recv.List[0].Type; recv != nil {
			if ident, ok := recv.(*ast.Ident); ok {
				symbol.Parent = ident.Name
			} else if star, ok := recv.(*ast.StarExpr); ok {
				if ident, ok := star.X.(*ast.Ident); ok {
					symbol.Parent = ident.Name
				}
			}
		}
	}

	// Extract doc comment
	if fn.Doc != nil {
		var docLines []string
		for _, comment := range fn.Doc.List {
			docLines = append(docLines, strings.TrimPrefix(comment.Text, "//"))
		}
		symbol.DocString = strings.Join(docLines, "\n")
	}

	return symbol
}

// extractGeneralDeclaration extracts symbols from general declarations (var, const, type)
func (p *GoTreeSitterParser) extractGeneralDeclaration(decl *ast.GenDecl) []*SymbolInfo {
	var symbols []*SymbolInfo

	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.ValueSpec:
			// Handle var and const declarations
			kind := SymbolKindVariable
			if decl.Tok == token.CONST {
				kind = SymbolKindConstant
			}

			for _, name := range s.Names {
				if name.Name == "_" {
					continue // Skip blank identifiers
				}

				pos := p.fset.Position(name.Pos())
				end := p.fset.Position(name.End())

				symbol := &SymbolInfo{
					Name:        name.Name,
					Kind:        kind,
					StartLine:   pos.Line,
					EndLine:     end.Line,
					StartColumn: pos.Column,
					EndColumn:   end.Column,
				}

				// Add type information if available
				if s.Type != nil {
					symbol.Signature = p.typeToString(s.Type)
				}

				symbols = append(symbols, symbol)
			}

		case *ast.TypeSpec:
			symbol := p.extractTypeSpec(s)
			if symbol != nil {
				symbols = append(symbols, symbol)
			}
		}
	}

	// Add doc comments to symbols
	if decl.Doc != nil && len(symbols) > 0 {
		var docLines []string
		for _, comment := range decl.Doc.List {
			docLines = append(docLines, strings.TrimPrefix(comment.Text, "//"))
		}
		docString := strings.Join(docLines, "\n")
		
		// Apply to first symbol (most common case)
		symbols[0].DocString = docString
	}

	return symbols
}

// extractTypeSpec extracts type symbols
func (p *GoTreeSitterParser) extractTypeSpec(spec *ast.TypeSpec) *SymbolInfo {
	if spec.Name == nil {
		return nil
	}

	pos := p.fset.Position(spec.Pos())
	end := p.fset.Position(spec.End())

	symbol := &SymbolInfo{
		Name:        spec.Name.Name,
		Kind:        SymbolKindType,
		StartLine:   pos.Line,
		EndLine:     end.Line,
		StartColumn: pos.Column,
		EndColumn:   end.Column,
	}

	// Determine specific type kind
	switch spec.Type.(type) {
	case *ast.StructType:
		symbol.Kind = SymbolKindStruct
		symbol.Signature = fmt.Sprintf("type %s struct", spec.Name.Name)
		
		// Extract struct fields as children
		if structType, ok := spec.Type.(*ast.StructType); ok {
			for _, field := range structType.Fields.List {
				for _, fieldName := range field.Names {
					symbol.Children = append(symbol.Children, fieldName.Name)
				}
			}
		}

	case *ast.InterfaceType:
		symbol.Kind = SymbolKindInterface
		symbol.Signature = fmt.Sprintf("type %s interface", spec.Name.Name)
		
		// Extract interface methods as children
		if ifaceType, ok := spec.Type.(*ast.InterfaceType); ok {
			for _, method := range ifaceType.Methods.List {
				for _, methodName := range method.Names {
					symbol.Children = append(symbol.Children, methodName.Name)
				}
			}
		}

	default:
		symbol.Signature = fmt.Sprintf("type %s %s", spec.Name.Name, p.typeToString(spec.Type))
	}

	return symbol
}

// buildFunctionSignature builds a function signature string
func (p *GoTreeSitterParser) buildFunctionSignature(fn *ast.FuncDecl) string {
	var sig strings.Builder
	
	sig.WriteString("func ")
	
	// Add receiver if method
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		sig.WriteString("(")
		for i, recv := range fn.Recv.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			if len(recv.Names) > 0 {
				sig.WriteString(recv.Names[0].Name)
				sig.WriteString(" ")
			}
			sig.WriteString(p.typeToString(recv.Type))
		}
		sig.WriteString(") ")
	}
	
	sig.WriteString(fn.Name.Name)
	
	// Add parameters
	if fn.Type.Params != nil {
		sig.WriteString("(")
		for i, param := range fn.Type.Params.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			if len(param.Names) > 0 {
				for j, name := range param.Names {
					if j > 0 {
						sig.WriteString(", ")
					}
					sig.WriteString(name.Name)
				}
				sig.WriteString(" ")
			}
			sig.WriteString(p.typeToString(param.Type))
		}
		sig.WriteString(")")
	}
	
	// Add return types
	if fn.Type.Results != nil {
		sig.WriteString(" ")
		if len(fn.Type.Results.List) > 1 {
			sig.WriteString("(")
		}
		for i, result := range fn.Type.Results.List {
			if i > 0 {
				sig.WriteString(", ")
			}
			if len(result.Names) > 0 {
				sig.WriteString(result.Names[0].Name)
				sig.WriteString(" ")
			}
			sig.WriteString(p.typeToString(result.Type))
		}
		if len(fn.Type.Results.List) > 1 {
			sig.WriteString(")")
		}
	}
	
	return sig.String()
}

// typeToString converts ast.Expr to string representation
func (p *GoTreeSitterParser) typeToString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}

	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return p.typeToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + p.typeToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + p.typeToString(t.Elt)
		}
		return "[" + p.exprToString(t.Len) + "]" + p.typeToString(t.Elt)
	case *ast.MapType:
		return "map[" + p.typeToString(t.Key) + "]" + p.typeToString(t.Value)
	case *ast.ChanType:
		prefix := "chan "
		switch t.Dir {
		case ast.RECV:
			prefix = "<-chan "
		case ast.SEND:
			prefix = "chan<- "
		}
		return prefix + p.typeToString(t.Value)
	case *ast.FuncType:
		return "func" // Simplified for now
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	default:
		return "unknown"
	}
}

// exprToString converts ast.Expr to string (simplified)
func (p *GoTreeSitterParser) exprToString(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.BasicLit:
		return e.Value
	default:
		return "..."
	}
}