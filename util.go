package gogm

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"reflect"
	"sync"
	"time"
)

func setUuidIfNeeded(val *reflect.Value, fieldName string) (bool, string, error){
	if val == nil{
		return false, "", errors.New("value can not be nil")
	}

	if reflect.TypeOf(*val).Kind() == reflect.Ptr{
		*val = val.Elem()
	}

	checkUuid := reflect.Indirect(*val).FieldByName(fieldName).Interface().(string)
	if checkUuid != ""{
		return false, checkUuid, nil
	}

	newUuid := uuid.New().String()

	reflect.Indirect(*val).FieldByName(fieldName).Set(reflect.ValueOf(newUuid))
	return true, newUuid, nil
}

func getTypeName(val reflect.Type) (string, error){
	if val.Kind() == reflect.Ptr{
		val = val.Elem()
	}

	if val.Kind() == reflect.Slice{
		val = val.Elem()
		if val.Kind() == reflect.Ptr{
			val = val.Elem()
		}
	}

	if val.Kind() == reflect.Struct{
		return val.Name(), nil
	} else {
		return "", fmt.Errorf("can not take name from kind {%s)", val.Kind().String())
	}
}

func toCypherParamsMap(val reflect.Value, config structDecoratorConfig) (map[string]interface{}, error){
	var err error
	defer func() {
		if r := recover(); r != nil{
			err = fmt.Errorf("%v", r)
		}
	}()

	if val.Type().Kind() == reflect.Interface || val.Type().Kind() == reflect.Ptr{
		val = val.Elem()
	}

	ret := map[string]interface{}{}

	for _, conf := range config.Fields{
		if conf.Relationship != "" || conf.Name == "id"{
			continue
		}

		if conf.IsTime {
			if conf.Type.Kind() == reflect.Int64{
				ret[conf.Name] = val.FieldByName(conf.FieldName).Interface()
			} else {
				dateInterface := val.FieldByName(conf.FieldName).Interface()

				dateObj, ok := dateInterface.(time.Time)
				if !ok {
					return nil, errors.New("cant convert date to time.Time")
				}

				ret[conf.Name] = dateObj.Format(time.RFC3339)
			}
		} else if conf.Properties {
			if conf.Type.Kind() == reflect.Map{
				propsMap, ok := val.Interface().(map[string]interface{})
				if !ok {
					for k, v := range propsMap{
						ret[conf.Name + "." + k] = v
					}
				} else {
					return nil, errors.New("unable to convert map to map[string]interface{}")
				}
			} else {
				return nil, errors.New("properties type is not a map")
			}
		} else {
			ret[conf.Name] = val.FieldByName(conf.FieldName).Interface()
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

	log.Infof("mapped relations [%s][%s][%v]", key, fieldType, len(r.configs[key][fieldType]))

	r.configs[key][fieldType] = append(r.configs[key][fieldType], dec)
}

func (r *relationConfigs) GetConfigs(nodeType, relationship, fieldType string) ([]decoratorConfig, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.configs == nil {
		return nil, errors.New("no configs provided")
	}

	key := r.getKey(nodeType, relationship)

	if _, ok := r.configs[key]; !ok {
		return nil, fmt.Errorf("no configs for key [%s]", key)
	}

	if _, ok := r.configs[key][fieldType]; !ok {
		return nil, fmt.Errorf("no configs for key [%s] and field type [%s]", key, fieldType)
	}

	return r.configs[key][fieldType], nil
}