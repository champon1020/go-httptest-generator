package std

import (
	"go/ast"
	"sort"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

// HandlerType is the format of handler declaration.
type HandlerType int

// Handler types of declaration.
const (
	InstanceH HandlerType = iota
	NewBuiltinH
	FuncDeclH
	FuncLitH
	HandlerFuncH
)

// Package includes package name and path.
type Package struct {
	Name string
	Path string
}

// Handler includes information of handler.
type Handler struct {
	Pkg     Package // Included pacakge
	File    string  // Included file name
	Name    string  // Handler function name
	URL     string  // Endpoint url
	Method  string  // Request method
	TypeFlg uint
}

// SetURLFromExpr assign url to Handler from expression.
func (h *Handler) SetURLFromExpr(expr ast.Expr) bool {
	basicLit, ok := expr.(*ast.BasicLit)
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

// NewHandler initializes struct handler.Context.
func NewHandler(pass *analysis.Pass) *Handler {
	h := Handler{}
	h.Method = "GET"
	h.Pkg.Name = pass.Pkg.Name()
	h.Pkg.Path = pass.Pkg.Path()
	return &h
}

// SortHandlers sort the slice of Hanlder.
// Sort with package name.
// If package name is same, sort with file name.
func SortHandlers(hs []*Handler) {
	sort.Slice(hs, func(i, j int) bool {
		if hs[i].Pkg.Name == hs[j].Pkg.Name {
			return hs[i].File < hs[j].File
		}

		return hs[i].Pkg.Name < hs[j].Pkg.Name
	})
}
