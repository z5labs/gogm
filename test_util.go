package gogm

import "github.com/neo4j/neo4j-go-driver/v4/neo4j"

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

type testRecord struct {
}

func (t *testRecord) Keys() []string {
	panic("implement me")
}

func (t *testRecord) Values() []interface{} {
	return []interface{}{
		testPath{
			nodes: []*testNode{
				{
					labels: []string{"f"},
					props: map[string]interface{}{
						"uuid": "0",
					},
					id: 0,
				},
				{
					labels: []string{"f"},
					props: map[string]interface{}{
						"uuid": "1",
					},
					id: 1,
				},
				{
					labels: []string{"f"},
					props: map[string]interface{}{
						"uuid": "2",
					},
					id: 2,
				},
			},
			relNodes: []*testRelationship{
				{
					id:      3,
					startId: 0,
					endId:   1,
					_type:   "test",
					props:   nil,
				},
				{
					id:      4,
					startId: 1,
					endId:   2,
					_type:   "test",
					props:   nil,
				},
			},
		},
	}

}

func (t testRecord) Get(key string) (interface{}, bool) {
	panic("implement me")
}

func (t testRecord) GetByIndex(index int) interface{} {
	panic("implement me")
}

type testResult struct {
	empty bool
	num   int
}

func (t *testResult) Keys() ([]string, error) {
	panic("implement me")
}

func (t *testResult) Next() bool {
	toRet := !t.empty && t.num == 0

	if !t.empty {
		t.num++
	}

	return toRet
}

func (t *testResult) Err() error {
	panic("implement me")
}

func (t *testResult) Record() neo4j.Record {
	return &testRecord{}
}

func (t *testResult) Summary() (neo4j.ResultSummary, error) {
	panic("implement me")
}

func (t *testResult) Consume() (neo4j.ResultSummary, error) {
	panic("implement me")
}
