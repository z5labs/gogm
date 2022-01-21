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

package gogm

import (
	"errors"
	"fmt"
	"reflect"

	dsl "github.com/mindstand/go-cypherdsl"
)

// Specifies query based load strategy
type LoadStrategy int

const (
	// PathLoadStrategy uses cypher path
	PATH_LOAD_STRATEGY LoadStrategy = iota
	// SchemaLoadStrategy generates queries specifically from generated schema
	SCHEMA_LOAD_STRATEGY
)

func (ls LoadStrategy) validate() error {
	switch ls {
	case PATH_LOAD_STRATEGY, SCHEMA_LOAD_STRATEGY:
		return nil
	default:
		return fmt.Errorf("invalid load strategy %d", ls)
	}
}

// PathLoadStrategyMany loads many using path strategy
func PathLoadStrategyMany(variable, label string, depth int, additionalConstraints dsl.ConditionOperator) (dsl.Cypher, error) {
	if variable == "" {
		return nil, errors.New("variable name cannot be empty")
	}

	if label == "" {
		return nil, errors.New("label can not be empty")
	}

	if depth < 0 {
		return nil, errors.New("depth can not be less than 0")
	}

	path := dsl.Path().
		P().
		V(dsl.V{Name: variable})

	if depth != 0 {
		path = path.
			E(dsl.E{Direction: dsl.DirectionNone, MinJumps: 0, MaxJumps: depth}).
			V(dsl.V{})
	}

	builder := dsl.QB().
		Match(path.Build())

	if additionalConstraints != nil {
		builder = builder.Where(additionalConstraints)
	}

	return builder.Return(false, dsl.ReturnPart{Name: "p"}), nil
}

// PathLoadStrategyOne loads one object using path strategy
func PathLoadStrategyOne(variable, label, fieldOn, paramName string, isGraphId bool, depth int, additionalConstraints dsl.ConditionOperator) (dsl.Cypher, error) {
	if variable == "" {
		return nil, errors.New("variable name cannot be empty")
	}

	if label == "" {
		return nil, errors.New("label can not be empty")
	}

	if depth < 0 {
		return nil, errors.New("depth can not be less than 0")
	}

	path := dsl.Path().
		P().
		V(dsl.V{Name: variable})

	if depth != 0 {
		path = path.
			E(dsl.E{Direction: dsl.DirectionNone, MinJumps: 0, MaxJumps: depth}).
			V(dsl.V{})
	}

	builder := dsl.QB().
		Match(path.Build())

	var condition *dsl.ConditionConfig
	if isGraphId {
		condition = &dsl.ConditionConfig{
			FieldManipulationFunction: "ID",
			Name:                      variable,
			ConditionOperator:         dsl.EqualToOperator,
			Check:                     dsl.ParamString("$" + paramName),
		}
	} else {
		condition = &dsl.ConditionConfig{
			Name:              variable,
			Field:             fieldOn,
			ConditionOperator: dsl.EqualToOperator,
			Check:             dsl.ParamString("$" + paramName),
		}
	}

	if additionalConstraints != nil {
		builder = builder.Where(additionalConstraints.And(condition))
	} else {
		builder = builder.Where(dsl.C(condition))
	}

	return builder.Return(false, dsl.ReturnPart{Name: "p"}), nil
}

// PathLoadStrategyEdgeConstraint is similar to load many, but requires that it is related to another node via some edge
func PathLoadStrategyEdgeConstraint(startVariable, startLabel, endLabel, endTargetField string, minJumps, maxJumps, depth int, additionalConstraints dsl.ConditionOperator) (dsl.Cypher, error) {
	if startVariable == "" {
		return nil, errors.New("variable name cannot be empty")
	}

	if startLabel == "" {
		return nil, errors.New("label can not be empty")
	}

	if endLabel == "" {
		return nil, errors.New("label can not be empty")
	}

	qp, err := dsl.ParamsFromMap(map[string]interface{}{
		endTargetField: dsl.ParamString(fmt.Sprintf("$%s", endTargetField)),
	})
	if err != nil {
		return nil, err
	}

	path := dsl.Path().P()

	if depth > 0 {
		path.V(dsl.V{}).
			E(dsl.E{
				Direction: dsl.DirectionNone,
				MinJumps:  0,
				MaxJumps:  depth,
			})
	}

	path.
		V(dsl.V{Name: startVariable, Type: startLabel}).
		E(dsl.E{MinJumps: minJumps, MaxJumps: maxJumps, Direction: dsl.DirectionNone}).
		V(dsl.V{Type: endLabel, Params: qp}).
		Build()

	builder := dsl.QB().Match(path)

	if additionalConstraints != nil {
		builder.Where(additionalConstraints)
	}

	return builder.Return(false, dsl.ReturnPart{Name: "p"}), nil
}

func getRelationshipsForLabel(gogm *Gogm, label string) ([]decoratorConfig, error) {
	raw, ok := gogm.mappedTypes.Get(label)
	if !ok {
		return nil, fmt.Errorf("struct config not found type (%s)", label)
	}

	config, ok := raw.(structDecoratorConfig)
	if !ok {
		return nil, errors.New("unable to cast into struct decorator config")
	}

	fields := []decoratorConfig{}
	for _, field := range config.Fields {
		if !field.Ignore && field.Relationship != "" {
			fields = append(fields, field)
		}
	}

	return fields, nil
}

