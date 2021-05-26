package gogm

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"reflect"
)

type GenerateNewIDFunc func() interface{}

type PrimaryKeyStrategy struct {
	// StrategyName is the name of strategy to map field
	StrategyName string
	// DBName for field in the database
	DBName    string
	FieldName string
	// Type of uuid
	Type reflect.Type
	// GenIDFunc function to generate new id
	GenIDFunc GenerateNewIDFunc

	noop bool
}

func (p *PrimaryKeyStrategy) validate() error {
	if p.StrategyName == "" {
		return errors.New("must have strategy name")
	}

	if p.DBName == "" {
		return errors.New("must have db name")
	}

	if p.Type == nil {
		return errors.New("must define type of primary key")
	}

	if p.GenIDFunc == nil {
		return fmt.Errorf("must define generate id function")
	}

	if p.noop {
		return nil
	}

	// validate that the gen id function generates the same type as p.Type
	testType := reflect.TypeOf(p.GenIDFunc())
	if testType != p.Type {
		return fmt.Errorf("GenIDFunc does not return same type as strategy.Type %s != %s", testType.Name(), p.Type.String())
	}

	return nil
}

var (
	UUIDPrimaryKeyStrategy = &PrimaryKeyStrategy{
		StrategyName: "UUID",
		DBName:       "uuid",
		FieldName:    "UUID",
		Type:         reflect.TypeOf(""),
		GenIDFunc: func() (id interface{}) {
			return uuid.New().String()
		},
		noop: false,
	}
	DefaultPrimaryKeyStrategy = &PrimaryKeyStrategy{
		StrategyName: "default",
		DBName:       "id",
		FieldName:    "Id",
		Type:         reflect.TypeOf(1),
		GenIDFunc: func() (id interface{}) {
			return ""
		},
		noop: true,
	}
)
