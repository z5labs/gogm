package gogm

import (
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
)

type LoadStrategy int

const (
	PATH_LOAD_STRATEGY LoadStrategy = iota
	SCHEMA_LOAD_STRATEGY
)

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

	builder := dsl.QB().
		Match(dsl.Path().
			P().
			V(dsl.V{Name: variable}).
			E(dsl.E{Direction: dsl.DirectionNone, MinJumps: 0, MaxJumps: depth}).
			V().Build())

	if additionalConstraints != nil {
		builder.Where(additionalConstraints)
	}

	return builder.Return(false, dsl.ReturnPart{Name: "p"}), nil
}

func PathLoadStrategyOne(variable, label string, depth int, additionalConstraints dsl.ConditionOperator) (dsl.Cypher, error) {
	if variable == "" {
		return nil, errors.New("variable name cannot be empty")
	}

	if label == "" {
		return nil, errors.New("label can not be empty")
	}

	if depth < 0 {
		return nil, errors.New("depth can not be less than 0")
	}

	builder := dsl.QB().
		Match(dsl.Path().
			P().
			V(dsl.V{Name: variable}).
			E(dsl.E{Direction: dsl.DirectionNone, MinJumps: 0, MaxJumps: depth}).
			V().Build())

	if additionalConstraints != nil {
		builder = builder.Where(additionalConstraints.And(&dsl.ConditionConfig{
			Name:              variable,
			Field:             "uuid",
			ConditionOperator: dsl.EqualToOperator,
			Check:             dsl.ParamString("{uuid}"),
		}))
	} else {
		builder = builder.Where(dsl.C(&dsl.ConditionConfig{
			Name:              variable,
			Field:             "uuid",
			ConditionOperator: dsl.EqualToOperator,
			Check:             dsl.ParamString("{uuid}"),
		}))
	}

	return builder.Return(false, dsl.ReturnPart{Name: "p"}), nil
}

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
		endTargetField: dsl.ParamString(fmt.Sprintf("{%s}", endTargetField)),
	})
	if err != nil {
		return nil, err
	}

	builder := dsl.QB().
		Match(dsl.Path().
			P().
			V(dsl.V{Name: startVariable, Type: startLabel}).
			E(dsl.E{MinJumps: minJumps, MaxJumps: maxJumps, Direction: dsl.DirectionNone}).
			V(dsl.V{Type: endLabel, Params: qp}).
			Build())

	if additionalConstraints != nil {
		builder.Where(additionalConstraints)
	}

	return builder.Return(false, dsl.ReturnPart{Name: "n"}), nil
}
