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
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

const defaultDepth = 1

const AccessModeRead = neo4j.AccessModeRead
const AccessModeWrite = neo4j.AccessModeWrite

type SessionConfig neo4j.SessionConfig

// Deprecated: Session will be removed in a later release in favor of SessionV2
type Session struct {
	gogm         *Gogm
	neoSess      neo4j.Session
	tx           neo4j.Transaction
	DefaultDepth int
	mode         neo4j.AccessMode
}

// uses global gogm
// Deprecated: Gogm.NewSession instead
func NewSession(readonly bool) (*Session, error) {
	return newSession(globalGogm, readonly)
}

func newSession(gogm *Gogm, readonly bool) (*Session, error) {
	if gogm == nil {
		return nil, errors.New("gogm instance cannot be nil")
	}

	if gogm.isNoOp {
		return nil, errors.New("please set global gogm instance with SetGlobalGogm()")
	}

	if gogm.driver == nil {
		return nil, errors.New("gogm driver not initialized")
	}

	session := &Session{
		gogm: gogm,
	}

	var mode neo4j.AccessMode

	if readonly {
		mode = AccessModeRead
	} else {
		mode = AccessModeWrite
	}

	neoSess := gogm.driver.NewSession(neo4j.SessionConfig{AccessMode: mode, FetchSize: neo4j.FetchDefault})

	session.neoSess = neoSess
	session.mode = mode

	session.DefaultDepth = defaultDepth

	return session, nil
}

// Deprecated: Gogm.NewSessionWithConfig instead
func NewSessionWithConfig(conf SessionConfig) (*Session, error) {
	return newSessionWithConfig(globalGogm, conf)
}

func newSessionWithConfig(gogm *Gogm, conf SessionConfig) (*Session, error) {
	if gogm == nil {
		return nil, errors.New("gogm instance is nil")
	}

	if gogm.isNoOp {
		return nil, errors.New("please set global gogm instance with SetGlobalGogm()")
	}

	if gogm.driver == nil {
		return nil, errors.New("gogm driver not initialized")
	}
	neoSess := gogm.driver.NewSession(neo4j.SessionConfig{
		AccessMode:   conf.AccessMode,
		Bookmarks:    conf.Bookmarks,
		DatabaseName: conf.DatabaseName,
		FetchSize:    neo4j.FetchDefault,
	})

	return &Session{
		neoSess:      neoSess,
		DefaultDepth: defaultDepth,
		mode:         conf.AccessMode,
		gogm:         gogm,
	}, nil
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

func (s *Session) LoadDepthFilter(respObj interface{}, id string, depth int, filter dsl.ConditionOperator, params map[string]interface{}) error {
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
	switch s.gogm.config.LoadStrategy {
	case PATH_LOAD_STRATEGY:
		query, err = PathLoadStrategyOne(varName, respObjName, "uuid", "uuid", false, depth, filter)
		if err != nil {
			return err
		}
	case SCHEMA_LOAD_STRATEGY:
		query, err = SchemaLoadStrategyOne(s.gogm, varName, respObjName, "uuid", "uuid", false, depth, filter)
		if err != nil {
			return err
		}
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
	cyp, err := query.ToCypher()
	if err != nil {
		return err
	}

	return s.runReadOnly(cyp, params, respObj)
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
	switch s.gogm.config.LoadStrategy {
	case PATH_LOAD_STRATEGY:
		query, err = PathLoadStrategyMany(varName, respObjName, depth, filter)
		if err != nil {
			return err
		}
	case SCHEMA_LOAD_STRATEGY:
		query, err = SchemaLoadStrategyMany(s.gogm, varName, respObjName, depth, filter)
		if err != nil {
			return err
		}
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
	cyp, err := query.ToCypher()
	if err != nil {
		return err
	}

	return s.runReadOnly(cyp, params, respObj)
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

	// there is no Schema Load Strategy implementation of EdgeConstraint as it would involve pathfinding within the schema (which would be expensive)
	query, err = PathLoadStrategyEdgeConstraint(varName, respObjName, endNodeType, endNodeField, minJumps, maxJumps, depth, filter)
	if err != nil {
		return err
	}

	// handle if in transaction
	cyp, err := query.ToCypher()
	if err != nil {
		return err
	}

	return s.runReadOnly(cyp, map[string]interface{}{
		endNodeField: edgeConstraint,
	}, respObj)
}

func (s *Session) runReadOnly(cyp string, params map[string]interface{}, respObj interface{}) error {
	// if in tx, run normally else run in managed tx
	if s.tx != nil {
		result, err := s.tx.Run(cyp, params)
		if err != nil {
			return err
		}

		return decode(s.gogm, result, respObj)
	}
	// run inside managed transaction if not already in a transaction
	_, err := s.neoSess.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		res, err := tx.Run(cyp, params)
		if err != nil {
			return nil, err
		}

		return nil, decode(s.gogm, res, respObj)
	})
	if err != nil {
		return fmt.Errorf("failed auto read tx, %w", err)
	}

	return nil
}

func (s *Session) Save(saveObj interface{}) error {
	return s.SaveDepth(saveObj, s.DefaultDepth)
}

func (s *Session) SaveDepth(saveObj interface{}, depth int) error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	// handle if in transaction
	return s.runWrite(saveDepth(s.gogm, saveObj, depth))
}

