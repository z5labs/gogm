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
	err = dropSess.Begin()
	if err != nil{
		return err
	}

	for _, constraint := range rowsStr{
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
	if mappedTypes == nil || len(mappedTypes) == 0{
		return errors.New("must have types to map")
	}
}

//verifies all indexes
func verifyAllIndexes() error{

}
