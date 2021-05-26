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
)

// todo replace this with go 1.13 errors

type InvalidDecoratorConfigError struct {
	Field string
	Issue string
}

func NewInvalidDecoratorConfigError(issue, field string) *InvalidDecoratorConfigError {
	return &InvalidDecoratorConfigError{
		Issue: issue,
		Field: field,
	}
}

func (i *InvalidDecoratorConfigError) Error() string {
	return fmt.Sprintf("issue: %s. occurred on field '%s'", i.Issue, i.Field)
}

type InvalidStructConfigError struct {
	issue string
}

func NewInvalidStructConfigError(issue string) *InvalidStructConfigError {
	return &InvalidStructConfigError{
		issue: issue,
	}
}

func (i *InvalidStructConfigError) Error() string {
	return i.issue
}

// base errors for gogm 1.13 errors, these are pretty self explanatory
var ErrNotFound = errors.New("gogm: data not found")
var ErrInternal = errors.New("gogm: internal error")
var ErrValidation = errors.New("gogm: struct validation error")
var ErrInvalidParams = errors.New("gogm: invalid params")
var ErrConfiguration = errors.New("gogm: configuration was malformed")
var ErrTransaction = errors.New("gogm: transaction error")
var ErrConnection = errors.New("gogm: connection error")
