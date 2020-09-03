package generator

import (
	"go/ast"
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

			// Parse handler block statement.
			switch arg1.(type) {
			case *ast.FuncLit:

			case *ast.Ident:
			default:
			}

			pass.Reportf(n.Pos(), "http.HandleFunc with %s", handlerInfo.URL)
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

func funcListHandler(call *ast.FuncLit) {

}
