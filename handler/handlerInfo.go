package handler

import (
	"go/ast"
	"sort"
	"strconv"
)

// Package includes package name and path.
type Package struct {
	Name string
	Path string
}

// HandlerInfo includes information of handler.
type HandlerInfo struct {
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

// NewHandlerInfo initializes HandlerInfo.
func NewHandlerInfo(pkgName string, pkgPath string) *HandlerInfo {
	handlerInfo := HandlerInfo{}
	handlerInfo.Method = "GET"
	handlerInfo.Pkg.Name = pkgName
	handlerInfo.Pkg.Path = pkgPath
	return &handlerInfo
}

// SortHandlersInfo sorts slice of HanlderInfo.
func SortHandlersInfo(h []*HandlerInfo) {
	sort.Slice(h, func(i, j int) bool {
		if h[i].Pkg.Name == h[j].Pkg.Name {
			return h[i].File < h[j].File
		}
		return h[i].Pkg.Name < h[j].Pkg.Name
	})
}

// SetURLFromExpr assign url to HandlerInfo from expression.
func (h *HandlerInfo) SetURLFromExpr(arg0 ast.Expr) bool {
	basicLit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return false
	}
	url, err := strconv.Unquote(basicLit.Value)
	if err != nil {
		/* handle error */
		return false
	}
	h.URL = url
	return true
}
