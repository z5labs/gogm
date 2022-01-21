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
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	dsl "github.com/mindstand/go-cypherdsl"
)

// checks if integer is in slice
func int64SliceContains(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// sets uuid for stuct if uuid field is empty
func handleNodeState(pkStrat *PrimaryKeyStrategy, val *reflect.Value) (isNew bool, id int64, relConfig map[string]*RelationConfig, err error) {
	if val == nil {
		return false, -1, nil, errors.New("value can not be nil")
	}

	if pkStrat == nil {
		return false, -1, nil, errors.New("pk strategy can not be nil")
	}

	if reflect.TypeOf(*val).Kind() == reflect.Ptr {
		*val = val.Elem()
	}

	loadVal := reflect.Indirect(*val).FieldByName("LoadMap")
	iConf := loadVal.Interface()

	var loadMap map[string]*RelationConfig
	var ok bool
	if iConf != nil && loadVal.Len() != 0 {
		loadMap, ok = iConf.(map[string]*RelationConfig)
		if !ok {
			return false, -1, nil, fmt.Errorf("unable to cast conf to [map[string]*RelationConfig], %w", ErrInternal)
		}
	}

	// handle the id
	if val.IsNil() {
		isNew = true
	} else {
		idVal := reflect.Indirect(*val).FieldByName(DefaultPrimaryKeyStrategy.FieldName)
		// idVal is a pointer
		idVal = idVal.Elem()
		if idVal.IsValid() {
			id = idVal.Int()
			isNew = false
		} else {
			isNew = true
		}

	}

	// using a pk strategy on top of default graph ids
	if pkStrat.StrategyName != DefaultPrimaryKeyStrategy.StrategyName {
		checkId := reflect.Indirect(*val).FieldByName(pkStrat.FieldName)
		if !checkId.IsZero() && !isNew {
			return false, id, loadMap, nil
		} else {
			// if id was not set by user, gen new id and set it
			if checkId.IsZero() {
				reflect.Indirect(*val).FieldByName(pkStrat.FieldName).Set(reflect.ValueOf(pkStrat.GenIDFunc()))
			}

			return true, -1, loadMap, nil
		}
	} else {
		return isNew, id, loadMap, nil
	}
}

// gets the type name from reflect type
func getTypeName(val reflect.Type) (string, error) {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() == reflect.Slice {
		val = val.Elem()
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
	}

	if val.Kind() == reflect.Struct {
		return val.Name(), nil
	} else {
		return "", fmt.Errorf("can not take name from kind {%s)", val.Kind().String())
	}
}

// converts struct fields to map that cypher can use
func toCypherParamsMap(gogm *Gogm, val reflect.Value, config structDecoratorConfig) (map[string]interface{}, error) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if val.Type().Kind() == reflect.Interface || val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}

	ret := map[string]interface{}{}

	for _, conf := range config.Fields {
		if conf.Relationship != "" || conf.Name == "id" || conf.Ignore {
			continue
		}

		field := val.FieldByName(conf.FieldName)

		if conf.Properties {
			//check if field is a map
			if conf.Type.Kind() == reflect.Map && field.Kind() == reflect.Map {
				for _, e := range field.MapKeys() {
					v := field.MapIndex(e)
					es := e.Interface().(string)
					ret[conf.Name+"."+es] = v.Interface()
				}
			} else if conf.Type.Kind() == reflect.Slice && field.Kind() == reflect.Slice {
				ret[conf.Name] = field.Interface()
			} else {
				return nil, fmt.Errorf("properties type is not a map or slice, %T", field.Interface())
			}
		} else {
			var val interface{}
			//check if field is type aliased
			if conf.IsTypeDef {
				val = field.Convert(conf.TypedefActual).Interface()
			} else {
				val = field.Interface()
			}

			if conf.PrimaryKey != "" {
				if conf.PrimaryKey == DefaultPrimaryKeyStrategy.StrategyName {
					// we dont want to write the id to the params map
					continue
				}
				ret[gogm.pkStrategy.DBName] = val
			} else {
				ret[conf.Name] = val
			}
		}
	}

	return ret, err
}

