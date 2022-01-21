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
	"log"
	"reflect"
	"testing"

	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/stretchr/testify/require"
)

func TestDecoratorConfig_Validate(t *testing.T) {
	req := require.New(t)
	gogm, err := getTestGogmWithDefaultStructs()
	req.Nil(err)
	req.NotNil(gogm)

	tests := []struct {
		Name       string
		Decorator  decoratorConfig
		ShouldPass bool
	}{
		{
			Name: "valid",
			Decorator: decoratorConfig{
				Properties: true,
				Type:       reflect.TypeOf(map[string]interface{}{}),
				Name:       "test",
			},
			ShouldPass: true,
		},
		{
			Name: "valid relationship",
			Decorator: decoratorConfig{
				FieldName:    "test_rel",
				Name:         "test_rel",
				Relationship: "rel",
				Type:         reflect.TypeOf([]interface{}{}),
			},
			ShouldPass: true,
		},
		{
			Name: "valid relationship with direction",
			Decorator: decoratorConfig{
				FieldName:    "test_rel",
				Name:         "test_rel",
				Relationship: "rel",
				Direction:    dsl.DirectionIncoming,
				Type:         reflect.TypeOf([]interface{}{}),
			},
			ShouldPass: true,
		},
		{
			Name: "valid pk (uuid)",
			Decorator: decoratorConfig{
				Name:       "uuid",
				Type:       reflect.TypeOf(""),
				PrimaryKey: UUIDPrimaryKeyStrategy.StrategyName,
			},
			ShouldPass: true,
		},
		{
			Name: "valid index",
			Decorator: decoratorConfig{
				Name:  "test_index",
				Type:  reflect.TypeOf(""),
				Index: true,
			},
			ShouldPass: true,
		},
		{
			Name: "valid unique",
			Decorator: decoratorConfig{
				Name:   "test_unique",
				Type:   reflect.TypeOf(""),
				Unique: true,
			},
			ShouldPass: true,
		},
		{
			Name: "valid plain",
			Decorator: decoratorConfig{
				Name: "test",
				Type: reflect.TypeOf(""),
			},
			ShouldPass: true,
		},
		{
			Name: "valid field pointer",
			Decorator: decoratorConfig{
				Name: "test",
				Type: reflect.PtrTo(reflect.TypeOf("")),
			},
			ShouldPass: true,
		},
		{
			Name: "invalid with wrong sig",
			Decorator: decoratorConfig{
				Properties: true,
				Type:       reflect.MapOf(reflect.TypeOf(decoratorConfig{}), reflect.TypeOf("")),
				Name:       "test",
			},
			ShouldPass: false,
		},
		{
			Name: "invalid prop extra decorator",
			Decorator: decoratorConfig{
				Properties: true,
				Type:       reflect.TypeOf(map[string]interface{}{}),
				Name:       "test",
				Unique:     true,
			},
			ShouldPass: false,
		},
		{
			Name: "invalid props decorator not specified",
			Decorator: decoratorConfig{
				Type: reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(map[string]interface{}{})),
				Name: "test",
			},
			ShouldPass: false,
		},
		{
			Name: "invalid relationship",
			Decorator: decoratorConfig{
				Relationship: "test",
				Name:         "test",
				Type:         reflect.TypeOf(""),
			},
			ShouldPass: false,
		},
		{
			Name: "invalid direction not defined",
			Decorator: decoratorConfig{
				Direction: dsl.DirectionOutgoing,
				Name:      "asdfa",
				Type:      reflect.TypeOf([]interface{}{}),
			},
			ShouldPass: false,
		},
		{
			Name: "invalid pk ptr str",
			Decorator: decoratorConfig{
				Name:       "uuid",
				PrimaryKey: UUIDPrimaryKeyStrategy.StrategyName,
				Type:       reflect.PtrTo(reflect.TypeOf("")),
			},
			ShouldPass: false,
		},
	}

	for _, test := range tests {
		t.Log("running test -", test.Name)
		err = test.Decorator.Validate(gogm)
		if test.ShouldPass {
			req.Nil(err)
		} else {
			req.NotNil(err)
		}
	}
}

