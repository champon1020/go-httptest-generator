package generator

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "handlerAnalyzer analyzes handlers and get handler information."

// HandlerAnalyzer analyzes handlers and get handler information.
var HandlerAnalyzer = &analysis.Analyzer{
	Name:     "hanlderAnalyzer",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		handlerInfo := NewHandlerInfo()
		call, _ := n.(*ast.CallExpr)
		if arg0, arg1, pkg, fn, ok := httpHandleFunc(pass, call); ok {
			// Parse url.
			basicLit, _ := arg0.(*ast.BasicLit)
			url, err := strconv.Unquote(basicLit.Value)
			if err != nil {
				/* handle error */
				return
			}
			handlerInfo.URL = url
			handlerInfo.Pkg = pkg
			handlerInfo.File = fn

			// Parse handler's block statement.
			switch arg1.(type) {
			case *ast.FuncLit:
				funcl, _ := arg1.(*ast.FuncLit)
				funcLitHandler(pass, funcl, handlerInfo)
			case *ast.Ident:
			default:
			}

			pass.Reportf(n.Pos(), "http.HandleFunc with %s %s", handlerInfo.URL, handlerInfo.Method)
		}
	})
	return nil, nil
}

// The CallExpr is whether `http.HandleFunc` or not.
// If so following values:
//   - the url of argument 0
//   - the handler of argument 1
//   - package name in which http.HandleFunc is called
//   - file name in which http.HandleFunc is called
func httpHandleFunc(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, ast.Expr, string, string, bool) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, nil, "", "", false
	}
	obj, ok := pass.TypesInfo.Uses[selector.Sel]
	if !ok {
		return nil, nil, "", "", false
	}
	fun, ok := obj.(*types.Func)
	if !ok || fun.Pkg().Path() != "net/http" || fun.Name() != "HandleFunc" {
		return nil, nil, "", "", false
	}

	pkg := fun.Pkg().Name()
	fn := pass.Fset.File(call.Lparen).Name()

	return call.Args[0], call.Args[1], pkg, fn, true
}

func funcLitHandler(pass *analysis.Pass, funcl *ast.FuncLit, handlerInfo *HandlerInfo) bool {
	params := funcl.Type.Params.List
	if len(params) != 2 {
		return false
	}
	//	req := funcl.Type.Params.List[1].Names[0] // name of parameter *http.Request
	for _, stmt := range funcl.Body.List {
		switch stmt.(type) {
		case *ast.IfStmt:
			ifStmt, _ := stmt.(*ast.IfStmt)
			searchMethodIfStmt(pass, ifStmt, handlerInfo)
		default:
		}
	}
	return true
}

func idnetHandler(pass *analysis.Pass, ident *ast.Ident, handlerinfo *HandlerInfo) bool {
	return true
}

func searchMethodIfStmt(pass *analysis.Pass, ifStmt *ast.IfStmt, handlerInfo *HandlerInfo) bool {
	binary, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok || binary.Op != token.NEQ {
		return false
	}

	selector, ok := binary.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}

	v, ok := pass.TypesInfo.Uses[ident]
	m, ok2 := pass.TypesInfo.Uses[selector.Sel]
	if !ok || !ok2 {
		return false
	}

	for _, n := range m.Pkg().Scope().Names() {
		obj := m.Pkg().Scope().Lookup(n)
		if obj == nil {
			continue
		}
		if types.Identical(types.NewPointer(obj.Type()).Underlying(), v.Type()) && m.Name() == "Method" {
			method, ok := binary.Y.(*ast.BasicLit)
			if !ok {
				continue
			}

			var err error
			handlerInfo.Method, err = strconv.Unquote(method.Value)
			if err != nil {
				continue
			}
			return true
		}
	}

	return false
}