func expandBootstrap(gogm *Gogm, variable, label string, depth int) (string, error) {
	clause := ""
	rels, err := getRelationshipsForLabel(gogm, label)
	if err != nil {
		return "", err
	}

	if depth > 0 {
		if len(rels) > 0 {
			clause += ", ["
		}

		expanded, err := expand(gogm, variable, label, rels, 1, depth-1)
		if err != nil {
			return "", err
		}
		clause += expanded

		if len(rels) > 0 {
			clause += "]"
		}
	}

	return clause, nil
}

func expand(gogm *Gogm, variable, label string, rels []decoratorConfig, level, depth int) (string, error) {
	clause := ""

	for i, rel := range rels {
		// check if a separator is needed
		if i > 0 {
			clause += ", "
		}

		ret, err := listComprehension(gogm, variable, label, rel, level, depth)
		if err != nil {
			return "", err
		}
		clause += ret
	}

	return clause, nil
}

func relString(variable string, rel decoratorConfig) string {
	start := "-"
	end := "-"

	if rel.Direction == dsl.DirectionIncoming {
		start = "<-"
	} else if rel.Direction == dsl.DirectionOutgoing {
		end = "->"
	}

	return fmt.Sprintf("%s[%s:%s]%s", start, variable, rel.Relationship, end)
}

func listComprehension(gogm *Gogm, fromNodeVar, label string, rel decoratorConfig, level, depth int) (string, error) {
	relVar := fmt.Sprintf("r_%c_%d", rel.Relationship[0], level)

	toNodeType := rel.Type.Elem()
	if rel.Type.Kind() == reflect.Slice {
		toNodeType = toNodeType.Elem()
	}

	toNodeLabel, err := traverseRelType(toNodeType, rel.Direction)
	if err != nil {
		return "", err
	}

	toNodeVar := fmt.Sprintf("n_%c_%d", toNodeLabel[0], level)

	clause := fmt.Sprintf("[(%s)%s(%s:%s) | [%s, %s", fromNodeVar, relString(relVar, rel), toNodeVar, toNodeLabel, relVar, toNodeVar)

	if depth > 0 {
		toNodeRels, err := getRelationshipsForLabel(gogm, label)
		if err != nil {
			return "", err
		}

		if len(toNodeRels) > 0 {
			toNodeExpansion, err := expand(gogm, toNodeVar, toNodeLabel, toNodeRels, level+1, depth-1)
			if err != nil {
				return "", err
			}
			clause += fmt.Sprintf(", [%s]", toNodeExpansion)
		}
	}

	clause += "]]"
	return clause, nil
}

// SchemaLoadStrategyMany loads many using schema strategy
func SchemaLoadStrategyMany(gogm *Gogm, variable, label string, depth int, additionalConstraints dsl.ConditionOperator) (dsl.Cypher, error) {
	if variable == "" {
		return nil, errors.New("variable name cannot be empty")
	}

	if label == "" {
		return nil, errors.New("label can not be empty")
	}

	if depth < 0 {
		return nil, errors.New("depth can not be less than 0")
	}

	builder := dsl.QB().Cypher(fmt.Sprintf("MATCH (%s:%s)", variable, label))

	if additionalConstraints != nil {
		builder = builder.Where(additionalConstraints)
	}

	builder = builder.Cypher("RETURN " + variable)

	if depth > 0 {
		clause, err := expandBootstrap(gogm, variable, label, depth)
		if err != nil {
			return nil, err
		}
		builder = builder.Cypher(clause)
	}

	return builder, nil
}

// SchemaLoadStrategyOne loads one object using schema strategy
func SchemaLoadStrategyOne(gogm *Gogm, variable, label, fieldOn, paramName string, isGraphId bool, depth int, additionalConstraints dsl.ConditionOperator) (dsl.Cypher, error) {
	if variable == "" {
		return nil, errors.New("variable name cannot be empty")
	}

	if label == "" {
		return nil, errors.New("label can not be empty")
	}

	if depth < 0 {
		return nil, errors.New("depth can not be less than 0")
	}

	builder := dsl.QB().Cypher(fmt.Sprintf("MATCH (%s:%s)", variable, label))

	var condition *dsl.ConditionConfig
	if isGraphId {
		condition = &dsl.ConditionConfig{
			FieldManipulationFunction: "ID",
			Name:                      variable,
			ConditionOperator:         dsl.EqualToOperator,
			Check:                     dsl.ParamString("$" + paramName),
		}
	} else {
		condition = &dsl.ConditionConfig{
			Name:              variable,
			Field:             fieldOn,
			ConditionOperator: dsl.EqualToOperator,
			Check:             dsl.ParamString("$" + paramName),
		}
	}

	if additionalConstraints != nil {
		builder = builder.Where(additionalConstraints.And(condition))
	} else {
		builder = builder.Where(dsl.C(condition))
	}

	builder = builder.Cypher("RETURN " + variable)

	if depth > 0 {
		clause, err := expandBootstrap(gogm, variable, label, depth)
		if err != nil {
			return nil, err
		}
		builder = builder.Cypher(clause)
	}

	return builder, nil
}
