package gogm

import (
	"errors"
	"github.com/cornelk/hashmap"
	"github.com/mindstand/golang-neo4j-bolt-driver/structures/graph"
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

func TestDecode(t *testing.T){
	if !testing.Short(){
		t.Skip()
		return
	}

	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}))

	req.EqualValues(3, mappedTypes.Len())

	query := `
		MATCH (n:a)
		WITH n
		MATCH (n)-[e*0..2]-(m)
		RETURN DISTINCT
			collect(extract(n in e | {StartNodeId: ID(startnode(n)), StartNodeType: labels(startnode(n))[0], EndNodeId: ID(endnode(n)), EndNodeType: labels(endnode(n))[0], Obj: n, Type: type(n)})) as Edges,
			collect(DISTINCT m) as Ends,
			collect(DISTINCT n) as Starts
	`

	err := dsl.Init(&dsl.ConnectionConfig{
		Username: "neo4j",
		Password: "password",
		Host: "0.0.0.0",
		Port: 7687,
		PoolSize: 15,
	})
	require.Nil(t, err)

	sess := dsl.NewSession()

	rows, err := sess.QueryReadOnly().Cypher(query).Query(nil)
	require.Nil(t, err)
	require.NotNil(t, rows)

	var stuff a
	require.Nil(t, decodeNeoRows(rows, &stuff))
	t.Log(stuff.Id)
	t.Log(stuff.UUID)
	t.Log(stuff.MultiSpecA[0].End.Id)
	req.NotEqual(0, stuff.Id)
	req.True(len(stuff.MultiSpecA) > 0)
}

type TestStruct struct {
	Id int64
	UUID string
	OtherField string

}

func toHashmap(m map[string]interface{}) *hashmap.HashMap{
	h := &hashmap.HashMap{}

	for k, v := range m{
		h.Set(k, v)
	}

	return h
}

func toHashmapStructdecconf(m map[string]structDecoratorConfig) *hashmap.HashMap{
	h := &hashmap.HashMap{}

	for k, v := range m{
		h.Set(k, v)
	}

	return h
}

