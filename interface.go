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
	"reflect"

	dsl "github.com/mindstand/go-cypherdsl"
)

// Edge specifies required functions for special edge nodes
type Edge interface {
	// GetStartNode gets start node of edge
	GetStartNode() interface{}
	// GetStartNodeType gets reflect type of start node
	GetStartNodeType() reflect.Type
	// SetStartNode sets start node of edge
	SetStartNode(v interface{}) error

	// GetEndNode gets end node of edge
	GetEndNode() interface{}
	// GetEndNodeType gets reflect type of end node
	GetEndNodeType() reflect.Type
	// SetEndNode sets end node of edge
	SetEndNode(v interface{}) error
}

//inspiration from -- https://github.com/neo4j/neo4j-ogm/blob/master/core/src/main/java/org/neo4j/ogm/session/Session.java

// ISession: V1 session object for ogm interactions
// Deprecated: use SessionV2 instead
type ISession interface {
	//transaction functions
	ITransaction

	//load single object
	Load(respObj interface{}, id string) error

	//load object with depth
	LoadDepth(respObj interface{}, id string, depth int) error

	//load with depth and filter
	LoadDepthFilter(respObj interface{}, id string, depth int, filter dsl.ConditionOperator, params map[string]interface{}) error

	//load with depth, filter and pagination
	LoadDepthFilterPagination(respObj interface{}, id string, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error

	//load slice of something
	LoadAll(respObj interface{}) error

	//load all of depth
	LoadAllDepth(respObj interface{}, depth int) error

	//load all of type with depth and filter
	LoadAllDepthFilter(respObj interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}) error

	//load all with depth, filter and pagination
	LoadAllDepthFilterPagination(respObj interface{}, depth int, filter dsl.ConditionOperator, params map[string]interface{}, pagination *Pagination) error

	// load all edge query
	// Deprecated: No equivalent function in SessionV2
	LoadAllEdgeConstraint(respObj interface{}, endNodeType, endNodeField string, edgeConstraint interface{}, minJumps, maxJumps, depth int, filter dsl.ConditionOperator) error

	//save object
	Save(saveObj interface{}) error

	//save object with depth
	SaveDepth(saveObj interface{}, depth int) error

	//delete
	Delete(deleteObj interface{}) error

	//delete uuid
	DeleteUUID(uuid string) error

	//specific query, responds to slice and single objects
	Query(query string, properties map[string]interface{}, respObj interface{}) error

	//similar to query, but returns raw rows/cols
	QueryRaw(query string, properties map[string]interface{}) ([][]interface{}, error)

	//delete everything, this will literally delete everything
	PurgeDatabase() error

	// closes session
	Close() error
}

// ITransaction specifies functions for Neo4j ACID transactions
// Deprecated: Use TransactionV2 instead
type ITransaction interface {
	// Begin begins transaction
	Begin() error
	// Rollback rolls back transaction
	Rollback() error
	// RollbackWithError wraps original error into rollback error if there is one
	RollbackWithError(err error) error
	// Commit commits transaction
	Commit() error
}
