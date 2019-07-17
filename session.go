package gogm

import (
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
	"reflect"
)

const defaultDepth = 1

type Session struct{
	conn *dsl.Session
	DefaultDepth int
	LoadStrategy LoadStrategy
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
	return s.LoadDepthFilterPagination(respObj, id, s.DefaultDepth, nil, nil,nil)
}

func (s *Session) LoadDepth(respObj interface{}, id string, depth int) error{
	return s.LoadDepthFilterPagination(respObj, id, depth, nil, nil,nil)
}

func (s *Session) LoadDepthFilter(respObj interface{}, id string, depth int, filter *dsl.ConditionBuilder, params map[string]interface{}) error{
		return s.LoadDepthFilterPagination(respObj, id, depth, filter, params,nil)
}

func (s *Session) LoadDepthFilterPagination(respObj interface{}, id string, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error {
	respType := reflect.TypeOf(respObj)

	//validate type is ptr
	if respType.Kind() != reflect.Ptr{
		return errors.New("respObj must be type ptr")
	}

	//"deref" reflect interface type
	respType = respType.Elem()

	//get the type name -- this maps directly to the label
	respObjName := respType.Name()

	//will need to keep track of these variables
	varName := "n"

	var query dsl.Cypher
	var err error

	//make the query based off of the load strategy
	switch s.LoadStrategy {
	case PATH_LOAD_STRATEGY:
		query, err = PathLoadStrategyOne(s.conn, varName, respObjName, depth, filter)
		if err != nil{
			return err
		}
	case SCHEMA_LOAD_STRATEGY:
		return errors.New("schema load strategy not supported yet")
	default:
		return errors.New("unknown load strategy")
	}


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

	if params == nil{
		params = map[string]interface{}{
			"uuid": id,
		}
	} else {
		params["uuid"] = id
	}

	rows, err := query.Query(params)
	if err != nil{
		return err
	}

	return decodeNeoRows(rows, respObj)
}

func (s *Session) LoadAll(respObj interface{}) error {
	return s.LoadAllDepthFilterPagination(respObj, s.DefaultDepth, nil, nil, nil)
}

func (s *Session) LoadAllDepth(respObj interface{}, depth int) error {
	return s.LoadAllDepthFilterPagination(respObj, depth, nil, nil, nil)
}

func (s *Session) LoadAllDepthFilter(respObj interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}) error {
	return s.LoadAllDepthFilterPagination(respObj, depth, filter, params, nil)
}

func (s *Session) LoadAllDepthFilterPagination(respObj interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error {
	rawRespType := reflect.TypeOf(respObj)

	if rawRespType.Kind() != reflect.Ptr{
		return fmt.Errorf("respObj must be a pointer to a slice, instead it is %T", respObj)
	}

	//deref to a slice
	respType := rawRespType.Elem()

	//validate type is ptr
	if respType.Kind() != reflect.Slice{
		return fmt.Errorf("respObj must be type slice, instead it is %T", respObj)
	}

	//"deref" reflect interface type
	respType = respType.Elem()

	//get the type name -- this maps directly to the label
	respObjName := respType.Name()

	//will need to keep track of these variables
	varName := "n"

	var query dsl.Cypher
	var err error

	//make the query based off of the load strategy
	switch s.LoadStrategy {
	case PATH_LOAD_STRATEGY:
		query, err = PathLoadStrategyMany(s.conn, varName, respObjName, depth, filter)
		if err != nil{
			return err
		}
	case SCHEMA_LOAD_STRATEGY:
		return errors.New("schema load strategy not supported yet")
	default:
		return errors.New("unknown load strategy")
	}


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

	rows, err := query.Query(params)
	if err != nil{
		return err
	}

	return decodeNeoRows(rows, respObj)
}

func (s *Session) Save(saveObj interface{}) error {
	return s.SaveDepth(saveObj, s.DefaultDepth)
}

func (s *Session) SaveDepth(saveObj interface{}, depth int) error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}

	return nil
}

func (s *Session) Delete(deleteObj interface{}) error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}

	//check if its an edge or a vertex

	//if its an edge delete it

	//if its a vertex detach delete  it

	return nil
}

func (s *Session) Query(query string, properties map[string]interface{}, respObj interface{}) error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}

	return nil
}

func (s *Session) PurgeDatabase() error {
	if s.conn == nil{
		return errors.New("neo4j connection not initialized")
	}

	_, err := s.conn.Query().Match(dsl.Path().V(dsl.V{Name: "n"}).Build()).Delete(true, "n").Exec(nil)
	return err
}

