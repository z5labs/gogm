package gogm

import (
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/mindstand/gogm/neo_encoder"
	"reflect"
)

const defaultDepth = 2

type Session struct{
	conn *dsl.Session
	DefaultDepth int
}

func NewSession() *Session{

	session := new(Session)

	session.conn = dsl.NewSession()

	session.DefaultDepth = defaultDepth

	return session
}

func (s *Session) Begin() error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}
	
	return s.conn.Begin()
}

func (s *Session) Rollback() error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}

	return s.conn.Rollback()
}

func (s *Session) Commit() error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}

	return s.conn.Commit()
}

func (s *Session) Load(respObj interface{}, id string) error {
	return s.LoadDepthFilterPagination(respObj, id, s.DefaultDepth, nil, nil)
}

func (s *Session) LoadDepth(respObj interface{}, id string, depth int) error{
	return s.LoadDepthFilterPagination(respObj, id, depth, nil, nil)
}

func (s *Session) LoadDepthFilter(respObj interface{}, id string, depth int, filter *dsl.ConditionBuilder) error{
		return s.LoadDepthFilterPagination(respObj, id, depth, filter, nil)
}

func (s *Session) LoadDepthFilterPagination(respObj interface{}, id string, depth int, filter dsl.ConditionOperator, pagination *Pagination) error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}

	respType := reflect.TypeOf(respObj)

	//validate type is ptr
	if respType.Kind() != reflect.Ptr{
		return errors.New("respObj must be type ptr")
	}

	//"deref" reflect interface type
	respType = respType.Elem()

	respObjName := respType.Name()

	//get config
	structConfig, ok := mappedTypes[respObjName]
	if !ok{
		return fmt.Errorf("unrecognized type '%s', ensure this is a mapped node", respObjName)
	}

	//will need to keep track of these variables
	varName := "n"
	edgeName := "e"

	//build match path
	path := dsl.Path().
		P().
		V(dsl.V{
			Name: varName,
			Type: structConfig.Labels[0], //should have 0, would have failed by now

		}).
		E(dsl.E{
			Name: edgeName,
			MaxJumps: depth,
		}).
		V(dsl.V{}).
		Build()

	//start query
	query := s.conn.QueryReadOnly().Match(path)

	//id condition
	idCondition := &dsl.ConditionConfig{
		Name: varName,
		Field: "uuid",
		ConditionOperator: dsl.EqualToOperator,
		Check: id,
	}

	//check filter, if not created initialize it with id condition
	if filter == nil{
		filter = dsl.C(idCondition)
	} else {
		filter = filter.And(idCondition)
	}

	//add where clause to the query
	query = query.Where(filter)

	//if the query requires pagination, set that up
	if pagination != nil{
		err := pagination.Validate()
		if err != nil{
			return err
		}

		query = query.
			OrderBy(dsl.OrderByConfig{
				Name: pagination.OrderByVarName,
				Member: pagination.OrderByField,
				Desc: pagination.OrderByDesc,
			}).
			Skip(pagination.LimitPerPage * pagination.PageNumber).
			Limit(pagination.LimitPerPage )
	}

	rows, err := query.Query(nil)
	if err != nil{
		return err
	}

	return neo_encoder.DecodeNeoRows(rows, respObj)
}

func (s *Session) LoadAll(respObj interface{}) error {
	return s.LoadAllDepthFilterPagination(respObj, s.DefaultDepth, nil, nil)
}

func (s *Session) LoadAllDepth(respObj interface{}, depth int) error {
	return s.LoadAllDepthFilterPagination(respObj, depth, nil, nil)
}

func (s *Session) LoadAllDepthFilter(respObj interface{}, depth int, filter *dsl.ConditionBuilder) error {
	return s.LoadAllDepthFilterPagination(respObj, depth, filter, nil)
}

func (s *Session) LoadAllDepthFilterPagination(respObj interface{}, depth int, filter *dsl.ConditionBuilder, pagination *Pagination) error {
	respType := reflect.TypeOf(respObj)

	//validate type is ptr
	if respType.Kind() != reflect.Ptr{
		return errors.New("respObj must be type ptr")
	}

	//"deref" reflect interface type
	respType = respType.Elem()

	respObjName := respType.Name()

	//get config
	structConfig, ok := mappedTypes[respObjName]
	if !ok{
		return fmt.Errorf("unrecognized type '%s', ensure this is a mapped node", respObjName)
	}

	//will need to keep track of these variables
	varName := "n"
	edgeName := "e"

	//build match path
	path := dsl.Path().
		P().
		V(dsl.V{
			Name: varName,
			Type: structConfig.Labels[0], //should have 0, would have failed by now

		}).
		E(dsl.E{
			Name: edgeName,
			MaxJumps: depth,
		}).
		V(dsl.V{}).
		Build()

	//start query
	query := s.conn.QueryReadOnly().Match(path)

	//check filter, if its there, add the where clause
	if filter != nil{
		query = query.Where(filter)
	}

	//add where clause to the query


	//if the query requires pagination, set that up
	if pagination != nil{
		err := pagination.Validate()
		if err != nil{
			return err
		}

		query = query.
			OrderBy(dsl.OrderByConfig{
				Name: pagination.OrderByVarName,
				Member: pagination.OrderByField,
				Desc: pagination.OrderByDesc,
			}).
			Skip(pagination.LimitPerPage * pagination.PageNumber).
			Limit(pagination.LimitPerPage )
	}

	rows, err := query.Query(nil)
	if err != nil{
		return err
	}

	return neo_encoder.DecodeNeoRows(rows, respObj)
}

func (s *Session) Save(saveObj interface{}) error {
	return s.SaveDepth(saveObj, s.DefaultDepth)
}

func (s *Session) SaveDepth(saveObj interface{}, depth int) error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}
}

func (s *Session) Delete(deleteObj interface{}) error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}

	//check if its an edge or a vertex

	//if its an edge delete it

	//if its a vertex detach delete  it
}

func (s *Session) Query(query string, properties map[string]interface{}, respObj interface{}) error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}
}

func (s *Session) QuerySlice(query string, properties map[string]interface{}, respObj interface{}) error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}
}

func (s *Session) PurgeDatabase() error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}

	_, err := s.conn.Query().Match(dsl.Path().V(dsl.V{Name: "n"}).Build()).Delete(true, "n").Exec(nil)
	return err
}

