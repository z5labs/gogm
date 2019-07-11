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
	if d.Ignore {
		return nil
	}

	//shouldn't happen, more of a sanity check
	if d.Name == ""{
		return NewInvalidStructConfigError("name must be defined", "")
	}

	kind := d.Type.Kind()

	//check for valid properties
	if kind == reflect.Map || d.Properties{
		if !d.Properties{
			return NewInvalidStructConfigError("properties must be added to gogm config on field with a map", d.Name)
		}

		if kind != reflect.Map || d.Type != reflect.TypeOf(map[string]interface{}{}){
			return NewInvalidStructConfigError("properties must be a map with signature map[string]interface{}", d.Name)
		}

		if d.PrimaryKey || d.Relationship != "" || d.Direction != "" || d.Index || d.Unique{
			return NewInvalidStructConfigError("field marked as properties can only have name defined", d.Name)
		}

		//valid properties
		return nil
	}

	//check if type is pointer
	if kind == reflect.Ptr{
		//if it is, get the type of the dereference
		kind = d.Type.Elem().Kind()
	}

	//check valid relationship
	if d.Direction != "" || d.Relationship != "" || kind == reflect.Struct || kind == reflect.Slice{
		if d.Relationship == ""{
			return NewInvalidStructConfigError("relationship has to be defined when creating a relationship", d.Name)
		}

		//check empty/undefined direction
		if d.Direction == ""{
			d.Direction = "outgoing" //default direction is outgoing
		} else {
			if !isValidDirection(d.Direction){
				return NewInvalidStructConfigError(fmt.Sprintf("invalid direction '%s'", d.Direction), d.Name)
			}
		}

		if kind != reflect.Struct && kind != reflect.Slice{
			return NewInvalidStructConfigError("relationship can only be defined on a struct or a slice", d.Name)
		}

		//check that it isn't defining anything else that shouldn't be defined
		if d.PrimaryKey || d.Properties || d.Index || d.Unique {
			return NewInvalidStructConfigError("can only define relationship, direction and name on a relationship", d.Name)
		}

		//relationship is valid now
		return nil
	}

	//standard field checks now

	//check pk and index and unique on the same field
	if d.PrimaryKey && (d.Index || d.Unique) {
		return NewInvalidStructConfigError("can not specify Index or Unique on primary key", d.Name)
	}

	if d.Index && d.Unique{
		return NewInvalidStructConfigError("can not specify Index and Unique on the same field", d.Name)
	}

	//validate pk
	if d.PrimaryKey{
		rootKind := d.Type.Kind()

		if rootKind != reflect.String && rootKind != reflect.Int64{
			return NewInvalidStructConfigError(fmt.Sprintf("invalid type for primary key %s", rootKind.String()), d.Name)
		}

		if rootKind == reflect.String{
			if d.Name != "uuid"{
				return NewInvalidStructConfigError("primary key with type string must be named 'uuid'", d.Name)
			}
		}

		if rootKind == reflect.Int64{
			if d.Name != "id"{
				return NewInvalidStructConfigError("primary key with type int64 must be named 'id'", d.Name)
			}
		}
	} else if d.Name == "id" || d.Name == "uuid"{
		return NewInvalidStructConfigError("name can not be id or uuid if not primary key", d.Name)
	}


	//should be good from here
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
type structDecoratorConfig struct{
	Fields   map[string]decoratorConfig
	IsVertex bool
}

//validates struct configuration
func (s *structDecoratorConfig) Validate() error{

	return nil
}

func getStructDecoratorConfig(i interface{}) (*structDecoratorConfig, error){
	toReturn := &structDecoratorConfig{}

	isEdge := false
	isVertex := false

	//check if its an edge
	if _, ok := i. (IEdge); ok{
		isEdge = true
	}

	if _, ok := i.(IVertex); ok{
		isVertex = true
	}

	if isEdge == isVertex{
		if !isVertex{
			return nil, errors.New("has to implement either IEdge or IVertex")
		}
		return nil, errors.New("can not implement both IEdge and IVertex")
	}

	toReturn.IsVertex = isVertex

	t := reflect.TypeOf(i)

	if t.NumField() == 0{
		return nil, errors.New("struct has no fields") //todo make error more thorough
	}

	toReturn.Fields = map[string]decoratorConfig{}

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

			toReturn.Fields[field.Name] = *config
		}
	}

	err := toReturn.Validate()
	if err != nil{
		return nil, err
	}

	return toReturn, nil
}