package gogm

import "github.com/neo4j/neo4j-go-driver/neo4j"

type testNode struct {
	id     int64
	labels []string
	props  map[string]interface{}
}

func (n testNode) Id() int64 {
	return n.id
}

func (n testNode) Labels() []string {
	return n.labels
}

func (n testNode) Props() map[string]interface{} {
	return n.props
}

type testRelationship struct {
	id      int64
	startId int64
	endId   int64
	_type   string
	props   map[string]interface{}
}

func (r testRelationship) Id() int64 {
	return r.id
}

func (r testRelationship) StartId() int64 {
	return r.startId
}

func (r testRelationship) EndId() int64 {
	return r.endId
}

func (r testRelationship) Type() string {
	return r._type
}

func (r testRelationship) Props() map[string]interface{} {
	return r.props
}

type testRelNode struct {
	id    int64
	_type string
	props map[string]interface{}
}

type testPath struct {
	nodes    []*testNode
	relNodes []*testRelationship
	indexes  []int
}

func (t testPath) Nodes() []neo4j.Node {
	nodes := make([]neo4j.Node, len(t.nodes))
	for i, n := range t.nodes {
		nodes[i] = n
	}
	return nodes
}

func (t testPath) Relationships() []neo4j.Relationship {
	nodes := make([]neo4j.Relationship, len(t.relNodes))
	for i, n := range t.relNodes {
		nodes[i] = n
	}
	return nodes
}
