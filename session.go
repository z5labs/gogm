// Copyright (c) 2020 MindStand Technologies, Inc
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
	"github.com/neo4j/neo4j-go-driver/neo4j"
	"reflect"
)

const defaultDepth = 1

type Session struct {
	neoSess      neo4j.Session
	tx           neo4j.Transaction
	DefaultDepth int
	LoadStrategy LoadStrategy
}

func NewSession(readonly bool) (*Session, error) {
	if driver == nil {
		return nil, errors.New("driver cannot be nil")
	}

	session := new(Session)

	var mode neo4j.AccessMode

	if readonly {
		mode = neo4j.AccessModeRead
	} else {
		mode = neo4j.AccessModeWrite
	}

	neoSess, err := driver.Session(mode)
	if err != nil {
		return nil, err
	}

	session.neoSess = neoSess

	session.DefaultDepth = defaultDepth

	return session, nil
}

func (s *Session) Begin() error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	if s.tx != nil {
		return fmt.Errorf("transaction already started: %w", ErrTransaction)
	}

	var err error

	s.tx, err = s.neoSess.BeginTransaction()
	if err != nil {
		return err
	}

	return nil
}

func (s *Session) Rollback() error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	if s.tx == nil {
		return fmt.Errorf("cannot rollback nil transaction: %w", ErrTransaction)
	}

	err := s.tx.Rollback()
	if err != nil {
		return err
	}

	s.tx = nil
	return nil
}

func (s *Session) RollbackWithError(originalError error) error {
	err := s.Rollback()
	if err != nil {
		return fmt.Errorf("original error: `%s`, rollback error: `%s`", originalError.Error(), err.Error())
	}

	return originalError
}

func (s *Session) Commit() error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	if s.tx == nil {
		return fmt.Errorf("cannot commit nil transaction: %w", ErrTransaction)
	}

	err := s.tx.Commit()
	if err != nil {
		return err
	}

	s.tx = nil
	return nil
}

func (s *Session) Load(respObj interface{}, id string) error {
	return s.LoadDepthFilterPagination(respObj, id, s.DefaultDepth, nil, nil, nil)
}

func (s *Session) LoadDepth(respObj interface{}, id string, depth int) error {
	return s.LoadDepthFilterPagination(respObj, id, depth, nil, nil, nil)
}

func (s *Session) LoadDepthFilter(respObj interface{}, id string, depth int, filter *dsl.ConditionBuilder, params map[string]interface{}) error {
	return s.LoadDepthFilterPagination(respObj, id, depth, filter, params, nil)
}

