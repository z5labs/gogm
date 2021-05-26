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

// gen provides code to generate link and unlink functions for gogm structs
package gen

import (
	"bytes"
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/mindstand/gogm/cmd/gogmcli/util"
	"go/format"
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Generate searches for all go source files, then generates link and unlink functions for all gogm structs
// takes in root directory and whether to log in debug mode
// note: Generate is not recursive, it only looks in the target directory
func Generate(directory string, debug bool) error {
	confs := map[string][]*relConf{}
	imps := map[string][]string{}
	var edges []string
	packageName := ""

	err := filepath.Walk(directory, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info == nil {
			return errors.New("file info is nil")
		}

		if info.IsDir() && filePath != directory {
			if debug {
				log.Printf("skipping [%s] as it is a directory\n", filePath)
			}
			return filepath.SkipDir
		}

		if path.Ext(filePath) == ".go" {
			if debug {
				log.Printf("parsing go file [%s]\n", filePath)
			}
			err := parseFile(filePath, &confs, &edges, imps, &packageName, debug)
			if err != nil {
				if debug {
					log.Printf("failed to parse go file [%s] with error '%s'\n", filePath, err.Error())
				}
				return err
			}
			if debug {
				log.Printf("successfully parsed go file [%s]\n", filePath)
			}
		} else if debug {
			log.Printf("skipping non go file [%s]\n", filePath)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if debug {
		log.Printf("found [%v] confs and [%v] rels", len(confs), len(edges))
	}

	var imports []string

	for _, imp := range imps {
		imports = append(imports, imp...)
	}

	imports = util.RemoveDuplicates(imports)

	for i := 0; i < len(imports); i++ {
		imports[i] = strings.Replace(imports[i], "\"", "", -1)
	}

	relations := make(map[string][]*relConf)

	if debug {
		log.Println("sorting relationships")
	}

	// sort out relationships
	for node, fields := range confs {
		if debug {
			log.Printf("sorting relationships for node [%s] with [%v] fields", node, len(fields))
		}
		for _, field := range fields {
			if field == nil {
				return errors.New("field can not be nil")
			}

			if debug {
				log.Printf("adding relationship [%s] from field [%s]", field.RelationshipName, field.Field)
			}

			if _, ok := relations[field.RelationshipName]; ok {
				relations[field.RelationshipName] = append(relations[field.RelationshipName], field)
			} else {
				relations[field.RelationshipName] = []*relConf{field}
			}
		}
	}

	if debug {
		log.Printf("there are [%v] relations\n", len(relations))
	}

	// validate relationships (i.e even number)
	for name, rel := range relations {
		if len(rel)%2 != 0 {
			return fmt.Errorf("relationship [%s] is invalid", name)
		}
	}

	funcs := make(map[string][]*tplRelConf)

	// set template stuff
	for _, rels := range relations {
		for _, rel := range rels {
			tplRel := &tplRelConf{
				StructName:        rel.NodeName,
				StructField:       rel.Field,
				OtherStructName:   rel.Type,
				StructFieldIsMany: rel.IsMany,
			}

			var isSpecialEdge bool

			if util.StringSliceContains(edges, rel.Type) {
				tplRel.UsesSpecialEdge = true
				tplRel.SpecialEdgeType = rel.Type
				tplRel.SpecialEdgeDirection = rel.Direction == dsl.DirectionOutgoing
				isSpecialEdge = true
			}

			err = parseDirection(rel, &rels, tplRel, isSpecialEdge, debug)
			if err != nil {
				return err
			}

			if tplRel.OtherStructField == "" {
				return fmt.Errorf("oposite side not found for node [%s] on relationship [%s] and field [%s]", rel.NodeName, rel.RelationshipName, rel.Field)
			}

			if debug {
				log.Printf("adding function to node [%s]", rel.NodeName)
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

	if debug {
		log.Printf("packageName: [%s]\n", packageName)
	}

	if len(funcs) == 0 {
		log.Printf("no functions to write, exiting")
		return nil
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

	// format generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	// create the file
	f, err := os.Create(path.Join(directory, "linking.go"))
	if err != nil {
		return err
	}

	// write code to the file
	lenBytes, err := f.Write(formatted)
	if err != nil {
		return err
	}

	if debug {
		log.Printf("done after writing [%v] bytes!", lenBytes)
	}

	// close the buffer
	err = f.Close()
	if err != nil {
		return err
	}

	log.Printf("wrote link functions to file [%s/linking.go]", directory)

	return nil
}

// parseDirection parses gogm struct tags and writes to a holder struct
func parseDirection(rel *relConf, rels *[]*relConf, tplRel *tplRelConf, isSpecialEdge, debug bool) error {
	for _, lookup := range *rels {
		// validate that these edges can point to each other
		if isSpecialEdge {
			if rel.Type != lookup.Type || lookup.RelationshipName != rel.RelationshipName {
				continue
			}
		} else {
			if rel.Type != lookup.NodeName || lookup.Type != rel.NodeName || lookup.RelationshipName != rel.RelationshipName {
				continue
			}
		}

		if debug {
			log.Printf("[%s]->[%s]", rel.Type, lookup.NodeName)
			log.Printf("[%s]<-[%s]", lookup.Type, rel.NodeName)
			log.Printf("[%s] and [%s]", lookup.RelationshipName, rel.RelationshipName)
		}

		switch rel.Direction {
		case dsl.DirectionOutgoing:
			if lookup.Direction == dsl.DirectionIncoming {
				tplRel.OtherStructField = lookup.Field
				tplRel.OtherStructFieldIsMany = lookup.IsMany
				if isSpecialEdge {
					tplRel.OtherStructName = lookup.NodeName
				}
				return nil
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
				return nil
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
				return nil
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
				return nil
			} else {
				continue
			}

		default:
			return fmt.Errorf("invalid direction [%v]", rel.Direction)
		}
	}

	return nil
}
