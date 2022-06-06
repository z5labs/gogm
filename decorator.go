// Copyright (c) 2022 MindStand Technologies, Inc
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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	dsl "github.com/mindstand/go-cypherdsl"
)

// defined the decorator name for struct tag
const decoratorName = "gogm"

// reflect type for go time.Time
var timeType = reflect.TypeOf(time.Time{})

//sub fields of the decorator
const (
	// specifies the name in neo4j
	//requires assignment (if specified)
	paramNameField = "name"

	// specifies the name of the relationship
	//requires assignment (if edge field)
	relationshipNameField = "relationship"

	//specifies direction, can only be (incoming|outgoing|both|none)
	//requires assignment (if edge field)
	directionField = "direction"

	//specifies if the field is to be indexed
	indexField = "index"

	//specifies if the field is unique
	uniqueField = "unique"

	//specifies is the field is a primary key
	primaryKeyField = "pk"

	//specifies if the field is map of type `map[string]interface{} or []<primitive>`
	propertiesField = "properties"

	// startField defines field which is start in a special edge
	startField = "start"

	// endField defines field which is end in a special edge
	endField = "end"

	//specifies if the field is to be ignored
	ignoreField = "-"

	//specifies deliminator between GoGM tags
	deliminator = ";"

	//assignment operator for GoGM tags
	assignmentOperator = "="
)

type propConfig struct {
	// IsMap if false assume slice
	IsMap        bool
	IsMapSlice   bool
	IsMapSliceTd bool
	MapSliceType reflect.Type
	SubType      reflect.Type
}

type specialEdgeConfig struct {
	Type reflect.Type
}

//decorator config defines configuration of GoGM field
type fieldDecoratorConfig struct {
	// ParentType holds the type of the parent object in order to validate some relationships
	ParentType reflect.Type
	// holds reflect type for the field
	Type reflect.Type
	// holds the name of the field for neo4j
	Name string
	// holds the name of the field in the struct
	FieldName string
	// holds the name of the relationship
	Relationship string
	// holds the direction
	Direction dsl.Direction
	// specifies if field is to be unique
	Unique bool
	// specifies if field is to be indexed
	Index bool
	// specifies if field represents many relationship
	ManyRelationship bool
	// uses edge specifies if the edge is a special node
	UsesEdgeNode bool
	// SpecialEdgeTag defines whether this field is a special edge tag (start or end)
	SpecialEdgeTag bool
	// SpecialEdgeStart defines that this field is start
	SpecialEdgeStart bool
	// SpecialEdgeEnd defines that this field is end
	SpecialEdgeEnd bool
	// SpecialEdgeConfig stores information about this field for runtime use
	SpecialEdgeConfig *specialEdgeConfig
	// specifies whether the field is the nodes primary key
	PrimaryKey string
	// specify if the field holds properties
	Properties bool

	PropConfig *propConfig
	// specifies if the field contains time value
	//	IsTime bool `json:"is_time"`
	// specifies if the field contains a typedef of another type
	IsTypeDef bool
	// holds the reflect type of the root type if typedefed
	TypedefActual reflect.Type
	// specifies whether to ignore the field
	Ignore bool
}

// equals checks equality of decorator configs
func (d *fieldDecoratorConfig) equals(comp *fieldDecoratorConfig) bool {
	if comp == nil {
		return false
	}

	return d.Name == comp.Name && d.FieldName == comp.FieldName && d.Relationship == comp.Relationship &&
		d.Direction == comp.Direction && d.Unique == comp.Unique && d.Index == comp.Index && d.ManyRelationship == comp.ManyRelationship &&
		d.UsesEdgeNode == comp.UsesEdgeNode && d.PrimaryKey == comp.PrimaryKey && d.Properties == comp.Properties &&
		d.IsTypeDef == comp.IsTypeDef && d.Ignore == comp.Ignore && d.SpecialEdgeTag == comp.SpecialEdgeTag && d.SpecialEdgeStart == comp.SpecialEdgeStart && d.SpecialEdgeEnd == comp.SpecialEdgeEnd
}

type structType int

const (
	typeVertex        structType = 0
	typeEdgeInterface structType = 1
	typeEdgeTag       structType = 2
)

