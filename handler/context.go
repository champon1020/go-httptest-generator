package handler

import (
	"go/ast"
	"sort"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

// Package includes package name and path.
type Package struct {
	Name string
	Path string
}

// Context includes information of handler.
type Context struct {
	pass *analysis.Pass

	Pkg    Package // Included pacakge
	File   string  // Included file name
	Name   string  // Handler name
	URL    string  // Endpoint url
	Method string  // Request method

	IsHandlerFunc bool
	IsFuncLit     bool
	IsFuncDecl    bool
	IsNew         bool
	IsInstance    bool
}

// NewContext initializes struct handler.Context.
func NewContext(pass *analysis.Pass) *Context {
	ctx := Context{}
	ctx.Method = "GET"
	ctx.pass = pass
	ctx.Pkg.Name = pass.Pkg.Name()
	ctx.Pkg.Path = pass.Pkg.Path()
	return &ctx
}

// SortContexts sorts slice of HanlderInfo.
func SortContexts(ctxs []*Context) {
	sort.Slice(ctxs, func(i, j int) bool {
		if ctxs[i].Pkg.Name == ctxs[j].Pkg.Name {
			return ctxs[i].File < ctxs[j].File
		}
		return ctxs[i].Pkg.Name < ctxs[j].Pkg.Name
	})
}

// SetURLFromExpr assign url to HandlerInfo from expression.
func (ctx *Context) SetURLFromExpr(arg0 ast.Expr) bool {
	basicLit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return false
	}
	url, err := strconv.Unquote(basicLit.Value)
	if err != nil {
		/* handle error */
		return false
	}
	ctx.URL = url
	return true
}