func TestStructDecoratorConfig_Validate(t *testing.T) {
	req := require.New(t)

	tests := []struct {
		Name       string
		Decorator  structDecoratorConfig
		ShouldPass bool
	}{
		{
			Name: "nil fields",
			Decorator: structDecoratorConfig{
				Fields:   nil,
				IsVertex: true,
			},
			ShouldPass: false,
		},
		{
			Name: "valid pk",
			Decorator: structDecoratorConfig{
				Fields: map[string]decoratorConfig{
					"uuid": {
						PrimaryKey: UUIDPrimaryKeyStrategy.StrategyName,
						Name:       "uuid",
						Type:       reflect.TypeOf(""),
					},
				},
				IsVertex: true,
			},
			ShouldPass: true,
		},
		{
			Name: "valid uuid with id",
			Decorator: structDecoratorConfig{
				Fields: map[string]decoratorConfig{
					"uuid": {
						PrimaryKey: UUIDPrimaryKeyStrategy.StrategyName,
						Name:       "uuid",
						Type:       reflect.TypeOf(""),
					},
					"id": {
						PrimaryKey: DefaultPrimaryKeyStrategy.StrategyName,
						Name:       "id",
						Type:       reflect.TypeOf(int64(1)),
					},
				},
				IsVertex: true,
			},
			ShouldPass: true,
		},
		{
			Name: "invalid relations",
			Decorator: structDecoratorConfig{
				Fields: map[string]decoratorConfig{
					"uuid": {
						PrimaryKey: UUIDPrimaryKeyStrategy.StrategyName,
						Name:       "uuid",
						Type:       reflect.TypeOf(""),
					},
					"rel_test": {
						Relationship: "test",
						Name:         "test",
						Type:         reflect.TypeOf([]interface{}{}),
					},
				},
				IsVertex: false,
			},
			ShouldPass: false,
		},
	}

	for _, test := range tests {
		t.Log("running test", test.Name)
		err := test.Decorator.Validate()
		if test.ShouldPass {
			req.Nil(err)
		} else {
			req.NotNil(err)
		}
	}
}

func TestNewDecoratorConfig(t *testing.T) {
	req := require.New(t)
	testGogm, err := getTestGogmWithDefaultStructs()
	req.Nil(err)
	req.NotNil(testGogm)
	var compare *decoratorConfig

	decName := "name=id"
	decNameStruct := decoratorConfig{
		Name:       "id",
		Type:       reflect.TypeOf(int64(1)),
		ParentType: reflect.TypeOf(a{}),
	}

	compare, err = newDecoratorConfig(testGogm, decName, "", reflect.TypeOf(int64(0)), reflect.TypeOf(a{}))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decNameStruct, *compare)

	decUUID := "pk=UUID"
	decUUIDStruct := decoratorConfig{
		Name:       "uuid",
		FieldName:  "UUID",
		PrimaryKey: UUIDPrimaryKeyStrategy.StrategyName,
		Type:       reflect.TypeOf(""),
		ParentType: reflect.TypeOf(a{}),
	}

	compare, err = newDecoratorConfig(testGogm, decUUID, "", reflect.TypeOf(""), reflect.TypeOf(a{}))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decUUIDStruct, *compare)

	decIndexField := "index;name=index_field"
	decIndexFieldStruct := decoratorConfig{
		Index:      true,
		Name:       "index_field",
		Type:       reflect.TypeOf(""),
		ParentType: reflect.TypeOf(a{}),
	}

	compare, err = newDecoratorConfig(testGogm, decIndexField, "", reflect.TypeOf(""), reflect.TypeOf(a{}))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decIndexFieldStruct, *compare)

	decUniqueField := "unique;name=unique_name"
	decUniqueFieldStruct := decoratorConfig{
		Unique:     true,
		Name:       "unique_name",
		Type:       reflect.TypeOf(""),
		ParentType: reflect.TypeOf(a{}),
	}

	compare, err = newDecoratorConfig(testGogm, decUniqueField, "", reflect.TypeOf(""), reflect.TypeOf(a{}))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decUniqueFieldStruct, *compare)

	decOne2One := "relationship=one2one;direction=incoming"
	decOne2OneStruct := decoratorConfig{
		FieldName:    "test_name",
		Name:         "test_name",
		Relationship: "one2one",
		Direction:    dsl.DirectionIncoming,
		Type:         reflect.TypeOf(a{}),
		ParentType:   reflect.TypeOf(a{}),
	}

	compare, err = newDecoratorConfig(testGogm, decOne2One, "test_name", reflect.TypeOf(a{}), reflect.TypeOf(a{}))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decOne2OneStruct, *compare)

	decProps := "properties;name=test"
	decPropsStruct := decoratorConfig{
		Properties: true,
		Name:       "test",
		Type:       reflect.TypeOf(map[string]interface{}{}),
		PropConfig: &propConfig{
			IsMap:      true,
			IsMapSlice: false,
			SubType:    emptyInterfaceType,
		},
		ParentType: reflect.TypeOf(a{}),
	}

	compare, err = newDecoratorConfig(testGogm, decProps, "", reflect.TypeOf(map[string]interface{}{}), reflect.TypeOf(a{}))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decPropsStruct, *compare)

	decIgnore := "-"

	compare, err = newDecoratorConfig(testGogm, decIgnore, "", reflect.TypeOf(int64(0)), reflect.TypeOf(a{}))
	req.Nil(err)
	req.NotNil(compare)
	req.True(compare.Ignore)

	decInvalidRelName := "relationship=A_REL;direction=incoming;name=ISHOULDNTBEHERE"

	compare, err = newDecoratorConfig(testGogm, decInvalidRelName, "TestFieldName", reflect.TypeOf(a{}), reflect.TypeOf(a{}))
	req.NotNil(err)
	req.Nil(compare)

	decInvalidIgnore := "-;index"

	compare, err = newDecoratorConfig(testGogm, decInvalidIgnore, "", reflect.TypeOf(int64(0)), reflect.TypeOf(a{}))
	req.NotNil(err)
	req.Nil(compare)

	// Both relationship on self
	compare, err = newDecoratorConfig(testGogm, "relationship=self2self;direction=both", "test_name", reflect.TypeOf(a{}), reflect.TypeOf(a{}))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decoratorConfig{
		ParentType:   reflect.TypeOf(a{}),
		FieldName:    "test_name",
		Name:         "test_name",
		Relationship: "self2self",
		Direction:    dsl.DirectionBoth,
		Type:         reflect.TypeOf(a{}),
	}, *compare)
}

