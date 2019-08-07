package gogm

import (
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestDecoratorConfig_Validate(t *testing.T) {
	req := require.New(t)

	validProps := decoratorConfig{
		Properties: true,
		Type:       reflect.TypeOf(map[string]interface{}{}),
		Name:       "test",
	}

	req.Nil(validProps.Validate())

	validRelationship := decoratorConfig{
		FieldName:    "test_rel",
		Name:         "test_rel",
		Relationship: "rel",
		Type:         reflect.TypeOf([]interface{}{}),
	}

	req.Nil(validRelationship.Validate())

	validRelationshipWithDirection := decoratorConfig{
		FieldName:    "test_rel",
		Name:         "test_rel",
		Relationship: "rel",
		Direction:    dsl.Incoming,
		Type:         reflect.TypeOf([]interface{}{}),
	}

	req.Nil(validRelationshipWithDirection.Validate())

	validStringPk := decoratorConfig{
		Name:       "uuid",
		Type:       reflect.TypeOf(""),
		PrimaryKey: true,
	}

	req.Nil(validStringPk.Validate())

	validInt64Pk := decoratorConfig{
		PrimaryKey: true,
		Type:       reflect.TypeOf(int64(1)),
		Name:       "id",
	}

	req.Nil(validInt64Pk.Validate())

	validFieldIndex := decoratorConfig{
		Name:  "test_index",
		Type:  reflect.TypeOf(""),
		Index: true,
	}

	req.Nil(validFieldIndex.Validate())

	validFieldUnique := decoratorConfig{
		Name:   "test_unique",
		Type:   reflect.TypeOf(""),
		Unique: true,
	}

	req.Nil(validFieldUnique.Validate())

	validPlainField := decoratorConfig{
		Name: "test",
		Type: reflect.TypeOf(""),
	}

	req.Nil(validPlainField.Validate())

	validFieldPtr := decoratorConfig{
		Name: "test",
		Type: reflect.PtrTo(reflect.TypeOf("")),
	}

	req.Nil(validFieldPtr.Validate())

	strType := reflect.TypeOf("")

	invalidPropsWrongSig := decoratorConfig{
		Properties: true,
		Type:       reflect.MapOf(strType, strType),
		Name:       "test",
	}

	req.NotNil(invalidPropsWrongSig)

	invalidPropsExtraDecorators := decoratorConfig{
		Properties: true,
		Type:       reflect.TypeOf(map[string]interface{}{}),
		Name:       "test",
		Unique:     true,
	}

	req.NotNil(invalidPropsExtraDecorators.Validate())

	invalidPropsDecoratorNotSpecified := decoratorConfig{
		Type: reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(map[string]interface{}{})),
		Name: "test",
	}

	req.NotNil(invalidPropsDecoratorNotSpecified.Validate())

	invalidRelationshipType := decoratorConfig{
		Relationship: "test",
		Name:         "test",
		Type:         strType,
	}

	req.NotNil(invalidRelationshipType.Validate())

	invalidDirectionDefinedNotRel := decoratorConfig{
		Direction: dsl.Outgoing,
		Name:      "asdfa",
		Type:      reflect.TypeOf([]interface{}{}),
	}

	req.NotNil(invalidDirectionDefinedNotRel.Validate())

	invalidPkPtrStr := decoratorConfig{
		Name:       "uuid",
		PrimaryKey: true,
		Type:       reflect.PtrTo(strType),
	}

	req.NotNil(invalidPkPtrStr.Validate())

	invalidPkPtrInt := decoratorConfig{
		Name:       "id",
		PrimaryKey: true,
		Type:       reflect.PtrTo(reflect.TypeOf(int64(1))),
	}

	req.NotNil(invalidPkPtrInt.Validate())
}

