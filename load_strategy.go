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
