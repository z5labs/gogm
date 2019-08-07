package gogm

import (
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
	"reflect"
	"strings"
	"time"
)

const decoratorName = "gogm"

var timeType = reflect.TypeOf(time.Time{})

//sub fields of the decorator
const (
	paramNameField        = "name"         //requires assignment
	relationshipNameField = "relationship" //requires assignment
	directionField        = "direction"    //requires assignment
	timeField             = "time"
	indexField            = "index"
	uniqueField           = "unique"
	primaryKeyField       = "pk"
	propertiesField       = "properties"
	ignoreField           = "-"
	deliminator           = ";"
	assignmentOperator    = "="
)

type decoratorConfig struct {
	Type             reflect.Type
	Name             string
	FieldName        string
	Relationship     string
	Direction        dsl.Direction
	Unique           bool
	Index            bool
	ManyRelationship bool
	UsesEdgeNode     bool
	PrimaryKey       bool
	Properties       bool
	IsTime           bool
	Ignore           bool
}

//have struct validate itself
func (d *decoratorConfig) Validate() error {
	if d.Ignore {
		if d.Relationship != "" || d.Unique || d.Index || d.ManyRelationship || d.UsesEdgeNode ||
			d.PrimaryKey || d.Properties || d.IsTime || d.Name != d.FieldName {
			log.Println(d)
			return NewInvalidDecoratorConfigError("ignore tag cannot be combined with any other tag", "")
		}

		return nil
	}

	//shouldn't happen, more of a sanity check
	if d.Name == "" {
		return NewInvalidDecoratorConfigError("name must be defined", "")
	}

	kind := d.Type.Kind()

	//check for valid properties
	if kind == reflect.Map || d.Properties {
		if !d.Properties {
			return NewInvalidDecoratorConfigError("properties must be added to gogm config on field with a map", d.Name)
		}

		if kind != reflect.Map || d.Type != reflect.TypeOf(map[string]interface{}{}) {
			return NewInvalidDecoratorConfigError("properties must be a map with signature map[string]interface{}", d.Name)
		}

		if d.PrimaryKey || d.Relationship != "" || d.Direction != 0 || d.Index || d.Unique {
			return NewInvalidDecoratorConfigError("field marked as properties can only have name defined", d.Name)
		}

		//valid properties
		return nil
	}

	//check if type is pointer
	if kind == reflect.Ptr {
		//if it is, get the type of the dereference
		kind = d.Type.Elem().Kind()
	}

	//check valid relationship
	if d.Direction != 0 || d.Relationship != "" || (kind == reflect.Struct && d.Type != timeType) || kind == reflect.Slice {
		if d.Relationship == "" {
			return NewInvalidDecoratorConfigError("relationship has to be defined when creating a relationship", d.FieldName)
		}

		//check empty/undefined direction
		if d.Direction == 0 {
			d.Direction = dsl.Outgoing //default direction is outgoing
		}

		if kind != reflect.Struct && kind != reflect.Slice {
			return NewInvalidDecoratorConfigError("relationship can only be defined on a struct or a slice", d.Name)
		}

		//check that it isn't defining anything else that shouldn't be defined
		if d.PrimaryKey || d.Properties || d.Index || d.Unique {
			return NewInvalidDecoratorConfigError("can only define relationship, direction and name on a relationship", d.Name)
		}

		// check that name is not defined (should be defaulted to field name)
		if d.Name != d.FieldName {
			return NewInvalidDecoratorConfigError("name tag can not be defined on a relationship (Name and FieldName must be the same)", d.Name)
		}

		//relationship is valid now
		return nil
	}

	//validate timeField
	if d.IsTime {
		if kind != reflect.Int64 && d.Type != timeType {
			return errors.New("can not be a time value and not be either an int64 or time.Time")
		}

		//time is valid
		return nil
	}

	//standard field checks now

	//check pk and index and unique on the same field
	if d.PrimaryKey && (d.Index || d.Unique) {
		return NewInvalidDecoratorConfigError("can not specify Index or Unique on primary key", d.Name)
	}

	if d.Index && d.Unique {
		return NewInvalidDecoratorConfigError("can not specify Index and Unique on the same field", d.Name)
	}

	//validate pk
	if d.PrimaryKey {
		rootKind := d.Type.Kind()

		if rootKind != reflect.String && rootKind != reflect.Int64 {
			return NewInvalidDecoratorConfigError(fmt.Sprintf("invalid type for primary key %s", rootKind.String()), d.Name)
		}

		if rootKind == reflect.String {
			if d.Name != "uuid" {
				return NewInvalidDecoratorConfigError("primary key with type string must be named 'uuid'", d.Name)
			}
		}

		if rootKind == reflect.Int64 {
			if d.Name != "id" {
				return NewInvalidDecoratorConfigError("primary key with type int64 must be named 'id'", d.Name)
			}
		}
	}

	//should be good from here
	return nil
}

var edgeType = reflect.TypeOf(new(IEdge)).Elem()

