// Copyright (c) 2021 MindStand Technologies, Inc
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
		Type:         reflect.TypeOf(int64Ptr(0)),
		GenIDFunc: func() (id interface{}) {
			return ""
		},
		noop: true,
	}
)