//structs with decorators for testing

type embedTest struct {
	Id   int64  `gogm:"name=id"`
	UUID string `gogm:"pk=UUID;name=uuid"`
}

type validStruct struct {
	embedTest
	IndexField             string                 `gogm:"index;name=index_field"`
	UniqueField            int                    `gogm:"unique;name=unique_field"`
	OneToOne               *validStruct           `gogm:"relationship=one2one;direction=incoming"`
	ManyToOne              []*a                   `gogm:"relationship=many2one;direction=outgoing"`
	SpecialOne             *c                     `gogm:"relationship=specC;direction=outgoing"`
	SpecialMany            []*c                   `gogm:"relationship=manyC;direction=outgoing"`
	PropsMapInterface      map[string]interface{} `gogm:"properties;name=props1"`
	PropsMapPrimitive      map[string]int         `gogm:"properties;name=props2"`
	PropsMapSlicePrimitive map[string][]int       `gogm:"properties;name=props3"`
	PropsSliceInterface    []string               `gogm:"properties;name=props4"`
	PropsPrimitive         []int                  `gogm:"properties;name=props5"`
	IgnoreMe               int                    `gogm:"-"`
}

func (v *validStruct) GetId() int64 {
	panic("implement me")
}

func (v *validStruct) SetId(i int64) {
	panic("implement me")
}

func (v *validStruct) GetUUID() string {
	panic("implement me")
}

func (v *validStruct) SetUUID(u string) {
	panic("implement me")
}

func (v *validStruct) GetLabels() []string {
	return []string{"validStruct"}
}

//issue is that it has no id defined
type mostlyValidStruct struct {
	IndexField  string `gogm:"index;name=index_field"`
	UniqueField int    `gogm:"unique;name=unique_field"`
}

func (m *mostlyValidStruct) GetLabels() []string {
	return []string{"mostlyValidStruct"}
}

//nothing defined
type emptyStruct struct{}

func (e *emptyStruct) GetLabels() []string {
	return []string{"emptyStruct"}
}

//has a valid field but also has a messed up one
type invalidStructDecorator struct {
	Id   int64  `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`

	MessedUp int `gogm:"sdfasdfasdfa"`
}

func (i *invalidStructDecorator) GetLabels() []string {
	return []string{"invalidStructDecorator"}
}

type invalidStructProperties struct {
	Id   int64  `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`

	Props  map[string]*validStruct   `gogm:"properties;name=props"`
	Props1 map[string][]*validStruct `gogm:"properties;name=props1"`
	Props2 []*validStruct            `gogm:"properties;name=props2"`
}

