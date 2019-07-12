package gogm

import (
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
)

//drops all known indexes
func dropAllIndexes() error{
	rows, err := dsl.QB(true).Cypher("CALL db.constraints").Query(nil)
	if err != nil{
		return err
	}

	rowsStr, err := dsl.RowsToStringArray(rows)
	if err != nil{
		return err
	}

	dropSess := dsl.NewSession()
	defer dropSess.Close()
	err = dropSess.Begin()
	if err != nil{
		return err
	}

	for _, constraint := range rowsStr{
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

	return dropSess.Commit()
}

//creates all indexes
func createAllIndexes() error{
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
func verifyAllIndexes() error{
	//validate that we have to do anything
	if mappedTypes == nil || len(mappedTypes) == 0{
		return errors.New("must have types to map")
	}

	rows, err := dsl.QB(true).Cypher("CALL db.constraints").Query(nil)
	if err != nil{
		return err
	}

	rowsStr, err := dsl.RowsToStringArray(rows)
	if err != nil{
		return err
	}


}
