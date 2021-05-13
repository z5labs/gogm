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
	"reflect"
	"strings"
	"time"

	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type SessionV2Impl struct {
	gogm *Gogm
	neoSess      neo4j.Session
	tx           neo4j.Transaction
	DefaultDepth int
	LoadStrategy LoadStrategy
	mode         neo4j.AccessMode
}

func newSessionWithConfigV2(gogm *Gogm, conf SessionConfig) (*SessionV2Impl, error) {
	if gogm == nil {
		return nil, errors.New("gogm instance can not be nil")
	}

	if globalGogm.isNoOp {
		return nil, errors.New("please set global gogm instance with SetGlobalGogm()")
	}

	neoSess := gogm.driver.NewSession(neo4j.SessionConfig{
		AccessMode:   conf.AccessMode,
		Bookmarks:    conf.Bookmarks,
		DatabaseName: conf.DatabaseName,
		FetchSize: conf.FetchSize,
	})

	return &SessionV2Impl{
		neoSess:      neoSess,
		DefaultDepth: defaultDepth,
	}, nil
}
func (s *SessionV2Impl) Begin() error {
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

func (s *SessionV2Impl) Rollback() error {
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

	err = s.tx.Close()
	if err != nil {
		return err
	}

	s.tx = nil
	return nil
}

func (s *SessionV2Impl) RollbackWithError(originalError error) error {
	err := s.Rollback()
	if err != nil {
		return fmt.Errorf("%s%w", err.Error(), originalError)
	}

	return originalError
}

func (s *SessionV2Impl) Commit() error {
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

	err = s.tx.Close()
	if err != nil {
		return err
	}

	s.tx = nil
	return nil
}

func (s *SessionV2Impl) Load(respObj interface{}, id string) error {
	return s.LoadDepthFilterPagination(respObj, id, s.DefaultDepth, nil, nil, nil)
}

func (s *SessionV2Impl) LoadDepth(respObj interface{}, id string, depth int) error {
	return s.LoadDepthFilterPagination(respObj, id, depth, nil, nil, nil)
}

func (s *SessionV2Impl) LoadDepthFilter(respObj interface{}, id string, depth int, filter *dsl.ConditionBuilder, params map[string]interface{}) error {
	return s.LoadDepthFilterPagination(respObj, id, depth, filter, params, nil)
}

func (s *SessionV2Impl) LoadDepthFilterPagination(respObj interface{}, id string, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error {
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

func (s *SessionV2Impl) LoadAll(respObj interface{}) error {
	return s.LoadAllDepthFilterPagination(respObj, s.DefaultDepth, nil, nil, nil)
}

func (s *SessionV2Impl) LoadAllDepth(respObj interface{}, depth int) error {
	return s.LoadAllDepthFilterPagination(respObj, depth, nil, nil, nil)
}

func (s *SessionV2Impl) LoadAllDepthFilter(respObj interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}) error {
	return s.LoadAllDepthFilterPagination(respObj, depth, filter, params, nil)
}

func (s *SessionV2Impl) LoadAllDepthFilterPagination(respObj interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error {
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

func (s *SessionV2Impl) LoadAllEdgeConstraint(respObj interface{}, endNodeType, endNodeField string, edgeConstraint interface{}, minJumps, maxJumps, depth int, filter dsl.ConditionOperator) error {
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

func (s *SessionV2Impl) Save(saveObj interface{}) error {
	return s.SaveDepth(saveObj, s.DefaultDepth)
}

func (s *SessionV2Impl) SaveDepth(saveObj interface{}, depth int) error {
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

func (s *SessionV2Impl) Delete(deleteObj interface{}) error {
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

func (s *SessionV2Impl) DeleteUUID(uuid string) error {
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

func (s *SessionV2Impl) Query(query string, properties map[string]interface{}, respObj interface{}) error {
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

func (s *SessionV2Impl) QueryRaw(query string, properties map[string]interface{}) ([][]interface{}, neo4j.ResultSummary, error) {
	if s.neoSess == nil {
		return nil, nil, errors.New("neo4j connection not initialized")
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
		return nil, nil, err
	}

	summary, err := res.Summary()
	if err != nil {
		return nil, nil, err
	}

	var result [][]interface{}

	// we have to wrap everything because the driver only exposes interfaces which are not serializable
	for res.Next() {
		valLen := len(res.Record().Values())
		valCap := cap(res.Record().Values())
		if valLen != 0 {
			vals := make([]interface{}, valLen, valCap)
			for i, val := range res.Record().Values() {
				switch val.(type) {
				case neo4j.Path:
					vals[i] = newPathWrap(val.(neo4j.Path))
					break
				case neo4j.Relationship:
					vals[i] = newRelationshipWrap(val.(neo4j.Relationship))
					break
				case neo4j.Node:
					vals[i] = newNodeWrap(val.(neo4j.Node))
					break
				default:
					vals[i] = val
					continue
				}
			}
			result = append(result, vals)
		}
	}

	return result, summary, nil
}

func (s *SessionV2Impl) PurgeDatabase() error {
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

func (s *SessionV2Impl) isTransientError(err error) bool {
	return strings.Contains(err.Error(), "Neo.TransientError.Transaction")
}

func (s *SessionV2Impl) reset() error {
	s.tx = nil

	if s.neoSess != nil {
		err := s.neoSess.Close()
		if err != nil {
			return err
		}

		s.neoSess = nil
	}

	var err error
	s.neoSess, err = driver.Session(s.mode)
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionV2Impl) ManagedTransaction(work TransactionWork) error {
	if work == nil {
		return errors.New("transaction work can not be nil")
	}

	if s.tx != nil {
		return errors.New("can not start managed transaction with pending transaction")
	}

	// add 1 to max retries to account for the first run not being a retry
	for i := 0; i < internalConfig.MaxRetries+1; i++ {
		err := s.Begin()
		if err != nil {
			return err
		}

		err = work(s)
		if err == nil {
			// tx wasn't closed
			if s.tx != nil {
				return s.RollbackWithError(errors.New("tx not closed in work function"))
			}
			return nil
		}

		//fmt.Printf("failed on pass %v with error %s\n", i, err.Error())
		if s.isTransientError(err) {
			// clean up tx
			err = s.reset()
			if err != nil {
				return fmt.Errorf("failed to rollback after transient error, %w", err)
			}

			if i+1 == internalConfig.MaxRetries {
				break
			}

			time.Sleep(internalConfig.RetryWaitDuration)
			continue
		}

		return err
	}

	return errors.New("max retries met, managed tx failed")
}

func (s *SessionV2Impl) Close() error {
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
