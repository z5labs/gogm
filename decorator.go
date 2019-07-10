package gogm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const decoratorName = "gogm"

//sub fields of the decorator
const (
	paramNameField = "name" //requires assignment
	relationshipNameField = "relationship" //requires assignment
	directionField = "direction" //requires assignment
	indexField = "index"
	uniqueField = "unique"
	primaryKeyField = "pk"
	propertiesField = "properties"
	ignoreField = "-"
	deliminator = ";"
	assignmentOperator = "="
)

type decoratorConfig struct{
	Type reflect.Type
	Name string
	Relationship string
	Direction string
	Unique bool
	Index bool
	PrimaryKey bool
	Properties bool
	Ignore bool
}

//have struct validate itself
func (d *decoratorConfig) Validate() error{
	k := d.Type.Kind()

	//check for valid properties
	if k == reflect.Map || d.Properties{
		if !d.Properties{
			return NewInvalidStructConfigError("properties must be added to gogm config on field with a map", d.Name)
		}

		var a interface{}

		if k != reflect.Map || d.Type != reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(a)){
			return NewInvalidStructConfigError("properties must be a map with signature map[string]interface{}", d.Name)
		}

		if d.PrimaryKey || d.Relationship != "" || d.Direction != "" || d.Index || d.Unique{
			return NewInvalidStructConfigError("field marked as properties can only have name defined", d.Name)
		}

		//valid properties
		return nil
	}

	return nil
}

func isValidDirection(d string) bool{
	lowerD := strings.ToLower(d)
	return lowerD == "incoming" || lowerD == "outgoing" || lowerD == "any"
}

func newDecoratorConfig(decorator, name string, varType reflect.Type) (*decoratorConfig, error){
	fields := strings.Split(decorator, deliminator)

	if len(fields) == 0{
		return nil, errors.New("decorator can not be empty")
	}

	//init bools to false
	toReturn := decoratorConfig{
		Unique: false,
		PrimaryKey: false,
		Ignore: false,
	}

	for _, field := range fields{

		//if its an assignment, further parsing is needed
		if strings.Contains(field, assignmentOperator){
			assign := strings.Split(field, assignmentOperator)
			if len(assign) != 2{
				return nil, errors.New("empty assignment") //todo replace with better error
			}

			key := assign[0]
			val := assign[1]

			switch key {
			case paramNameField:
				toReturn.Name = val
				continue
			case relationshipNameField:
				toReturn.Relationship = val
				continue
			case directionField:
				if !isValidDirection(val){
					return nil, fmt.Errorf("%s is not a valid direction", val)
				}
				toReturn.Direction = strings.ToLower(val)
				continue
			default:
				return nil, errors.New("unknown field") //todo replace with better errors
			}
		}

		//simple bool check
		switch field {
		case uniqueField:
			toReturn.Unique = true
			continue
		case primaryKeyField:
			toReturn.PrimaryKey = true
			continue
		case ignoreField:
			toReturn.Ignore = true
			continue
		case propertiesField:
			toReturn.Properties = true
			continue
		case indexField:
			toReturn.Index = true
			continue
		default:
			return nil, errors.New("unknown field") //todo replace with better error
		}
	}

	//use var name if name is not set explicitly
	if toReturn.Name == ""{
		toReturn.Name = name
	}

	toReturn.Type = varType

	//ensure config complies with constraints
	err := toReturn.Validate()
	if err != nil{
		return nil, err
	}

	return &toReturn, nil
}

// field name : decorator configuration
type structDecoratorConfig map[string]decoratorConfig

//validates struct configuration
func (s *structDecoratorConfig) Validate() error{
	//todo write validator
	return nil
}

func getDecoratorConfig(i interface{}) (structDecoratorConfig, error){
	toReturn := structDecoratorConfig{}

	t := reflect.TypeOf(i)

	if t.NumField() == 0{
		return nil, errors.New("struct has no fields") //todo make error more thorough
	}

	//iterate through fields and get their configuration
	for i := 0; i < t.NumField(); i++{
		field := t.Field(i)

		tag := field.Tag.Get(decoratorName)

		if tag != ""{
			config, err := newDecoratorConfig(tag, field.Name, field.Type)
			if err != nil{
				return nil, err
			}

			if config == nil{
				return nil, errors.New("config is nil") //todo better error
			}

			toReturn[field.Name] = *config
		}
	}

	err := toReturn.Validate()
	if err != nil{
		return nil, err
	}

	return toReturn, nil
}