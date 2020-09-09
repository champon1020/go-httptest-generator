package handler

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

// Parse handler.
func parseHandler(obj types.Object, handlerInfo *HandlerInfo, pass *analysis.Pass) bool {
	// Handler must be at the toplevel scope of application pacakge.
	if obj.Parent() != obj.Pkg().Scope() {
		return false
	}

	// Eithre underlying handler or handler variable must be exported.
	for ident, o := range pass.TypesInfo.Uses {
		// If handler variable is exported, it's ok.
		if obj.Exported() {
			handlerInfo.IsInstance = true
			handlerInfo.Name = obj.Name()
			return parseServeHttp(obj.Type().Underlying(), handlerInfo, pass)
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
			if !ok {
				continue
			}
			if !ident.IsExported() {
				continue
			}
			handlerInfo.IsNew = true
			handlerInfo.Name = ident.Name
			return parseServeHttp(obj.Type().Underlying(), handlerInfo, pass)
		}
	}

	return false
}

// Parse any handler with builtin new.
func parseHandlerWithNew(h ast.Expr, handlerInfo *HandlerInfo, pass *analysis.Pass) bool {
	hIdent, ok := h.(*ast.Ident)
	if !ok {
		return false
	}

	// types.TypeName
	handler := pass.TypesInfo.Uses[hIdent]
	if !handler.Exported() {
		return false
	}

	hIface := httpHandlerObj.Type().Underlying().(*types.Interface)
	if !types.Implements(handler.Type(), hIface) &&
		!types.Implements(types.NewPointer(handler.Type()), hIface) {
		return false
	}

	ok = parseServeHttp(handler.Type(), handlerInfo, pass)
	return ok
}

// Parse ServeHTTP method of http.Handler interface.
func parseServeHttp(typ types.Type, handlerInfo *HandlerInfo, pass *analysis.Pass) bool {
	m := analysisutil.MethodOf(typ, "ServeHTTP")
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	var flg bool
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		fDecl, _ := n.(*ast.FuncDecl)
		if m == pass.TypesInfo.Defs[fDecl.Name] {
			if parseHandlerBlock(fDecl, handlerInfo, pass) {
				flg = true
				return
			}
		}
	})
	return flg
}

// Parse handler function.
func parseHandlerFunc(obj types.Object, handlerInfo *HandlerInfo, pass *analysis.Pass) bool {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.GenDecl)(nil),
	}

	var flg bool
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		// Index
		case *ast.FuncDecl:
			// Search same object.
			if obj != pass.TypesInfo.ObjectOf(n.Name) {
				return
			}

			// Function declaration must be exported.
			if !n.Name.IsExported() {
				return
			}

			// Parse function block statement.
			if parseHandlerBlock(n, handlerInfo, pass) {
				handlerInfo.IsFuncDecl = true
				handlerInfo.Name = n.Name.Name
				flg = true
				break
			}
		// IndexVar
		case *ast.GenDecl:
			for _, s := range n.Specs {
				vSpec, ok := s.(*ast.ValueSpec)
				if !ok {
					continue
				}

				var decideName bool
				for i, ident := range vSpec.Names {
					// Search same object.
					if obj != pass.TypesInfo.ObjectOf(ident) {
						continue
					}

					// Either function variable or underlygin function declaration must be exported.
					// If function literal is exported and the scope is toplevel of application package,
					// it's ok to use test.
					if ident.IsExported() && obj.Parent() == obj.Pkg().Scope() {
						handlerInfo.IsFuncLit = true
						handlerInfo.Name = ident.Name
						decideName = true
					} else {
						handlerInfo.IsFuncDecl = true
						ident, ok := vSpec.Values[i].(*ast.Ident)
						if !ok {
							continue
						}

						if ident.IsExported() {
							handlerInfo.Name = ident.Name
							decideName = true
						}
					}

					// Parse function block statement.
					if decideName && parseHandlerBlock(vSpec.Values[i], handlerInfo, pass) {
						flg = true
						break
					}
				}
			}
		}
	})

	return flg
}

// Parse handler block statement and assign handler information to HandlerInfo.
func parseHandlerBlock(n ast.Node, handlerInfo *HandlerInfo, pass *analysis.Pass) bool {
	switch n := n.(type) {
	case *ast.FuncLit:
		funcLitHandler(pass, n, handlerInfo)
	case *ast.FuncDecl:
		funcDeclHandler(pass, n, handlerInfo)
	case *ast.Ident:
		decl, ok := n.Obj.Decl.(*ast.FuncDecl)
		if !ok {
			return false
		}
		funcDeclHandler(pass, decl, handlerInfo)
	default:
	}
	return true
}

// Parse function block whose type is ast.FuncLit.Body.
func funcLitHandler(pass *analysis.Pass, funcl *ast.FuncLit, handlerInfo *HandlerInfo) bool {
	params := funcl.Type.Params.List
	if len(params) != 2 {
		return false
	}
	parseBlockStmt(pass, funcl.Body, handlerInfo)
	return true
}

// Parse function block whose type is ast.FuncDecl.Body.
func funcDeclHandler(pass *analysis.Pass, fDecl *ast.FuncDecl, handlerInfo *HandlerInfo) bool {
	params := fDecl.Type.Params.List
	if len(params) != 2 {
		return false
	}
	parseBlockStmt(pass, fDecl.Body, handlerInfo)
	return true
}

// Parse ast.BlockStmt and assign hander information to handlerInfo.
func parseBlockStmt(pass *analysis.Pass, body *ast.BlockStmt, handlerInfo *HandlerInfo) {
	for _, stmt := range body.List {
		switch stmt.(type) {
		case *ast.IfStmt:
			ifStmt, _ := stmt.(*ast.IfStmt)
			searchMethodIfStmt(pass, ifStmt, handlerInfo)
		default:
		}
	}
}

// Search statement `if (*http.Request).Method != <Method Name>`.
// If not exists, default method is 'GET'.
func searchMethodIfStmt(pass *analysis.Pass, ifStmt *ast.IfStmt, handlerInfo *HandlerInfo) bool {
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
		handlerInfo.Method, err = strconv.Unquote(method.Value)
		if err != nil {
			return false
		}
	}

	return true
}
