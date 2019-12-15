// Copyright (c) 2019 MindStand Technologies, Inc
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
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestSetUuidIfNeeded(t *testing.T) {
	val := &a{}

	_, _, _, err := handleNodeState(nil, "UUID")
	require.NotNil(t, err)

	v := reflect.ValueOf(val)
	isNew, _, _, err := handleNodeState(&v, "UUID")
	require.Nil(t, err)
	require.True(t, isNew)

	val.UUID = "dasdfasd"

	v = reflect.ValueOf(val)
	isNew, _, _, err = handleNodeState(&v, "UUID")
	require.Nil(t, err)
	require.True(t, isNew)

	val.UUID = "dasdfasd"
	val.LoadMap = map[string]*RelationConfig{}

	v = reflect.ValueOf(val)
	isNew, _, _, err = handleNodeState(&v, "UUID")
	require.Nil(t, err)
	require.True(t, isNew)

	val.UUID = "dasdfasd"
	val.LoadMap = nil

	v = reflect.ValueOf(val)
	isNew, _, _, err = handleNodeState(&v, "UUID")
	require.Nil(t, err)
	require.True(t, isNew)

	val.UUID = "dasdfasd"
	val.LoadMap = map[string]*RelationConfig{
		"dasdfasd": {
			Ids:          []int64{69},
			RelationType: Single,
		},
	}

	v = reflect.ValueOf(val)
	isNew, _, _, err = handleNodeState(&v, "UUID")
	require.Nil(t, err)
	require.False(t, isNew)

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
		BaseNode: BaseNode{
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
