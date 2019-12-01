package gen

import (
	"bytes"
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/mindstand/gogm/cmd/gogmcli/util"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Generate(directory string) error {
	confs := map[string][]*relConf{}
	imps := map[string][]string{}
	var edges []string
	packageName := ""

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return errors.New("file info is nil")
		}

		if info.IsDir() {
			return nil
		}

		if err != nil {
			return err
		}

		if strings.Contains(path, ".go") {
			err := parseFile(path, &confs, &edges, imps, &packageName)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	var imports []string

	for _, imp := range imps {
		imports = append(imports, imp...)
	}

	imports = util.SliceUniqMap(imports)

	for i := 0; i < len(imports); i++ {
		imports[i] = strings.Replace(imports[i], "\"", "", -1)
	}

	relations := make(map[string][]*relConf)

	// sort out relationships
	for _, fields := range confs {
		for _, field := range fields {
			if field == nil {
				return errors.New("field can not be nil")
			}

			if _, ok := relations[field.RelationshipName]; ok {
				relations[field.RelationshipName] = append(relations[field.RelationshipName], field)
			} else {
				relations[field.RelationshipName] = []*relConf{field}
			}
		}
	}

	// validate relationships (i.e even number)
	for name, rel := range relations {
		if len(rel) % 2 != 0 {
			return fmt.Errorf("relationship [%s] is invalid", name)
		}
	}

	funcs := make(map[string][]*tplRelConf)

	// set template stuff
	for _, rels := range relations {
		for _, rel := range rels {
			tplRel := &tplRelConf{
				StructName:             rel.NodeName,
				StructField:            rel.Field,
				OtherStructName:        rel.Type,
				StructFieldIsMany:      rel.IsMany,
			}

			var isSpecialEdge bool

			if util.StringSliceContains(edges, rel.Type) {
				tplRel.UsesSpecialEdge = true
				tplRel.SpecialEdgeType = rel.Type
				tplRel.SpecialEdgeDirection = rel.Direction == dsl.DirectionIncoming
				isSpecialEdge = true
			}

			searchLoop:
			for _, lookup := range rels {
				//check special edge
				 if rel.Type != lookup.NodeName && !isSpecialEdge{
					continue
				}

				switch rel.Direction {
				case dsl.DirectionOutgoing:
					if lookup.Direction == dsl.DirectionIncoming {
						tplRel.OtherStructField = lookup.Field
						tplRel.OtherStructFieldIsMany = lookup.IsMany
						if isSpecialEdge {
							tplRel.OtherStructName = lookup.NodeName
						}
						break searchLoop
					} else {
						continue
					}

				case dsl.DirectionIncoming:
					if lookup.Direction == dsl.DirectionOutgoing {
						tplRel.OtherStructField = lookup.Field
						tplRel.OtherStructFieldIsMany = lookup.IsMany
						if isSpecialEdge {
							tplRel.OtherStructName = lookup.NodeName
						}
						break searchLoop
					} else {
						continue
					}

				case dsl.DirectionNone:
					if lookup.Direction == dsl.DirectionNone {
						tplRel.OtherStructField = lookup.Field
						tplRel.OtherStructFieldIsMany = lookup.IsMany
						if isSpecialEdge {
							tplRel.OtherStructName = lookup.NodeName
						}
						break searchLoop
					} else {
						continue
					}

				case dsl.DirectionBoth:
					if lookup.Direction == dsl.DirectionBoth {
						tplRel.OtherStructField = lookup.Field
						tplRel.OtherStructFieldIsMany = lookup.IsMany
						if isSpecialEdge {
							tplRel.OtherStructName = lookup.NodeName
						}
						break searchLoop
					} else {
						continue
					}

				default:
					return fmt.Errorf("invalid direction [%v]", rel.Direction)
				}
			}

			if tplRel.OtherStructField == "" {
				return fmt.Errorf("oposite side not found for node [%s]", rel.NodeName)
			}

			if _, ok := funcs[rel.NodeName]; ok {
				funcs[rel.NodeName] = append(funcs[rel.NodeName], tplRel)
			} else {
				funcs[rel.NodeName] = []*tplRelConf{tplRel}
			}
		}
	}

	//write templates out
	tpl := template.New("linkFile")

	//register templates
	for _, templateString := range []string{singleLink, linkMany, linkSpec, unlinkSingle, unlinkMulti, unlinkSpec, masterTpl} {
		tpl, err = tpl.Parse(templateString)
		if err != nil {
			return err
		}
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, templateConfig{
		Imports:     imports,
		PackageName: packageName,
		Funcs:       funcs,
	})
	if err != nil {
		return err
	}

	f, err := os.Create(fmt.Sprintf("%s/linking.go", directory))
	if err != nil {
		return err
	}

	lenBytes, err := f.Write(buf.Bytes())
	if err != nil {
		return err
	}

	log.Printf("done after writing [%v] bytes!", lenBytes)

	return f.Close()
}
