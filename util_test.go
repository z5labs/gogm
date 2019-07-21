package gogm

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestSetUuidIfNeeded(t *testing.T){
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

func TestGetTypeName(t *testing.T){
	val := &a{}

	name, err := getTypeName(reflect.TypeOf(val))
	require.Nil(t, err)
	require.EqualValues(t, "a", name)

	val1 := []a{}

	name, err = getTypeName(reflect.TypeOf(val1))
	require.Nil(t, err)
	require.EqualValues(t, "a", name)
}

func TestToCypherParamsMap(t *testing.T){
	val := a{
		Id: 0,
		UUID: "testuuid",
		TestField: "testvalue",
	}

	config, _, err := getStructDecoratorConfig(&val)
	require.Nil(t, err)

	params, err := toCypherParamsMap(reflect.ValueOf(val), *config)
	require.Nil(t, err)
	require.EqualValues(t, map[string]interface{}{
		"uuid": "testuuid",
		"test_field": "testvalue",
	}, params)
}