func TestStructDecoratorConfig_Validate(t *testing.T) {
	req := require.New(t)

	//nil fields
	test := structDecoratorConfig{
		Fields:   nil,
		IsVertex: true,
	}

	req.NotNil(test.Validate())

	//valid pk
	test = structDecoratorConfig{
		Fields: map[string]decoratorConfig{
			"uuid": {
				PrimaryKey: true,
				Name:       "uuid",
				Type:       reflect.TypeOf(""),
			},
		},
		IsVertex: true,
	}

	req.Nil(test.Validate())

	//invalid pk
	test = structDecoratorConfig{
		Fields: map[string]decoratorConfig{
			"uuid": {
				PrimaryKey: true,
				Name:       "uuid",
				Type:       reflect.TypeOf(""),
			},
			"id": {
				PrimaryKey: true,
				Name:       "id",
				Type:       reflect.TypeOf(int64(1)),
			},
		},
		IsVertex: true,
	}

	req.NotNil(test.Validate())

	//invalid rels
	test = structDecoratorConfig{
		Fields: map[string]decoratorConfig{
			"uuid": {
				PrimaryKey: true,
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
	}

	req.NotNil(test.Validate())
}

func TestNewDecoratorConfig(t *testing.T) {
	req := require.New(t)
	var err error
	var compare *decoratorConfig

	decName := "name=id"
	decNameStruct := decoratorConfig{
		Name: "id",
		Type: reflect.TypeOf(int64(1)),
	}

	compare, err = newDecoratorConfig(decName, "", reflect.TypeOf(int64(0)))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decNameStruct, *compare)

	decUUID := "pk;name=uuid"
	decUUIDStruct := decoratorConfig{
		Name:       "uuid",
		PrimaryKey: true,
		Type:       reflect.TypeOf(""),
	}

	compare, err = newDecoratorConfig(decUUID, "", reflect.TypeOf(""))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decUUIDStruct, *compare)

	decIndexField := "index;name=index_field"
	decIndexFieldStruct := decoratorConfig{
		Index: true,
		Name:  "index_field",
		Type:  reflect.TypeOf(""),
	}

	compare, err = newDecoratorConfig(decIndexField, "", reflect.TypeOf(""))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decIndexFieldStruct, *compare)

	decUniqueField := "unique;name=unique_name"
	decUniqueFieldStruct := decoratorConfig{
		Unique: true,
		Name:   "unique_name",
		Type:   reflect.TypeOf(""),
	}

	compare, err = newDecoratorConfig(decUniqueField, "", reflect.TypeOf(""))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decUniqueFieldStruct, *compare)

	decOne2One := "relationship=one2one;direction=incoming"
	decOne2OneStruct := decoratorConfig{
		FieldName:    "test_name",
		Name:         "test_name",
		Relationship: "one2one",
		Direction:    dsl.Incoming,
		Type:         reflect.TypeOf(a{}),
	}

	compare, err = newDecoratorConfig(decOne2One, "test_name", reflect.TypeOf(a{}))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decOne2OneStruct, *compare)

	decProps := "properties;name=test"
	decPropsStruct := decoratorConfig{
		Properties: true,
		Name:       "test",
		Type:       reflect.TypeOf(map[string]interface{}{}),
	}

	compare, err = newDecoratorConfig(decProps, "", reflect.TypeOf(map[string]interface{}{}))
	req.Nil(err)
	req.NotNil(compare)
	req.EqualValues(decPropsStruct, *compare)

	decIgnore := "-"

	compare, err = newDecoratorConfig(decIgnore, "", reflect.TypeOf(int64(0)))
	req.Nil(err)
	req.NotNil(compare)
	req.True(compare.Ignore)

	decInvalidRelName := "relationship=A_REL;direction=incoming;name=ISHOULDNTBEHERE"

	compare, err = newDecoratorConfig(decInvalidRelName, "TestFieldName", reflect.TypeOf(a{}))
	req.NotNil(err)
	req.Nil(compare)

	decInvalidIgnore := "-;index"

	compare, err = newDecoratorConfig(decInvalidIgnore, "", reflect.TypeOf(int64(0)))
	req.NotNil(err)
	req.Nil(compare)
}

//structs with decorators for testing

type validStruct struct {
	Id          int64                  `gogm:"name=id"`
	UUID        string                 `gogm:"pk;name=uuid"`
	IndexField  string                 `gogm:"index;name=index_field"`
	UniqueField int                    `gogm:"unique;name=unique_field"`
	OneToOne    *validStruct           `gogm:"relationship=one2one;direction=incoming"`
	ManyToOne   []interface{}          `gogm:"relationship=many2one;direction=outgoing"`
	Props       map[string]interface{} `gogm:"properties;name=props"`
	IgnoreMe    int                    `gogm:"-"`
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

	Props map[string]string `gogm:"name=props"` //should have properties decorator
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

func TestGetStructDecoratorConfig(t *testing.T) {
	req := require.New(t)

	conf, _, err := getStructDecoratorConfig(&validStruct{})
	req.Nil(err)
	req.NotNil(conf)
	checkObj := structDecoratorConfig{
		IsVertex: true,
		Type:     reflect.TypeOf(validStruct{}),
		Label:    "validStruct",
		Fields: map[string]decoratorConfig{
			"Id": {
				Name:      "id",
				FieldName: "Id",
				Type:      reflect.TypeOf(int64(0)),
			},
			"UUID": {
				Name:       "uuid",
				FieldName:  "UUID",
				PrimaryKey: true,
				Type:       reflect.TypeOf(""),
			},
			"IndexField": {
				FieldName: "IndexField",
				Name:      "index_field",
				Index:     true,
				Type:      reflect.TypeOf(""),
			},
			"UniqueField": {
				FieldName: "UniqueField",
				Unique:    true,
				Name:      "unique_field",
				Type:      reflect.TypeOf(int(1)),
			},
			"OneToOne": {
				FieldName:    "OneToOne",
				Name:         "OneToOne",
				Relationship: "one2one",
				Direction:    dsl.Incoming,
				Type:         reflect.TypeOf(&validStruct{}),
			},
			"ManyToOne": {
				FieldName:        "ManyToOne",
				Name:             "ManyToOne",
				Relationship:     "many2one",
				Direction:        dsl.Outgoing,
				ManyRelationship: true,
				Type:             reflect.TypeOf([]interface{}{}),
			},
			"Props": {
				FieldName:  "Props",
				Properties: true,
				Name:       "props",
				Type:       reflect.TypeOf(map[string]interface{}{}),
			},
			"IgnoreMe": {
				FieldName: "IgnoreMe",
				Name:      "IgnoreMe",
				Ignore:    true,
				Type:      reflect.TypeOf(int(1)),
			},
		},
	}
	req.EqualValues(checkObj, *conf)

	conf, _, err = getStructDecoratorConfig(&mostlyValidStruct{})
	req.NotNil(err)
	req.Nil(conf)

	conf, _, err = getStructDecoratorConfig(&emptyStruct{})
	req.NotNil(err)
	req.Nil(conf)

	conf, _, err = getStructDecoratorConfig(&invalidStructDecorator{})
	req.NotNil(err)
	req.Nil(conf)

	conf, _, err = getStructDecoratorConfig(&invalidStructProperties{})
	req.NotNil(err)
	req.Nil(conf)

	conf, _, err = getStructDecoratorConfig(&invalidEdge{})
	req.NotNil(err)
	req.Nil(conf)

	conf, _, err = getStructDecoratorConfig(&invalidNameStruct{})
	req.NotNil(err)
	req.Nil(conf)

	conf, _, err = getStructDecoratorConfig(&invalidIgnoreStruct{})
	req.NotNil(err)
	req.Nil(conf)
}
