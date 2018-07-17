package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

type goMethod struct {
	name     string
	receiver string
	params   []string
	results  []string
}

type goStruct struct {
	name    string
	methods *[]goMethod
}

func generateStructs(target string) map[string]goStruct {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, target, nil, 0)
	if err != nil {
		panic(err)
	}
	goStructs := make(map[string]goStruct)
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			lines := []string{}
			fileName := fs.Position(file.Pos()).Filename
			f, err := os.Open(fileName)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if fn.Name.IsExported() == false {
						continue
					}
					funcName := fn.Name.Name
					receiver := ""
					params := []string{}
					results := []string{}
					if fn.Recv != nil {
						for _, p := range fn.Recv.List {
							switch xv := p.Type.(type) {
							case *ast.StarExpr:
								if si, ok := xv.X.(*ast.Ident); ok {
									receiver = si.Name
								}
							case *ast.Ident:
								receiver = xv.Name
							}
						}
						if fn.Type.Params != nil {
							for _, p := range fn.Type.Params.List {

								line := fs.Position(p.Type.Pos()).Line
								begin := fs.Position(p.Type.Pos()).Column - 1
								end := fs.Position(p.Type.End()).Column - 1
								s := lines[line-1][begin:end]
								params = append(params, s)
							}
						}
						if fn.Type.Results != nil {
							for _, p := range fn.Type.Results.List {
								line := fs.Position(p.Type.Pos()).Line
								begin := fs.Position(p.Type.Pos()).Column - 1
								end := fs.Position(p.Type.End()).Column - 1
								s := lines[line-1][begin:end]
								results = append(results, s)
							}
						}
						m := goMethod{
							name:     funcName,
							params:   params,
							results:  results,
							receiver: receiver,
						}
						if _, ok := goStructs[m.receiver]; ok {
							methods := goStructs[m.receiver].methods
							*methods = append(*methods, m)
						} else {
							methods := make([]goMethod, 1, 100)
							methods[0] = m
							goStructs[m.receiver] = goStruct{
								name:    m.receiver,
								methods: &methods,
							}
						}
					}
				}
			}
		}
	}
	return goStructs
}

func generateInterfaces(s map[string]goStruct) []byte {
	output := ""
	for i, gs := range s {
		output += fmt.Sprintf("// %sInterface is an interface for the %s struct\n", i, i)
		output += fmt.Sprintf("type %sInterface interface {\n", i)
		for _, m := range *gs.methods {
			output += "    "
			output += m.name
			output += "("
			output += strings.Join(m.params, ", ")
			output += ")"
			if len(m.results) > 0 {
				if len(m.results) > 1 {
					output += " ("
					output += strings.Join(m.results, ", ")
					output += ")"
				} else {
					output += " "
					output += m.results[0]
				}
			}
			output += "\n"
		}
		output += "}\n\n"
	}
	return []byte(output)
}

func main() {
	args := os.Args[1:]
	var target string
	if len(args) > 0 {
		target = args[0]
	} else {
		target = "."
	}

	s := generateStructs(target)
	output := generateInterfaces(s)
	fmt.Printf("%s", string(output))

}
