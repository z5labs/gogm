package gogm

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"reflect"
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
		if conf.Relationship == "" && conf.Name != "id"{
			ret[conf.Name] = val.FieldByName(conf.FieldName).Interface()
		}
	}

	return ret, err
}