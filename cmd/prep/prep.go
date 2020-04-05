package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/loader"
)

type (
	queryFinder struct {
		packageInfo *types.Info
		queries     []string
	}
)

func main() {
	var (
		sourcePackage = flag.String("f", "", "source package import path, i.e. github.com/my/package")
	)
	flag.Parse()

	if *sourcePackage == "" {
		flag.PrintDefaults()
		return
	}

	conf := loader.Config{
		TypeChecker: types.Config{
			FakeImportC:              false,
			DisableUnusedImportCheck: true,
			Error:                    func(err error) {},
		},
		TypeCheckFuncBodies: func(path string) bool {
			return strings.HasPrefix(path, *sourcePackage)
		},
	}

	conf.Import(*sourcePackage)

	prog, err := conf.Load()
	if err != nil {
		log.Fatalf("prep: failed to load package: %v\n", err)
	}
	pkg := prog.Package(*sourcePackage)

	finder := &queryFinder{
		packageInfo: &pkg.Info,
	}

	for _, file := range pkg.Files {
		ast.Walk(finder, file)
	}

	path, err := getPathToPackage(*sourcePackage)
	if err != nil {
		log.Fatalf("prep: %v", err)
	}

	outputFileName := filepath.Join(path, "prepared_statements.go")

	queries := uniqueStrings(finder.queries)

	if len(queries) == 0 {
		log.Fatalf("prep: no SQL queries found")
	}

	code := generateCode(pkg.Pkg.Name(), *sourcePackage, queries)
	file, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("prep: failed to create file: %v", err)
	}
	defer file.Close()

	if _, err := file.Write(code); err != nil {
		log.Fatalf("prep: failed to write generated code to the file: %v", err)
	}
}

func getPathToPackage(importPath string) (string, error) {
	p, err := build.Default.Import(importPath, "", build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("failed to detect absolute path of the package %q: %v", importPath, err)
	}

	return filepath.Join(p.SrcRoot, p.ImportPath), nil
}

func generateCode(packageName, importPath string, queries []string) []byte {
	buf := bytes.NewBuffer([]byte{})

	if len(queries) == 0 {
		fmt.Fprintf(buf,
			"//go:generate prep -f %s\n\npackage %s\n\nfunc init() {\n\tprepStatements = []string{}\n}",
			importPath, packageName)

		return buf.Bytes()
	}

	fmt.Fprintf(buf,
		"//go:generate prep -f %s\n\npackage %s\n\nfunc init() {\n\tprepStatements = []string{\n\t\t%s,\n\t}\n}",
		importPath, packageName, strings.Join(queries, ",\n\t\t"))
	return buf.Bytes()
}

// uniqueStrings returns a sorted slice of the unique strings
// from the given strings slice
func uniqueStrings(strings []string) []string {
	m := make(map[string]struct{})
	for _, s := range strings {
		m[s] = struct{}{}
	}

	var unique []string
	for s := range m {
		unique = append(unique, s)
	}

	sort.Strings(unique)
	return unique
}

// maps method name to the interface it implements
var methodImplements = map[string]string{
	"ExecContext":         "ExecContext",
	"QueryContext":        "QueryContext",
	"QueryRowContext":     "QueryRowContext",
	"NamedExecContext":    "NamedExecContext",
	"GetContext":          "GetContext",
	"SelectContext":       "SelectContext",
	"NamedQueryContext":   "NamedQueryContext",
	"PrepareContext":      "PrepareContext",
	"PrepareNamedContext": "PrepareNamedContext",
}

// Visit implements ast.Visitor interface
func (f *queryFinder) Visit(node ast.Node) ast.Visitor {
	fCall, ok := node.(*ast.CallExpr)
	if !ok {
		return f
	}

	selector, ok := fCall.Fun.(*ast.SelectorExpr)
	if !ok {
		return f
	}

	interfaceName := methodImplements[selector.Sel.Name]
	if interfaceName == "" {
		return f
	}

	var query string
	switch selector.Sel.Name {
	case "ExecContext", "QueryContext", "QueryRowContext", "NamedExecContext", "NamedQueryContext", "PrepareContext", "PrepareNamedContext":
		query = f.processQuery(fCall.Args[1])
	case "GetContext", "SelectContext":
		query = f.processQuery(fCall.Args[2])
	}

	if query != "" {
		f.queries = append(f.queries, query)
	}

	return nil
}

// processQuery returns a string value of the expression if the
// expression is either a string literal or a string constant otherwise
// an empty string is returned
func (f *queryFinder) processQuery(queryArg ast.Expr) string {
	switch q := queryArg.(type) {
	case *ast.BasicLit:
		return q.Value
	case *ast.Ident:
		obj := f.packageInfo.ObjectOf(q)

		if constant, ok := obj.(*types.Const); ok {
			return constant.Val().ExactString()
		}
	}

	return ""
}