func newDecoratorConfig(decorator, name string, varType reflect.Type) (*decoratorConfig, error) {
	fields := strings.Split(decorator, deliminator)

	if len(fields) == 0 {
		return nil, errors.New("decorator can not be empty")
	}

	//init bools to false
	toReturn := decoratorConfig{
		Unique:     false,
		PrimaryKey: false,
		Ignore:     false,
		Direction:  0,
		IsTime:     false,
		Type:       varType,
		FieldName:  name,
	}

	for _, field := range fields {

		//if its an assignment, further parsing is needed
		if strings.Contains(field, assignmentOperator) {
			assign := strings.Split(field, assignmentOperator)
			if len(assign) != 2 {
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
				if varType.Kind() == reflect.Slice {
					toReturn.ManyRelationship = true
					toReturn.UsesEdgeNode = reflect.PtrTo(varType.Elem()).Implements(edgeType)
				} else {
					toReturn.ManyRelationship = false
					toReturn.UsesEdgeNode = varType.Implements(edgeType)
				}

				continue
			case directionField:
				dir := strings.ToLower(val)
				switch strings.ToLower(dir) {
				case "incoming":
					toReturn.Direction = dsl.Incoming
					continue
				case "outgoing":
					toReturn.Direction = dsl.Outgoing
					continue
				case "any":
					toReturn.Direction = dsl.Any
					continue
				default:
					toReturn.Direction = dsl.Any
					continue
				}
			default:
				return nil, fmt.Errorf("key '%s' is not recognized", key) //todo replace with better errors
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
		case timeField:
			toReturn.IsTime = true
		default:
			return nil, fmt.Errorf("key '%s' is not recognized", field) //todo replace with better error
		}
	}

	//use var name if name is not set explicitly
	if toReturn.Name == "" {
		toReturn.Name = name
	} else if toReturn.Relationship != "" {
		// check that name is never defined on a relationship
		return nil, errors.New("name tag can not be defined on a relationship")
	}

	//ensure config complies with constraints
	err := toReturn.Validate()
	if err != nil {
		return nil, err
	}

	return &toReturn, nil
}

type structDecoratorConfig struct {
	// field name : decorator configuration
	Fields   map[string]decoratorConfig
	Label    string
	IsVertex bool
	Type     reflect.Type
}

//validates struct configuration
func (s *structDecoratorConfig) Validate() error {
	if s.Fields == nil {
		return errors.New("no fields defined")
	}

	pkCount := 0
	rels := 0

	for _, conf := range s.Fields {
		if conf.PrimaryKey {
			pkCount++
		}

		if conf.Relationship != "" {
			rels++
		}
	}

	if pkCount == 0 {
		if s.IsVertex {
			return NewInvalidStructConfigError("primary key required on node " + s.Label)
		}
	} else if pkCount > 1 {
		return NewInvalidStructConfigError("too many primary keys defined")
	}

	//edge specific check
	if !s.IsVertex {
		if rels > 0 {
			return NewInvalidStructConfigError("relationships can not be defined on edges")
		}
	}

	//good now
	return nil
}

func getStructDecoratorConfig(i interface{}, mappedRelations *relationConfigs) (*structDecoratorConfig, error) {
	toReturn := &structDecoratorConfig{}

	t := reflect.TypeOf(i)

	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("must pass pointer to struct, instead got %T", i)
	}

	t = t.Elem()

	isEdge := false

	//check if its an edge
	if _, ok := i.(IEdge); ok {
		isEdge = true
	}

	toReturn.IsVertex = !isEdge

	toReturn.Label = t.Name()

	toReturn.Type = t

	if t.NumField() == 0 {
		return nil, errors.New("struct has no fields") //todo make error more thorough
	}

	toReturn.Fields = map[string]decoratorConfig{}

	//iterate through fields and get their configuration
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		tag := field.Tag.Get(decoratorName)

		if tag != "" {
			config, err := newDecoratorConfig(tag, field.Name, field.Type)
			if err != nil {
				return nil, err
			}

			if config == nil {
				return nil, errors.New("config is nil") //todo better error
			}

			if config.Relationship != "" {
				var endType reflect.Type

				if field.Type.Kind() == reflect.Ptr {
					endType = field.Type.Elem()
				} else if field.Type.Kind() == reflect.Slice {
					temp := field.Type.Elem()
					if temp.Kind() == reflect.Ptr {
						temp = temp.Elem()
					}
					endType = temp
				} else {
					endType = field.Type
				}

				endTypeName := ""
				if reflect.PtrTo(endType).Implements(edgeType) {
					log.Info(endType.Name())
					endVal := reflect.New(endType)
					var endTypeVal []reflect.Value

					//log.Info(endVal.String())

					if config.Direction == dsl.Outgoing {
						endTypeVal = endVal.MethodByName("GetEndNodeType").Call(nil)
					} else {
						endTypeVal = endVal.MethodByName("GetStartNodeType").Call(nil)
					}

					if len(endTypeVal) != 1 {
						return nil, errors.New("GetEndNodeType failed")
					}

					if endTypeVal[0].IsNil() {
						return nil, errors.New("GetEndNodeType() can not return a nil value")
					}

					convertedType, ok := endTypeVal[0].Interface().(reflect.Type)
					if !ok {
						return nil, errors.New("cannot convert to type reflect.Type")
					}

					if convertedType.Kind() == reflect.Ptr{
						endTypeName = convertedType.Elem().Name()
					} else {
						endTypeName = convertedType.Name()
					}
				} else {
					endTypeName = endType.Name()
				}

				mappedRelations.Add(toReturn.Label, config.Relationship, endTypeName, *config)
			}

			toReturn.Fields[field.Name] = *config
		}
	}

	err := toReturn.Validate()
	if err != nil {
		return nil, err
	}

	return toReturn, nil
}
