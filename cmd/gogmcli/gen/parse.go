package gen

import (
	"bytes"
	"errors"
	go_cypherdsl "github.com/mindstand/go-cypherdsl"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"strings"
)

type relConf struct {
	NodeName string
	Field string
	RelationshipName string
	Type string
	IsMany bool
	Direction go_cypherdsl.Direction
}

func parseFile(filePath string, confs *map[string][]*relConf, edges *[]string, imports map[string][]string, packageName *string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	if node.Scope != nil {
		*packageName = node.Name.Name
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

				//check if its a special edge
				isEdge, err := parseGogmEdge(node, label)
				if err != nil {
					return err
				}

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

	//check if its an edge node
	if !GetStartNode || !GetStartNodeType || !SetStartNode || !GetEndNode || !GetEndNodeType || !SetEndNode {
		return false, nil
	}

	return true, nil
}

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
					part = strings.Replace(strings.Replace(part, "`gogm:", "", -1), "\"", "", -1)
					if strings.Contains(part, "relationship") && strings.Contains(part, "direction") {
						gogmParts := strings.Split(part, ";")

						var dir go_cypherdsl.Direction
						var relName string
						for _, p := range gogmParts {
							if strings.Contains(p, "direction") {
								str := strings.ToLower(strings.Replace(strings.Replace(p, "direction=", "", -1), "\"", "", -1))
								switch str{
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
								relName = strings.ToLower(strings.Replace(strings.Replace(p, "relationship=", "", -1), "\"", "", -1))
							}
						}

						var typeNameBuf bytes.Buffer

						err := printer.Fprint(&typeNameBuf, fset, field.Type)
						if err != nil {
							log.Fatal(err)
						}

						t := typeNameBuf.String()

						(*confs)[label] = append((*confs)[label], &relConf{
							Field:  field.Names[0].Name,
							RelationshipName: relName,
							Type:   strings.Replace(strings.Replace(t, "[]", "", -1), "*", "", -1),
							IsMany: strings.Contains(t, "[]"),
							Direction: dir,
							NodeName: label,
						})
					}
				}
			}
		}
	}

	return nil
}