// specifies configuration on GoGM node
type structDecoratorConfig struct {
	// Holds fields -> their configurations
	// field name : decorator configuration
	Fields map[string]fieldDecoratorConfig
	// holds label for the node, maps to struct name
	Label string
	// specifies if the node is a vertex or an edge (if true, it's a vertex)
	Type structType
	// holds the reflect type of the struct
	ReflectType reflect.Type
	// defines if there is a pk
	HasPK bool
}

// equals checks equality of structDecoratorConfigs
func (s *structDecoratorConfig) equals(comp *structDecoratorConfig) bool {
	if comp == nil {
		return false
	}

	if comp.Fields != nil && s.Fields != nil {
		for field, decConfig := range s.Fields {
			if compConfig, ok := comp.Fields[field]; ok {
				if !compConfig.equals(&decConfig) {
					return false
				}
			} else {
				return false
			}
		}
	} else {
		return false
	}

	return s.Type == comp.Type && s.Label == comp.Label
}

// validate checks if the configuration is valid
func (d *fieldDecoratorConfig) validate(gogm *Gogm) error {
	if d.Ignore {
		if d.Relationship != "" || d.Unique || d.Index || d.ManyRelationship || d.UsesEdgeNode || d.SpecialEdgeTag ||
			d.PrimaryKey != "" || d.Properties || d.Name != d.FieldName {
			return NewInvalidFieldDecoratorConfigError("ignore tag cannot be combined with any other tag", "")
		}

		return nil
	}

	//shouldn't happen, more of a sanity check
	if d.Name == "" {
		return NewInvalidFieldDecoratorConfigError("name must be defined", "")
	}

	// check if this is a special edge field
	if d.SpecialEdgeTag {
		// verify that only start or end are defined
		if (d.SpecialEdgeStart || d.SpecialEdgeEnd) && !(d.SpecialEdgeStart && d.SpecialEdgeEnd) {
			// we're good
		} else {
			return NewInvalidFieldDecoratorConfigError("End and Start can not both be defined on the same tag, must be one or the other", d.Name)
		}

		// ensure nothing else is defined on this
		if d.Index || d.Unique || d.Properties || d.Relationship != "" || d.PrimaryKey != "" {
			return NewInvalidFieldDecoratorConfigError("Can not provide any additional tags on field marked as end or start", d.Name)
		}

		// special edge field is valid
		return nil
	}

	kind := d.Type.Kind()

	// properties supports map and slices
	if (kind == reflect.Map || kind == reflect.Slice) && d.Properties && d.Relationship == "" {
		if d.PrimaryKey != "" || d.Relationship != "" || d.Direction != 0 || d.Index || d.Unique {
			return NewInvalidFieldDecoratorConfigError("field marked as properties can only have name defined", d.Name)
		}

		if kind == reflect.Slice {
			sliceType := reflect.SliceOf(d.Type)
			sliceKind := sliceType.Elem().Elem().Kind()
			if _, err := getPrimitiveType(sliceKind); err != nil && sliceKind != reflect.Interface {
				return NewInvalidFieldDecoratorConfigError("property slice not of type <primitive>", d.Name)
			}
		} else if kind == reflect.Map {
			if d.Type.Key().Kind() != reflect.String {
				return NewInvalidFieldDecoratorConfigError("property map key not of type string", d.Name)
			}
			mapType := d.Type.Elem()
			mapKind := mapType.Kind()
			// check if the key is a string

			if mapKind == reflect.Slice {
				mapElem := mapType.Elem().Kind()
				if _, err := getPrimitiveType(mapElem); err != nil {
					return NewInvalidFieldDecoratorConfigError("property map not of type <primitive> or []<primitive>", d.Name)
				}
			} else if _, err := getPrimitiveType(mapKind); err != nil && mapType.Kind() != reflect.Interface {
				return NewInvalidFieldDecoratorConfigError("property map not of type <primitive> or []<primitive> or interface{} or []interface{}", d.Name)
			}
		} else {
			return NewInvalidFieldDecoratorConfigError("property muss be map[string]<primitive> or map[string][]<primitive> or []primitive", d.Name)
		}
	} else if d.Properties {
		return NewInvalidFieldDecoratorConfigError("property must be map[string]<primitive> or map[string][]<primitive> or []primitive", d.Name)
	} else if kind == reflect.Map {
		return NewInvalidFieldDecoratorConfigError("field with map must be marked as a property", d.Name)
	}

	//check if type is pointer
	if kind == reflect.Ptr {
		//if it is, get the type of the dereference
		kind = d.Type.Elem().Kind()
	}

	//check valid relationship
	if d.Direction != 0 || d.Relationship != "" || (kind == reflect.Struct && d.Type != timeType) || (kind == reflect.Slice && !d.Properties) {
		if d.Relationship == "" {
			return NewInvalidFieldDecoratorConfigError("relationship has to be defined when creating a relationship", d.FieldName)
		}

		//check empty/undefined direction
		if d.Direction == 0 {
			d.Direction = dsl.DirectionOutgoing //default direction is outgoing
		}

		if kind != reflect.Struct && kind != reflect.Slice {
			return NewInvalidFieldDecoratorConfigError("relationship can only be defined on a struct or a slice", d.Name)
		}

		//check that it isn't defining anything else that shouldn't be defined
		if d.PrimaryKey != "" || d.Properties || d.Index || d.Unique {
			return NewInvalidFieldDecoratorConfigError("can only define relationship, direction and name on a relationship", d.Name)
		}

		// check that name is not defined (should be defaulted to field name)
		if d.Name != d.FieldName {
			return NewInvalidFieldDecoratorConfigError("name tag can not be defined on a relationship (Name and FieldName must be the same)", d.Name)
		}

		//relationship is valid now
		return nil
	}

	//standard field checks now

	//check pk and index and unique on the same field
	if d.PrimaryKey != "" && (d.Index || d.Unique) {
		return NewInvalidFieldDecoratorConfigError("can not specify Index or Unique on primary key", d.Name)
	}

	if d.Index && d.Unique {
		return NewInvalidFieldDecoratorConfigError("can not specify Index and Unique on the same field", d.Name)
	}

	//validate pk
	// ignore default since everything should have that
	if d.PrimaryKey != "" && d.PrimaryKey != DefaultPrimaryKeyStrategy.StrategyName {
		// validate strategy matches
		if d.PrimaryKey != gogm.pkStrategy.StrategyName {
			return fmt.Errorf("trying to use strategy '%s' when '%s' is registered", d.PrimaryKey, gogm.pkStrategy.StrategyName)
		}

		// validate type is correct
		if d.Type != gogm.pkStrategy.Type {
			return fmt.Errorf("struct defined type (%s) different than register pk type (%s)", d.Type.Name(), gogm.pkStrategy.Type.Name())
		}
	}

	//should be good from here
	return nil
}