func (i *invalidStructProperties) GetLabels() []string {
	return []string{"invalidStructProperties"}
}

type invalidEdge struct {
	UUID string      `gogm:"pk;name=uuid"`
	Rel  interface{} `gogm:"relationship=should_fail"`
}

func (i *invalidEdge) GetLabels() []string {
	return []string{"invalidEdge"}
}

type invalidNameStruct struct {
	Id   int64  `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`
	// relationship cannot be named
	InvalidRel *invalidNameStruct `gogm:"relationship=ONE_TO_ONE;direction=incoming;name=AAAAAA"`
}

func (i *invalidNameStruct) GetLabels() []string {
	return []string{"invalidNameStruct"}
}

type invalidIgnoreStruct struct {
	Id   int64  `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`
	// should fail because ignore struct has additional tags
	IgnoreMe int64 `gogm:"-;unique"`
}

func (i *invalidIgnoreStruct) GetLabels() []string {
	return []string{"invalidIgnoreStruct"}
}

type uuidlessEdge struct {
	SomeProps map[string]interface{} `gogm:"name=props;properties"`
}

func (i *uuidlessEdge) GetLabels() []string {
	return []string{"uuidlessEdge"}
}

func TestGetStructDecoratorConfig_RelDirectionBoth(t *testing.T) {
	req := require.New(t)

	// Rel within single type
	type TypeWithRelWithinType struct {
		BaseUUIDNode
		Bidirectional []*TypeWithRelWithinType `gogm:"relationship=BIDIRECTIONAL;direction=both"`
	}

	testGogm, err := getTestGogm(&TypeWithRelWithinType{})
	req.Nil(err)
	req.NotNil(testGogm)

	type UnrequitingRelType struct {
		BaseUUIDNode
		// relationship not returned :(
	}

	// Invalid both config
	type UnrequitedRelType struct {
		BaseUUIDNode
		Unrequited []*UnrequitingRelType `gogm:"relationship=BIDIRECTIONAL;direction=both"`
	}

	testGogm, err = getTestGogm(&UnrequitedRelType{}, &UnrequitingRelType{})
	req.NotNil(err)
	req.Nil(testGogm)
}