type relationConfigs struct {
	// [type-relationship][fieldType][]decoratorConfig
	configs map[string]map[string][]decoratorConfig

	mutex sync.Mutex
}

func (r *relationConfigs) getKey(nodeType, relationship string) string {
	return fmt.Sprintf("%s-%s", nodeType, relationship)
}

func (r *relationConfigs) Add(nodeType, relationship, fieldType string, dec decoratorConfig) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.configs == nil {
		r.configs = map[string]map[string][]decoratorConfig{}
	}

	key := r.getKey(nodeType, relationship)

	if _, ok := r.configs[key]; !ok {
		r.configs[key] = map[string][]decoratorConfig{}
	}

	if _, ok := r.configs[key][fieldType]; !ok {
		r.configs[key][fieldType] = []decoratorConfig{}
	}

	//log.Debugf("mapped relations [%s][%s][%v]", key, fieldType, len(r.configs[key][fieldType]))

	r.configs[key][fieldType] = append(r.configs[key][fieldType], dec)
}

func (r *relationConfigs) GetConfigs(startNodeType, startNodeFieldType, endNodeType, endNodeFieldType, relationship string) (start, end *decoratorConfig, err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.configs == nil {
		return nil, nil, errors.New("no configs provided")
	}

	start, err = r.getConfig(startNodeType, relationship, startNodeFieldType, dsl.DirectionOutgoing)
	if err != nil {
		return nil, nil, err
	}

	end, err = r.getConfig(endNodeType, relationship, endNodeFieldType, dsl.DirectionIncoming)
	if err != nil {
		return nil, nil, err
	}

	return start, end, nil
}

func (r *relationConfigs) getConfig(nodeType, relationship, fieldType string, direction dsl.Direction) (*decoratorConfig, error) {
	if r.configs == nil {
		return nil, errors.New("no configs provided")
	}

	key := r.getKey(nodeType, relationship)

	if _, ok := r.configs[key]; !ok {
		return nil, fmt.Errorf("no configs for key [%s]", key)
	}

	var ok bool
	var confs []decoratorConfig

	if confs, ok = r.configs[key][fieldType]; !ok {
		return nil, fmt.Errorf("no configs for key [%s] and field type [%s]", key, fieldType)
	}

	if len(confs) == 1 {
		return &confs[0], nil
	} else if len(confs) > 1 {
		for _, c := range confs {
			if c.Direction == direction {
				return &c, nil
			}
		}
		return nil, errors.New("relation with correct direction not found")
	} else {
		return nil, fmt.Errorf("config not found, %w", ErrInternal)
	}
}

type validation struct {
	Incoming []string
	Outgoing []string
	None     []string
	Both     []string
	BothSelf []string
}

func (r *relationConfigs) Validate() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	checkMap := map[string]*validation{}

	for title, confMap := range r.configs {
		parts := strings.Split(title, "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid length for parts [%v] should be 2. Rel is [%s], %w", len(parts), title, ErrValidation)
		}

		// vType := parts[0]
		relType := parts[1]

		for field, configs := range confMap {
			for _, config := range configs {
				if _, ok := checkMap[relType]; !ok {
					checkMap[relType] = &validation{
						Incoming: []string{},
						Outgoing: []string{},
						None:     []string{},
						Both:     []string{},
						BothSelf: []string{},
					}
				}

				validate := checkMap[relType]

				switch config.Direction {
				case dsl.DirectionIncoming:
					validate.Incoming = append(validate.Incoming, field)
				case dsl.DirectionOutgoing:
					validate.Outgoing = append(validate.Outgoing, field)
				case dsl.DirectionNone:
					validate.None = append(validate.None, field)
				case dsl.DirectionBoth:
					if field == config.ParentType.Name() {
						validate.BothSelf = append(validate.BothSelf, field)
					} else {
						validate.Both = append(validate.Both, field)
					}
				default:
					return fmt.Errorf("unrecognized direction [%s], %w", config.Direction.ToString(), ErrValidation)
				}
			}
		}
	}

	for relType, validateConfig := range checkMap {
		//check normal
		if len(validateConfig.Outgoing) != len(validateConfig.Incoming) {
			return fmt.Errorf("invalid directional configuration on relationship [%s], %w", relType, ErrValidation)
		}

		//check both direction
		if len(validateConfig.Both) != 0 {
			if len(validateConfig.Both)%2 != 0 {
				return fmt.Errorf("invalid length for 'both' validation, %w", ErrValidation)
			}
		}

		//check none direction
		if len(validateConfig.None) != 0 {
			if len(validateConfig.None)%2 != 0 {
				return fmt.Errorf("invalid length for 'none' validation, %w", ErrValidation)
			}
		}
	}
	return nil
}

