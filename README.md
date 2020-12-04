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

### Run Sample
```
cd sample
rm sample1_test.go sample2_test.go sample3_test.go
sh run.sh
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
