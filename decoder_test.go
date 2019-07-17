package gogm

import (
	"errors"
	"github.com/cornelk/hashmap"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/structures/graph"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

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
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`
	TestField string `gogm:"name=test_field"`
	Single *b `gogm:"relationship=test_rel"`
	Multi []b `gogm:"relationship=multib"`
	SingleSpec *c `gogm:"relationship=special_single"`
	MultiSpec []c `gogm:"relationship=special_multi"`
}

type b struct{
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`
	TestField string `gogm:"name=test_field"`
	Single *a `gogm:"relationship=test_rel"`
	Multi []a `gogm:"relationship=multib"`
	SingleSpec *c `gogm:"relationship=special_single"`
	MultiSpec []c `gogm:"relationship=special_multi"`
}

type c struct{
	Start a
	End b
	Test string
}

func (c *c) GetStartNode() interface{} {
	return c.Start
}

func (c *c) SetStartNode(v interface{}) error {
	var ok bool
	c.Start, ok = v.(a)
	if !ok{
		return errors.New("unable to cast to a")
	}

	return nil
}

func (c *c) GetEndNode() interface{} {
	return c.End
}

func (c *c) SetEndNode(v interface{}) error {
	var ok bool
	c.End, ok = v.(b)
	if !ok{
		return errors.New("unable to cast to b")
	}

	return nil
}

func TestDecoder(t *testing.T){
	req := require.New(t)

	req.Nil(setupInit(true, nil, &a{}, &b{}, &c{}))

	req.EqualValues(3, mappedTypes.Len())

	vars := [][]interface{}{
		{
			neoEdgeConfig{
				Type: "test_rel",
				StartNodeType: "a",
				StartNodeId: 1,
				EndNodeType: "b",
				EndNodeId: 2,
			},
		},
		{
			graph.Node{
				Labels: []string{"b"},
				Properties: map[string]interface{}{
					"test_field": "test",
					"uuid": "dasdfas",
				},
				NodeIdentity: 2,
			},
		},
		{
			graph.Node{
				Labels: []string{"a"},
				Properties: map[string]interface{}{
					"test_field": "test",
					"uuid": "dasdfasd",
				},
				NodeIdentity: 1,
			},
		},
	}

	var readin a

	comp := a{
		TestField: "test",
		Id: 1,
		UUID: "dasdfasd",
		Single: &b{
			TestField: "test",
			UUID: "dasdfas",
			Id: 2,
		},
	}

	req.Nil(decode(vars, &readin))
	req.EqualValues(comp.TestField, readin.TestField)
	req.EqualValues(comp.UUID, readin.UUID)
	req.EqualValues(comp.Id, readin.Id)
	req.EqualValues(comp.Single.Id, readin.Single.Id)
	req.EqualValues(comp.Single.UUID, readin.Single.UUID)
	req.EqualValues(comp.Single.TestField, readin.Single.TestField)
}