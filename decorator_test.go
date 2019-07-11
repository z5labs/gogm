package gogm

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestDecoratorConfig_Validate(t *testing.T) {
	req := require.New(t)

	validProps := decoratorConfig{
		Properties: true,
		Type: reflect.TypeOf(map[string]interface{}{}),
		Name: "test",
	}

	req.Nil(validProps.Validate())

	validRelationship := decoratorConfig{
		Name: "test_rel",
		Relationship: "rel",
		Type: reflect.TypeOf([]interface{}{}),
	}

	req.Nil(validRelationship.Validate())

	validRelationshipWithDirection := decoratorConfig{
		Name: "test_rel",
		Relationship: "rel",
		Direction: "incoming",
		Type: reflect.TypeOf([]interface{}{}),
	}

	req.Nil(validRelationshipWithDirection.Validate())

	validStringPk := decoratorConfig{
		Name: "uuid",
		Type: reflect.TypeOf(""),
		PrimaryKey: true,
	}

	req.Nil(validStringPk.Validate())

	validInt64Pk := decoratorConfig{
		PrimaryKey: true,
		Type: reflect.TypeOf(int64(1)),
		Name: "id",
	}

	req.Nil(validInt64Pk.Validate())

	validFieldIndex := decoratorConfig{
		Name: "test_index",
		Type: reflect.TypeOf(""),
		Index: true,
	}

	req.Nil(validFieldIndex.Validate())

	validFieldUnique := decoratorConfig{
		Name: "test_unique",
		Type: reflect.TypeOf(""),
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
		Type: reflect.MapOf(strType, strType),
		Name: "test",
	}

	req.NotNil(invalidPropsWrongSig)

	invalidPropsExtraDecorators := decoratorConfig{
		Properties: true,
		Type: reflect.TypeOf(map[string]interface{}{}),
		Name: "test",
		Unique: true,
	}

	req.NotNil(invalidPropsExtraDecorators.Validate())

	invalidPropsDecoratorNotSpecified := decoratorConfig{
		Type: reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(map[string]interface{}{})),
		Name: "test",
	}

	req.NotNil(invalidPropsDecoratorNotSpecified.Validate())

	invalidRelationshipType := decoratorConfig{
		Relationship: "test",
		Name: "test",
		Type: strType,
	}

	req.NotNil(invalidRelationshipType.Validate())

	invalidDirectionDefinedNotRel := decoratorConfig{
		Direction: "outgoing",
		Name: "asdfa",
		Type: reflect.TypeOf([]interface{}{}),
	}

	req.NotNil(invalidDirectionDefinedNotRel.Validate())

	invalidPkPtrStr := decoratorConfig{
		Name: "uuid",
		PrimaryKey: true,
		Type: reflect.PtrTo(strType),
	}

	req.NotNil(invalidPkPtrStr.Validate())

	invalidPkPtrInt := decoratorConfig{
		Name: "id",
		PrimaryKey: true,
		Type: reflect.PtrTo(reflect.TypeOf(int64(1))),
	}

	req.NotNil(invalidPkPtrInt.Validate())
}

func TestStructDecoratorConfig_Validate(t *testing.T) {

}

func TestNewDecoratorConfig(t *testing.T){

}

func TestGetStructDecoratorConfig(t *testing.T){

}

//structs with decorators for testing

type validStruct struct{
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`
	IndexField string `gogm:"index;name=index_field"`
	UniqueField int `gogm:"unique;name=unique_field"`
	OneToOne interface{} `gogm:"relationship=one2one;direction=incoming"`
	ManyToOne []interface{} `gogm:"relationship=many2one;direction=outgoing"`
	Props map[string]string `gogm:"properties"`
	IgnoreMe int `gogm:"-"`
}

//issue is that it has no id defined
type mostlyValidStruct struct{
	IndexField string `gogm:"index;name=index_field"`
	UniqueField int `gogm:"unique;name=unique_field"`
}

//nothing defined
type emptyStruct struct {}

//has a valid field but also has a messed up one
type invalidStructDecorator struct{
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`

	MessedUp int `gogm:"sdfasdfasdfa"`
}

type invalidStructProperties struct {
	Id int64 `gogm:"name=id"`
	UUID string `gogm:"pk;name=uuid"`

	Props map[string]string `gogm:"name=props"` //should have properties decorator
}