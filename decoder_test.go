package gogm

import (
	"errors"
	"github.com/cornelk/hashmap"
	"github.com/mindstand/golang-neo4j-bolt-driver/structures/graph"
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

	bn := graph.Node{
		NodeIdentity: 10,
		Properties: map[string]interface{}{
			"uuid":        "dadfasdfasdf",
			"other_field": "dafsdfasd",
		},
		Labels: []string{"TestStruct"},
	}

	val, err := convertNodeToValue(bn)
	req.Nil(err)
	req.NotNil(val)
	req.EqualValues(TestStruct{
		Id:         10,
		UUID:       "dadfasdfasdf",
		OtherField: "dafsdfasd",
	}, val.Interface().(TestStruct))

	bn = graph.Node{
		NodeIdentity: 10,
		Properties: map[string]interface{}{
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
	TestField         string   `gogm:"name=test_field"`
	TestTypeDefString tdString `gogm:"name=test_type_def_string"`
	TestTypeDefInt    tdInt    `gogm:"name=test_type_def_int"`
	SingleA           *b       `gogm:"direction=incoming;relationship=test_rel"`
	ManyA             []*b     `gogm:"direction=incoming;relationship=testm2o"`
	MultiA            []*b     `gogm:"direction=incoming;relationship=multib"`
	SingleSpecA       *c       `gogm:"direction=outgoing;relationship=special_single"`
	MultiSpecA        []*c     `gogm:"direction=outgoing;relationship=special_multi"`
}

type b struct {
	BaseNode
	TestField  string    `gogm:"name=test_field"`
	TestTime   time.Time `gogm:"time;name=test_time"`
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
	Id    int64                  `gogm:"name=id"`
	UUID  string                 `gogm:"pk;name=uuid"`
	Props map[string]interface{} `gogm:"name=props;properties"`
}

func TestDecoder(t *testing.T) {
	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}, &f{}, &propsTest{}))

	//	req.EqualValues(3, mappedTypes.Len())

	vars10 := [][]interface{}{
		{
			graph.Path{
				Nodes: []graph.Node{
					graph.Node{
						Labels: []string{"f"},
						Properties: map[string]interface{}{
							"uuid": "0",
						},
						NodeIdentity: 0,
					},
					graph.Node{
						Labels: []string{"f"},
						Properties: map[string]interface{}{
							"uuid": "1",
						},
						NodeIdentity: 1,
					},
					graph.Node{
						Labels: []string{"f"},
						Properties: map[string]interface{}{
							"uuid": "2",
						},
						NodeIdentity: 2,
					},
				},
				Relationships: []graph.UnboundRelationship{
					graph.UnboundRelationship{
						RelIdentity: 3,
						Type:        "test",
					},
					graph.UnboundRelationship{
						RelIdentity: 4,
						Type:        "test",
					},
				},
				Sequence: []int{1, 1 /*#*/, 2, 2},
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
	req.Nil(decode(vars10, &readin10))
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
			graph.Path{
				Nodes: []graph.Node{
					graph.Node{
						Labels: []string{"b"},
						Properties: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime.Format(time.RFC3339),
						},
						NodeIdentity: 2,
					},
					graph.Node{
						Labels: []string{"a"},
						Properties: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						NodeIdentity: 1,
					},
				},
				Relationships: []graph.UnboundRelationship{
					graph.UnboundRelationship{
						RelIdentity: 1,
						Type:        "test_rel",
						Properties:  nil,
					},
				},
				Sequence: []int{1, 1},
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

	req.Nil(decode(vars, &readin))
	req.EqualValues(comp.TestField, readin.TestField)
	req.EqualValues(comp.UUID, readin.UUID)
	req.EqualValues(comp.Id, readin.Id)
	req.EqualValues(comp.SingleA.Id, readin.SingleA.Id)
	req.EqualValues(comp.SingleA.UUID, readin.SingleA.UUID)
	req.EqualValues(comp.SingleA.TestField, readin.SingleA.TestField)

	var readinSlicePtr []*a

	req.Nil(decode(vars, &readinSlicePtr))
	req.EqualValues(comp.TestField, readinSlicePtr[0].TestField)
	req.EqualValues(comp.UUID, readinSlicePtr[0].UUID)
	req.EqualValues(comp.Id, readinSlicePtr[0].Id)
	req.EqualValues(comp.SingleA.Id, readinSlicePtr[0].SingleA.Id)
	req.EqualValues(comp.SingleA.UUID, readinSlicePtr[0].SingleA.UUID)
	req.EqualValues(comp.SingleA.TestField, readinSlicePtr[0].SingleA.TestField)

	var readinSlice []a

	req.Nil(decode(vars, &readinSlice))
	req.EqualValues(comp.TestField, readinSlice[0].TestField)
	req.EqualValues(comp.UUID, readinSlice[0].UUID)
	req.EqualValues(comp.Id, readinSlice[0].Id)
	req.EqualValues(comp.SingleA.Id, readinSlice[0].SingleA.Id)
	req.EqualValues(comp.SingleA.UUID, readinSlice[0].SingleA.UUID)
	req.EqualValues(comp.SingleA.TestField, readinSlice[0].SingleA.TestField)

	vars2 := [][]interface{}{
		{
			graph.Path{
				Nodes: []graph.Node{
					graph.Node{
						Labels: []string{"a"},
						Properties: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						NodeIdentity: 1,
					},
					graph.Node{
						Labels: []string{"b"},
						Properties: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime.Format(time.RFC3339),
						},
						NodeIdentity: 2,
					},
				},
				Relationships: []graph.UnboundRelationship{
					graph.UnboundRelationship{
						RelIdentity: 5,
						Type:        "special_single",
						Properties: map[string]interface{}{
							"test": "testing",
							"uuid": "asdfasdafsd",
						},
					},
				},
				Sequence: []int{1, 1},
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

	req.Nil(decode(vars2, &readin2))
	req.EqualValues(comp2.TestField, readin2.TestField)
	req.EqualValues(comp2.UUID, readin2.UUID)
	req.EqualValues(comp2.Id, readin2.Id)
	req.EqualValues(comp2.SingleSpecA.End.Id, readin2.SingleSpecA.End.Id)
	req.EqualValues(comp2.SingleSpecA.End.UUID, readin2.SingleSpecA.End.UUID)
	req.EqualValues(comp2.SingleSpecA.End.TestField, readin2.SingleSpecA.End.TestField)

	vars3 := [][]interface{}{
		{
			graph.Path{
				Nodes: []graph.Node{
					graph.Node{
						Labels: []string{"a"},
						Properties: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						NodeIdentity: 1,
					},
					graph.Node{
						Labels: []string{"b"},
						Properties: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime.Format(time.RFC3339),
						},
						NodeIdentity: 2,
					},
				},
				Relationships: []graph.UnboundRelationship{
					graph.UnboundRelationship{
						RelIdentity: 5,
						Type:        "multib",
						Properties:  nil,
					},
				},
				Sequence: []int{1, 1},
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

	req.Nil(decode(vars3, &readin3))
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
			graph.Path{
				Nodes: []graph.Node{
					graph.Node{
						Labels: []string{"a"},
						Properties: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfasd",
						},
						NodeIdentity: 1,
					},
					graph.Node{
						Labels: []string{"b"},
						Properties: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime.Format(time.RFC3339),
						},
						NodeIdentity: 2,
					},
				},
				Relationships: []graph.UnboundRelationship{
					graph.UnboundRelationship{
						RelIdentity: 5,
						Type:        "special_multi",
						Properties: map[string]interface{}{
							"test": "testing",
							"uuid": "asdfasdafsd",
						},
					},
				},
				Sequence: []int{1, 1},
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

	req.Nil(decode(vars4, &readin4))
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
			graph.Path{
				Nodes: []graph.Node{
					graph.Node{
						NodeIdentity: 1,
						Labels:       []string{"propsTest"},
						Properties: map[string]interface{}{
							"uuid":            var5uuid,
							"props.test.test": "test",
							"props.test2":     "test2",
							"props.test3":     "test3",
						},
					},
				},
				Relationships: nil,
				Sequence:      nil,
			},
		},
	}

	var readin5 propsTest

	r := propsTest{
		Id:   1,
		UUID: var5uuid,
		Props: map[string]interface{}{
			"test.test": "test",
			"test2":     "test2",
			"test3":     "test3",
		},
	}

	req.Nil(decode(vars5, &readin5))
	req.EqualValues(r.Id, readin5.Id)
	req.EqualValues(r.UUID, readin5.UUID)
	req.EqualValues(r.Props["test"], readin5.Props["test"])
	req.EqualValues(r.Props["test2"], readin5.Props["test2"])
	req.EqualValues(r.Props["test3"], readin5.Props["test3"])

	//multi single
	vars6 := [][]interface{}{
		{
			graph.Path{
				Nodes: []graph.Node{
					graph.Node{
						Labels: []string{"b"},
						Properties: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime.Format(time.RFC3339),
						},
						NodeIdentity: 2,
					},
					graph.Node{
						Labels: []string{"b"},
						Properties: map[string]interface{}{
							"test_field": "test",
							"uuid":       "dasdfas",
							"test_time":  fTime.Format(time.RFC3339),
						},
						NodeIdentity: 3,
					},
				},
				Relationships: nil,
				Sequence:      nil,
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

	req.Nil(decode(vars6, &readin6))
	req.True(len(readin6) == 2)

	vars7 := [][]interface{}{
		{
			graph.Path{
				Nodes:         nil,
				Relationships: nil,
				Sequence:      nil,
			},
		},
	}

	var readin7 []*b

	emptyErr := decode(vars7, &readin7)

	req.NotNil(emptyErr)
	req.True(errors.As(emptyErr, &ErrNotFound))
	req.True(len(readin7) == 0)

	vars8 := [][]interface{}{
		{
			graph.Path{
				Nodes:         nil,
				Relationships: nil,
				Sequence:      nil,
			},
		},
	}

	var readin8 b

	emptyErr = decode(vars8, &readin8)

	req.NotNil(emptyErr)
	req.True(errors.As(emptyErr, &ErrNotFound))
	req.Zero(readin8)

	vars9 := [][]interface{}{
		{
			graph.Node{
				NodeIdentity: 55,
				Labels:       []string{"b"},
				Properties: map[string]interface{}{
					"test_field": "test",
					"uuid":       "dasdfas",
					"test_time":  fTime.Format(time.RFC3339),
				},
			},
		},
	}
	var readin9 b
	req.Nil(decode(vars9, &readin9))
	req.Equal("test", readin9.TestField)
	req.Equal(int64(55), readin9.Id)
	req.Equal("dasdfas", readin9.UUID)

}
