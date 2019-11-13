package main

import (
	"log"
	"reflect"
)

type L2 struct {
	SomeString string `json:"some_string"`
}

type L1 struct {
	L2
	AnotherString string `json:"another_string"`
}

type L0 struct {
	L1
	MoreStrings []string `json:"more_strings"`
}

func main() {
	l := new(L0)
	fields := getFields(reflect.TypeOf(l))
	log.Println(fields)
}

func getFields(val reflect.Type) []string {
	var fields []string
	if val.Kind() == reflect.Ptr {
		return getFields(val.Elem())
	}

	for i := 0; i < val.NumField(); i++ {
		tempField := val.Field(i)
		if tempField.Anonymous && tempField.Type.Kind() == reflect.Struct{
			fields = append(fields, getFields(tempField.Type)...)
		} else {
			fields = append(fields, tempField.Name)
		}
	}

	return fields
}