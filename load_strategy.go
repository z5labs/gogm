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

var edgesPart = `collect(extract(n in e | {StartNodeId: ID(startnode(n)), StartNodeType: labels(startnode(n)), EndNodeId: ID(endnode(n)), EndNode: labels(endnode(n)), Obj: n, Type: type(n)})) as Edges,`

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
		builder = builder.Where(additionalConstraints)
	}

	builder.
		Match(dsl.Path().
			V(dsl.V{Name: variable}).
			E(dsl.E{Name: "e", Direction:dsl.DirectionPtr(dsl.Any), MinJumps: 0, MaxJumps: depth}).
			V(dsl.V{Name: "m"}).Build()).
		Return(true,
			dsl.ReturnPart{Name: edgesPart},
			dsl.ReturnPart{Name: "collect(DISTINCT m) as Ends"},
			dsl.ReturnPart{Name: fmt.Sprintf("collect(DISTINCT %s) as Starts", variable)},
		)

	return builder, nil
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
			Check: dsl.ParamString("{uuid}"),
		}))
	} else {
		builder = builder.Where(dsl.C(&dsl.ConditionConfig{
			Name: variable,
			Check: dsl.ParamString("{uuid}"),
		}))
	}

	builder.
		Match(dsl.Path().
			V(dsl.V{Name: variable}).
			E(dsl.E{Name: "e", Direction:dsl.DirectionPtr(dsl.Any), MinJumps: 0, MaxJumps: depth}).
			V(dsl.V{Name: "m"}).Build()).
		Return(true,
			dsl.ReturnPart{Name: edgesPart},
			dsl.ReturnPart{Name: "collect(DISTINCT m) as Ends"},
			dsl.ReturnPart{Name: fmt.Sprintf("collect(DISTINCT %s) as Starts", variable)},
		)

	return builder, nil
}