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

var (
	httpServeMux types.Object
	httpRequest  types.Object
)

func initObj(imports []*types.Package) bool {
	var flg bool
	for _, p := range imports {
		if p.Path() == "net/http" {
			httpServeMux = p.Scope().Lookup("ServeMux")
			httpRequest = p.Scope().Lookup("Request")
			flg = true
			break
		}
	}
	return flg
}

func run(pass *analysis.Pass) (interface{}, error) {
	if ok := initObj(pass.Pkg.Imports()); !ok {
		/* handle error */
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		handlerInfo := NewHandlerInfo(pass.Pkg.Name())
		call, _ := n.(*ast.CallExpr)

		// http.HandleFunc
		if arg0, arg1, fn, ok := httpHandleFunc(pass, call); ok {
			handlerInfo.File = fn
			if ok := parseURL(arg0, handlerInfo); !ok {
				return
			}
			if ok := parseHandlerBlock(arg1, handlerInfo, pass); !ok {
				return
			}

			pass.Reportf(n.Pos(), "HandleFunc %s %s", handlerInfo.URL, handlerInfo.Method)
		}

		// http.ServeMux.Handle
		if arg0, arg1, fn, ok := muxHandle(pass, call); ok {
			handlerInfo.File = fn
			if ok := parseURL(arg0, handlerInfo); !ok {
				return
			}
			if ok := parseHandlerBlock(arg1, handlerInfo, pass); !ok {
				return
			}

			pass.Reportf(n.Pos(), "mux %s %s", handlerInfo.URL, handlerInfo.Method)
		}
	})
	return nil, nil
}

// Parse URL and assign to HandlerInfo.
func parseURL(arg0 ast.Expr, handlerInfo *HandlerInfo) bool {
	basicLit, _ := arg0.(*ast.BasicLit)
	url, err := strconv.Unquote(basicLit.Value)
	if err != nil {
		/* handle error */
		return false
	}
	handlerInfo.URL = url
	return true
}

// Parse handler block statement and assign handler information to HandlerInfo.
func parseHandlerBlock(arg1 ast.Expr, handlerInfo *HandlerInfo, pass *analysis.Pass) bool {
	switch arg1.(type) {
	case *ast.FuncLit:
		funcl, _ := arg1.(*ast.FuncLit)
		funcLitHandler(pass, funcl, handlerInfo)
	case *ast.Ident:
	case *ast.CallExpr:
	default:
	}
	return true
}

// The CallExpr is whether `http.HandleFunc` or not.
// If so following values:
//   - the url of argument 0
//   - the handler of argument 1
//   - file name in which http.HandleFunc is called
func httpHandleFunc(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, ast.Expr, string, bool) {
	falseReturn := func() (ast.Expr, ast.Expr, string, bool) {
		return nil, nil, "", false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return falseReturn()
	}
	obj, ok := pass.TypesInfo.Uses[selector.Sel]
	if !ok {
		return falseReturn()
	}
	fun, ok := obj.(*types.Func)
	if !ok || fun.Pkg().Path() != "net/http" || fun.Name() != "HandleFunc" {
		return falseReturn()
	}

	fn := pass.Fset.File(call.Lparen).Name()

	return call.Args[0], call.Args[1], fn, true
}

// The CallExpr is whether `mux.HandleFunc` or not.
// Return values is same to httpHandleFunc.
func muxHandle(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, ast.Expr, string, bool) {
	falseReturn := func() (ast.Expr, ast.Expr, string, bool) {
		return nil, nil, "", false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return falseReturn()
	}
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return falseReturn()
	}

	v, ok := pass.TypesInfo.Uses[ident]
	m, ok2 := pass.TypesInfo.Uses[selector.Sel]
	if !ok || !ok2 || m.Name() != "Handle" {
		return falseReturn()
	}

	ptr, ok := v.Type().Underlying().(*types.Pointer)
	if !ok || !types.Identical(httpServeMux.Type(), ptr.Elem()) {
		return falseReturn()
	}

	fn := pass.Fset.File(call.Lparen).Name()

	return call.Args[0], call.Args[1], fn, true
}

func funcLitHandler(pass *analysis.Pass, funcl *ast.FuncLit, handlerInfo *HandlerInfo) bool {
	params := funcl.Type.Params.List
	if len(params) != 2 {
		return false
	}
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
	if ok && types.Identical(httpRequest.Type(), ptr.Elem()) {
		var err error
		handlerInfo.Method, err = strconv.Unquote(method.Value)
		if err != nil {
			return false
		}
	}

	return true
}
