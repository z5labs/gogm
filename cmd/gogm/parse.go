package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"strings"
)

type relConf struct {
	Field string
	Type string
	IsMany bool
}

func parseFile(filePath string, confs *map[string][]*relConf, imports map[string][]string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	if node.Scope != nil {
		if node.Scope.Objects != nil && len(node.Scope.Objects) != 0 {
			for label, config := range node.Scope.Objects {
				tSpec, ok := config.Decl.(*ast.TypeSpec)
				if !ok {
					continue
				}

				strType, ok := tSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				if node.Imports != nil && len(node.Imports) != 0 {
					var imps []string

					for _, impSpec := range node.Imports {
						imps = append(imps, impSpec.Path.Value)
					}

					imports[label] = imps
				}

				(*confs)[label] = []*relConf{}

				if strType.Fields != nil && strType.Fields.List != nil && len(strType.Fields.List) != 0 {
					for _, field := range strType.Fields.List {
						if field.Tag != nil && field.Tag.Value != "" {
							if strings.Contains(field.Tag.Value, "relationship") && strings.Contains(field.Tag.Value, "direction") {
								var typeNameBuf bytes.Buffer

								err = printer.Fprint(&typeNameBuf, fset, field.Type)
								if err != nil {
									log.Fatal(err)
								}

								t := typeNameBuf.String()

								(*confs)[label] = append((*confs)[label], &relConf{
									Field:  field.Names[0].Name,
									Type:   t,
									IsMany: strings.Contains(t, "[]"),
								})
							}
						}
					}
				}
			}
		}
	}

	return nil
}