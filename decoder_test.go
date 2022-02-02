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
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/cornelk/hashmap"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/stretchr/testify/require"
)

func TestTraverseResultRecordValues(t *testing.T) {
	req := require.New(t)

	// empty case
	pArr, rArr, nArr := traverseResultRecordValues([]interface{}{})
	req.Len(pArr, 0)
	req.Len(rArr, 0)
	req.Len(nArr, 0)

	// garbage record case
	pArr, rArr, nArr = traverseResultRecordValues([]interface{}{"hello", []interface{}{"there"}})
	req.Len(pArr, 0)
	req.Len(rArr, 0)
	req.Len(nArr, 0)

	// define our test paths, rels, and nodes
	p1 := neo4j.Path{
		Nodes: []neo4j.Node{
			{
				Id:     1,
				Labels: []string{"start"},
			},
			{
				Id:     2,
				Labels: []string{"end"},
			},
		},
		Relationships: []neo4j.Relationship{
			{
				Id:      3,
				StartId: 1,
				EndId:   2,
				Type:    "someType",
			},
		},
	}

	n1 := neo4j.Node{
		Id:     4,
		Labels: []string{"start"},
	}

	n2 := neo4j.Node{
		Id:     5,
		Labels: []string{"end"},
	}

	r1 := neo4j.Relationship{
		Id:      6,
		StartId: 4,
		EndId:   5,
		Type:    "someType",
	}

	// normal case (paths, nodes, and rels, but no nested results)
	pArr, rArr, nArr = traverseResultRecordValues([]interface{}{p1, n1, n2, r1})
	req.Equal(pArr[0], p1)
	req.Equal(rArr[0], r1)
	req.ElementsMatch(nArr, []interface{}{n1, n2})

	// case with nested nodes and rels
	pArr, rArr, nArr = traverseResultRecordValues([]interface{}{p1, []interface{}{n1, n2, r1}})
	req.Equal(pArr[0], p1)
	req.Equal(rArr[0], r1)
	req.ElementsMatch(nArr, []interface{}{n1, n2})
}

type TestStruct struct {
	Id         *int64
	UUID       string
	OtherField string
}

func toHashmapStructdecconf(m map[string]structDecoratorConfig) *hashmap.HashMap {
	h := &hashmap.HashMap{}

	for k, v := range m {
		h.Set(k, v)
	}

	return h
}

func TestConvertNodeToValue(t *testing.T) {
	req := require.New(t)
	gogm, err := getTestGogmWithDefaultStructs()
	req.Nil(err)
	req.NotNil(gogm)
	mappedTypes := toHashmapStructdecconf(map[string]structDecoratorConfig{
		"TestStruct": {
			Type: reflect.TypeOf(TestStruct{}),
			Fields: map[string]decoratorConfig{
				"UUID": {
					Type:       reflect.TypeOf(""),
					PrimaryKey: UUIDPrimaryKeyStrategy.StrategyName,
					Name:       "uuid",
				},
				"Id": {
					Type: reflect.TypeOf(int64(0)),
					Name: "id",
				},
				"OtherField": {
					Type: reflect.TypeOf(""),
					Name: "other_field",
				},
			},
			Label:    "TestStruct",
			IsVertex: true,
		},
	})
	gogm.mappedTypes = mappedTypes

	bn := neo4j.Node{
		Id: 10,
		Props: map[string]interface{}{
			"uuid":        "dadfasdfasdf",
			"other_field": "dafsdfasd",
		},
		Labels: []string{"TestStruct"},
	}

	val, err := convertNodeToValue(gogm, bn)
	req.Nil(err)
	req.NotNil(val)
	req.EqualValues(TestStruct{
		Id:         int64Ptr(10),
		UUID:       "dadfasdfasdf",
		OtherField: "dafsdfasd",
	}, val.Interface().(TestStruct))

	bn = neo4j.Node{
		Id: 10,
		Props: map[string]interface{}{
			"uuid":        "dadfasdfasdf",
			"other_field": "dafsdfasd",
			"t":           "dadfasdf",
		},
		Labels: []string{"TestStruct"},
	}

	var te structDecoratorConfig
	temp, ok := mappedTypes.Get("TestStruct")
	req.True(ok)

	te, ok = temp.(structDecoratorConfig)
	req.True(ok)

	te.Fields["tt"] = decoratorConfig{
		Type: reflect.TypeOf(""),
		Name: "test",
	}
	mappedTypes.Set("TestStruct", te)
	val, err = convertNodeToValue(gogm, bn)
	req.Nil(err)
	req.NotNil(val)
}

