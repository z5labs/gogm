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
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestParseStruct(t *testing.T) {
	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}))

	parseO2O(req)

	parseM2O(req)

	parseM2M(req)
}

func parseO2O(req *require.Assertions) {
	//test single save
	comp1 := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		BaseNode: BaseNode{
			Id:   1,
			UUID: "comp1uuid",
			LoadMap: map[string]*RelationConfig{
				"SingleSpecA": {
					Ids:          []int64{2},
					RelationType: Single,
				},
			},
		},
	}

	b1 := &b{
		BaseNode: BaseNode{
			Id:   2,
			UUID: "b1uuid",
			LoadMap: map[string]*RelationConfig{
				"SingleSpec": {
					Ids:          []int64{1},
					RelationType: Single,
				},
			},
		},
		TestField: "test",
	}

	c1 := &c{
		Start: comp1,
		End:   b1,
		Test:  "testing",
	}

	comp1.SingleSpecA = c1
	b1.SingleSpec = c1

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}
	oldRels := map[string]map[string]*RelationConfig{}
	curRels := map[string]map[string]*RelationConfig{}
	ids := []*string{}

	val := reflect.ValueOf(comp1)
	nodeRef := map[string]*reflect.Value{}

	req.Nil(parseStruct("", "", false, dsl.DirectionBoth, nil, &val, 0, 5, &nodes, &relations, &oldRels, &ids, &nodeRef, &[]string{}))
	req.Nil(generateCurRels("", &val, 0, 5, &curRels))
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
	req.Equal(2, len(oldRels))
	req.Equal(2, len(curRels))
	req.Equal(int64(2), curRels["comp1uuid"]["SingleSpecA"].Ids[0])
	req.Equal(int64(1), curRels["b1uuid"]["SingleSpec"].Ids[0])
	req.EqualValues(oldRels, curRels)
}

func parseM2O(req *require.Assertions) {
	//test single save
	a1 := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		BaseNode: BaseNode{
			Id:   1,
			UUID: "a1uuid",
			LoadMap: map[string]*RelationConfig{
				"ManyA": {
					Ids:          []int64{2},
					RelationType: Multi,
				},
			},
		},
		ManyA: []*b{},
	}

	b1 := &b{
		TestField: "test",
		BaseNode: BaseNode{
			Id:   2,
			UUID: "b1uuid",
			LoadMap: map[string]*RelationConfig{
				"ManyB": {
					Ids:          []int64{1},
					RelationType: Single,
				},
			},
		},
	}

	b1.ManyB = a1
	a1.ManyA = append(a1.ManyA, b1)

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}
	oldRels := map[string]map[string]*RelationConfig{}
	curRels := map[string]map[string]*RelationConfig{}
	ids := []*string{}

	val := reflect.ValueOf(a1)
	nodeRef := map[string]*reflect.Value{}
	req.Nil(parseStruct("", "", false, dsl.DirectionBoth, nil, &val, 0, 5, &nodes, &relations, &oldRels, &ids, &nodeRef, &[]string{}))
	req.Nil(generateCurRels("", &val, 0, 5, &curRels))
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
	req.EqualValues(oldRels, curRels)
}

func parseM2M(req *require.Assertions) {
	//test single save
	a1 := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		BaseNode: BaseNode{
			Id:   1,
			UUID: "a1uuid",
			LoadMap: map[string]*RelationConfig{
				"MultiA": {
					Ids:          []int64{2},
					RelationType: Multi,
				},
			},
		},
		ManyA:  []*b{},
		MultiA: []*b{},
	}

	b1 := &b{
		TestField: "test",
		BaseNode: BaseNode{
			Id:   2,
			UUID: "b1uuid",
			LoadMap: map[string]*RelationConfig{
				"Multi": {
					Ids:          []int64{1},
					RelationType: Multi,
				},
			},
		},
		Multi: []*a{},
	}

	b1.Multi = append(b1.Multi, a1)
	a1.MultiA = append(a1.MultiA, b1)

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}
	oldRels := map[string]map[string]*RelationConfig{}
	curRels := map[string]map[string]*RelationConfig{}
	ids := []*string{}

	nodeRef := map[string]*reflect.Value{}

	val := reflect.ValueOf(a1)

	req.Nil(parseStruct("", "", false, dsl.DirectionBoth, nil, &val, 0, 5, &nodes, &relations, &oldRels, &ids, &nodeRef, &[]string{}))
	req.Nil(generateCurRels("", &val, 0, 5, &curRels))
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
	req.EqualValues(oldRels, curRels)
}

func TestCalculateCurRels(t *testing.T) {
	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}))

	//test single save
	a1 := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		BaseNode: BaseNode{
			Id:   1,
			UUID: "a1uuid",
			LoadMap: map[string]*RelationConfig{
				"MultiA": {
					Ids:          []int64{2},
					RelationType: Multi,
				},
			},
		},
		ManyA:  []*b{},
		MultiA: []*b{},
	}

	//b1 := &b{
	//	TestField: "test",
	//	BaseNode: BaseNode{
	//		Id:   2,
	//		UUID: "b1uuid",
	//		LoadMap: map[string]*RelationConfig{
	//			"Multi": {
	//				Ids:          []int64{1},
	//				RelationType: Multi,
	//			},
	//		},
	//	},
	//	Multi: []*a{},
	//}

	//b1.Multi = append(b1.Multi, a1)
	//a1.MultiA = append(a1.MultiA, b1)

	nodes := map[string]map[string]nodeCreateConf{}
	relations := map[string][]relCreateConf{}
	oldRels := map[string]map[string]*RelationConfig{}
	curRels := map[string]map[string]*RelationConfig{}
	ids := []*string{}

	nodeRef := map[string]*reflect.Value{}

	val := reflect.ValueOf(a1)

	req.Nil(parseStruct("", "", false, dsl.DirectionBoth, nil, &val, 0, 5, &nodes, &relations, &oldRels, &ids, &nodeRef, &[]string{}))
	req.Nil(generateCurRels("", &val, 0, 5, &curRels))
	req.Equal(1, len(curRels))
}

func TestCalculateDels(t *testing.T) {
	req := require.New(t)

	//test node removed
	dels := calculateDels(map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
	})

	req.EqualValues(map[string][]int64{
		"node1": {2},
	}, dels)

	//test field removed
	dels = calculateDels(map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
		"node2": {
			"RelFieldNew": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
	})

	req.EqualValues(map[string][]int64{
		"node1": {2},
		"node2": {1},
	}, dels)

	//test field empty
	dels = calculateDels(map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
	})

	req.EqualValues(map[string][]int64{
		"node1": {2},
		"node2": {1},
	}, dels)

	//test nothing changed
	dels = calculateDels(map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[string]map[string]*RelationConfig{
		"node1": {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		"node2": {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	})

	req.EqualValues(map[string][]int64{}, dels)
}