func TestGetStructDecoratorConfig(t *testing.T) {
	req := require.New(t)
	testGogm, err := getTestGogmWithDefaultStructs()
	req.Nil(err)
	req.NotNil(testGogm)
	mappedRelations := &relationConfigs{}

	conf, err := getStructDecoratorConfig(testGogm, &validStruct{}, mappedRelations)
	req.Nil(err)
	req.NotNil(conf)
	checkObj := structDecoratorConfig{
		IsVertex: true,
		Type:     reflect.TypeOf(validStruct{}),
		Label:    "validStruct",
		Fields: map[string]decoratorConfig{
			"Id": {
				Name:       "id",
				FieldName:  "Id",
				Type:       reflect.TypeOf(int64(0)),
				ParentType: reflect.TypeOf(validStruct{}),
			},
			"UUID": {
				Name:       "uuid",
				FieldName:  "UUID",
				PrimaryKey: UUIDPrimaryKeyStrategy.StrategyName,
				Type:       reflect.TypeOf(""),
				ParentType: reflect.TypeOf(validStruct{}),
			},
			"IndexField": {
				FieldName:  "IndexField",
				Name:       "index_field",
				Index:      true,
				Type:       reflect.TypeOf(""),
				ParentType: reflect.TypeOf(validStruct{}),
			},
			"UniqueField": {
				FieldName:  "UniqueField",
				Unique:     true,
				Name:       "unique_field",
				Type:       reflect.TypeOf(int(1)),
				ParentType: reflect.TypeOf(validStruct{}),
			},
			"OneToOne": {
				FieldName:    "OneToOne",
				Name:         "OneToOne",
				Relationship: "one2one",
				Direction:    dsl.DirectionIncoming,
				Type:         reflect.TypeOf(&validStruct{}),
				ParentType:   reflect.TypeOf(validStruct{}),
			},
			"SpecialOne": {
				FieldName:    "SpecialOne",
				Name:         "SpecialOne",
				Relationship: "specC",
				Direction:    dsl.DirectionOutgoing,
				UsesEdgeNode: true,
				Type:         reflect.TypeOf(&c{}),
				ParentType:   reflect.TypeOf(validStruct{}),
			},
			"SpecialMany": {
				FieldName:        "SpecialMany",
				Name:             "SpecialMany",
				Relationship:     "manyC",
				Direction:        dsl.DirectionOutgoing,
				UsesEdgeNode:     true,
				ManyRelationship: true,
				Type:             reflect.TypeOf([]*c{}),
				ParentType:       reflect.TypeOf(validStruct{}),
			},
			"ManyToOne": {
				FieldName:        "ManyToOne",
				Name:             "ManyToOne",
				Relationship:     "many2one",
				Direction:        dsl.DirectionOutgoing,
				ManyRelationship: true,
				Type:             reflect.TypeOf([]*a{}),
				ParentType:       reflect.TypeOf(validStruct{}),
			},
			"PropsMapInterface": {
				FieldName:  "PropsMapInterface",
				Properties: true,
				Name:       "props1",
				Type:       reflect.TypeOf(map[string]interface{}{}),
				PropConfig: &propConfig{
					IsMap:      true,
					IsMapSlice: false,
					SubType:    emptyInterfaceType,
				},
				ParentType: reflect.TypeOf(validStruct{}),
			},
			"PropsMapPrimitive": {
				FieldName:  "PropsMapPrimitive",
				Properties: true,
				Name:       "props2",
				Type:       reflect.TypeOf(map[string]int{}),
				PropConfig: &propConfig{
					IsMap:      true,
					IsMapSlice: false,
					SubType:    reflect.TypeOf(int(0)),
				},
				ParentType: reflect.TypeOf(validStruct{}),
			},
			"PropsMapSlicePrimitive": {
				FieldName:  "PropsMapSlicePrimitive",
				Properties: true,
				Name:       "props3",
				Type:       reflect.TypeOf(map[string][]int{}),
				PropConfig: &propConfig{
					IsMap:        true,
					IsMapSlice:   true,
					SubType:      reflect.TypeOf(int(0)),
					MapSliceType: reflect.TypeOf([]int{}),
				},
				ParentType: reflect.TypeOf(validStruct{}),
			},
			"PropsSliceInterface": {
				FieldName:  "PropsSliceInterface",
				Properties: true,
				Name:       "props4",
				Type:       reflect.TypeOf([]string{}),
				PropConfig: &propConfig{
					IsMap:      false,
					IsMapSlice: false,
					SubType:    reflect.TypeOf(""),
				},
				ParentType: reflect.TypeOf(validStruct{}),
			},
			"PropsPrimitive": {
				FieldName:  "PropsPrimitive",
				Properties: true,
				Name:       "props5",
				Type:       reflect.TypeOf([]int{}),
				PropConfig: &propConfig{
					IsMap:      false,
					IsMapSlice: false,
					SubType:    reflect.TypeOf(int(0)),
				},
				ParentType: reflect.TypeOf(validStruct{}),
			},
			"IgnoreMe": {
				FieldName:  "IgnoreMe",
				Name:       "IgnoreMe",
				Ignore:     true,
				Type:       reflect.TypeOf(int(1)),
				ParentType: reflect.TypeOf(validStruct{}),
			},
		},
	}
	req.EqualValues(checkObj, *conf)

	conf, err = getStructDecoratorConfig(testGogm, &mostlyValidStruct{}, mappedRelations)
	req.NotNil(err)
	req.Nil(conf)

	conf, err = getStructDecoratorConfig(testGogm, &emptyStruct{}, mappedRelations)
	req.NotNil(err)
	req.Nil(conf)

	conf, err = getStructDecoratorConfig(testGogm, &invalidStructDecorator{}, mappedRelations)
	req.NotNil(err)
	req.Nil(conf)

	conf, err = getStructDecoratorConfig(testGogm, &invalidStructProperties{}, mappedRelations)
	req.NotNil(err)
	req.Nil(conf)

	conf, err = getStructDecoratorConfig(testGogm, &invalidEdge{}, mappedRelations)
	req.NotNil(err)
	req.Nil(conf)

	conf, err = getStructDecoratorConfig(testGogm, &invalidNameStruct{}, mappedRelations)
	req.NotNil(err)
	req.Nil(conf)

	conf, err = getStructDecoratorConfig(testGogm, &invalidIgnoreStruct{}, mappedRelations)
	req.NotNil(err)
	req.Nil(conf)

	conf, err = getStructDecoratorConfig(testGogm, &uuidlessEdge{}, mappedRelations)
	log.Println("ERR::", err)
	req.NotNil(err)
	req.Nil(conf)
}
