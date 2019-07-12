package gogm

import (
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/mindstand/gogm/util"
)

//drops all known indexes
func dropAllIndexesAndConstraints() error{
	constraintRows, err := dsl.QB(true).Cypher("CALL db.constraints").Query(nil)
	if err != nil{
		return err
	}

	constraints, err := dsl.RowsToStringArray(constraintRows)
	if err != nil{
		return err
	}

	indexRows, err := dsl.QB(true).Cypher("CALL db.indexes()").Query(nil)
	if err != nil{
		return err
	}

	indexes, err := dsl.RowsTo2dStringArray(indexRows)
	if err != nil{
		return err
	}

	dropSess := dsl.NewSession()
	defer dropSess.Close()
	err = dropSess.Begin()
	if err != nil{
		return err
	}

	for _, constraint := range constraints {
		log.Debugf("dropping constraint '%s'", constraint)
		_, err := dropSess.Query().Cypher(fmt.Sprintf("DROP %s", constraint)).Exec(nil)
		if err != nil{
			oerr := err
			err = dropSess.Rollback()
			if err != nil{
				return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
			}

			return oerr
		}
	}

	for _, index := range indexes{
		if len(index) == 0{
			return errors.New("invalid index config")
		}

		_, err := dropSess.Query().Cypher(fmt.Sprintf("DROP %s", index[0])).Exec(nil)
		if err != nil{
			oerr := err
			err = dropSess.Rollback()
			if err != nil{
				return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
			}

			return oerr
		}
	}

	return dropSess.Commit()
}

//creates all indexes
func createAllIndexesAndConstraints() error{
	//validate that we have to do anything
	if mappedTypes == nil || len(mappedTypes) == 0{
		return errors.New("must have types to map")
	}

	numIndexCreated := 0

	//setup session
	sess := dsl.NewSession()
	defer sess.Close()
	err := sess.Begin()
	if err != nil{
		return err
	}

	//index and/or create unique constraints wherever necessary
	for node, structConfig := range mappedTypes{
		if structConfig.Fields == nil || len(structConfig.Fields) == 0{
			continue
		}

		var indexFields []string

		for fieldName, config := range structConfig.Fields{
			//pk is a special unique key
			if config.PrimaryKey || config.Unique{
				for _, label := range structConfig.Labels{
					numIndexCreated++

					_, err := sess.Query().Create(dsl.NewConstraint(&dsl.ConstraintConfig{
						Unique: true,
						Name: node,
						Type: label,
						Field: fieldName,
					})).Exec(nil)
					if err != nil{
						oerr := err
						err = sess.Rollback()
						if err != nil{
							return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
						}

						return oerr
					}
				}

			} else if config.Index{
				indexFields = append(indexFields)
			}
		}

		//create composite index
		if len(indexFields) > 0{
			for _, label := range structConfig.Labels{
				numIndexCreated++
				_, err := sess.Query().Create(dsl.NewIndex(&dsl.IndexConfig{
					Type: label,
					Fields: indexFields,
				})).Exec(nil)
				if err != nil{
					oerr := err
					err = sess.Rollback()
					if err != nil{
						return fmt.Errorf("failed to rollback, original error was %s", oerr.Error())
					}

					return oerr
				}
			}
		}
	}

	log.Debugf("created (%v) indexes", numIndexCreated)

	return sess.Commit()
}

//verifies all indexes
func verifyAllIndexesAndConstraints() error{
	//validate that we have to do anything
	if mappedTypes == nil || len(mappedTypes) == 0{
		return errors.New("must have types to map")
	}

	var constraints []string
	var indexes []string

	//build constraint strings
	for node, structConfig := range mappedTypes{
		if structConfig.Fields == nil || len(structConfig.Fields) == 0{
			continue
		}

		fields := []string{}

		for fieldName, config := range structConfig.Fields{

			if config.PrimaryKey || config.Unique{
				for _, label := range structConfig.Labels{
					t := fmt.Sprintf("CONSTRAINT ON (%s:%s) ASSERT %s.%s IS UNIQUE", node, label, node, config.Name)
					constraints = append(constraints, t)
				}
			} else if config.Index{
				fields = append(fields, fieldName)
			}
		}

		f := "("
		for _, field := range fields{
			f += field
		}

		f += ")"

		for _, label := range structConfig.Labels{
			indexes = append(indexes, fmt.Sprintf("INDEX ON :%s (%s)", label, f))
		}
	}

	//get whats there now
	constRows, err := dsl.QB(true).Cypher("CALL db.constraints").Query(nil)
	if err != nil{
		return err
	}

	foundConstraints, err := dsl.RowsToStringArray(constRows)
	if err != nil{
		return err
	}

	var foundIndexes []string

	indexRows, err := dsl.QB(true).Cypher("CALL db.indexes()").Query(nil)
	if err != nil{
		return err
	}

	findexes, err := dsl.RowsTo2dStringArray(indexRows)
	if err != nil{
		return err
	}

	for _, index := range findexes{
		if len(index) == 0{
			return errors.New("invalid index config")
		}

		foundIndexes = append(foundIndexes, index[0])
	}

	//verify from there
	_, found := util.Difference(foundIndexes, indexes)
	if found{
		return errors.New("found differences in remote vs ogm for found indexes")
	}

	_, found = util.Difference(foundConstraints, constraints)
	if found{
		return errors.New("found differences in remote vs ogm for found indexes")
	}

	return nil
}
