package gogm

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestSetUuidIfNeeded(t *testing.T) {
	val := &a{}

	_, _, err := setUuidIfNeeded(nil, "UUID")
	require.NotNil(t, err)

	v := reflect.ValueOf(val)
	isCreated, _, err := setUuidIfNeeded(&v, "UUID")
	require.Nil(t, err)
	require.True(t, isCreated)

	val.UUID = "dasdfasd"

	v = reflect.ValueOf(val)
	isCreated, _, err = setUuidIfNeeded(&v, "UUID")
	require.Nil(t, err)
	require.False(t, isCreated)
}

func TestGetTypeName(t *testing.T) {
	val := &a{}

	name, err := getTypeName(reflect.TypeOf(val))
	require.Nil(t, err)
	require.EqualValues(t, "a", name)

	val1 := []a{}

	name, err = getTypeName(reflect.TypeOf(val1))
	require.Nil(t, err)
	require.EqualValues(t, "a", name)
}

func TestToCypherParamsMap(t *testing.T) {
	val := a{
		embedTest: embedTest{
			Id:   0,
			UUID: "testuuid",
		},
		TestField: "testvalue",
	}

	config, err := getStructDecoratorConfig(&val, mappedRelations)
	require.Nil(t, err)

	params, err := toCypherParamsMap(reflect.ValueOf(val), *config)
	require.Nil(t, err)
	require.EqualValues(t, map[string]interface{}{
		"uuid":                 "testuuid",
		"test_type_def_int":    0,
		"test_type_def_string": "",
		"test_field":           "testvalue",
	}, params)

	p := propsTest{
		Id:   1,
		UUID: "testuuid",
		Props: map[string]interface{}{
			"test": "testvalue",
		},
	}

	config, err = getStructDecoratorConfig(&p, mappedRelations)
	require.Nil(t, err)

	params, err = toCypherParamsMap(reflect.ValueOf(&p), *config)
	require.Nil(t, err)
	require.EqualValues(t, map[string]interface{}{
		"uuid":       "testuuid",
		"props.test": "testvalue",
	}, params)
}

type TypeDefString string
type TypeDefInt int
type TypeDefInt64 int64

func TestTypeDefStuff(t *testing.T) {
	stringType := reflect.TypeOf("")
	//tdStringType := reflect.TypeOf(TypeDefString(""))
	//intType := reflect.TypeOf(0)
	//int64Type := reflect.TypeOf(int64(0))

	//t.Log(tdStringType.Kind().String() == tdStringType.Name())

	td := TypeDefString("test")
	stringTd := "test"

	te := reflect.ValueOf(td).Convert(stringType).Interface()

	comp, ok := te.(string)
	if !ok {
		t.FailNow()
		return
	}

	t.Log(comp == stringTd)
}
