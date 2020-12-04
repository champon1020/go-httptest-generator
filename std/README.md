# Documentation with standard http package

## Description

You can generate test templates for api which is created with standard http package 
(For example, ```http.Handle``` or ```http.HandleFunc```).

## Code Limitation

If you use this module to generate test templates, there are some api code limitations.

### 1. Export handler or handler function
```go
type AnyHandler struct{} // OK
type anyHandler struct{} // BAD
func HFunc(w http.ResponseWriter, r *http.Request){} // OK
func hFunc(w http.ResponseWriter, r *http.Request){} // OK

var HFunc2 = hFunc // OK
var hFunc2 = HFunc // Ignore
```


### 2. Write handler or handler function at the top level scope
You have to declare the handler struct or handler function in the top level scope of your application package.
This is because the test functions need to access to your handler or handler function which is implemented out of the test files.

```go
type AnyHandler struct{} // OK
func HFunc(w http.ResponseWriter, r *http.Request){} // OK

func main(){
    type AnyHandler struct{} // BAD
    var HFunc2 = func(w http.ResponseWriter, r *http.Request){} // BAD
    var HFunc3 = HFunc // Ignore
}
```

### 3. Write if statement to check the request method
You have to write if statement to check the request method in handler function.
If this statement is not implemented, it generates as "GET" method.

```go
func Index(w http.ResponseWriter, r *http.Request){
    if r.Method != "POST" {
        return
    }
    ...
}
```

## Supported Format Examples
```
```
