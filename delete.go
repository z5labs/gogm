// Copyright (c) 2019 MindStand Technologies, Inc
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
	"github.com/mindstand/go-bolt/connection"
	dsl "github.com/mindstand/go-cypherdsl"
	"reflect"
)

// deleteNode is used to remove nodes from the database
func deleteNode(conn connection.IConnection, deleteObj interface{}) error {
	rawType := reflect.TypeOf(deleteObj)

	if rawType.Kind() != reflect.Ptr && rawType.Kind() != reflect.Slice {
		return errors.New("delete obj can only be ptr or slice")
	}

	var ids []int64

	if rawType.Kind() == reflect.Ptr {
		delValue := reflect.ValueOf(deleteObj).Elem()
		id, ok := delValue.FieldByName("Id").Interface().(int64)
		if !ok {
			return errors.New("unable to cast id to int64")
		}

		ids = append(ids, id)
	} else {
		slType := rawType.Elem()

		extraElem := false

		if slType.Kind() == reflect.Ptr {
			extraElem = true
		}

		slVal := reflect.ValueOf(deleteObj)

		slLen := slVal.Len()

		for i := 0; i < slLen; i++ {
			val := slVal.Index(i)
			if extraElem {
				val = val.Elem()
			}

			id, ok := val.FieldByName("Id").Interface().(int64)
			if !ok {
				return errors.New("unable to cast id to int64")
			}

			ids = append(ids, id)
		}
	}

	return deleteByIds(conn, ids...)
}

// deleteByIds deletes node by graph ids
func deleteByIds(conn connection.IConnection, ids ...int64) error {
	_, err := dsl.QB().
		Cypher("UNWIND {rows} as row").
		Match(dsl.Path().V(dsl.V{Name: "n"}).Build()).
		Where(dsl.C(&dsl.ConditionConfig{
			FieldManipulationFunction: "ID",
			Name:                      "n",
			ConditionOperator:         dsl.EqualToOperator,
			Check:                     dsl.ParamString("row"),
		})).
		Delete(true, "n").
		WithNeo(conn).
		Exec(map[string]interface{}{
			"rows": ids,
		})
	if err != nil {
		return err
	}

	return nil
}

// deleteByUuids deletes nodes by uuids
func deleteByUuids(conn connection.IConnection, ids ...string) error {
	_, err := dsl.QB().
		Cypher("UNWIND {rows} as row").
		Match(dsl.Path().V(dsl.V{Name: "n"}).Build()).
		Where(dsl.C(&dsl.ConditionConfig{
			Name:              "n",
			Field:             "uuid",
			ConditionOperator: dsl.EqualToOperator,
			Check:             dsl.ParamString("row"),
		})).
		Delete(true, "n").
		WithNeo(conn).
		Exec(map[string]interface{}{
			"rows": ids,
		})
	if err != nil {
		return err
	}

	return nil
}