func (s *Session) Delete(deleteObj interface{}) error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	if deleteObj == nil {
		return errors.New("deleteObj can not be nil")
	}

	// handle if in transaction
	workFunc, err := deleteNode(deleteObj)
	if err != nil {
		return fmt.Errorf("failed to generate work func for delete, %w", err)
	}

	return s.runWrite(workFunc)
}

func (s *Session) DeleteUUID(uuid string) error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	// handle if in transaction
	return s.runWrite(deleteByUuids(uuid))
}

func (s *Session) runWrite(work neo4j.TransactionWork) error {
	// if already in a transaction
	if s.tx != nil {
		_, err := work(s.tx)
		if err != nil {
			return fmt.Errorf("failed to save in manual tx, %w", err)
		}

		return nil
	}

	_, err := s.neoSess.WriteTransaction(work)
	if err != nil {
		return fmt.Errorf("failed to save in auto transaction, %w", err)
	}

	return nil
}

func (s *Session) Query(query string, properties map[string]interface{}, respObj interface{}) error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	if s.mode == neo4j.AccessModeRead {
		return s.runReadOnly(query, properties, respObj)
	}

	return s.runWrite(func(tx neo4j.Transaction) (interface{}, error) {
		res, err := tx.Run(query, properties)
		if err != nil {
			return nil, err
		}

		return nil, decode(s.gogm, res, respObj)
	})
}

func (s *Session) QueryRaw(query string, properties map[string]interface{}) ([][]interface{}, error) {
	if s.neoSess == nil {
		return nil, errors.New("neo4j connection not initialized")
	}
	var err error
	if s.tx != nil {
		res, err := s.tx.Run(query, properties)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query, %w", err)
		}

		return s.parseResult(res), nil
	} else {
		var ires interface{}
		if s.mode == AccessModeRead {
			ires, err = s.neoSess.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
				res, err := tx.Run(query, properties)
				if err != nil {
					return nil, err
				}

				return s.parseResult(res), nil
			})
		} else {
			ires, err = s.neoSess.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
				res, err := tx.Run(query, properties)
				if err != nil {
					return nil, err
				}

				return s.parseResult(res), nil
			})
		}
		if err != nil {
			return nil, fmt.Errorf("failed to run auto transaction, %w", err)
		}

		result, ok := ires.([][]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to cast %T to [][]interface{}", ires)
		}

		return result, nil
	}
}

func (s *Session) parseResult(res neo4j.Result) [][]interface{} {
	var result [][]interface{}

	// we have to wrap everything because the driver only exposes interfaces which are not serializable
	for res.Next() {
		valLen := len(res.Record().Values)
		valCap := cap(res.Record().Values)
		if valLen != 0 {
			vals := make([]interface{}, valLen, valCap)
			for i, val := range res.Record().Values {
				switch v := val.(type) {
				case neo4j.Path:
					vals[i] = v
				case neo4j.Relationship:
					vals[i] = v
				case neo4j.Node:
					vals[i] = v
				default:
					vals[i] = v
				}
			}
			result = append(result, vals)
		}
	}

	return result
}

func (s *Session) PurgeDatabase() error {
	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	// handle if in transaction
	cyp, err := dsl.QB().Match(dsl.Path().V(dsl.V{Name: "n"}).Build()).Delete(true, "n").ToCypher()
	if err != nil {
		return err
	}

	return s.runWrite(func(tx neo4j.Transaction) (interface{}, error) {
		return tx.Run(cyp, nil)
	})
}

func (s *Session) Close() error {
	if s.neoSess == nil {
		return fmt.Errorf("cannot close nil connection: %w", ErrInternal)
	}

	// handle tx
	if s.tx != nil {
		s.gogm.logger.Warn("attempting to close a session with a pending transaction. Tx is being rolled back")
		err := s.tx.Rollback()
		if err != nil {
			return err
		}
		s.tx = nil
	}

	return s.neoSess.Close()
}