func (s *Session) LoadDepthFilterPagination(respObj interface{}, id string, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error {
	respType := reflect.TypeOf(respObj)

	//validate type is ptr
	if respType.Kind() != reflect.Ptr {
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
		query, err = PathLoadStrategyOne(varName, respObjName, depth, filter)
		if err != nil {
			return err
		}
	case SCHEMA_LOAD_STRATEGY:
		return errors.New("schema load strategy not supported yet")
	default:
		return errors.New("unknown load strategy")
	}

	//if the query requires pagination, set that up
	if pagination != nil {
		err := pagination.Validate()
		if err != nil {
			return err
		}

		query = query.
			OrderBy(dsl.OrderByConfig{
				Name:   pagination.OrderByVarName,
				Member: pagination.OrderByField,
				Desc:   pagination.OrderByDesc,
			}).
			Skip(pagination.LimitPerPage * pagination.PageNumber).
			Limit(pagination.LimitPerPage)
	}

	if params == nil {
		params = map[string]interface{}{
			"uuid": id,
		}
	} else {
		params["uuid"] = id
	}

	// handle if in transaction
	var rf neoRunFunc
	if s.tx != nil {
		rf = s.tx.Run
	} else {
		rf = runWrap(s.neoSess)
	}

	cyp, err := query.ToCypher()
	if err != nil {
		return err
	}

	result, err := rf(cyp, params)
	if err != nil {
		return err
	}

	return decode(result, respObj)
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

	if rawRespType.Kind() != reflect.Ptr {
		return fmt.Errorf("respObj must be a pointer to a slice, instead it is %T", respObj)
	}

	//deref to a slice
	respType := rawRespType.Elem()

	//validate type is ptr
	if respType.Kind() != reflect.Slice {
		return fmt.Errorf("respObj must be type slice, instead it is %T", respObj)
	}

	//"deref" reflect interface type
	respType = respType.Elem()

	if respType.Kind() == reflect.Ptr {
		//slice of pointers
		respType = respType.Elem()
	}

	//get the type name -- this maps directly to the label
	respObjName := respType.Name()

	//will need to keep track of these variables
	varName := "n"

	var query dsl.Cypher
	var err error

	//make the query based off of the load strategy
	switch s.LoadStrategy {
	case PATH_LOAD_STRATEGY:
		query, err = PathLoadStrategyMany(varName, respObjName, depth, filter)
		if err != nil {
			return err
		}
	case SCHEMA_LOAD_STRATEGY:
		return errors.New("schema load strategy not supported yet")
	default:
		return errors.New("unknown load strategy")
	}

	//if the query requires pagination, set that up
	if pagination != nil {
		err := pagination.Validate()
		if err != nil {
			return err
		}

		query = query.
			OrderBy(dsl.OrderByConfig{
				Name:   pagination.OrderByVarName,
				Member: pagination.OrderByField,
				Desc:   pagination.OrderByDesc,
			}).
			Skip(pagination.LimitPerPage * pagination.PageNumber).
			Limit(pagination.LimitPerPage)
	}

	// handle if in transaction
	var rf neoRunFunc
	if s.tx != nil {
		rf = s.tx.Run
	} else {
		rf = runWrap(s.neoSess)
	}

	cyp, err := query.ToCypher()
	if err != nil {
		return err
	}

	result, err := rf(cyp, params)
	if err != nil {
		return err
	}

	return decode(result, respObj)
}

func (s *Session) LoadAllEdgeConstraint(respObj interface{}, endNodeType, endNodeField string, edgeConstraint interface{}, minJumps, maxJumps, depth int, filter dsl.ConditionOperator) error {
	rawRespType := reflect.TypeOf(respObj)

	if rawRespType.Kind() != reflect.Ptr {
		return fmt.Errorf("respObj must be a pointer to a slice, instead it is %T", respObj)
	}

	//deref to a slice
	respType := rawRespType.Elem()

	//validate type is ptr
	if respType.Kind() != reflect.Slice {
		return fmt.Errorf("respObj must be type slice, instead it is %T", respObj)
	}

	//"deref" reflect interface type
	respType = respType.Elem()

	if respType.Kind() == reflect.Ptr {
		//slice of pointers
		respType = respType.Elem()
	}

	//get the type name -- this maps directly to the label
	respObjName := respType.Name()

	//will need to keep track of these variables
	varName := "n"

	var query dsl.Cypher
	var err error

	//make the query based off of the load strategy
	switch s.LoadStrategy {
	case PATH_LOAD_STRATEGY:
		query, err = PathLoadStrategyEdgeConstraint(varName, respObjName, endNodeType, endNodeField, minJumps, maxJumps, depth, filter)
		if err != nil {
			return err
		}
	case SCHEMA_LOAD_STRATEGY:
		return errors.New("schema load strategy not supported yet")
	default:
		return errors.New("unknown load strategy")
	}

	// handle if in transaction
	var rf neoRunFunc
	if s.tx != nil {
		rf = s.tx.Run
	} else {
		rf = runWrap(s.neoSess)
	}

	cyp, err := query.ToCypher()
	if err != nil {
		return err
	}

	result, err := rf(cyp, map[string]interface{}{
		endNodeField: edgeConstraint,
	})
	if err != nil {
		return err
	}

	return decode(result, respObj)
}

func (s *Session) Save(saveObj interface{}) error {
	return s.SaveDepth(saveObj, s.DefaultDepth)
}

func (s *Session) SaveDepth(saveObj interface{}, depth int) error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	// handle if in transaction
	var rf neoRunFunc
	if s.tx != nil {
		rf = s.tx.Run
	} else {
		rf = runWrap(s.neoSess)
	}

	return saveDepth(rf, saveObj, depth)
}

func (s *Session) Delete(deleteObj interface{}) error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	if deleteObj == nil {
		return errors.New("deleteObj can not be nil")
	}

	// handle if in transaction
	var rf neoRunFunc
	if s.tx != nil {
		rf = s.tx.Run
	} else {
		rf = runWrap(s.neoSess)
	}

	return deleteNode(rf, deleteObj)
}

func (s *Session) DeleteUUID(uuid string) error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	// handle if in transaction
	var rf neoRunFunc
	if s.tx != nil {
		rf = s.tx.Run
	} else {
		rf = runWrap(s.neoSess)
	}

	return deleteByUuids(rf, uuid)
}

func (s *Session) Query(query string, properties map[string]interface{}, respObj interface{}) error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	// handle if in transaction
	var rf neoRunFunc
	if s.tx != nil {
		rf = s.tx.Run
	} else {
		rf = runWrap(s.neoSess)
	}

	res, err := rf(query, properties)
	if err != nil {
		return err
	}

	return decode(res, respObj)
}

func (s *Session) QueryRaw(query string, properties map[string]interface{}) ([][]interface{}, error) {
	if s.neoSess == nil {
		return nil, errors.New("neo4j connection not initialized")
	}

	// handle if in transaction
	var rf neoRunFunc
	if s.tx != nil {
		rf = s.tx.Run
	} else {
		rf = runWrap(s.neoSess)
	}

	res, err := rf(query, properties)
	if err != nil {
		return nil, err
	}

	var result [][]interface{}

	for res.Next() {
		result = append(result, res.Record().Values())
	}

	return result, nil
}

func (s *Session) PurgeDatabase() error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	// handle if in transaction
	var rf neoRunFunc
	if s.tx != nil {
		rf = s.tx.Run
	} else {
		rf = runWrap(s.neoSess)
	}

	cyp, err := dsl.QB().Match(dsl.Path().V(dsl.V{Name: "n"}).Build()).Delete(true, "n").ToCypher()
	if err != nil {
		return err
	}

	_, err = rf(cyp, nil)
	return err
}

func (s *Session) Close() error {
	if s.neoSess == nil {
		return fmt.Errorf("cannot close nil connection: %w", ErrInternal)
	}

	// handle tx
	if s.tx != nil {
		log.Warn("attempting to close a session with a pending transaction. Tx is being rolled back")
		err := s.tx.Rollback()
		if err != nil {
			return err
		}
		s.tx = nil
	}

	return s.neoSess.Close()
}
