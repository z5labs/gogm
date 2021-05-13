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
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func resultToStringArrV4(isConstraint bool, res neo4j.Result) ([]string, error) {
	if res == nil {
		return nil, errors.New("result is nil")
	}

	var result []string

	var i int
	if isConstraint {
		i = 0
	} else {
		i = 1
	}

	for res.Next() {
		val := res.Record().Values()
		// nothing to parse
		if val == nil || len(val) == 0 {
			continue
		}

		str, ok := val[i].(string)
		if !ok {
			return nil, fmt.Errorf("unable to parse [%T] to string. Value is %v: %w", val[i], val[i], ErrInternal)
		}

		result = append(result, str)
	}

	return result, nil
}

//drops all known indexes
func dropAllIndexesAndConstraintsV4() error {
	for _, db := range internalConfig.TargetDbs {
		sess, err := driver.NewSession(neo4j.SessionConfig{
			AccessMode:   neo4j.AccessModeWrite,
			Bookmarks:    nil,
			DatabaseName: db,
		})
		if err != nil {
			return err
		}
		defer sess.Close()

		res, err := sess.Run("CALL db.constraints()", nil)
		if err != nil {
			return err
		}

		constraints, err := resultToStringArrV4(true, res)
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
				res, err := tx.Run(fmt.Sprintf("DROP CONSTRAINT %s IF EXISTS", constraint), nil)
				if err != nil {
					oerr := err
					err = tx.Rollback()
					if err != nil {
						return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
					}

					return oerr
				} else if err = res.Err(); err != nil {
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
				oerr := err
				err = tx.Rollback()
				if err != nil {
					return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
				}

				return oerr
			}
		}

		res, err = sess.Run("CALL db.indexes()", nil)
		if err != nil {
			return err
		} else if err = res.Err(); err != nil {
			return err
		}

		indexes, err := resultToStringArrV4(false, res)
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

				res, err := tx.Run(fmt.Sprintf("DROP INDEX %s IF EXISTS", index), nil)
				if err != nil {
					oerr := err
					err = tx.Rollback()
					if err != nil {
						return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
					}

					return oerr
				} else if err = res.Err(); err != nil {
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
				oerr := err
				err = tx.Rollback()
				if err != nil {
					return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
				}

				return oerr
			}
		} else {
			continue
		}
	}
	return nil
}

//creates all indexes
func createAllIndexesAndConstraintsV4(mappedTypes *hashmap.HashMap) error {
	for _, db := range internalConfig.TargetDbs {
		sess, err := driver.NewSession(neo4j.SessionConfig{
			AccessMode:   neo4j.AccessModeWrite,
			Bookmarks:    nil,
			DatabaseName: db,
		})
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

					cyp, err := dsl.QB().Create(dsl.NewConstraint(&dsl.ConstraintConfig{
						Unique: true,
						Name:   node,
						Type:   structConfig.Label,
						Field:  config.Name,
					})).ToCypher()
					if err != nil {
						return err
					}

					res, err := tx.Run(cyp, nil)
					if err != nil {
						oerr := err
						err = tx.Rollback()
						if err != nil {
							return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
						}

						return oerr
					} else if err = res.Err(); err != nil {
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
				cyp, err := dsl.QB().Create(dsl.NewIndex(&dsl.IndexConfig{
					Type:   structConfig.Label,
					Fields: indexFields,
				})).ToCypher()
				if err != nil {
					return err
				}

				res, err := tx.Run(cyp, nil)
				if err != nil {
					oerr := err
					err = tx.Rollback()
					if err != nil {
						return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
					}

					return oerr
				} else if err = res.Err(); err != nil {
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

		err = tx.Commit()
		if err != nil {
			oerr := err
			err = tx.Rollback()
			if err != nil {
				return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
			}

			return oerr
		}
	}
	return nil
}

//verifies all indexes
func verifyAllIndexesAndConstraintsV4(mappedTypes *hashmap.HashMap) error {
	for _, db := range internalConfig.TargetDbs {
		sess, err := driver.NewSession(neo4j.SessionConfig{
			AccessMode:   neo4j.AccessModeWrite,
			Bookmarks:    nil,
			DatabaseName: db,
		})
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
		foundResult, err := sess.Run("CALL db.constraints", nil)
		if err != nil {
			return err
		} else if err = foundResult.Err(); err != nil {
			return err
		}

		foundConstraints, err := resultToStringArrV4(true, foundResult)
		if err != nil {
			return err
		}

		foundInxdexResult, err := sess.Run("CALL db.indexes()", nil)
		if err != nil {
			return err
		} else if err = foundInxdexResult.Err(); err != nil {
			return err
		}

		foundIndexes, err := resultToStringArrV4(false, foundInxdexResult)
		if err != nil {
			return err
		}

		//verify from there
		delta, found := arrayOperations.Difference(foundIndexes, indexes)
		if !found {
			return fmt.Errorf("found differences in remote vs ogm for found indexes, %v", delta)
		}

		log.Debug(delta)

		var founds []string

		for _, constraint := range foundConstraints {
			founds = append(founds, constraint)
		}

		delta, found = arrayOperations.Difference(founds, constraints)
		if !found {
			return fmt.Errorf("found differences in remote vs ogm for found constraints, %v", delta)
		}

		log.Debug(delta)
	}

	return nil
}
