package handler

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Parse http.HandlerFunc.
func parseHttpHandlerFunc(obj types.Object, handlerInfo *HandlerInfo, pass *analysis.Pass) bool {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}

	var flg bool
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
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

					// http.HandlerFunc
					call, ok := vSpec.Values[i].(*ast.CallExpr)
					if !ok {
						continue
					}

					// Either function variable or underlygin function declaration must be exported.
					// If function literal is exported and the scope is toplevel of application package,
					// it's ok to use test.
					if ident.IsExported() && obj.Parent() == obj.Pkg().Scope() {
						handlerInfo.IsHandlerFunc = true
						handlerInfo.Name = ident.Name
						decideName = true
					} else {
						// argment of http.HandlerFunc
						_, ok := call.Args[0].(*ast.Ident)
						if !ok {
							continue
						}
					}

					// Parse function block statement.
					if decideName && parseHandlerBlock(call.Args[0], handlerInfo, pass) {
						flg = true
						break
					}
				}
			}
		}
	})

	return flg
}

// The CallExpr is whether `http.HandlerFunc` or not.
func isHttpHandlerFunc(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, ast.Expr, string, bool) {
	return searchFuncInNetHttp(pass, call, "HandlerFunc")
}
