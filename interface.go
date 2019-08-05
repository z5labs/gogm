package gogm

import dsl "github.com/mindstand/go-cypherdsl"

type IEdge interface {
	GetStartNode() interface{}
	SetStartNode(v interface{}) error

	GetEndNode() interface{}
	SetEndNode(v interface{}) error
}

//inspiration from -- https://github.com/neo4j/neo4j-ogm/blob/master/core/src/main/java/org/neo4j/ogm/session/Session.java

//session object for ogm interactions
type ISession interface {
	//transaction functions
	ITransaction

	//load single object
	Load(respObj interface{}, id string) error

	//load object with depth
	LoadDepth(respObj interface{}, id string, depth int) error

	//load with depth and filter
	LoadDepthFilter(respObj interface{}, id string, depth int, filter *dsl.ConditionBuilder, params map[string]interface{}) error

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

	//load all edge query
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

	Close() error
}

type ITransaction interface {
	Begin() error
	Rollback() error
	RollbackWithError(err error) error
	Commit() error
}