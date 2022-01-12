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

	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

//session version 2 is experimental to start trying breaking changes
type SessionV2 interface {
	//transaction functions
	TransactionV2

	// Begin begins transaction
	Begin(ctx context.Context) error

	// ManagedTransaction runs tx work managed for retry
	ManagedTransaction(ctx context.Context, work TransactionWork) error

	// closes session
	Close() error
}

// TransactionV2 specifies functions for Neo4j ACID transactions
type TransactionV2 interface {
	// Rollback rolls back transaction
	Rollback(ctx context.Context) error
	// RollbackWithError wraps original error into rollback error if there is one
	RollbackWithError(ctx context.Context, err error) error
	// Commit commits transaction
	Commit(ctx context.Context) error

	// functions the tx can do
	ogmFunctions
}

type ogmFunctions interface {
	//load single object
	Load(ctx context.Context, respObj, id interface{}) error

	//load object with depth
	LoadDepth(ctx context.Context, respObj, id interface{}, depth int) error

	//load with depth and filter
	LoadDepthFilter(ctx context.Context, respObj, id interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}) error

	//load with depth, filter and pagination
	LoadDepthFilterPagination(ctx context.Context, respObj, id interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error

	//load slice of something
	LoadAll(ctx context.Context, respObj interface{}) error

	//load all of depth
	LoadAllDepth(ctx context.Context, respObj interface{}, depth int) error

	//load all of type with depth and filter
	LoadAllDepthFilter(ctx context.Context, respObj interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}) error

	//load all with depth, filter and pagination
	LoadAllDepthFilterPagination(ctx context.Context, respObj interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error

	//save object at default depth
	Save(ctx context.Context, saveObj interface{}) error

	//save object with depth
	SaveDepth(ctx context.Context, saveObj interface{}, depth int) error

	//delete
	Delete(ctx context.Context, deleteObj interface{}) error

	//delete uuid
	DeleteUUID(ctx context.Context, uuid string) error

	//specific query, responds to slice and single objects
	Query(ctx context.Context, query string, properties map[string]interface{}, respObj interface{}) error

	//similar to query, but returns raw rows/cols
	QueryRaw(ctx context.Context, query string, properties map[string]interface{}) ([][]interface{}, neo4j.ResultSummary, error)
}

type TransactionWork func(tx TransactionV2) error
