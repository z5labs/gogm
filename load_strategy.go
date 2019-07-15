package gogm

import (
	"errors"
	dsl "github.com/mindstand/go-cypherdsl"
)

type LoadStrategy int

const (
	PATH_LOAD_STRATEGY LoadStrategy = iota
	SCHEMA_LOAD_STRATEGY
)

/*
example
MATCH (n:`OrganizationNode`) WHERE n.`uuid` = { id } WITH n MATCH p=(n)-[e*0..2]-(m) RETURN ID(n) as N_ID, n, ID(m) AS M_ID, m, LABELS(m), {type: type(e[0]), sn: ID(startnode(e[0])), en: ID(endNode(e[0]))}
*/

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

	e := "e"
	m := "m"

	query := sess.Query().
		Match(dsl.Path().V(dsl.V{Name: variable}).Build()).
		With(&dsl.WithConfig{
			Parts: []dsl.WithPart{
				{
					Name: variable,
				},
			},
		})

	//add conditional if needed
	if additionalConstraints != nil{
		query = query.Where(additionalConstraints)
	}

	return query.
		Match(dsl.Path().
			P().
			V(dsl.V{Name: variable}).
			E(dsl.E{
				Name: e,
				MinJumps: 0,
				MaxJumps: depth,
			}).
			V(dsl.V{
				Name: m,
			}).
			Build()).
		Return(true,
			dsl.ReturnPart{
				Function: &dsl.FunctionConfig{
					Name: "ID",
					Params: []interface{}{
						variable,
					},
				},
				Alias: "N_ID",
			},
			dsl.ReturnPart{
				Name: variable,
			},
			dsl.ReturnPart{
				Function: &dsl.FunctionConfig{
					Name: "ID",
					Params: []interface{}{
						m,
					},
				},
				Alias: "M_ID",
			},
			dsl.ReturnPart{
				Name: m,
			},
			dsl.ReturnPart{
				Function: &dsl.FunctionConfig{
					Name: "LABELS",
					Params: []interface{}{
						m,
					},
				},
			},
			dsl.ReturnPart{
				Name: "{Type: type(e[0]), StartNode: ID(startnode(e[0])), EndNode: ID(endnode(e[0]))}",
				Alias: "edge_config",
			},
		), nil
}

func PathLoadStrategyOne(sess *dsl.Session, variable, label string, depth int, uuid string, additionalConstraints dsl.ConditionOperator) (dsl.Cypher, error) {
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

	if uuid == ""{
		return nil, errors.New("uuid can not be empty")
	}

	e := "e"
	m := "m"

	if additionalConstraints == nil{
		additionalConstraints = dsl.C(&dsl.ConditionConfig{
			Name: variable,
			Field: "uuid",
			ConditionOperator: dsl.EqualToOperator,
			Check: dsl.ParamString(uuid),
		})
	} else {
		additionalConstraints = additionalConstraints.And(&dsl.ConditionConfig{
			Name: variable,
			Field: "uuid",
			ConditionOperator: dsl.EqualToOperator,
			Check: dsl.ParamString("{uuid}"),
		})
	}

	return sess.Query().
		Match(dsl.Path().V(dsl.V{Name: variable}).Build()).
		With(&dsl.WithConfig{
			Parts: []dsl.WithPart{
				{
					Name: variable,
				},
			},
		}).
		Where(additionalConstraints).
		Match(dsl.Path().
			P().
			V(dsl.V{Name: variable}).
			E(dsl.E{
				Name: e,
				MinJumps: 0,
				MaxJumps: depth,
			}).
			V(dsl.V{
				Name: m,
			}).
			Build()).
		Return(true,
			dsl.ReturnPart{
				Function: &dsl.FunctionConfig{
					Name: "ID",
					Params: []interface{}{
						variable,
					},
				},
				Alias: "N_ID",
			},
			dsl.ReturnPart{
				Name: variable,
			},
			dsl.ReturnPart{
				Function: &dsl.FunctionConfig{
					Name: "ID",
					Params: []interface{}{
						m,
					},
				},
				Alias: "M_ID",
			},
			dsl.ReturnPart{
				Name: m,
			},
			dsl.ReturnPart{
				Function: &dsl.FunctionConfig{
					Name: "LABELS",
					Params: []interface{}{
						m,
					},
				},
			},
			dsl.ReturnPart{
				Name: "{Type: type(e[0]), StartNode: ID(startnode(e[0])), EndNode: ID(endnode(e[0]))}",
				Alias: "edge_config",
			},
		), nil
}