package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"

	"golang.org/x/tools/go/packages"
)

type Enum string

const (
	One Enum = "one"
	Two Enum = "two"
)

const (
	Three Enum = "three"
	Four  Enum = "four"
)

type SecondEnum string

const (
	SecondOne   SecondEnum = "one"
	SecondTwo   SecondEnum = "two"
	SecondThree SecondEnum = "three"
	SecondFour  SecondEnum = "four"
)

type MyEnum int

const (
	Ant MyEnum = iota
	Fly

	Cat
	TRex

	Door
	Couch
)

func main() {
	flag.Parse()

	// Many tools pass their command-line arguments (after any flags)
	// uninterpreted to packages.Load so that it can interpret them
	// according to the conventions of the underlying build system.
	cfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedSyntax}
	pkgs, err := packages.Load(cfg, flag.Args()...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	// Print the names of the source files
	// for each package listed on the command line.
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			m := make(map[string]*EnumInfo)
			m["Enum"] = &EnumInfo{
				Name:     "Enum",
				TypeName: "string",
			}
			m["MyEnum"] = &EnumInfo{
				Name:     "MyEnum",
				TypeName: "int",
			}
			PopulateEnumInfo(m, file)

			for k, v := range m {
				log.Println(k, "-", v.TypeName)
				for _, c := range v.Consts {
					log.Println(c.Name, "-", c.Value)
				}
			}
		}
	}
}

type EnumInfo struct {
	Name     string
	TypeName string // base type
	Consts   []ConstValue
}

type ConstValue struct {
	Name  string
	Value any // int or string
}

func PopulateEnumInfo(enumTypesMap map[string]*EnumInfo, file *ast.File) {
	// phase 1: iterate scope objects to get the values
	var nameValues = make(map[string]any)

	for _, object := range file.Scope.Objects {
		if object.Kind == ast.Con {
			nameValues[object.Name] = object.Data
		}
	}

	// phase 2: iterate decls to get the type and names in order
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if genDecl.Tok != token.CONST {
			continue
		}
		var enumInfo *EnumInfo
		for _, spec := range genDecl.Specs {
			valSpec := spec.(*ast.ValueSpec)
			if typeIdent, ok := valSpec.Type.(*ast.Ident); ok {
				enumInfo = enumTypesMap[typeIdent.String()]
			}
			if enumInfo != nil {
				for _, nameIdent := range valSpec.Names {
					name := nameIdent.String()
					if name == "_" {
						continue
					}
					value := nameValues[name]
					enumInfo.Consts = append(enumInfo.Consts, ConstValue{
						Name:  name,
						Value: value,
					})
				}
			}
		}
	}
}
