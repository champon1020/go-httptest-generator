package generator

import (
	"flag"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"github.com/gostaticanalysis/analysisutil"
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
	httpHandlerObj     types.Object
	httpHandlerFuncObj types.Object
	httpRequestObj     types.Object
)

var newObj = types.Universe.Lookup("new")

func initObj(imports []*types.Package) bool {
	var flg bool
	for _, p := range imports {
		if p.Path() == "net/http" {
			httpHandlerObj = p.Scope().Lookup("Handler")
			httpHandlerFuncObj = p.Scope().Lookup("HandlerFunc")
			httpRequestObj = p.Scope().Lookup("Request")
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

	handlersInfo := []*HandlerInfo{}
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		handlerInfo := NewHandlerInfo(pass.Pkg.Name(), pass.Pkg.Path())
		call, _ := n.(*ast.CallExpr)

		// http.Handle
		if arg0, arg1, fn, ok := isHttpHandle(pass, call); ok {
			handlerInfo.File = fn
			if analyzeHttpHandle(pass, handlerInfo, arg0, arg1) {
				handlersInfo = append(handlersInfo, handlerInfo)
				pass.Reportf(n.Pos(), "Handle %s %s %s", handlerInfo.URL, handlerInfo.Method, handlerInfo.Name)
			}
			return
		}

		// http.HandleFunc
		if arg0, arg1, fn, ok := isHttpHandleFunc(pass, call); ok {
			handlerInfo.File = fn
			if analyzeHttpHandleFunc(pass, handlerInfo, arg0, arg1) {
				handlersInfo = append(handlersInfo, handlerInfo)
				pass.Reportf(n.Pos(), "HandleFunc %s %s %s", handlerInfo.URL, handlerInfo.Method, handlerInfo.Name)
			}
			return
		}
	})

	// If this is not test, generate test files.
	if flag.Lookup("test.v") == nil {
		GenerateAllTests(handlersInfo)
	}

	return nil, nil
}

// Analyze when using http.Handle.
func analyzeHttpHandle(pass *analysis.Pass, handlerInfo *HandlerInfo, arg0 ast.Expr, arg1 ast.Expr) bool {
	if !parseURL(arg0, handlerInfo) {
		return false
	}

	switch arg1 := arg1.(type) {
	case *ast.CallExpr:
		// Examples:
		// http.Hanle("url", http.HandlerFunc(Index))  // OK
		// http.Hanle("url", http.HandlerFunc(index))  // Ignore
		if selector, ok := arg1.Fun.(*ast.SelectorExpr); ok {
			obj := pass.TypesInfo.ObjectOf(selector.Sel)
			if types.Identical(httpHandlerFuncObj.Type(), obj.Type()) {
				ident, ok := arg1.Args[0].(*ast.Ident)
				if !ok {
					return false
				}
				if parseHandlerFunc(pass.TypesInfo.ObjectOf(ident), handlerInfo, pass) {
					break
				}
			}
		}
		if ident, ok := arg1.Fun.(*ast.Ident); ok {
			// Examples:
			// http.Handle("url", new(AnyHandler)) // OK
			// http.Handle("url", new(anyHandler)) // Ignore
			obj := pass.TypesInfo.Uses[ident]
			if types.Identical(newObj.Type(), obj.Type()) &&
				parseAnyHandlerWithNew(arg1.Args[0], handlerInfo, pass) {
				hIdent, _ := arg1.Args[0].(*ast.Ident)
				handlerInfo.Name = hIdent.Name
				handlerInfo.IsNew = true
				break
			}
		}
		return false
	case *ast.Ident:
		obj := pass.TypesInfo.Uses[arg1]
		// Examples:
		// http.Handle("url", H) // OK
		// http.Handle("url", h) // Ignore
		// http.Handle("url", H2) // OK
		// http.Handle("url", h2) // Ignore
		if types.Identical(obj.Type(), httpHandlerFuncObj.Type()) &&
			parseHttpHandlerFunc(obj, handlerInfo, pass) {
			break
		}

		// Examples:
		// http.Handle("url", A)  // OK
		// http.Handle("url", AA) // Ignore
		// http.Handle("url", a)  // OK
		// http.Handle("url", aa) // Ignore
		if parseAnyHandler(obj, handlerInfo, pass) {
			break
		}
		return false
	}

	return true
}

// Analyze when using http.HandleFunc.
func analyzeHttpHandleFunc(pass *analysis.Pass, handlerInfo *HandlerInfo, arg0 ast.Expr, arg1 ast.Expr) bool {
	if !parseURL(arg0, handlerInfo) {
		return false
	}

	// Examples:
	// http.HandleFunc("url", Index) // OK
	// http.HandleFunc("url", index) // Ignore
	// http.HandleFunc("url", IndexVar) // OK
	// http.HandleFunc("url", IndexVar2) // OK
	// http.HandleFunc("url", IndexVar3) // Ignore
	// http.HandleFunc("url", IndexVar4) // OK
	// http.HandleFunc("url", IndexVar5) // Ignore
	ident, ok := arg1.(*ast.Ident)
	if !ok {
		return false
	}
	obj := pass.TypesInfo.ObjectOf(ident)
	if !parseHandlerFunc(obj, handlerInfo, pass) {
		return false
	}

	return true
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

// Parse URL and assign to HandlerInfo.
func parseURL(arg0 ast.Expr, handlerInfo *HandlerInfo) bool {
	basicLit, ok := arg0.(*ast.BasicLit)
	if !ok {
		return false
	}
	url, err := strconv.Unquote(basicLit.Value)
	if err != nil {
		/* handle error */
		return false
	}
	handlerInfo.URL = url
	return true
}

// Parse any handler.
func parseAnyHandler(obj types.Object, handlerInfo *HandlerInfo, pass *analysis.Pass) bool {
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
func parseAnyHandlerWithNew(h ast.Expr, handlerInfo *HandlerInfo, pass *analysis.Pass) bool {
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

// The CallExpr is whether `http.Handle` or not.
func isHttpHandle(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, ast.Expr, string, bool) {
	return searchFuncInNetHttp(pass, call, "Handle")
}

// The CallExpr is whether `http.HandleFunc` or not.
func isHttpHandleFunc(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, ast.Expr, string, bool) {
	return searchFuncInNetHttp(pass, call, "HandleFunc")
}

// The CallExpr is whether `http.HandlerFunc` or not.
func isHttpHandlerFunc(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, ast.Expr, string, bool) {
	return searchFuncInNetHttp(pass, call, "HandlerFunc")
}

// Search called function the name is `funcName`.
// If exists, return the following values:
//   - the url of argument 0
//   - the handler of argument 1
//   - file name in which http.HandleFunc is called
func searchFuncInNetHttp(pass *analysis.Pass, call *ast.CallExpr, funcName string) (ast.Expr, ast.Expr, string, bool) {
	var (
		arg0 ast.Expr
		arg1 ast.Expr
	)

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

	switch obj := obj.(type) {
	case *types.Func:
		if obj.Pkg().Path() != "net/http" || obj.Name() != funcName {
			return falseReturn()
		}
	case *types.TypeName:
		if !types.Identical(obj.Type(), httpHandlerFuncObj.Type()) {
			return falseReturn()
		}
	}

	fn := pass.Fset.File(call.Lparen).Name()
	if len(call.Args) > 0 {
		arg0 = call.Args[0]
	}
	if len(call.Args) > 1 {
		arg1 = call.Args[1]
	}

	return arg0, arg1, fn, true
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
