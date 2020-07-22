// Copyright (c) 2020 MindStand Technologies, Inc
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
	"github.com/adam-hanna/arrayOperations"
	"github.com/cornelk/hashmap"
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func resultToStringArr(res neo4j.Result) ([]string, error) {
	if res == nil {
		return nil, errors.New("result is nil")
	}

	var result []string

	for res.Next() {
		val := res.Record().Values()
		// nothing to parse
		if val == nil || len(val) == 0 {
			continue
		}

		str, ok := val[0].(string)
		if !ok {
			return nil, fmt.Errorf("unable to parse [%T] to string, %w", val[0], ErrInternal)
		}

		result = append(result, str)
	}

	return result, nil
}

//drops all known indexes
func dropAllIndexesAndConstraints() error {
	sess, err := driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer sess.Close()

	res, err := sess.Run("CALL db.constraints", nil)
	if err != nil {
		return err
	}

	constraints, err := resultToStringArr(res)
	if err != nil {
		return err
	}

	//if there is anything, get rid of it
	if len(constraints) != 0 {
		tx, err := sess.BeginTransaction()
		if err != nil {
			return err
		}

		for _, constraint := range constraints {
			log.Debugf("dropping constraint '%s'", constraint)
			_, err := tx.Run(fmt.Sprintf("DROP %s", constraint), nil)
			if err != nil {
				oerr := err
				err = tx.Rollback()
				if err != nil {
					return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
				}

				return oerr
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}
	}

	res, err = sess.Run("CALL db.indexes()", nil)
	if err != nil {
		return err
	}

	indexes, err := resultToStringArr(res)
	if err != nil {
		return err
	}

	//if there is anything, get rid of it
	if len(indexes) != 0 {
		tx, err := sess.BeginTransaction()
		if err != nil {
			return err
		}

		for _, index := range indexes {
			if len(index) == 0 {
				return errors.New("invalid index config")
			}

			_, err := tx.Run(fmt.Sprintf("DROP %s", index), nil)
			if err != nil {
				oerr := err
				err = tx.Rollback()
				if err != nil {
					return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
				}

				return oerr
			}
		}

		return tx.Commit()
	} else {
		return nil
	}
}

//creates all indexes
func createAllIndexesAndConstraints(mappedTypes *hashmap.HashMap) error {
	sess, err := driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer sess.Close()

	//validate that we have to do anything
	if mappedTypes == nil || mappedTypes.Len() == 0 {
		return errors.New("must have types to map")
	}

	numIndexCreated := 0

	tx, err := sess.BeginTransaction()
	if err != nil {
		return err
	}

	//index and/or create unique constraints wherever necessary
	//for node, structConfig := range mappedTypes{
	for nodes := range mappedTypes.Iter() {
		node := nodes.Key.(string)
		structConfig := nodes.Value.(structDecoratorConfig)
		if structConfig.Fields == nil || len(structConfig.Fields) == 0 {
			continue
		}

		var indexFields []string

		for _, config := range structConfig.Fields {
			//pk is a special unique key
			if config.PrimaryKey || config.Unique {
				numIndexCreated++

				_, err := dsl.QB().WithNeo(conn).Create(dsl.NewConstraint(&dsl.ConstraintConfig{
					Unique: true,
					Name:   node,
					Type:   structConfig.Label,
					Field:  config.Name,
				})).Exec(nil)
				if err != nil {
					oerr := err
					err = tx.Rollback()
					if err != nil {
						return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
					}

					return oerr
				}
			} else if config.Index {
				indexFields = append(indexFields, config.Name)
			}
		}

		//create composite index
		if len(indexFields) > 0 {
			numIndexCreated++
			_, err := dsl.QB().WithNeo(conn).Create(dsl.NewIndex(&dsl.IndexConfig{
				Type:   structConfig.Label,
				Fields: indexFields,
			})).Exec(nil)
			if err != nil {
				oerr := err
				err = tx.Rollback()
				if err != nil {
					return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
				}

				return oerr
			}
		}
	}

	log.Debugf("created (%v) indexes", numIndexCreated)

	return tx.Commit()
}

//verifies all indexes
func verifyAllIndexesAndConstraints(mappedTypes *hashmap.HashMap) error {
	sess, err := driver.Session(neo4j.AccessModeWrite)
	if err != nil {
		return err
	}
	defer sess.Close()

	//validate that we have to do anything
	if mappedTypes == nil || mappedTypes.Len() == 0 {
		return errors.New("must have types to map")
	}

	var constraints []string
	var indexes []string

	//build constraint strings
	for nodes := range mappedTypes.Iter() {
		node := nodes.Key.(string)
		structConfig := nodes.Value.(structDecoratorConfig)

		if structConfig.Fields == nil || len(structConfig.Fields) == 0 {
			continue
		}

		fields := []string{}

		for _, config := range structConfig.Fields {

			if config.PrimaryKey || config.Unique {
				t := fmt.Sprintf("CONSTRAINT ON (%s:%s) ASSERT %s.%s IS UNIQUE", node, structConfig.Label, node, config.Name)
				constraints = append(constraints, t)

				indexes = append(indexes, fmt.Sprintf("INDEX ON :%s(%s)", structConfig.Label, config.Name))

			} else if config.Index {
				fields = append(fields, config.Name)
			}
		}

		f := "("
		for _, field := range fields {
			f += field
		}

		f += ")"

		indexes = append(indexes, fmt.Sprintf("INDEX ON :%s%s", structConfig.Label, f))

	}

	//get whats there now
	foundConstraints, err := dsl.QB().WithNeo(conn).Cypher("CALL db.constraints").Query(nil)
	if err != nil {
		return err
	}

	var foundIndexes []string

	findexes, err := dsl.QB().WithNeo(conn).Cypher("CALL db.indexes()").Query(nil)
	if err != nil {
		return err
	}

	if len(findexes) != 0 {
		for _, index := range findexes {
			if len(index) == 0 {
				return errors.New("invalid index config")
			}

			foundIndexes = append(foundIndexes, index[0].(string))
		}
	}

	//verify from there
	delta, found := arrayOperations.Difference(foundIndexes, indexes)
	if !found {
		return fmt.Errorf("found differences in remote vs ogm for found indexes, %v", delta)
	}

	log.Debug(delta)

	var founds []string

	for _, constraint := range foundConstraints {
		if len(constraint) != 0 {
			val, ok := constraint[0].(string)
			if !ok {
				return fmt.Errorf("unable to convert [%T] to [string]", val)
			}

			founds = append(founds, val)
		}
	}

	delta, found = arrayOperations.Difference(founds, constraints)
	if !found {
		return fmt.Errorf("found differences in remote vs ogm for found constraints, %v", delta)
	}

	log.Debug(delta)

	return nil
}