func TestConvertNodeToValue(t *testing.T){

	req := require.New(t)

	mappedTypes = toHashmapStructdecconf(map[string]structDecoratorConfig{
		"TestStruct": {
			Type: reflect.TypeOf(TestStruct{}),
			Fields: map[string]decoratorConfig{
				"UUID": {
					Type: reflect.TypeOf(""),
					PrimaryKey: true,
					Name: "uuid",
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
			Label: "TestStruct",
			IsVertex: true,
		},
	})

	bn := graph.Node{
		NodeIdentity: 10,
		Properties: map[string]interface{}{
			"uuid": "dadfasdfasdf",
			"other_field": "dafsdfasd",
		},
		Labels: []string{"TestStruct"},
	}

	val, err := convertNodeToValue(bn)
	req.Nil(err)
	req.NotNil(val)
	req.EqualValues(TestStruct{
		Id: 10,
		UUID: "dadfasdfasdf",
		OtherField: "dafsdfasd",
	}, val.Interface().(TestStruct))

	bn = graph.Node{
		NodeIdentity: 10,
		Properties: map[string]interface{}{
			"uuid": "dadfasdfasdf",
			"other_field": "dafsdfasd",
			"t": "dadfasdf",
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
	req.NotNil(err)
	req.Nil(val)
}

type a struct{
	Id          int64  `gogm:"name=id"`
	UUID        string `gogm:"pk;name=uuid"`
	TestField   string `gogm:"name=test_field"`
	SingleA     *b     `gogm:"direction=incoming;relationship=test_rel"`
	MultiA      []b    `gogm:"direction=incoming;relationship=multib"`
	SingleSpecA *c     `gogm:"direction=outgoing;relationship=special_single"`
	MultiSpecA  []c    `gogm:"direction=outgoing;relationship=special_multi"`
}

type b struct{
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`
	TestField string `gogm:"name=test_field"`
	TestTime time.Time `gogm:"time;name=test_time"`
	Single *a `gogm:"direction=outgoing;relationship=test_rel"`
	Multi []a `gogm:"direction=outgoing;relationship=multib"`
	SingleSpec *c `gogm:"direction=incoming;relationship=special_single"`
	MultiSpec []c `gogm:"direction=incoming;relationship=special_multi"`
}

type c struct{
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`
	Start *a
	End *b
	Test string `gogm:"name=test"`
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
	if !ok{
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
	if !ok{
		return errors.New("unable to cast to b")
	}

	return nil
}

type propsTest struct {
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`
	Props map[string]interface{} `gogm:"name=props;properties"`
}

func TestDecoder(t *testing.T){
	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}, &propsTest{}))

//	req.EqualValues(3, mappedTypes.Len())

	Type := "Type"
	StartNodeType := "StartNodeType"
	StartNodeId := "StartNodeId"
	EndNodeType := "EndNodeType"
	EndNodeId := "EndNodeId"
	Obj := "Obj"

	fTime := time.Now().UTC()

	vars := [][]interface{}{
		{
			[]interface{}{
				[]interface{}{neoEdgeConfig{}},
				[]interface{}{
					map[string]interface{}{
						Type: "test_rel",
						StartNodeType: "b",
						StartNodeId: int64(2),
						EndNodeType: "a",
						EndNodeId: int64(1),
					},
				},
			},
			[]interface{}{
				graph.Node{
					Labels: []string{"a"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfasd",
					},
					NodeIdentity: 1,
				},
				graph.Node{
					Labels: []string{"b"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfas",
						"test_time": fTime.Format(time.RFC3339),
					},
					NodeIdentity: 2,
				},
			},
			[]interface{}{
				graph.Node{
					Labels: []string{"a"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfasd",
					},
					NodeIdentity: 1,
				},
			},
		},
	}

	var readin a

	comp := &a{
		TestField: "test",
		Id: 1,
		UUID: "dasdfasd",
	}

	comp22 := &b{
		TestField: "test",
		UUID: "dasdfas",
		TestTime: fTime,
		Id: 2,
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
			[]interface{}{
				[]interface{}{neoEdgeConfig{}},
				[]interface{}{
					map[string]interface{}{
						Type: "special_single",
						StartNodeType: "a",
						StartNodeId: int64(1),
						EndNodeType: "b",
						EndNodeId: int64(2),
						Obj: map[string]interface{}{
							"test": "testing",
							"uuid": "asdfasdafsd",
						},
					},
				},
			},
			[]interface{}{
				graph.Node{
					Labels: []string{"b"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfas",
						"test_time": fTime.Format(time.RFC3339),
					},
					NodeIdentity: 2,
				},
				graph.Node{
					Labels: []string{"a"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfasd",
					},
					NodeIdentity: 1,
				},
			},
			[]interface{}{
				graph.Node{
					Labels: []string{"a"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfasd",
					},
					NodeIdentity: 1,
				},
			},
		},
	}

	var readin2 a

	comp2 := &a{
		TestField: "test",
		Id:        1,
		UUID:      "dasdfasd",
	}

	b2 := &b{
		TestField: "test",
		UUID: "dasdfas",
		TestTime: fTime,
		Id: 2,
	}

	c1 := &c{
		UUID: "asdfasdafsd",
		Id: 420,
		Start: comp2,
		End: b2,
		Test: "testing",
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
			[]interface{}{
				[]interface{}{neoEdgeConfig{}},
				[]interface{}{
					map[string]interface{}{
						Type: "multib",
						StartNodeType: "a",
						StartNodeId: int64(1),
						EndNodeType: "b",
						EndNodeId: int64(2),
					},
				},
			},
			[]interface{}{
				graph.Node{
					Labels: []string{"b"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfas",
						"test_time": fTime.Format(time.RFC3339),
					},
					NodeIdentity: 2,
				},
				graph.Node{
					Labels: []string{"a"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfasd",
					},
					NodeIdentity: 1,
				},
			},
			[]interface{}{
				graph.Node{
					Labels: []string{"a"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfasd",
					},
					NodeIdentity: 1,
				},
			},
		},
	}

	var readin3 a

	comp3 := a{
		TestField: "test",
		Id: 1,
		UUID: "dasdfasd",
		MultiA: []b{
			{
				TestField: "test",
				UUID: "dasdfas",
				Id: 2,
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
			[]interface{}{
				[]interface{}{
					neoEdgeConfig{},
				},
				[]interface{}{
					map[string]interface{}{
						Type: "special_multi",
						Obj: map[string]interface{}{
							"test": "testing",
							"uuid": "asdfasdafsd",
						},
						StartNodeType: "a",
						StartNodeId: int64(1),
						EndNodeType: "b",
						EndNodeId: int64(2),
					},
				},
			},
			[]interface{}{
				graph.Node{
					Labels: []string{"b"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfas",
						"test_time": fTime.Format(time.RFC3339),
					},
					NodeIdentity: 2,
				},
				graph.Node{
					Labels: []string{"a"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfasd",
					},
					NodeIdentity: 1,
				},
			},
			[]interface{}{
				graph.Node{
					Labels: []string{"b"},
					Properties: map[string]interface{}{
						"test_field": "test",
						"uuid": "dasdfas",
						"test_time": fTime.Format(time.RFC3339),
					},
					NodeIdentity: 2,
				},
			},
		},
	}

	var readin4 b

	comp4 := &a{
		TestField: "test",
		Id:        1,
		UUID:      "dasdfasd",
	}

	b3 := &b{
		TestField: "test",
		UUID: "dasdfas",
		TestTime: fTime,
		Id: 2,
	}

	c4 := c{
		UUID: "asdfasdafsd",
		Start: comp4,
		End: b3,
		Test: "testing",
	}

	comp4.MultiSpecA = append(comp4.MultiSpecA, c4)
	b3.MultiSpec = append(b3.MultiSpec, c4)

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

	vars5 := [][]interface{} {
		[]interface{}{
			[]interface{}{},
			[]interface{}{
				graph.Node{
					NodeIdentity: 1,
					Labels: []string{ "propsTest" },
					Properties: map[string]interface{}{
						"uuid": var5uuid,
						"props.test": "test",
						"props.test2": "test2",
						"props.test3": "test3",
					},
				},
			},
			[]interface{}{
				graph.Node{
					NodeIdentity: 1,
					Labels: []string{ "propsTest" },
					Properties: map[string]interface{}{
						"uuid": var5uuid,
						"props.test": "test",
						"props.test2": "test2",
						"props.test3": "test3",
					},
				},
			},
		},
	}

	var readin5 propsTest

	r := propsTest{
		Id:    1,
		UUID:  var5uuid,
		Props: map[string]interface{}{
			"test": "test",
			"test2": "test2",
			"test3": "test3",
		},
	}

	req.Nil(decode(vars5, &readin5))
	req.EqualValues(r.Id, readin5.Id)
	req.EqualValues(r.UUID, readin5.UUID)
	req.EqualValues(r.Props["test"], readin5.Props["test"])
	req.EqualValues(r.Props["test2"], readin5.Props["test2"])
	req.EqualValues(r.Props["test3"], readin5.Props["test3"])
}