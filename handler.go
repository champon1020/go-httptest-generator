package generator

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "handler analyzer is ..."

// HandlerAnalyzer is ...
var HandlerAnalyzer = &analysis.Analyzer{
	Name: "handler analyzer",
	Doc:  doc,
	Run:  run,
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
	inspect := inspector.New(pass.Files)
	nodeFilter := []ast.Node{
		new(ast.CallExpr),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		handlerInfo := NewHandlerInfo()
		if arg0, arg1, ok := httpHandleFunc(n); ok {
			// Parse url.
			basicLit, _ := arg0.(*ast.BasicLit)
			url, err := strconv.Unquote(basicLit.Value)
			if err != nil {
				/* handle error */
				return
			}
			handlerInfo.URL = url

			// Parse handler block statement.
			switch arg1.(type) {
			case *ast.FuncLit:
				funcLit, _ := arg1.(*ast.FuncLit)
				parseHandleFuncBlock(funcLit.Body, handlerInfo)
			case *ast.Ident:
				// ident, _ := arg1.(*ast.Ident)
			default:
			}

			fmt.Println(handlerInfo)
			pass.Reportf(n.Pos(), "http.HandleFunc with %s", handlerInfo.URL)
		}
	})

	return nil, nil
}

func httpHandleFunc(n ast.Node) (ast.Expr, ast.Expr, bool) {
	callExpr, ok := n.(*ast.CallExpr)
	if !ok {
		return nil, nil, false
	}

	v, m, ok := accessFieldOrMethod(callExpr.Fun)
	if !ok || v.Name != "http" || m.Name != "HandleFunc" {
		return nil, nil, false
	}

	if len(callExpr.Args) != 2 {
		return nil, nil, false
	}

	return callExpr.Args[0], callExpr.Args[1], true
}

func parseHandleFuncBlock(block *ast.BlockStmt, handlerInfo *HandlerInfo) {
	for _, stmt := range block.List {
		switch stmt.(type) {
		case *ast.IfStmt:
			ifStmt, _ := stmt.(*ast.IfStmt)
			binaryExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
			if ok && binaryExpr.Op == token.NEQ {
				v, m, ok := accessFieldOrMethod(binaryExpr.X)
				if ok && v.Name == "r" && m.Name == "Method" {
					basicLit, ok := binaryExpr.Y.(*ast.BasicLit)
					if ok {
						handlerInfo.Method, _ = strconv.Unquote(basicLit.Value)
					}
				}
			}
		default:
		}
	}
}

func accessFieldOrMethod(expr ast.Expr) (*ast.Ident, *ast.Ident, bool) {
	selectorExpr, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return nil, nil, false
	}

	x, ok := selectorExpr.X.(*ast.Ident)
	if !ok {
		return nil, nil, false
	}

	return x, selectorExpr.Sel, true
}