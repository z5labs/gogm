// Copyright (c) 2021 MindStand Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package gen

import (
	"bytes"
	"errors"
	go_cypherdsl "github.com/mindstand/go-cypherdsl"
	"github.com/mindstand/gogm/v2/cmd/gogmcli/util"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"strings"
)

type relConf struct {
	NodeName         string
	Field            string
	RelationshipName string
	Type             string
	IsMany           bool
	Direction        go_cypherdsl.Direction
}

// parses each file using ast looking for nodes to handle
func parseFile(filePath string, confs *map[string][]*relConf, edges *[]string, imports map[string][]string, packageName *string, debug bool) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	if node.Scope != nil {
		*packageName = node.Name.Name
		if node.Scope.Objects != nil && len(node.Scope.Objects) != 0 {
			for label, config := range node.Scope.Objects {
				log.Println("checking ", label)
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

				//check if its a special edge
				isEdge, err := parseGogmEdge(node, label)
				if err != nil {
					return err
				}

				if debug {
					log.Printf("node [%s] is edge [%v]", label, isEdge)
				}

				// if its not an edge, parse it as a gogm struct
				if !isEdge {
					(*confs)[label] = []*relConf{}
					err = parseGogmNode(strType, confs, label, fset)
					if err != nil {
						return err
					}
				} else {
					*edges = append(*edges, label)
				}

			}
		}
	}

	return nil
}

//parseGogmEdge: checks if node implements `IEdge`
func parseGogmEdge(node *ast.File, label string) (bool, error) {
	if node == nil {
		return false, errors.New("node can not be nil")
	}

	var GetStartNode, GetStartNodeType, SetStartNode, GetEndNode, GetEndNodeType, SetEndNode bool

	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl != nil {
			if funcDecl.Recv != nil {
				if funcDecl.Recv.List != nil {
					if len(funcDecl.Recv.List) != 0 {
						if len(funcDecl.Recv.List[0].Names) != 0 {
							decl, ok := funcDecl.Recv.List[0].Names[0].Obj.Decl.(*ast.Field)
							if !ok {
								continue
							}

							startType, ok := decl.Type.(*ast.StarExpr)
							if !ok {
								continue
							}

							x, ok := startType.X.(*ast.Ident)
							if !ok {
								continue
							}

							//check that the function is the right type
							if x.Name != label {
								continue
							}
						}
					} else {
						continue
					}

					switch funcDecl.Name.Name {
					case "GetStartNode":
						GetStartNode = true
						break
					case "GetStartNodeType":
						GetStartNodeType = true
						break
					case "SetStartNode":
						SetStartNode = true
						break
					case "GetEndNode":
						GetEndNode = true
						break
					case "GetEndNodeType":
						GetEndNodeType = true
						break
					case "SetEndNode":
						SetEndNode = true
						break
					default:
						continue
					}
				}
			}
		}
	}
	//check if its an edge node
	return GetStartNode && GetStartNodeType && SetStartNode && GetEndNode && GetEndNodeType && SetEndNode, nil
}

// parseGogmNode generates configuration for GoGM struct
func parseGogmNode(strType *ast.StructType, confs *map[string][]*relConf, label string, fset *token.FileSet) error {
	if strType.Fields != nil && strType.Fields.List != nil && len(strType.Fields.List) != 0 {
	fieldLoop:
		for _, field := range strType.Fields.List {
			if field.Tag != nil && field.Tag.Value != "" {
				parts := strings.Split(field.Tag.Value, " ")
				for _, part := range parts {
					if !strings.Contains(part, "gogm") {
						continue
					}
					part = util.RemoveFromString(part, "gogm:", "\"", "`")
					if strings.Contains(part, "relationship") && strings.Contains(part, "direction") {
						gogmParts := strings.Split(part, ";")

						var dir go_cypherdsl.Direction
						var relName string
						for _, p := range gogmParts {
							if strings.Contains(p, "direction") {
								str := util.RemoveFromString(p, "direction=", "\"")
								switch str {
								case "incoming":
									dir = go_cypherdsl.DirectionIncoming
									break
								case "outgoing":
									dir = go_cypherdsl.DirectionOutgoing
									break
								case "both":
									dir = go_cypherdsl.DirectionBoth
									break
								case "none":
									dir = go_cypherdsl.DirectionNone
									break
								default:
									log.Printf("direction %s not found", str)
									continue fieldLoop
								}
							} else if strings.Contains(part, "relationship") {
								relName = strings.ToLower(util.RemoveFromString(p, "relationship=", "\"", "`"))
							}
						}

						var typeNameBuf bytes.Buffer

						err := printer.Fprint(&typeNameBuf, fset, field.Type)
						if err != nil {
							log.Fatal(err)
						}

						t := typeNameBuf.String()

						(*confs)[label] = append((*confs)[label], &relConf{
							Field:            field.Names[0].Name,
							RelationshipName: relName,
							Type:             strings.Replace(strings.Replace(t, "[]", "", -1), "*", "", -1),
							IsMany:           strings.Contains(t, "[]"),
							Direction:        dir,
							NodeName:         label,
						})
					}
				}
			}
		}
	}

	return nil
}
