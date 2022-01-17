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
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"

	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type SessionV2Impl struct {
	gogm         *Gogm
	neoSess      neo4j.Session
	tx           neo4j.Transaction
	DefaultDepth int
	conf         SessionConfig
	lastBookmark string
}

func newSessionWithConfigV2(gogm *Gogm, conf SessionConfig) (*SessionV2Impl, error) {
	if gogm == nil {
		return nil, errors.New("gogm instance can not be nil")
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

	return &SessionV2Impl{
		neoSess:      neoSess,
		DefaultDepth: defaultDepth,
		conf:         conf,
		gogm:         gogm,
	}, nil
}
func (s *SessionV2Impl) Begin(ctx context.Context) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.Begin")
		defer span.Finish()
	} else {
		span = nil
	}

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

func (s *SessionV2Impl) Rollback(ctx context.Context) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.Rollback")
		defer span.Finish()
	} else {
		span = nil
	}

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

func (s *SessionV2Impl) RollbackWithError(ctx context.Context, originalError error) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.RollbackWithError")
		defer span.Finish()
	} else {
		span = nil
	}

	err := s.Rollback(ctx)
	if err != nil {
		return fmt.Errorf("%s%w", err.Error(), originalError)
	}

	return originalError
}

func (s *SessionV2Impl) Commit(ctx context.Context) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.Commit")
		defer span.Finish()
	} else {
		span = nil
	}

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

func (s *SessionV2Impl) Load(ctx context.Context, respObj, id interface{}) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.Load")
		defer span.Finish()
	} else {
		span = nil
	}

	return s.LoadDepthFilterPagination(ctx, respObj, id, s.DefaultDepth, nil, nil, nil)
}

func (s *SessionV2Impl) LoadDepth(ctx context.Context, respObj, id interface{}, depth int) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.LoadDepth")
		defer span.Finish()
	} else {
		span = nil
	}

	return s.LoadDepthFilterPagination(ctx, respObj, id, depth, nil, nil, nil)
}

func (s *SessionV2Impl) LoadDepthFilter(ctx context.Context, respObj, id interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.LoadDepthFilter")
		defer span.Finish()
	} else {
		span = nil
	}

	return s.LoadDepthFilterPagination(ctx, respObj, id, depth, filter, params, nil)
}

func (s *SessionV2Impl) LoadDepthFilterPagination(ctx context.Context, respObj, id interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.LoadDepthFilterPagination")
		defer span.Finish()
	} else {
		span = nil
	}

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

	paramName := "idprm"
	isGraphId := s.gogm.pkStrategy.StrategyName == DefaultPrimaryKeyStrategy.StrategyName
	field := s.gogm.pkStrategy.DBName
	//make the query based off of the load strategy
	switch s.gogm.config.LoadStrategy {
	case PATH_LOAD_STRATEGY:
		query, err = PathLoadStrategyOne(varName, respObjName, field, paramName, isGraphId, depth, filter)
		if err != nil {
			return err
		}
	case SCHEMA_LOAD_STRATEGY:
		query, err = SchemaLoadStrategyOne(s.gogm, varName, respObjName, field, paramName, isGraphId, depth, filter)
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
			paramName: id,
		}
	} else {
		params[paramName] = id
	}

	cyp, err := query.ToCypher()
	if err != nil {
		return err
	}

	return s.runReadOnly(ctx, cyp, params, respObj)
}

func (s *SessionV2Impl) LoadAll(ctx context.Context, respObj interface{}) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.LoadAll")
		defer span.Finish()
	} else {
		span = nil
	}
	return s.LoadAllDepthFilterPagination(ctx, respObj, s.DefaultDepth, nil, nil, nil)
}

func (s *SessionV2Impl) LoadAllDepth(ctx context.Context, respObj interface{}, depth int) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.LoadAllDepth")
		defer span.Finish()
	} else {
		span = nil
	}
	return s.LoadAllDepthFilterPagination(ctx, respObj, depth, nil, nil, nil)
}

