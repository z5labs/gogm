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
	"errors"
	"github.com/cornelk/hashmap"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

type TestStruct struct {
	Id         int64
	UUID       string
	OtherField string
}

func toHashmap(m map[string]interface{}) *hashmap.HashMap {
	h := &hashmap.HashMap{}

	for k, v := range m {
		h.Set(k, v)
	}

	return h
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

	mappedTypes = toHashmapStructdecconf(map[string]structDecoratorConfig{
		"TestStruct": {
			Type: reflect.TypeOf(TestStruct{}),
			Fields: map[string]decoratorConfig{
				"UUID": {
					Type:       reflect.TypeOf(""),
					PrimaryKey: true,
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

	bn := testNode{
		id: 10,
		props: map[string]interface{}{
			"uuid":        "dadfasdfasdf",
			"other_field": "dafsdfasd",
		},
		labels: []string{"TestStruct"},
	}

	val, err := convertNodeToValue(bn)
	req.Nil(err)
	req.NotNil(val)
	req.EqualValues(TestStruct{
		Id:         10,
		UUID:       "dadfasdfasdf",
		OtherField: "dafsdfasd",
	}, val.Interface().(TestStruct))

	bn = testNode{
		id: 10,
		props: map[string]interface{}{
			"uuid":        "dadfasdfasdf",
			"other_field": "dafsdfasd",
			"t":           "dadfasdf",
		},
		labels: []string{"TestStruct"},
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
	val, err = convertNodeToValue(bn)
	req.Nil(err)
	req.NotNil(val)
}

type tdString string
type tdInt int

type f struct {
	BaseNode
	Parents  []*f `gogm:"direction=outgoing;relationship=test"`
	Children []*f `gogm:"direction=incoming;relationship=test"`
}

type a struct {
	BaseNode
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
	BaseNode
	TestField  string    `gogm:"name=test_field"`
	TestTime   time.Time `gogm:"name=test_time"`
	Single     *a        `gogm:"direction=outgoing;relationship=test_rel"`
	ManyB      *a        `gogm:"direction=outgoing;relationship=testm2o"`
	Multi      []*a      `gogm:"direction=outgoing;relationship=multib"`
	SingleSpec *c        `gogm:"direction=incoming;relationship=special_single"`
	MultiSpec  []*c      `gogm:"direction=incoming;relationship=special_multi"`
}

type c struct {
	BaseNode
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
	Id         int64                  `gogm:"name=id"`
	UUID       string                 `gogm:"pk;name=uuid"`
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

func TestDecode(t *testing.T) {

	req := require.New(t)
	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}, &f{}, &propsTest{}))

	var fNode f
	t1 := testResult{
		empty: true,
	}

	req.True(errors.Is(decode(&t1, &fNode), ErrNotFound))

	t1.empty = false

	req.Nil(decode(&t1, &fNode))
}

func TestInnerDecode(t *testing.T) {
	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}, &f{}, &propsTest{}))

	//	req.EqualValues(3, mappedTypes.Len())

	vars10 := [][]interface{}{
		{
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
		},
	}

	f0 := f{
		BaseNode: BaseNode{
			Id:   0,
			UUID: "0",
		},
	}

	f1 := f{
		BaseNode: BaseNode{
			Id:   1,
			UUID: "1",
		},
	}

	f2 := f{
		BaseNode: BaseNode{
			Id:   2,
			UUID: "2",
		},
	}

	f0.Parents = []*f{&f1}
	f1.Children = []*f{&f0}
	f1.Parents = []*f{&f2}
	f2.Children = []*f{&f1}

	var readin10 []*f
	req.Nil(innerDecode(vars10, &readin10))
	req.True(len(readin10) == 3)
	for _, r := range readin10 {
		if r.Id == 0 {
			req.True(len(r.Parents) == 1)
			req.True(r.LoadMap["Parents"].Ids[0] == 1)
			req.True(len(r.Children) == 0)
		} else if r.Id == 1 {
			req.True(len(r.Parents) == 1)
			req.True(r.LoadMap["Parents"].Ids[0] == 2)
			req.True(len(r.Children) == 1)
			req.True(r.LoadMap["Children"].Ids[0] == 0)
		} else if r.Id == 2 {
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
			testPath{
				nodes: []*testNode{
					{
						labels: []string{"b"},
						props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						id: 2,
					},
					{
						labels: []string{"a"},
						props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						id: 1,
					},
				},
				relNodes: []*testRelationship{
					{
						id:      1,
						startId: 1,
						endId:   2,
						_type:   "test_rel",
						props:   nil,
					},
				},
			},
		},
	}

	var readin a

	comp := &a{
		BaseNode: BaseNode{
			Id:   1,
			UUID: "dasdfasd",
		},
		TestField:         "test",
		TestTypeDefInt:    600,
		TestTypeDefString: "TDs",
	}

	comp22 := &b{
		BaseNode: BaseNode{
			Id:   2,
			UUID: "dasdfas",
		},
		TestField: "test",
		TestTime:  fTime,
	}

	comp.SingleA = comp22
	comp22.Single = comp

	req.Nil(innerDecode(vars, &readin))
	req.EqualValues(comp.TestField, readin.TestField)
	req.EqualValues(comp.UUID, readin.UUID)
	req.EqualValues(comp.Id, readin.Id)
	req.EqualValues(comp.SingleA.Id, readin.SingleA.Id)
	req.EqualValues(comp.SingleA.UUID, readin.SingleA.UUID)
	req.EqualValues(comp.SingleA.TestField, readin.SingleA.TestField)

	var readinSlicePtr []*a

	req.Nil(innerDecode(vars, &readinSlicePtr))
	req.EqualValues(comp.TestField, readinSlicePtr[0].TestField)
	req.EqualValues(comp.UUID, readinSlicePtr[0].UUID)
	req.EqualValues(comp.Id, readinSlicePtr[0].Id)
	req.EqualValues(comp.SingleA.Id, readinSlicePtr[0].SingleA.Id)
	req.EqualValues(comp.SingleA.UUID, readinSlicePtr[0].SingleA.UUID)
	req.EqualValues(comp.SingleA.TestField, readinSlicePtr[0].SingleA.TestField)

	var readinSlice []a

	req.Nil(innerDecode(vars, &readinSlice))
	req.EqualValues(comp.TestField, readinSlice[0].TestField)
	req.EqualValues(comp.UUID, readinSlice[0].UUID)
	req.EqualValues(comp.Id, readinSlice[0].Id)
	req.EqualValues(comp.SingleA.Id, readinSlice[0].SingleA.Id)
	req.EqualValues(comp.SingleA.UUID, readinSlice[0].SingleA.UUID)
	req.EqualValues(comp.SingleA.TestField, readinSlice[0].SingleA.TestField)

	vars2 := [][]interface{}{
		{
			testPath{
				nodes: []*testNode{
					{
						labels: []string{"a"},
						props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						id: 1,
					},
					{
						labels: []string{"b"},
						props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						id: 2,
					},
				},
				relNodes: []*testRelationship{
					{
						id:      5,
						startId: 1,
						endId:   2,
						_type:   "special_single",
						props: map[string]interface{}{
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
		BaseNode: BaseNode{
			Id:   1,
			UUID: "dasdfasd",
		},
		TestField: "test",
	}

	b2 := &b{
		BaseNode: BaseNode{
			Id:   2,
			UUID: "dasdfas",
		},
		TestField: "test",
		TestTime:  fTime,
	}

	c1 := &c{
		BaseNode: BaseNode{
			Id:   34,
			UUID: "asdfasdafsd",
		},
		Start: comp2,
		End:   b2,
		Test:  "testing",
	}

	comp2.SingleSpecA = c1
	b2.SingleSpec = c1

	req.Nil(innerDecode(vars2, &readin2))
	req.EqualValues(comp2.TestField, readin2.TestField)
	req.EqualValues(comp2.UUID, readin2.UUID)
	req.EqualValues(comp2.Id, readin2.Id)
	req.EqualValues(comp2.SingleSpecA.End.Id, readin2.SingleSpecA.End.Id)
	req.EqualValues(comp2.SingleSpecA.End.UUID, readin2.SingleSpecA.End.UUID)
	req.EqualValues(comp2.SingleSpecA.End.TestField, readin2.SingleSpecA.End.TestField)

	vars3 := [][]interface{}{
		{
			testPath{
				nodes: []*testNode{
					{
						labels: []string{"a"},
						props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						id: 1,
					},
					{
						labels: []string{"b"},
						props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						id: 2,
					},
				},
				relNodes: []*testRelationship{
					{
						id:      5,
						startId: 1,
						endId:   2,
						_type:   "multib",
						props:   nil,
					},
				},
			},
		},
	}

	var readin3 a

	comp3 := a{
		BaseNode: BaseNode{
			Id:   1,
			UUID: "dasdfasd",
		},
		TestField: "test",
		MultiA: []*b{
			{
				TestField: "test",
				BaseNode: BaseNode{
					Id:   2,
					UUID: "dasdfas",
				},
				TestTime: fTime,
			},
		},
	}

	req.Nil(innerDecode(vars3, &readin3))
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
			testPath{
				nodes: []*testNode{
					{
						labels: []string{"a"},
						props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						id: 1,
					},
					{
						labels: []string{"b"},
						props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						id: 2,
					},
				},
				relNodes: []*testRelationship{
					{
						id:      5,
						startId: 1,
						endId:   2,
						_type:   "special_multi",
						props: map[string]interface{}{
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
		BaseNode: BaseNode{
			Id:   1,
			UUID: "dasdfasd",
		},
	}

	b3 := &b{
		TestField: "test",
		BaseNode: BaseNode{
			Id:   2,
			UUID: "dasdfas",
		},
		TestTime: fTime,
	}

	c4 := c{
		BaseNode: BaseNode{
			UUID: "asdfasdafsd",
		},
		Start: comp4,
		End:   b3,
		Test:  "testing",
	}

	comp4.MultiSpecA = append(comp4.MultiSpecA, &c4)
	b3.MultiSpec = append(b3.MultiSpec, &c4)

	req.Nil(innerDecode(vars4, &readin4))
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
			testPath{
				nodes: []*testNode{
					{
						id:     1,
						labels: []string{"propsTest"},
						props: map[string]interface{}{
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
		Id:   1,
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

	req.Nil(innerDecode(vars5, &readin5))
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
			testPath{
				nodes: []*testNode{
					{
						labels: []string{"b"},
						props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						id: 2,
					},
					{
						labels: []string{"b"},
						props: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime,
						},
						id: 3,
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

	req.Nil(innerDecode(vars6, &readin6))
	req.True(len(readin6) == 2)

	vars7 := [][]interface{}{
		{
			testPath{
				nodes:    nil,
				relNodes: nil,
				indexes:  nil,
			},
		},
	}

	var readin7 []*b

	emptyErr := innerDecode(vars7, &readin7)

	req.NotNil(emptyErr)
	req.True(errors.As(emptyErr, &ErrNotFound))
	req.True(len(readin7) == 0)

	vars8 := [][]interface{}{
		{
			testPath{
				nodes:    nil,
				relNodes: nil,
				indexes:  nil,
			},
		},
	}

	var readin8 b

	emptyErr = innerDecode(vars8, &readin8)

	req.NotNil(emptyErr)
	req.True(errors.As(emptyErr, &ErrNotFound))
	req.Zero(readin8)

	vars9 := [][]interface{}{
		{
			testNode{
				id:     55,
				labels: []string{"b"},
				props: map[string]interface{}{
					"test_field": "test",
					"uuid":       "dasdfas",
					"test_time":  fTime,
				},
			},
		},
	}
	var readin9 b
	req.Nil(innerDecode(vars9, &readin9))
	req.Equal("test", readin9.TestField)
	req.Equal(int64(55), readin9.Id)
	req.Equal("dasdfas", readin9.UUID)

}
