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

const loadMapField = "LoadMap"

// BaseNode contains fields that ALL GoGM nodes are required to have
type BaseNode struct {
	// Id is the GraphId that neo4j uses internally
	Id *int64 `json:"-" gogm:"pk=default"`

	// LoadMap represents the state of how a node was loaded for neo4j.
	// This is used to determine if relationships are removed on save
	// field -- relations
	LoadMap map[string]*RelationConfig `json:"-" gogm:"-"`
}

type BaseUUIDNode struct {
	BaseNode
	// UUID is the unique identifier GoGM uses as a primary key
	UUID string `gogm:"pk=UUID"`
}

// Specifies Type of testRelationship
type RelationType int

const (
	// Side of relationship can only point to 0 or 1 other nodes
	Single RelationType = 0

	// Side of relationship can point to 0+ other nodes
	Multi RelationType = 1
)

// RelationConfig specifies how relationships are loaded
type RelationConfig struct {
	// stores graph ids
	Ids []int64 `json:"-" gomg:"-"`
	// specifies relationship type
	RelationType RelationType `json:"-"  gomg:"-"`
}

// specifies how edges are loaded
type neoEdgeConfig struct {
	Id int64

	StartNodeId   int64
	StartNodeType string

	EndNodeId   int64
	EndNodeType string

	Obj map[string]interface{}

	Type string
}