func (s *SessionV2Impl) LoadAllDepthFilter(ctx context.Context, respObj interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.LoadAllDepthFilter")
		defer span.Finish()
	} else {
		span = nil
	}
	return s.LoadAllDepthFilterPagination(ctx, respObj, depth, filter, params, nil)
}

func (s *SessionV2Impl) LoadAllDepthFilterPagination(ctx context.Context, respObj interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.LoadAllDepthFilterPagination")
		defer span.Finish()
	} else {
		span = nil
	}

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

	cyp, err := query.ToCypher()
	if err != nil {
		return err
	}

	return s.runReadOnly(ctx, cyp, params, respObj)
}

func (s *SessionV2Impl) runReadOnly(ctx context.Context, cyp string, params map[string]interface{}, respObj interface{}) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.runReadOnly")
		defer span.Finish()
		span.LogKV("query", cyp)
		if s.gogm.config.EnableLogParams {
			s.gogm.logger.Warn("logging params")
			span.LogKV("params", params)
		}
	} else {
		span = nil
	}

	// if in tx, run normally else run in managed tx
	if s.tx != nil {
		if span != nil {
			span.LogKV("info", "running in existing transaction")
		}
		result, err := s.tx.Run(cyp, params)
		if err != nil {
			return err
		}

		return decode(s.gogm, result, respObj)
	}
	// run inside managed transaction if not already in a transaction
	if span != nil {
		span.LogKV("info", "running in driver managed transaction")
	}
	_, err := s.neoSess.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		res, err := tx.Run(cyp, params)
		if err != nil {
			return nil, err
		}

		return nil, decode(s.gogm, res, respObj)
	}, neo4j.WithTxTimeout(time.Until(s.getDeadline(ctx))))
	if err != nil {
		return fmt.Errorf("failed auto read tx, %w", err)
	}

	return nil
}

func (s *SessionV2Impl) Save(ctx context.Context, saveObj interface{}) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.Save")
		defer span.Finish()
	} else {
		span = nil
	}

	return s.SaveDepth(ctx, saveObj, s.DefaultDepth)
}

func (s *SessionV2Impl) SaveDepth(ctx context.Context, saveObj interface{}, depth int) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.SaveDepth")
		defer span.Finish()
	} else {
		span = nil
	}

	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	return s.runWrite(ctx, saveDepth(s.gogm, saveObj, depth))
}

func (s *SessionV2Impl) Delete(ctx context.Context, deleteObj interface{}) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.Delete")
		defer span.Finish()
	} else {
		span = nil
	}

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

	return s.runWrite(ctx, workFunc)
}

func (s *SessionV2Impl) DeleteUUID(ctx context.Context, uuid string) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.DeleteUUID")
		defer span.Finish()
	} else {
		span = nil
	}

	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	// handle if in transaction
	return s.runWrite(ctx, deleteByUuids(uuid))
}

func (s *SessionV2Impl) runWrite(ctx context.Context, work neo4j.TransactionWork) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.runWrite")
		defer span.Finish()
	} else {
		span = nil
	}

	// if already in a transaction
	if s.tx != nil {
		_, err := work(s.tx)
		if err != nil {
			return fmt.Errorf("failed to save in manual tx, %w", err)
		}

		return nil
	}

	s.gogm.logger.Debug("running in managed write transaction")
	_, err := s.neoSess.WriteTransaction(work, neo4j.WithTxTimeout(time.Until(s.getDeadline(ctx))))
	if err != nil {
		return fmt.Errorf("failed to save in auto transaction, %w", err)
	}

	return nil
}

func (s *SessionV2Impl) Query(ctx context.Context, query string, properties map[string]interface{}, respObj interface{}) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.Query")
		defer span.Finish()
	} else {
		span = nil
	}

	if s.neoSess == nil {
		return errors.New("neo4j connection not initialized")
	}

	if s.conf.AccessMode == neo4j.AccessModeRead {
		return s.runReadOnly(ctx, query, properties, respObj)
	}

	if s.conf.AccessMode == AccessModeRead {
		return s.runReadOnly(ctx, query, properties, respObj)
	}

	return s.runWrite(ctx, func(tx neo4j.Transaction) (interface{}, error) {
		res, err := tx.Run(query, properties)
		if err != nil {
			return nil, err
		}

		return nil, decode(s.gogm, res, respObj)
	})
}

