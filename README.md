# go-httptest-generator
HTTP Test Generator for Go.

## Description
go-httptest-generator is the tool which generates httptest template files.

## Usage

### Build
```
go build cmd/main.go
```

### Generate
Execute following command on your application project.
```
go vet -vettool /path/to/build/file pkgName
```

### Test
```
go test -v
```

## Code Limitation
To use this tool, you have to take care when you write code.
For example, either handler or handler function must be exported.
And I recommend you to write handler or handler function on the top level scope of your application pacakge.
These are because test files need to access to your handler or handler function from outside(geenrated test files).

### Examples:
```
/*
  http.HandleFunc("url", index)
*/
func Index(w http.ResponseWriter, r *http.Request){}  // OK
func index(w http.ResponseWriter, r *http.Request){}  // Ignore

var IndexVar = Index  // OK
var IndexVar2 = index  // OK
var indexVar = Index  // OK
var indexVar2= index  // Ignore
{
    var IndexVar3 = Index  // OK
    var IndexVar4 = index  // Ignore
}

/*
  http.Handle("url", new(AnyHandler))
  http.Handle("url", A)
  http.Handle("url", H)
*/
type AnyHandler struct {}  // OK
func (a *AnyHandler) ServeHTTP(w http.ResponseWriter, r *http.Requese){}

type anyHandler struct {}  // Ignore
func (a *anyHandler) ServeHTTP(w http.ResponseWriter, r *http.Requese){}

var A = AnyHandler  // OK
var A2 = anyHandler  // OK
var a = AnyHandler  // OK
var a2 = anyHandler  // Ignore
{
    var A3 = AnyHandler  // Ignore
}

var H = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){})  // OK
var h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){})  // Ignore
var H2 = http.HandlerFunc(index)  // OK
var h2 = http.HandlerFunc(Index)  // Ignore
{
    var H3 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){})  // Ignore
    var H4 = http.HandlerFunc(Index)  // OK
    var H5 = http.HandlerFunc(index)  // Ignore
}
```
