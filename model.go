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

import "github.com/neo4j/neo4j-go-driver/neo4j"

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

type NodeWrap struct {
	Id     int64                  `json:"id"`
	Labels []string               `json:"labels"`
	Props  map[string]interface{} `json:"props"`
}

func newNodeWrap(node neo4j.Node) *NodeWrap {
	return &NodeWrap{
		Id:     node.Id(),
		Labels: node.Labels(),
		Props:  node.Props(),
	}
}

type PathWrap struct {
	Nodes    []*NodeWrap         `json:"nodes"`
	RelNodes []*RelationshipWrap `json:"rel_nodes"`
}

func newPathWrap(path neo4j.Path) *PathWrap {
	pw := new(PathWrap)
	nodes := path.Nodes()
	if nodes != nil && len(nodes) != 0 {
		nds := make([]*NodeWrap, len(nodes), cap(nodes))
		for i, n := range nodes {
			nds[i] = newNodeWrap(n)
		}

		pw.Nodes = nds
	}

	rels := path.Relationships()
	if rels != nil && len(rels) != 0 {
		newRels := make([]*RelationshipWrap, len(rels), cap(rels))
		for i, rel := range rels {
			newRels[i] = newRelationshipWrap(rel)
		}

		pw.RelNodes = newRels
	}

	return pw
}

type RelationshipWrap struct {
	Id      int64                  `json:"id"`
	StartId int64                  `json:"start_id"`
	EndId   int64                  `json:"end_id"`
	Type    string                 `json:"type"`
	Props   map[string]interface{} `json:"props"`
}

func newRelationshipWrap(rel neo4j.Relationship) *RelationshipWrap {
	return &RelationshipWrap{
		Id:      rel.Id(),
		StartId: rel.StartId(),
		EndId:   rel.EndId(),
		Type:    rel.Type(),
		Props:   rel.Props(),
	}
}
