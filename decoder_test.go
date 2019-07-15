package gogm

import (
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

func TestConvertNodeToValue(t *testing.T){

	req := require.New(t)

	mappedTypes = map[string]structDecoratorConfig{
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
	}

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
	mappedTypes["TestStruct"].Fields["tt"] = decoratorConfig{
		Type: reflect.TypeOf(""),
		Name: "test",
	}
	val, err = convertNodeToValue(bn)
	req.NotNil(err)
	req.Nil(val)
}