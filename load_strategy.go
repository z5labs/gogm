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

var edgesPart = `collect(extract(n in %s | {StartNodeId: ID(startnode(n)), StartNodeType: labels(startnode(n))[0], EndNodeId: ID(endnode(n)), EndNodeType: labels(endnode(n))[0], Obj: n, Type: type(n)})) as Edges`

/*
example
MATCH (n:OrganizationNode)
		WITH n
		MATCH (n)-[e*0..1]-(m)
		RETURN DISTINCT
			collect(extract(n in e | {StartNodeId: ID(startnode(n)), StartNodeType: labels(startnode(n)), EndNodeId: ID(endnode(n)), EndNode: labels(endnode(n)), Obj: n, Type: type(n)})) as Edges,
			collect(DISTINCT m) as Ends,
			collect(DISTINCT n) as Starts*/

func PathLoadStrategyMany(sess *dsl.Session, variable, label string, depth int, additionalConstraints dsl.ConditionOperator) (dsl.Cypher, error){
	if sess == nil{
		return nil, errors.New("session can not be nil")
	}

	if variable == ""{
		return nil, errors.New("variable name cannot be empty")
	}

	if label == ""{
		return nil, errors.New("label can not be empty")
	}

	if depth < 0{
		return nil, errors.New("depth can not be less than 0")
	}

	builder := sess.QueryReadOnly().
		Match(dsl.Path().V(dsl.V{Name: variable, Type: label}).Build()).
		With(&dsl.WithConfig{
			Parts: []dsl.WithPart{
				{
					Name: variable,
				},
			},
		})

	if additionalConstraints != nil{
		builder.Where(additionalConstraints)
	}

	if depth == 0 {
		builder.With(&dsl.WithConfig{
			Parts: []dsl.WithPart{
				{
					Name: "n",
				},
				{
					Name: "[]",
					As: "e",
				},
				{
					Name: "[]",
					As: "m",
				},
			},
		})
	} else {
		builder.
			OptionalMatch(dsl.Path().
				V(dsl.V{Name: variable}).
				E(dsl.E{Name: "e", Direction: dsl.DirectionNone, MinJumps: 0, MaxJumps: depth}).
				V(dsl.V{Name: "m"}).Build())
	}

	BuildReturnQuery(builder, "n", "m", "e")

	return builder, nil
}

func BuildReturnQuery(builder dsl.Cypher, startSide, endSide, edge string){
	builder.Return(true,
		dsl.ReturnPart{Name: fmt.Sprintf(edgesPart, edge)},
		dsl.ReturnPart{Name: fmt.Sprintf("collect(DISTINCT %s) as Ends", endSide)},
		dsl.ReturnPart{Name: fmt.Sprintf("collect(DISTINCT %s) as Starts", startSide)},
	)
}

func PathLoadStrategyOne(sess *dsl.Session, variable, label string, depth int, additionalConstraints dsl.ConditionOperator) (dsl.Cypher, error) {
	if sess == nil{
		return nil, errors.New("session can not be nil")
	}

	if variable == ""{
		return nil, errors.New("variable name cannot be empty")
	}

	if label == ""{
		return nil, errors.New("label can not be empty")
	}

	if depth < 0{
		return nil, errors.New("depth can not be less than 0")
	}

	builder := sess.QueryReadOnly().
		Match(dsl.Path().V(dsl.V{Name: variable, Type: label}).Build()).
		With(&dsl.WithConfig{
			Parts: []dsl.WithPart{
				{
					Name: variable,
				},
			},
		})

	if additionalConstraints != nil{
		builder = builder.Where(additionalConstraints.And(&dsl.ConditionConfig{
			Name: variable,
			Field: "uuid",
			ConditionOperator: dsl.EqualToOperator,
			Check: dsl.ParamString("{uuid}"),
		}))
	} else {
		builder = builder.Where(dsl.C(&dsl.ConditionConfig{
			Name: variable,
			Field: "uuid",
			ConditionOperator: dsl.EqualToOperator,
			Check: dsl.ParamString("{uuid}"),
		}))
	}

	if depth == 0 {
		builder.With(&dsl.WithConfig{
			Parts: []dsl.WithPart{
				{
					Name: variable,
				},
				{
					Name: "[]",
					As: "e",
				},
				{
					Name: "[]",
					As: "m",
				},
			},
		})
	} else {
		builder.
			OptionalMatch(dsl.Path().
				V(dsl.V{Name: variable}).
				E(dsl.E{Name: "e", Direction: dsl.DirectionNone, MinJumps: 0, MaxJumps: depth}).
				V(dsl.V{Name: "m"}).Build())
	}


	BuildReturnQuery(builder, "n", "m", "e")

	return builder, nil
}

func PathLoadStrategyEdgeConstraint(sess *dsl.Session, startVariable, startLabel, endLabel, endTargetField string, minJumps, maxJumps, depth int, additionalConstraints dsl.ConditionOperator) (dsl.Cypher, error) {
	if sess == nil{
		return nil, errors.New("session can not be nil")
	}

	if startVariable == ""{
		return nil, errors.New("variable name cannot be empty")
	}

	if startLabel == ""{
		return nil, errors.New("label can not be empty")
	}

	if endLabel == ""{
		return nil, errors.New("label can not be empty")
	}

	qp, err := dsl.ParamsFromMap(map[string]interface{}{
		endTargetField: dsl.ParamString(fmt.Sprintf("{%s}", endTargetField)),
	})
	if err != nil {
		return nil, err
	}

	builder := sess.QueryReadOnly().
		Match(dsl.Path().
			V(dsl.V{Name: startVariable, Type: startLabel}).
			E(dsl.E{MinJumps: minJumps, MaxJumps: maxJumps, Direction: dsl.DirectionNone}).
			V(dsl.V{Type: endLabel, Params: qp}).
			Build()).

		With(&dsl.WithConfig{
			Parts: []dsl.WithPart{
				{
					Name: startVariable,
				},
			},
		})
	if additionalConstraints != nil{
		builder.Where(additionalConstraints)
	}

	if depth == 0 {
		builder.With(&dsl.WithConfig{
			Parts: []dsl.WithPart{
				{
					Name: startVariable,
				},
				{
					Name: "[]",
					As: "e",
				},
				{
					Name: "[]",
					As: "m",
				},
			},
		})
	} else {
		builder.
			OptionalMatch(dsl.Path().
				V(dsl.V{Name: startVariable}).
				E(dsl.E{Name: "e", Direction: dsl.DirectionNone, MinJumps: 0, MaxJumps: depth}).
				V(dsl.V{Name: "m"}).Build())
	}


	BuildReturnQuery(builder, "n", "m", "e")

	return builder, nil
}