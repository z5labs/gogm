package gogm

import (
	"errors"
	"reflect"
	"strings"
)

const decoratorName = "gogm"

//sub fields of the decorator
const (
	paramNameField = "name" //requires assignment
	relationshipNameField = "relationship" //requires assignment
	directionField = "direction" //requires assignment
	uniqueField = "unique"
	primaryKeyField = "pk"
	primaryKeyTypeField = "pk_type" //requires assignment
	ignoreField = "-"
	deliminator = ";"
	assignmentOperator = "="
)

type decoratorConfig struct{
	Name string
	Relationship string
	Direction string
	Unique bool
	PrimaryKey bool
	PrimaryKeyType string
	Ignore bool
}

func newDecoratorConfig(decorator string) (*decoratorConfig, error){
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
				toReturn.Direction = val //todo validate direction
				continue
			case primaryKeyTypeField:
				toReturn.PrimaryKeyType = val //todo validate direction
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
		default:
			return nil, errors.New("unknown field") //todo replace with better error
		}
	}

	return &toReturn, nil
}

// field name : decorator configuration
type structDecoratorConfig map[string]decoratorConfig

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
			config, err := newDecoratorConfig(tag)
			if err != nil{
				return nil, err
			}

			if config == nil{
				return nil, errors.New("config is nil") //todo better error
			}

			toReturn[field.Name] = *config
		}
	}

	return toReturn, nil
}