type tdString string
type tdInt int

type f struct {
	BaseUUIDNode
	Parents  []*f `gogm:"direction=outgoing;relationship=test"`
	Children []*f `gogm:"direction=incoming;relationship=test"`
}

type a struct {
	BaseUUIDNode
	PropTest0         map[string]interface{} `gogm:"properties;name=props0"`
	PropTest1         map[string]string      `gogm:"properties;name=props1"`
	PropsTest2        []string               `gogm:"properties;name=props2"`
	PropsTest3        []int                  `gogm:"properties;name=props3"`
	TestField         string                 `gogm:"name=test_field"`
	TestTypeDefString tdString               `gogm:"name=test_type_def_string"`
	TestTypeDefInt    tdInt                  `gogm:"name=test_type_def_int"`
	SingleA           *b                     `gogm:"direction=incoming;relationship=test_rel"`
	ManyA             []*b                   `gogm:"direction=incoming;relationship=testm2o"`
	MultiA            []*b                   `gogm:"direction=incoming;relationship=multib"`
	SingleSpecA       *c                     `gogm:"direction=outgoing;relationship=special_single"`
	MultiSpecA        []*c                   `gogm:"direction=outgoing;relationship=special_multi"`
	Created           time.Time              `gogm:"name=created"`
}

type b struct {
	BaseUUIDNode
	TestField  string    `gogm:"name=test_field"`
	TestTime   time.Time `gogm:"name=test_time"`
	Single     *a        `gogm:"direction=outgoing;relationship=test_rel"`
	ManyB      *a        `gogm:"direction=outgoing;relationship=testm2o"`
	Multi      []*a      `gogm:"direction=outgoing;relationship=multib"`
	SingleSpec *c        `gogm:"direction=incoming;relationship=special_single"`
	MultiSpec  []*c      `gogm:"direction=incoming;relationship=special_multi"`
}

type c struct {
	BaseUUIDNode
	Start *a
	End   *b
	Test  string `gogm:"name=test"`
}

func (c *c) GetStartNode() interface{} {
	return c.Start
}

func (c *c) GetStartNodeType() reflect.Type {
	return reflect.TypeOf(&a{})
}

func (c *c) SetStartNode(v interface{}) error {
	var ok bool
	c.Start, ok = v.(*a)
	if !ok {
		return errors.New("unable to cast to a")
	}

	return nil
}

func (c *c) GetEndNode() interface{} {
	return c.End
}

func (c *c) GetEndNodeType() reflect.Type {
	return reflect.TypeOf(&b{})
}

func (c *c) SetEndNode(v interface{}) error {
	var ok bool
	c.End, ok = v.(*b)
	if !ok {
		return errors.New("unable to cast to b")
	}

	return nil
}

type propsTest struct {
	Id         *int64                 `gogm:"pk=default"`
	UUID       string                 `gogm:"pk=UUID;name=uuid"`
	PropTest0  map[string]interface{} `gogm:"properties;name=props0"`
	PropTest1  map[string]string      `gogm:"properties;name=props1"`
	PropsTest2 []string               `gogm:"properties;name=props2"`
	PropsTest3 []int                  `gogm:"properties;name=props3"`
	PropsTest4 tdArr                  `gogm:"name=props4;properties"`
	PropsTest5 tdArrOfTd              `gogm:"name=props5;properties"`
	PropsTest6 tdMap                  `gogm:"name=props6;properties"`
	PropsTest7 tdMapTdSlice           `gogm:"name=props7;properties"`
	PropsTest8 tdMapTdSliceOfTd       `gogm:"name=props8;properties"`
}

func getTestGogmWithDefaultStructs() (*Gogm, error) {
	return getTestGogm(&a{}, &b{}, &c{}, &f{}, &propsTest{})
}

func getTestGogm(types ...interface{}) (*Gogm, error) {
	g := &Gogm{
		config: &Config{
			Logger:   GetDefaultLogger(),
			LogLevel: "DEBUG",
		},
		pkStrategy:       UUIDPrimaryKeyStrategy,
		logger:           GetDefaultLogger(),
		boltMajorVersion: 4,
		mappedTypes:      &hashmap.HashMap{},
		driver:           nil,
		mappedRelations:  &relationConfigs{},
		ogmTypes:         types,
		isNoOp:           false,
	}

	err := g.parseOgmTypes()
	if err != nil {
		return nil, err
	}

	return g, nil
}

