package std

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"github.com/gostaticanalysis/analysisutil"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// parseHandler analyze the constructed handler object.
func parseHandler(pass *analysis.Pass, h *Handler, obj types.Object) bool {
	// Handler must be at the toplevel scope of application pacakge.
	if obj.Parent() != obj.Pkg().Scope() {
		return false
	}

	// Either underlying handler or handler variable must be exported.
	for ident, o := range pass.TypesInfo.Uses {
		// If handler variable is exported, ok.
		if obj.Exported() {
			h.TypeFlg |= (1 << InstanceH)
			h.Name = obj.Name()
			return parseServeHTTP(pass, h, obj.Type().Underlying())
		}
		if o != obj {
			continue
		}

		vSpec, ok := ident.Obj.Decl.(*ast.ValueSpec)
		if !ok {
			continue
		}

		for _, v := range vSpec.Values {
			uExpr, ok := v.(*ast.UnaryExpr)
			if !ok {
				continue
			}

			compLit, ok := uExpr.X.(*ast.CompositeLit)
			if !ok {
				continue
			}

			ident, ok := compLit.Type.(*ast.Ident)
			if !ok || !ident.IsExported() {
				continue
			}

			h.TypeFlg |= (1 << NewBuiltinH)
			h.Name = ident.Name
			return parseServeHTTP(pass, h, obj.Type().Underlying())
		}
	}

	return false
}

// parseHandlerWithNewBuiltin parse the handler with new builtin statement.
func parseHandlerWithNewBuiltin(pass *analysis.Pass, h *Handler, expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}

	handler := pass.TypesInfo.Uses[ident]
	if !handler.Exported() {
		return false
	}

	iface := httpHandlerObj.Type().Underlying().(*types.Interface)
	if !types.Implements(handler.Type(), iface) &&
		!types.Implements(types.NewPointer(handler.Type()), iface) {
		return false
	}

	ok = parseServeHTTP(pass, h, handler.Type())
	return ok
}

// parseHandlerFunc parse the handler function.
func parseHandlerFunc(pass *analysis.Pass, h *Handler, obj types.Object) bool {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.GenDecl)(nil),
	}

	done := false
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		// Parse handler function which is declared as function.
		// Example:
		//      func Index(w http.ResponseWriter, r *http.Request) { ... }
		case *ast.FuncDecl:
			// Search same object.
			if obj != pass.TypesInfo.ObjectOf(n.Name) {
				return
			}

			// Function declaration must be exported.
			if !n.Name.IsExported() {
				return
			}

			// Parse function block.
			if parseHandlerFuncBlock(pass, h, n) {
				h.TypeFlg |= (1 << FuncDeclH)
				h.Name = n.Name.Name
				done = true
				break
			}
		// Parse handler function which is declared as variable.
		// Example:
		//      Index := func(w http.ResponseWriter, r *http.Request) { ... }
		case *ast.GenDecl:
			for _, s := range n.Specs {
				vSpec, ok := s.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for i, ident := range vSpec.Names {
					// Search same object.
					if obj != pass.TypesInfo.ObjectOf(ident) {
						continue
					}

					// Either function variable or underlygin function declaration must be exported.
					// If function literal is exported and the scope is toplevel of application package,
					// it's ok to use test.
					if ident.IsExported() && obj.Parent() == obj.Pkg().Scope() {
						h.TypeFlg |= (1 << FuncLitH)
						h.Name = ident.Name
					} else {
						ident, ok := vSpec.Values[i].(*ast.Ident)
						if !ok {
							continue
						}

						h.TypeFlg |= (1 << FuncDeclH)
						if ident.IsExported() {
							h.Name = ident.Name
						}
					}

					// Parse function block.
					if h.Name != "" && parseHandlerFuncBlock(pass, h, vSpec.Values[i]) {
						done = true
						break
					}
				}
			}
		}
	})

	return done
}

// parseServeHTTP parse the ServeHTTP method of http.Handler interface.
func parseServeHTTP(pass *analysis.Pass, h *Handler, typ types.Type) bool {
	m := analysisutil.MethodOf(typ, "ServeHTTP")
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	done := false
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		fDecl, _ := n.(*ast.FuncDecl)
		if m == pass.TypesInfo.Defs[fDecl.Name] {
			if parseHandlerFuncBlock(pass, h, fDecl) {
				done = true
				return
			}
		}
	})
	return done
}

// Parse handler function block and assign handler information to pointer of Handler.
func parseHandlerFuncBlock(pass *analysis.Pass, h *Handler, n ast.Node) bool {
	done := false
	switch n := n.(type) {
	case *ast.FuncLit:
		done = funcLitHandler(pass, h, n)
	case *ast.FuncDecl:
		done = funcDeclHandler(pass, h, n)
	case *ast.Ident:
		decl, ok := n.Obj.Decl.(*ast.FuncDecl)
		if !ok {
			return false
		}
		done = funcDeclHandler(pass, h, decl)
	}

	return done
}

// Parse function block whose type is ast.FuncLit.Body.
func funcLitHandler(pass *analysis.Pass, h *Handler, funcl *ast.FuncLit) bool {
	params := funcl.Type.Params.List
	if len(params) != 2 {
		return false
	}

	done := parseBlockStmt(pass, h, funcl.Body)
	return done
}

// Parse function block whose type is ast.FuncDecl.Body.
func funcDeclHandler(pass *analysis.Pass, h *Handler, fDecl *ast.FuncDecl) bool {
	params := fDecl.Type.Params.List
	if len(params) != 2 {
		return false
	}

	done := parseBlockStmt(pass, h, fDecl.Body)
	return done
}

// Parse ast.BlockStmt and assign hander information to handlerInfo.
func parseBlockStmt(pass *analysis.Pass, h *Handler, body *ast.BlockStmt) bool {
	done := false
	for _, stmt := range body.List {
		switch stmt.(type) {
		case *ast.IfStmt:
			ifStmt, _ := stmt.(*ast.IfStmt)
			done = parseMethodIfStmt(pass, h, ifStmt)
		}
	}
	return done
}

// Parse statement `if (*http.Request).Method != <Method Name>`.
// If not exists, default method is 'GET'.
func parseMethodIfStmt(pass *analysis.Pass, h *Handler, ifStmt *ast.IfStmt) bool {
	binary, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok || binary.Op != token.NEQ {
		return false
	}

	selector, ok := binary.X.(*ast.SelectorExpr)
	method, ok2 := binary.Y.(*ast.BasicLit)
	if !ok || !ok2 {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	v, ok := pass.TypesInfo.Uses[ident]
	m, ok2 := pass.TypesInfo.Uses[selector.Sel]
	if !ok || !ok2 || m.Name() != "Method" {
		return false
	}

	ptr, ok := v.Type().Underlying().(*types.Pointer)
	if ok && types.Identical(httpRequestObj.Type(), ptr.Elem()) {
		var err error
		h.Method, err = strconv.Unquote(method.Value)
		if err != nil {
			return false
		}
	}

	return true
}