//isDifferentType, differentType, error
func getActualTypeIfAliased(iType reflect.Type) (bool, reflect.Type, error) {
	if iType == nil {
		return false, nil, errors.New("iType can not be nil")
	}

	if iType.Kind() == reflect.Ptr {
		iType = iType.Elem()
	}

	//check if its a struct or an interface, we can skip that
	if iType.Kind() == reflect.Struct || iType.Kind() == reflect.Interface || iType.Kind() == reflect.Slice || iType.Kind() == reflect.Map {
		return false, nil, nil
	}

	//type is the same as the kind
	if iType.Kind().String() == iType.Name() {
		return false, nil, nil
	}

	actualType, err := getPrimitiveType(iType.Kind())
	if err != nil {
		return false, nil, err
	}

	return true, actualType, nil
}

func getPrimitiveType(k reflect.Kind) (reflect.Type, error) {
	switch k {
	case reflect.Int:
		return reflect.TypeOf(0), nil
	case reflect.Int64:
		return reflect.TypeOf(int64(0)), nil
	case reflect.Int32:
		return reflect.TypeOf(int32(0)), nil
	case reflect.Int16:
		return reflect.TypeOf(int16(0)), nil
	case reflect.Int8:
		return reflect.TypeOf(int8(0)), nil
	case reflect.Uint64:
		return reflect.TypeOf(uint64(0)), nil
	case reflect.Uint32:
		return reflect.TypeOf(uint32(0)), nil
	case reflect.Uint16:
		return reflect.TypeOf(uint16(0)), nil
	case reflect.Uint8:
		return reflect.TypeOf(uint8(0)), nil
	case reflect.Uint:
		return reflect.TypeOf(uint(0)), nil
	case reflect.Bool:
		return reflect.TypeOf(false), nil
	case reflect.Float64:
		return reflect.TypeOf(float64(0)), nil
	case reflect.Float32:
		return reflect.TypeOf(float32(0)), nil
	case reflect.String:
		return reflect.TypeOf(""), nil
	default:
		return nil, fmt.Errorf("[%s] not supported", k.String())
	}
}

func int64Ptr(n int64) *int64 {
	return &n
}

// traverseRelType finds the label of a node from a relationship (decoratorConfig).
// if a special edge is passed in, the linked node's label is returned.
func traverseRelType(endType reflect.Type, direction dsl.Direction) (string, error) {
	if !reflect.PtrTo(endType).Implements(edgeType) {
		return endType.Name(), nil
	}

	endVal := reflect.New(endType)
	var endTypeVal []reflect.Value

	if direction == dsl.DirectionOutgoing {
		endTypeVal = endVal.MethodByName("GetEndNodeType").Call(nil)
	} else {
		endTypeVal = endVal.MethodByName("GetStartNodeType").Call(nil)
	}

	if len(endTypeVal) != 1 {
		return "", errors.New("GetEndNodeType failed")
	}

	if endTypeVal[0].IsNil() {
		return "", errors.New("GetEndNodeType() can not return a nil value")
	}

	convertedType, ok := endTypeVal[0].Interface().(reflect.Type)
	if !ok {
		return "", errors.New("cannot convert to type reflect.Type")
	}

	if convertedType.Kind() == reflect.Ptr {
		return convertedType.Elem().Name(), nil
	} else {
		return convertedType.Name(), nil
	}
}
