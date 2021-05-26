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
	"time"
)

func TestParseStruct(t *testing.T) {
	req := require.New(t)

	gogm, err := getTestGogm()
	req.Nil(err)
	req.NotNil(gogm)

	parseO2O(gogm, req)

	parseM2O(gogm, req)

	parseM2M(gogm, req)
}

func parseO2O(gogm *Gogm, req *require.Assertions) {
	//test single save
	comp1 := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		BaseUUIDNode: BaseUUIDNode{
			UUID: "comp1uuid",
			BaseNode: BaseNode{
				Id: 1,
				LoadMap: map[string]*RelationConfig{
					"SingleSpecA": {
						Ids:          []int64{2},
						RelationType: Single,
					},
				},
			},
		},
	}

	b1 := &b{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "b1uuid",
			BaseNode: BaseNode{
				Id: 2,
				LoadMap: map[string]*RelationConfig{
					"SingleSpec": {
						Ids:          []int64{1},
						RelationType: Single,
					},
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

	var (
		// [LABEL][uintptr]{config}
		nodes = map[string]map[uintptr]*nodeCreate{}
		// [LABEL] []{config}
		relations = map[string][]*relCreate{}
		// node id -- [field] config
		oldRels = map[uintptr]map[string]*RelationConfig{}
		// node id -- [field] config
		curRels = map[int64]map[string]*RelationConfig{}
		// id to reflect value
		nodeIdRef = map[uintptr]int64{}
		// uintptr to reflect value (for new nodes that dont have a graph id yet)
		nodeRef = map[uintptr]*reflect.Value{}
	)

	val := reflect.ValueOf(comp1)

	req.Nil(parseStruct(gogm, 0, "", false, dsl.DirectionBoth, nil, &val, 0, 5,
		nodes, relations, nodeIdRef, nodeRef, oldRels))
	req.Nil(generateCurRels(gogm, 0, &val, 0, 5, curRels))
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
	req.Equal(2, len(oldRels))
	req.Equal(2, len(curRels))
	// req.Equal(int64(2), curRels["comp1uuid"]["SingleSpecA"].Ids[0])
	// req.Equal(int64(1), curRels["b1uuid"]["SingleSpec"].Ids[0])
	// todo better way to test this specifically
	// req.EqualValues(oldRels, curRels)
}

func parseM2O(gogm *Gogm, req *require.Assertions) {
	//test single save
	a1 := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,
		BaseUUIDNode: BaseUUIDNode{
			UUID: "a1uuid",
			BaseNode: BaseNode{
				Id: 1,
				LoadMap: map[string]*RelationConfig{
					"ManyA": {
						Ids:          []int64{2},
						RelationType: Multi,
					},
				},
			},
		},
		ManyA: []*b{},
	}

	b1 := &b{
		TestField: "test",
		BaseUUIDNode: BaseUUIDNode{
			UUID: "b1uuid",
			BaseNode: BaseNode{
				Id: 2,
				LoadMap: map[string]*RelationConfig{
					"ManyB": {
						Ids:          []int64{1},
						RelationType: Single,
					},
				},
			},
		},
	}

	b1.ManyB = a1
	a1.ManyA = append(a1.ManyA, b1)

	var (
		// [LABEL][int64 (graphid) or uintptr]{config}
		nodes = map[string]map[uintptr]*nodeCreate{}
		// [LABEL] []{config}
		relations = map[string][]*relCreate{}
		// node id -- [field] config
		oldRels = map[uintptr]map[string]*RelationConfig{}
		// node id -- [field] config
		curRels = map[int64]map[string]*RelationConfig{}
		// id to reflect value
		nodeIdRef = map[uintptr]int64{}
		// uintptr to reflect value (for new nodes that dont have a graph id yet)
		nodeRef = map[uintptr]*reflect.Value{}
	)

	val := reflect.ValueOf(a1)
	req.Nil(parseStruct(gogm, 0, "", false, dsl.DirectionBoth, nil, &val, 0, 5,
		nodes, relations, nodeIdRef, nodeRef, oldRels))
	req.Nil(generateCurRels(gogm, 0, &val, 0, 5, curRels))
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
	// todo better way to test this
	// req.EqualValues(oldRels, curRels)
}

func parseM2M(gogm *Gogm, req *require.Assertions) {
	//test single save
	a1 := &a{
		TestField:         "test",
		TestTypeDefString: "dasdfas",
		TestTypeDefInt:    600,

		BaseUUIDNode: BaseUUIDNode{
			UUID: "a1uuid",
			BaseNode: BaseNode{
				Id: 1,

				LoadMap: map[string]*RelationConfig{
					"MultiA": {
						Ids:          []int64{2},
						RelationType: Multi,
					},
				},
			},
		},
		ManyA:  []*b{},
		MultiA: []*b{},
	}

	b1 := &b{
		TestField: "test",
		BaseUUIDNode: BaseUUIDNode{
			UUID: "b1uuid",
			BaseNode: BaseNode{
				Id: 2,
				LoadMap: map[string]*RelationConfig{
					"Multi": {
						Ids:          []int64{1},
						RelationType: Multi,
					},
				},
			},
		},

		Multi: []*a{},
	}

	b1.Multi = append(b1.Multi, a1)
	a1.MultiA = append(a1.MultiA, b1)

	var (
		// [LABEL][int64 (graphid) or uintptr]{config}
		nodes = map[string]map[uintptr]*nodeCreate{}
		// [LABEL] []{config}
		relations = map[string][]*relCreate{}
		// node id -- [field] config
		oldRels = map[uintptr]map[string]*RelationConfig{}
		// node id -- [field] config
		curRels = map[int64]map[string]*RelationConfig{}
		// id to reflect value
		nodeIdRef = map[uintptr]int64{}
		// uintptr to reflect value (for new nodes that dont have a graph id yet)
		nodeRef = map[uintptr]*reflect.Value{}
	)

	val := reflect.ValueOf(a1)
	req.Nil(parseStruct(gogm, 0, "", false, dsl.DirectionBoth, nil, &val, 0, 5,
		nodes, relations, nodeIdRef, nodeRef, oldRels))
	req.Nil(generateCurRels(gogm, 0, &val, 0, 5, curRels))
	req.Equal(2, len(nodes))
	req.Equal(1, len(nodes["a"]))
	req.Equal(1, len(nodes["b"]))
	req.Equal(1, len(relations))
	// todo better way to test this
	// req.EqualValues(oldRels, curRels)
}

func TestCalculateCurRels(t *testing.T) {
	req := require.New(t)

	gogm, err := getTestGogm()
	req.Nil(err)
	req.NotNil(gogm)

	cases := []struct {
		Name       string
		Expected   map[int64]map[string]*RelationConfig
		Value      interface{}
		ShouldPass bool
		Depth      int
	}{
		{
			Name: "Basic test",
			Expected: map[int64]map[string]*RelationConfig{
				1: {
					"MultiA": {
						Ids:          []int64{2},
						RelationType: Multi,
					},
				},
				2: {
					"Multi": {
						Ids:          []int64{1},
						RelationType: Multi,
					},
				},
			},
			Value: func() interface{} {
				a1 := &a{
					TestField:         "test",
					TestTypeDefString: "dasdfas",
					TestTypeDefInt:    600,
					BaseUUIDNode: BaseUUIDNode{
						UUID: "a1uuid",
						BaseNode: BaseNode{
							Id: 1,
							LoadMap: map[string]*RelationConfig{
								"MultiA": {
									Ids:          []int64{2},
									RelationType: Multi,
								},
							},
						},
					},
					ManyA:  []*b{},
					MultiA: []*b{},
				}

				b1 := &b{
					TestField: "test",
					BaseUUIDNode: BaseUUIDNode{
						BaseNode: BaseNode{
							Id: 2,
							LoadMap: map[string]*RelationConfig{
								"Multi": {
									Ids:          []int64{1},
									RelationType: Multi,
								},
							},
						},
						UUID: "b1uuid",
					},
					Multi: []*a{},
				}

				b1.Multi = append(b1.Multi, a1)
				a1.MultiA = append(a1.MultiA, b1)

				return a1
			}(),
			ShouldPass: true,
			Depth:      1,
		},
		{
			Name: "from integration test",
			Expected: map[int64]map[string]*RelationConfig{
				1: {
					"SingleSpecA": {
						Ids:          []int64{2},
						RelationType: Single,
					},
					"ManyA": {
						Ids:          []int64{3},
						RelationType: Multi,
					},
				},
				2: {
					"SingleSpec": {
						Ids:          []int64{1},
						RelationType: Single,
					},
				},
				3: {
					"ManyB": {
						Ids:          []int64{1},
						RelationType: Single,
					},
				},
			},
			Value: func() interface{} {
				a2 := &a{
					BaseUUIDNode: BaseUUIDNode{
						BaseNode: BaseNode{
							Id: 1,
						},
					},
					TestField: "test",
					PropTest0: map[string]interface{}{
						"test.test": "test",
						"test2":     1,
					},
					PropTest1: map[string]string{
						"test": "test",
					},
					PropsTest2: []string{"test", "test"},
					PropsTest3: []int{1, 2},
				}

				b2 := &b{
					BaseUUIDNode: BaseUUIDNode{
						BaseNode: BaseNode{
							Id: 2,
						},
					},
					TestField: "test",
					TestTime:  time.Now().UTC(),
				}

				b3 := &b{
					BaseUUIDNode: BaseUUIDNode{
						BaseNode: BaseNode{
							Id: 3,
						},
					},
					TestField: "dasdfasd",
				}

				edgeC1 := &c{
					BaseUUIDNode: BaseUUIDNode{
						BaseNode: BaseNode{
							Id: 4,
						},
					},
					Start: a2,
					End:   b2,
					Test:  "testing",
				}

				a2.SingleSpecA = edgeC1
				a2.ManyA = []*b{b3}
				b2.SingleSpec = edgeC1
				b3.ManyB = a2
				// a2 -> b2
				// a2 -> b3
				return a2
			}(),
			ShouldPass: true,
			Depth:      2,
		},
	}

	for _, _case := range cases {
		var (
			// [LABEL][int64 (graphid) or uintptr]{config}
			nodes = map[string]map[uintptr]*nodeCreate{}
			// [LABEL] []{config}
			relations = map[string][]*relCreate{}
			// node id -- [field] config
			oldRels = map[uintptr]map[string]*RelationConfig{}
			// node id -- [field] config
			curRels = map[int64]map[string]*RelationConfig{}
			// id to reflect value
			nodeIdRef = map[uintptr]int64{}
			// uintptr to reflect value (for new nodes that dont have a graph id yet)
			nodeRef = map[uintptr]*reflect.Value{}
		)
		t.Log("running test -", _case.Name)
		val := reflect.ValueOf(_case.Value)
		req.Nil(parseStruct(gogm, 0, "", false, dsl.DirectionBoth, nil, &val, 0, _case.Depth,
			nodes, relations, nodeIdRef, nodeRef, oldRels))
		err = generateCurRels(gogm, 0, &val, 0, _case.Depth, curRels)
		if _case.ShouldPass {
			req.Nil(err)
			req.Equal(_case.Expected, curRels, "Expected rels not equal to generated")
		} else {
			req.NotNil(err)
		}
	}
}

func TestCalculateDels(t *testing.T) {
	req := require.New(t)

	//test node removed
	dels, err := calculateDels(map[uintptr]map[string]*RelationConfig{
		uintptr(1): {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		uintptr(2): {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[int64]map[string]*RelationConfig{
		1: {
			"RelField": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
	}, map[uintptr]int64{
		uintptr(1): 1,
		uintptr(2): 2,
	})
	req.Nil(err)

	req.EqualValues(map[int64][]int64{
		1: {2},
	}, dels)

	//test field removed
	dels, err = calculateDels(map[uintptr]map[string]*RelationConfig{
		uintptr(1): {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		uintptr(2): {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[int64]map[string]*RelationConfig{
		1: {
			"RelField": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
		2: {
			"RelFieldNew": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
	}, map[uintptr]int64{
		uintptr(1): 1,
		uintptr(2): 2,
	})
	req.Nil(err)

	req.EqualValues(map[int64][]int64{
		1: {2},
		2: {1},
	}, dels)

	//test field empty
	dels, err = calculateDels(map[uintptr]map[string]*RelationConfig{
		uintptr(1): {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		uintptr(2): {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[int64]map[string]*RelationConfig{
		1: {
			"RelField": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
		2: {
			"RelField2": {
				Ids:          []int64{},
				RelationType: Single,
			},
		},
	}, map[uintptr]int64{
		uintptr(1): 1,
		uintptr(2): 2,
	})

	req.EqualValues(map[int64][]int64{
		1: {2},
		2: {1},
	}, dels)

	//test nothing changed
	dels, err = calculateDels(map[uintptr]map[string]*RelationConfig{
		uintptr(1): {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		uintptr(2): {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[int64]map[string]*RelationConfig{
		1: {
			"RelField": {
				Ids:          []int64{2},
				RelationType: Single,
			},
		},
		2: {
			"RelField2": {
				Ids:          []int64{1},
				RelationType: Single,
			},
		},
	}, map[uintptr]int64{
		uintptr(1): 1,
		uintptr(2): 2,
	})
	req.Nil(err)

	req.EqualValues(map[int64][]int64{}, dels)
}