func (s *SessionV2Impl) QueryRaw(ctx context.Context, query string, properties map[string]interface{}) ([][]interface{}, neo4j.ResultSummary, error) {
	if s.neoSess == nil {
		return nil, nil, errors.New("neo4j connection not initialized")
	}
	var err error
	if s.tx != nil {
		res, err := s.tx.Run(query, properties)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to execute query, %w", err)
		}

		parsedResult := s.parseResult(res)

		sum, err := res.Consume()
		if err != nil {
			return nil, nil, err
		}

		return parsedResult, sum, nil
	} else {
		var ires interface{}
		var sum neo4j.ResultSummary
		if s.conf.AccessMode == AccessModeRead {
			ires, err = s.neoSess.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
				res, err := tx.Run(query, properties)
				if err != nil {
					return nil, err
				}

				pres := s.parseResult(res)

				sum, err = res.Consume()
				if err != nil {
					return nil, err
				}

				return pres, nil
			})
		} else {
			ires, err = s.neoSess.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
				res, err := tx.Run(query, properties)
				if err != nil {
					return nil, err
				}

				pres := s.parseResult(res)

				sum, err = res.Consume()
				if err != nil {
					return nil, err
				}

				return pres, nil
			})
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to run auto transaction, %w", err)
		}

		result, ok := ires.([][]interface{})
		if !ok {
			return nil, nil, fmt.Errorf("failed to cast %T to [][]interface{}", ires)
		}

		return result, sum, nil
	}
}

func (s *SessionV2Impl) parseResult(res neo4j.Result) [][]interface{} {
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

	s.neoSess = s.gogm.driver.NewSession(neo4j.SessionConfig{
		AccessMode:   s.conf.AccessMode,
		Bookmarks:    s.conf.Bookmarks,
		DatabaseName: s.conf.DatabaseName,
		FetchSize:    s.conf.FetchSize,
	})

	return nil
}

func (s *SessionV2Impl) getDeadline(ctx context.Context) time.Time {
	var deadline time.Time
	var ok bool

	// handle timeout info
	if ctx != nil {
		deadline, ok = ctx.Deadline()
		if !ok {
			deadline = time.Now().Add(s.gogm.config.DefaultTransactionTimeout)
		}
	} else {
		deadline = time.Now().Add(s.gogm.config.DefaultTransactionTimeout)
	}

	return deadline
}

func (s *SessionV2Impl) ManagedTransaction(ctx context.Context, work TransactionWork) error {
	var span opentracing.Span
	if ctx != nil && s.gogm.config.OpentracingEnabled {
		span, ctx = opentracing.StartSpanFromContext(ctx, "gogm.SessionV2Impl.ManagedTransaction")
		defer span.Finish()
	} else {
		span = nil
	}

	if work == nil {
		return errors.New("transaction work can not be nil")
	}

	if s.tx != nil {
		return errors.New("can not start managed transaction with pending transaction")
	}

	txWork := func(tx neo4j.Transaction) (interface{}, error) {
		s.tx = tx
		return nil, work(s)
	}

	defer s.clearTx()
	// handle timeout info
	deadline := s.getDeadline(ctx)

	if s.conf.AccessMode == AccessModeWrite {
		_, err := s.neoSess.WriteTransaction(txWork, neo4j.WithTxTimeout(time.Until(deadline)))
		if err != nil {
			return fmt.Errorf("failed managed write tx, %w", err)
		}

		s.lastBookmark = s.neoSess.LastBookmark()

		return nil
	}

	_, err := s.neoSess.ReadTransaction(txWork)
	if err != nil {
		return fmt.Errorf("failed managed write tx, %w", err)
	}

	s.lastBookmark = s.neoSess.LastBookmark()

	return nil
}

func (s *SessionV2Impl) clearTx() {
	s.tx = nil
}

func (s *SessionV2Impl) Close() error {
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