var edgeType = reflect.TypeOf(new(Edge)).Elem()

// newDecoratorConfig generates decorator config for field
// takes in the raw tag, name of the field and reflect type
// varType has NOT called .Elem()
func newDecoratorConfig(gogm *Gogm, decorator, name string, varType reflect.Type, parentType reflect.Type) (*fieldDecoratorConfig, error) {
	tags := strings.Split(decorator, deliminator)

	if len(tags) == 0 {
		return nil, errors.New("decorator can not be empty")
	}

	//init bools to false
	toReturn := fieldDecoratorConfig{
		ParentType: parentType,
		Unique:     false,
		Ignore:     false,
		Direction:  0,
		Type:       varType,
		FieldName:  name,
	}

	for _, tag := range tags {

		//if it's an assignment, further parsing is needed
		if strings.Contains(tag, assignmentOperator) {
			assign := strings.Split(tag, assignmentOperator)
			if len(assign) != 2 {
				return nil, errors.New("empty assignment") //todo replace with better error
			}

			key := assign[0]
			val := assign[1]

			switch key {
			case paramNameField:
				toReturn.Name = val
				continue
			case primaryKeyField:
				toReturn.PrimaryKey = val
				// set other stuff related to the pk strategy
				if gogm.pkStrategy.StrategyName == val {
					toReturn.Name = gogm.pkStrategy.DBName
					toReturn.FieldName = gogm.pkStrategy.FieldName
				}
				continue
			case relationshipNameField:
				toReturn.Relationship = val
				if varType.Kind() == reflect.Slice {
					toReturn.ManyRelationship = true
					if varType.Elem().Kind() != reflect.Ptr {
						return nil, errors.New("slice must be of pointer type")
					}
					toReturn.UsesEdgeNode = varType.Elem().Implements(edgeType)
				} else {
					toReturn.ManyRelationship = false
					toReturn.UsesEdgeNode = varType.Implements(edgeType)
				}

				continue
			case directionField:
				dir := strings.ToLower(val)
				switch strings.ToLower(dir) {
				case "incoming":
					toReturn.Direction = dsl.DirectionIncoming
					continue
				case "outgoing":
					toReturn.Direction = dsl.DirectionOutgoing
					continue
				case "none":
					toReturn.Direction = dsl.DirectionNone
					continue
				case "both":
					toReturn.Direction = dsl.DirectionBoth
					continue
				default:
					toReturn.Direction = dsl.DirectionNone
					continue
				}
			default:
				return nil, fmt.Errorf("key '%s' is not recognized", key) //todo replace with better errors
			}
		}

		//simple bool check
		switch tag {
		case uniqueField:
			toReturn.Unique = true
			continue
		case ignoreField:
			toReturn.Ignore = true
			continue
		case propertiesField:
			conf := propConfig{}
			conf.IsMapSlice = false
			k := varType.Kind()
			if k == reflect.Slice {
				conf.IsMap = false
				conf.SubType = varType.Elem()
			} else if k == reflect.Map {
				conf.IsMap = true
				sub := varType.Elem()
				if sub.Kind() == reflect.Slice {
					conf.IsMapSlice = true
					// check if actual slice is type deffed
					isAliased, aliasType, err := getActualTypeIfAliased(sub)
					if err != nil {
						return nil, err
					}
					if !isAliased {
						conf.MapSliceType = sub
					} else if aliasType != nil && isAliased {
						conf.MapSliceType = aliasType
						conf.IsMapSliceTd = true
					} else {
						return nil, fmt.Errorf("type found to be aliased but alias type nil")
					}
					conf.SubType = sub.Elem()
				} else {
					conf.SubType = sub
				}
			}
			toReturn.PropConfig = &conf
			toReturn.Properties = true
			continue
		case indexField:
			toReturn.Index = true
			continue
		case startField:
			toReturn.SpecialEdgeTag = true
			toReturn.SpecialEdgeStart = true
			// check if varType is a pointer, if it's not we can't continue
			if varType.Kind() != reflect.Pointer {
				return nil, fmt.Errorf("unable to get type from special edge start since the kind is not a pointer, but a %s: %w", varType.Kind().String(), ErrConfiguration)
			}

			toReturn.SpecialEdgeConfig = &specialEdgeConfig{
				Type: varType.Elem(),
			}
			continue
		case endField:
			toReturn.SpecialEdgeTag = true
			toReturn.SpecialEdgeEnd = true
			// check if varType is a pointer, if it's not we can't continue
			if varType.Kind() != reflect.Pointer {
				return nil, fmt.Errorf("unable to get type from special edge start since the kind is not a pointer, but a %s: %w", varType.Kind().String(), ErrConfiguration)
			}

			toReturn.SpecialEdgeConfig = &specialEdgeConfig{
				Type: varType.Elem(),
			}
			continue
		default:
			return nil, fmt.Errorf("key '%s' is not recognized", tag) //todo replace with better error
		}
	}

	//if its not a relationship, check if the tag was typedeffed
	if toReturn.Relationship == "" {
		//check if tag is type def
		isTypeDef, newType, err := getActualTypeIfAliased(varType)
		if err != nil {
			return nil, err
		}

		//handle if it is
		if isTypeDef {
			if newType == nil {
				return nil, errors.New("new type can not be nil")
			}

			toReturn.IsTypeDef = true
			toReturn.TypedefActual = newType
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
	err := toReturn.validate(gogm)
	if err != nil {
		return nil, err
	}

	return &toReturn, nil
}

//validates if struct decorator is valid
func (s *structDecoratorConfig) validate() error {
	if s.Fields == nil {
		return errors.New("no fields defined")
	}

	pkCount := 0
	rels := 0
	defaultPkFound := false
	startsFound := 0
	endsFound := 0
	markedAsTagEdge := false

	for _, conf := range s.Fields {
		// ignore default, we only care about custom pk's (like uuid)
		if conf.PrimaryKey != "" {
			if conf.PrimaryKey == DefaultPrimaryKeyStrategy.StrategyName {
				defaultPkFound = true
			} else {
				pkCount++
			}

		}

		if conf.SpecialEdgeTag {
			markedAsTagEdge = true
		}

		if conf.SpecialEdgeStart {
			startsFound++
		}

		if conf.SpecialEdgeEnd {
			endsFound++
		}

		if conf.Relationship != "" {
			rels++
		}
	}

	// validate correct numbers of start and end
	if markedAsTagEdge {
		if startsFound != 1 {
			return NewInvalidStructConfigError(fmt.Sprintf("invalid number of fields marked as `start`. Expected 1 found %v", startsFound))
		}

		if endsFound != 1 {
			return NewInvalidStructConfigError(fmt.Sprintf("invalid number of fields marked as `end`. Expected 1 found %v", endsFound))
		}
	}

	s.HasPK = pkCount > 0 || defaultPkFound

	//edge specific check
	if s.Type == typeEdgeInterface || s.Type == typeEdgeTag {
		if rels > 0 {
			return NewInvalidStructConfigError("relationships can not be defined on edges")
		}

		// validate that there are no primary keys the edge
	} else {
		// pks only required on vertexes
		if pkCount == 0 && !defaultPkFound {
			return NewInvalidStructConfigError("primary key required on node/edge " + s.Label)
		} else if pkCount > 1 {
			return NewInvalidStructConfigError("too many primary keys defined")
		}
	}

	//good now
	return nil
}

// getStructDecoratorConfig generates structDecoratorConfig for struct
func getStructDecoratorConfig(gogm *Gogm, i interface{}, mappedRelations *relationConfigs) (*structDecoratorConfig, error) {
	toReturn := &structDecoratorConfig{}

	t := reflect.TypeOf(i)

	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("must pass pointer to struct, instead got %T", i)
	}

	t = t.Elem()

	toReturn.Label = t.Name()

	toReturn.ReflectType = t

	if t.NumField() == 0 {
		return nil, fmt.Errorf("struct has no fields: %w", ErrConfiguration)
	}

	toReturn.Fields = map[string]fieldDecoratorConfig{}

	fields := getFields(t)

	if len(fields) == 0 {
		return nil, errors.New("failed to parse fields")
	}

	// check if any of the fields are for start defined edge
	fieldDefinedSpecialEdge := false

	//iterate through fields and get their configuration
	for _, field := range fields {
		tag := field.Tag.Get(decoratorName)

		if tag != "" {
			config, err := newDecoratorConfig(gogm, tag, field.Name, field.Type, t)
			if err != nil {
				return nil, err
			}

			if config.SpecialEdgeTag {
				fieldDefinedSpecialEdge = true
			}

			if config.Relationship != "" {
				var endType reflect.Type

				if field.Type.Kind() == reflect.Ptr {
					endType = field.Type.Elem()
				} else if field.Type.Kind() == reflect.Slice {
					temp := field.Type.Elem()
					if strings.Contains(temp.String(), "interface") {
						return nil, fmt.Errorf("relationship field [%s] on type [%s] can not be a slice of generic interface", config.Name, toReturn.Label)
					}
					if temp.Kind() == reflect.Ptr {
						temp = temp.Elem()
					} else {
						return nil, fmt.Errorf("relationship field [%s] on type [%s] must a slice[]*%s", config.Name, toReturn.Label, temp.String())
					}
					endType = temp
				} else {
					endType = field.Type
				}

				endTypeName, err := traverseRelType(endType, config.Direction)
				if err != nil {
					return nil, err
				}

				mappedRelations.Add(toReturn.Label, config.Relationship, endTypeName, *config)
			}

			toReturn.Fields[field.Name] = *config
		}
	}

	//check if its an edge
	_, implementsEdge := i.(Edge)
	if implementsEdge && fieldDefinedSpecialEdge {
		return nil, NewInvalidStructConfigError("can not define start/end with tags and implement Edge interface. It is recommended to delete interface implementation functions since Edge is now deprecated")
	}

	if implementsEdge {
		toReturn.Type = typeEdgeInterface
	} else if fieldDefinedSpecialEdge {
		toReturn.Type = typeEdgeTag
	} else {
		toReturn.Type = typeVertex
	}

	err := toReturn.validate()
	if err != nil {
		return nil, err
	}

	return toReturn, nil
}

// getFields gets all fields in a struct, specifically also gets fields from embedded structs
func getFields(val reflect.Type) []*reflect.StructField {
	var fields []*reflect.StructField
	if val.Kind() == reflect.Ptr {
		return getFields(val.Elem())
	}

	for i := 0; i < val.NumField(); i++ {
		tempField := val.Field(i)
		if tempField.Anonymous && tempField.Type.Kind() == reflect.Struct {
			fields = append(fields, getFields(tempField.Type)...)
		} else {
			fields = append(fields, &tempField)
		}
	}

	return fields
}