func TestDecode(t *testing.T) {
	req := require.New(t)
	gogm, err := getTestGogmWithDefaultStructs()
	req.Nil(err)
	req.NotNil(gogm)

	var fNode f
	t1 := testResult{
		empty: true,
	}

	req.True(errors.Is(decode(gogm, &t1, &fNode), ErrNotFound))

	t1.empty = false

	req.Contains(decode(gogm, &t1, &fNode).Error(), "no primary nodes to return")
}

func TestDecode2(t *testing.T) {
	req := require.New(t)

	gogm, err := getTestGogmWithDefaultStructs()
	req.Nil(err)
	req.NotNil(gogm)

	//	req.EqualValues(3, mappedTypes.Len())
	vars10 := [][]interface{}{
		{
			neo4j.Path{
				Nodes: []neo4j.Node{
					{
						Labels: []string{"f"},
						Props: map[string]interface{}{
							"uuid": "0",
						},
						Id: 0,
					},
					{
						Labels: []string{"f"},
						Props: map[string]interface{}{
							"uuid": "1",
						},
						Id: 1,
					},
					{
						Labels: []string{"f"},
						Props: map[string]interface{}{
							"uuid": "2",
						},
						Id: 2,
					},
				},
				Relationships: []neo4j.Relationship{
					{
						Id:      3,
						StartId: 0,
						EndId:   1,
						Type:    "test",
						Props:   nil,
					},
					{
						Id:      4,
						StartId: 1,
						EndId:   2,
						Type:    "test",
						Props:   nil,
					},
				},
			},
		},
	}

	f0 := f{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "0",
			BaseNode: BaseNode{
				Id: int64Ptr(0),
			},
		},
	}

	f1 := f{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "1",
			BaseNode: BaseNode{
				Id: int64Ptr(1),
			},
		},
	}

	f2 := f{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "2",
			BaseNode: BaseNode{
				Id: int64Ptr(2),
			},
		},
	}

	f0.Parents = []*f{&f1}
	f1.Children = []*f{&f0}
	f1.Parents = []*f{&f2}
	f2.Children = []*f{&f1}

	var readin10 []*f
	req.Nil(decode(gogm, newMockResult(vars10), &readin10))
	req.True(len(readin10) == 3)
	for _, r := range readin10 {
		if *r.Id == 0 {
			req.True(len(r.Parents) == 1)
			req.True(r.LoadMap["Parents"].Ids[0] == 1)
			req.True(len(r.Children) == 0)
		} else if *r.Id == 1 {
			req.True(len(r.Parents) == 1)
			req.True(r.LoadMap["Parents"].Ids[0] == 2)
			req.True(len(r.Children) == 1)
			req.True(r.LoadMap["Children"].Ids[0] == 0)
		} else if *r.Id == 2 {
			req.True(len(r.Parents) == 0)
			req.True(len(r.Children) == 1)
			req.True(r.LoadMap["Children"].Ids[0] == 1)
		} else {
			t.FailNow()
		}
	}

	fTime := time.Now().UTC()

	vars := [][]interface{}{
		{
			neo4j.Path{
				Nodes: []neo4j.Node{
					{
						Labels: []string{"b"},
						Props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						Id: 2,
					},
					{
						Labels: []string{"a"},
						Props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						Id: 1,
					},
				},
				Relationships: []neo4j.Relationship{
					{
						Id:      1,
						StartId: 1,
						EndId:   2,
						Type:    "test_rel",
						Props:   nil,
					},
				},
			},
		},
	}

	var readin a

	comp := &a{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "dasdfasd",
			BaseNode: BaseNode{
				Id: int64Ptr(1),
			},
		},
		TestField:         "test",
		TestTypeDefInt:    600,
		TestTypeDefString: "TDs",
	}

	comp22 := &b{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "dasdfas",
			BaseNode: BaseNode{
				Id: int64Ptr(2),
			},
		},
		TestField: "test",
		TestTime:  fTime,
	}

	comp.SingleA = comp22
	comp22.Single = comp

	req.Nil(decode(gogm, newMockResult(vars), &readin))
	req.EqualValues(comp.TestField, readin.TestField)
	req.EqualValues(comp.UUID, readin.UUID)
	req.EqualValues(comp.Id, readin.Id)
	req.EqualValues(comp.SingleA.Id, readin.SingleA.Id)
	req.EqualValues(comp.SingleA.UUID, readin.SingleA.UUID)
	req.EqualValues(comp.SingleA.TestField, readin.SingleA.TestField)

	var readinSlicePtr []*a

	req.Nil(decode(gogm, newMockResult(vars), &readinSlicePtr))
	req.EqualValues(comp.TestField, readinSlicePtr[0].TestField)
	req.EqualValues(comp.UUID, readinSlicePtr[0].UUID)
	req.EqualValues(comp.Id, readinSlicePtr[0].Id)
	req.EqualValues(comp.SingleA.Id, readinSlicePtr[0].SingleA.Id)
	req.EqualValues(comp.SingleA.UUID, readinSlicePtr[0].SingleA.UUID)
	req.EqualValues(comp.SingleA.TestField, readinSlicePtr[0].SingleA.TestField)

	var readinSlice []a

	req.Nil(decode(gogm, newMockResult(vars), &readinSlice))
	req.EqualValues(comp.TestField, readinSlice[0].TestField)
	req.EqualValues(comp.UUID, readinSlice[0].UUID)
	req.EqualValues(comp.Id, readinSlice[0].Id)
	req.EqualValues(comp.SingleA.Id, readinSlice[0].SingleA.Id)
	req.EqualValues(comp.SingleA.UUID, readinSlice[0].SingleA.UUID)
	req.EqualValues(comp.SingleA.TestField, readinSlice[0].SingleA.TestField)

	vars2 := [][]interface{}{
		{
			neo4j.Path{
				Nodes: []neo4j.Node{
					{
						Labels: []string{"a"},
						Props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						Id: 1,
					},
					{
						Labels: []string{"b"},
						Props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						Id: 2,
					},
				},
				Relationships: []neo4j.Relationship{
					{
						Id:      5,
						StartId: 1,
						EndId:   2,
						Type:    "special_single",
						Props: map[string]interface{}{
							"test": "testing",
							"uuid": "asdfasdafsd",
						},
					},
				},
			},
		},
	}

	var readin2 a

	comp2 := &a{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "dasdfasd",
			BaseNode: BaseNode{
				Id: int64Ptr(1),
			},
		},
		TestField: "test",
	}

	b2 := &b{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "dasdfas",
			BaseNode: BaseNode{
				Id: int64Ptr(2),
			},
		},
		TestField: "test",
		TestTime:  fTime,
	}

	c1 := &c{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "asdfasdafsd",
			BaseNode: BaseNode{
				Id: int64Ptr(34),
			},
		},
		Start: comp2,
		End:   b2,
		Test:  "testing",
	}

	comp2.SingleSpecA = c1
	b2.SingleSpec = c1

	req.Nil(decode(gogm, newMockResult(vars2), &readin2))
	req.EqualValues(comp2.TestField, readin2.TestField)
	req.EqualValues(comp2.UUID, readin2.UUID)
	req.EqualValues(comp2.Id, readin2.Id)
	req.EqualValues(comp2.SingleSpecA.End.Id, readin2.SingleSpecA.End.Id)
	req.EqualValues(comp2.SingleSpecA.End.UUID, readin2.SingleSpecA.End.UUID)
	req.EqualValues(comp2.SingleSpecA.End.TestField, readin2.SingleSpecA.End.TestField)

	vars3 := [][]interface{}{
		{
			neo4j.Path{
				Nodes: []neo4j.Node{
					{
						Labels: []string{"a"},
						Props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						Id: 1,
					},
					{
						Labels: []string{"b"},
						Props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						Id: 2,
					},
				},
				Relationships: []neo4j.Relationship{
					{
						Id:      5,
						StartId: 1,
						EndId:   2,
						Type:    "multib",
						Props:   nil,
					},
				},
			},
		},
	}

	var readin3 a

	comp3 := a{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "dasdfasd",
			BaseNode: BaseNode{
				Id: int64Ptr(1),
			},
		},
		TestField: "test",
		MultiA: []*b{
			{
				TestField: "test",
				BaseUUIDNode: BaseUUIDNode{
					UUID: "dasdfas",
					BaseNode: BaseNode{
						Id: int64Ptr(2),
					},
				},
				TestTime: fTime,
			},
		},
	}

	req.Nil(decode(gogm, newMockResult(vars3), &readin3))
	req.EqualValues(comp3.TestField, readin3.TestField)
	req.EqualValues(comp3.UUID, readin3.UUID)
	req.EqualValues(comp3.Id, readin3.Id)
	req.NotNil(readin3.MultiA)
	req.EqualValues(1, len(readin3.MultiA))
	req.EqualValues(comp3.MultiA[0].Id, readin3.MultiA[0].Id)
	req.EqualValues(comp3.MultiA[0].UUID, readin3.MultiA[0].UUID)
	req.EqualValues(comp3.MultiA[0].TestField, readin3.MultiA[0].TestField)

	vars4 := [][]interface{}{
		{
			neo4j.Path{
				Nodes: []neo4j.Node{
					{
						Labels: []string{"a"},
						Props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						Id: 1,
					},
					{
						Labels: []string{"b"},
						Props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						Id: 2,
					},
				},
				Relationships: []neo4j.Relationship{
					{
						Id:      5,
						StartId: 1,
						EndId:   2,
						Type:    "special_multi",
						Props: map[string]interface{}{
							"test": "testing",
							"uuid": "asdfasdafsd",
						},
					},
				},
			},
		},
	}

	var readin4 b

	comp4 := &a{
		TestField: "test",
		BaseUUIDNode: BaseUUIDNode{
			UUID: "dasdfasd",
			BaseNode: BaseNode{
				Id: int64Ptr(1),
			},
		},
	}

	b3 := &b{
		TestField: "test",
		BaseUUIDNode: BaseUUIDNode{
			UUID: "dasdfas",
			BaseNode: BaseNode{
				Id: int64Ptr(2),
			},
		},
		TestTime: fTime,
	}

	c4 := c{
		BaseUUIDNode: BaseUUIDNode{
			UUID: "asdfasdafsd",
		},
		Start: comp4,
		End:   b3,
		Test:  "testing",
	}

	comp4.MultiSpecA = append(comp4.MultiSpecA, &c4)
	b3.MultiSpec = append(b3.MultiSpec, &c4)

	req.Nil(decode(gogm, newMockResult(vars4), &readin4))
	req.EqualValues(b3.TestField, readin4.TestField)
	req.EqualValues(b3.UUID, readin4.UUID)
	req.EqualValues(b3.Id, readin4.Id)
	req.NotNil(readin4.MultiSpec)
	req.EqualValues(1, len(readin4.MultiSpec))
	req.EqualValues(b3.MultiSpec[0].End.Id, readin4.MultiSpec[0].End.Id)
	req.EqualValues(b3.MultiSpec[0].End.UUID, readin4.MultiSpec[0].End.UUID)
	req.EqualValues(b3.MultiSpec[0].End.TestField, readin4.MultiSpec[0].End.TestField)

	var5uuid := "dasdfasdf"

	vars5 := [][]interface{}{
		{
			neo4j.Path{
				Nodes: []neo4j.Node{
					{
						Id:     1,
						Labels: []string{"propsTest"},
						Props: map[string]interface{}{
							"uuid":             var5uuid,
							"props0.test.test": "test",
							"props0.test2":     1,
							"props1.test":      "test",
							"props2":           []interface{}{"test"},
							"props3":           []interface{}{1, 2},
							"props4":           []interface{}{"test1", "test2"},
							"props5":           []interface{}{"tdtest"},
							"props6.test":      1,
							"props7.test":      []interface{}{"test1", "test2"},
							// "props8.test3":     []interface{}{"test1", "test"},
						},
					},
				},
			},
		},
	}

	var readin5 propsTest

	r := propsTest{
		Id:   int64Ptr(1),
		UUID: var5uuid,
		PropTest0: map[string]interface{}{
			"test.test": "test",
			"test2":     1,
		},
		PropTest1: map[string]string{
			"test": "test",
		},
		PropsTest2: []string{"test"},
		PropsTest3: []int{1, 2},
		PropsTest4: []string{"test1", "test2"},
		PropsTest5: []tdString{"tdtest"},
		PropsTest6: map[string]interface{}{
			"test": 1,
		},
		PropsTest7: map[string]tdArr{
			"test": []string{"test1", "test2"},
		},
		PropsTest8: map[string]tdArrOfTd{},
	}

	req.Nil(decode(gogm, newMockResult(vars5), &readin5))
	req.EqualValues(r.Id, readin5.Id)
	req.EqualValues(r.UUID, readin5.UUID)
	req.EqualValues(r.PropTest0["test"], readin5.PropTest0["test"])
	req.EqualValues(r.PropTest0["test2"], readin5.PropTest0["test2"])
	req.EqualValues(r.PropTest1["test"], readin5.PropTest1["test"])
	req.EqualValues(r.PropsTest2, readin5.PropsTest2)
	req.EqualValues(r.PropsTest3, readin5.PropsTest3)
	req.EqualValues(r.PropsTest4, readin5.PropsTest4)
	req.EqualValues(r.PropsTest5, readin5.PropsTest5)
	req.EqualValues(r.PropsTest6, readin5.PropsTest6)
	req.EqualValues(r.PropsTest7, readin5.PropsTest7)
	req.EqualValues(r.PropsTest8, readin5.PropsTest8)

	//multi single
	vars6 := [][]interface{}{
		{
			neo4j.Path{
				Nodes: []neo4j.Node{
					{
						Labels: []string{"b"},
						Props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						Id: 2,
					},
					{
						Labels: []string{"b"},
						Props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						Id: 3,
					},
				},
			},
		},
	}

	var readin6 []*b

	//b31 := &b{
	//	TestField: "test",
	//	UUID: "dasdfas",
	//	TestTime: fTime,
	//	Id: 2,
	//}
	//b32 := &b{
	//	TestField: "test",
	//	UUID: "dasdfas",
	//	TestTime: fTime,
	//	Id: 3,
	//}

	req.Nil(decode(gogm, newMockResult(vars6), &readin6))
	req.True(len(readin6) == 2)

	vars7 := [][]interface{}{
		{
			neo4j.Path{
				Nodes:         nil,
				Relationships: nil,
			},
		},
	}

	var readin7 []*b

	emptyErr := decode(gogm, newMockResult(vars7), &readin7)

	req.NotNil(emptyErr)
	req.True(errors.As(emptyErr, &ErrNotFound))
	req.True(len(readin7) == 0)

	vars8 := [][]interface{}{
		{
			neo4j.Path{
				Nodes:         nil,
				Relationships: nil,
			},
		},
	}

	var readin8 b

	emptyErr = decode(gogm, newMockResult(vars8), &readin8)

	req.NotNil(emptyErr)
	req.True(errors.As(emptyErr, &ErrNotFound))
	req.Zero(readin8)

	vars9 := [][]interface{}{
		{
			neo4j.Node{
				Id:     55,
				Labels: []string{"b"},
				Props: map[string]interface{}{
					"test_field": "test",
					"uuid":       "dasdfas",
					"test_time":  fTime,
				},
			},
		},
	}
	var readin9 b
	req.Nil(decode(gogm, newMockResult(vars9), &readin9))
	req.Equal("test", readin9.TestField)
	req.Equal(int64(55), *readin9.Id)
	req.Equal("dasdfas", readin9.UUID)

	// decode should be able to handle queries that return nested lists of paths, relationships, and nodes
	decodeResultNested := [][]interface{}{
		{
			neo4j.Node{
				Id:     18,
				Labels: []string{"a"},
				Props: map[string]interface{}{
					"uuid": "2588baca-7561-43f8-9ddb-9c7aecf87284",
				},
			},
		},
		{
			[]interface{}{
				[]interface{}{
					[]interface{}{
						[]interface{}{
							neo4j.Relationship{
								Id:      0,
								StartId: 19,
								EndId:   18,
								Type:    "testm2o",
							},
							neo4j.Node{
								Id:     19,
								Labels: []string{"b"},
								Props: map[string]interface{}{
									"test_fielda": "1234",
									"uuid":        "b6d8c2ab-06c2-43d0-8452-89d6c4ec5d40",
								},
							},
						},
					},
					[]interface{}{},
					[]interface{}{
						[]interface{}{
							neo4j.Relationship{
								Id:      1,
								StartId: 18,
								EndId:   19,
								Type:    "special_single",
								Props: map[string]interface{}{
									"test": "testing",
								},
							},
							neo4j.Node{
								Id:     19,
								Labels: []string{"b"},
								Props: map[string]interface{}{
									"test_fielda": "1234",
									"uuid":        "b6d8c2ab-06c2-43d0-8452-89d6c4ec5d40",
								},
							},
						},
					},
				},
				[]interface{}{},
				[]interface{}{},
			},
		},
	}
	var readinNested a
	req.Nil(decode(gogm, newMockResult(decodeResultNested), &readinNested))
	req.Equal("2588baca-7561-43f8-9ddb-9c7aecf87284", readinNested.UUID)
	req.Len(readinNested.ManyA, 1)
	req.Equal("b6d8c2ab-06c2-43d0-8452-89d6c4ec5d40", readinNested.ManyA[0].UUID)
	req.Equal(readinNested.ManyA[0], readinNested.SingleSpecA.End, "Two rels should have the same node instance")
}
