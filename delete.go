package gogm

import (
	"errors"
	dsl "github.com/mindstand/go-cypherdsl"
	"reflect"
)

func deleteNode(conn *dsl.Session, deleteObj interface{}) error{
	rawType := reflect.TypeOf(deleteObj)

	if rawType.Kind() != reflect.Ptr && rawType.Kind() != reflect.Slice{
		return errors.New("delete obj can only be ptr or slice")
	}

	var ids []int64

	if rawType.Kind() == reflect.Ptr{
		delValue := reflect.ValueOf(deleteObj).Elem()
		id, ok := delValue.FieldByName("Id").Interface().(int64)
		if !ok{
			return errors.New("unable to cast id to int64")
		}

		ids = append(ids, id)
	} else {
		slType := rawType.Elem()

		extraElem := false

		if slType.Kind() == reflect.Ptr{
			extraElem = true
		}

		slVal := reflect.ValueOf(deleteObj)

		slLen := slVal.Len()

		for i := 0; i < slLen; i++{
			val := slVal.Index(i)
			if extraElem {
				val = val.Elem()
			}

			id, ok := val.FieldByName("Id").Interface().(int64)
			if !ok{
				return errors.New("unable to cast id to int64")
			}

			ids = append(ids, id)
		}
	}

	return deleteByIds(conn, ids...)
}

func deleteByIds(conn *dsl.Session, ids ...int64) error{
	rows, err := conn.Query().
		Cypher("UNWIND {rows} as row").
		Match(dsl.Path().V(dsl.V{Name: "n"}).Build()).
		Where(dsl.C(&dsl.ConditionConfig{
			FieldManipulationFunction: "ID",
			Name: "n",
			ConditionOperator: dsl.EqualToOperator,
			Check: dsl.ParamString("row"),
		})).
		Delete(true, "n").
		Exec(map[string]interface{}{
			"rows": ids,
		})
	if err != nil{
		return err
	}

	if numRows, err := rows.RowsAffected(); err != nil{
		return err
	} else if numRows == 0{
		return errors.New("nothing got deleted")
	}

	return nil
}

func deleteByUuids(conn *dsl.Session, ids ...string) error{
	rows, err := conn.Query().
		Cypher("UNWIND {rows} as row").
		Match(dsl.Path().V(dsl.V{Name: "n"}).Build()).
		Where(dsl.C(&dsl.ConditionConfig{
			Name: "n",
			Field: "uuid",
			ConditionOperator: dsl.EqualToOperator,
			Check: dsl.ParamString("row"),
		})).
		Delete(true, "n").
		Exec(map[string]interface{}{
			"rows": ids,
		})
	if err != nil{
		return err
	}

	if numRows, err := rows.RowsAffected(); err != nil{
		return err
	} else if numRows == 0{
		return errors.New("nothing got deleted")
	}

	return nil
}