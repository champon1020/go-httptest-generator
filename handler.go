package generator

import (
	"fmt"
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "handler analyzer is ..."

// HandlerAnalyzer is ...
var HandlerAnalyzer = &analysis.Analyzer{
	Name:     "handler analyzer",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

type HandlerInfo struct {
	URL    string
	Method string
}

func NewHandlerInfo() *HandlerInfo {
	handlerInfo := HandlerInfo{}
	handlerInfo.Method = "GET"
	return &handlerInfo
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		call, _ := n.(*ast.CallExpr)
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		obj, ok := pass.TypesInfo.Uses[selector.Sel]
		if !ok {
			return
		}
		fun, ok := obj.(*types.Func)
		if !ok || fun.Pkg().Path() != "net/http" || fun.Name() != "HandleFunc" {
			return
		}
		fmt.Println(obj)
	})
	return nil, nil
}
