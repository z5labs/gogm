// Copyright (c) 2022 MindStand Technologies, Inc
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

// InvalidDecoratorConfigError defines an error for a malformed struct tag
type InvalidDecoratorConfigError struct {
	Field string
	Issue string
}

// NewInvalidDecoratorConfigError creates an InvalidDecoratorConfigError structure
func NewInvalidDecoratorConfigError(issue, field string) *InvalidDecoratorConfigError {
	return &InvalidDecoratorConfigError{
		Issue: issue,
		Field: field,
	}
}

// Error() implements builtin Error() interface
func (i *InvalidDecoratorConfigError) Error() string {
	return fmt.Sprintf("issue: %s. occurred on field '%s'", i.Issue, i.Field)
}

// InvalidStructConfigError defines an error for a malformed gogm structure
type InvalidStructConfigError struct {
	issue string
}

// NewInvalidStructConfigError creates an InvalidStructConfigError structure
func NewInvalidStructConfigError(issue string) *InvalidStructConfigError {
	return &InvalidStructConfigError{
		issue: issue,
	}
}

// Error() implements builtin Error() interface
func (i *InvalidStructConfigError) Error() string {
	return i.issue
}

var (
	// ErrNotFound is returned when gogm is unable to find data
	ErrNotFound = errors.New("gogm: data not found")

	// ErrInternal is returned for general internal gogm errors
	ErrInternal = errors.New("gogm: internal error")

	// ErrValidation is returned when there is a validation error
	ErrValidation = errors.New("gogm: struct validation error")

	// ErrInvalidParams is returned when params to a function are invalid
	ErrInvalidParams = errors.New("gogm: invalid params")

	// ErrConfiguration is returned for configuration errors
	ErrConfiguration = errors.New("gogm: configuration was malformed")

	// ErrTransaction is returned for errors related to gogm transactions
	ErrTransaction = errors.New("gogm: transaction error")

	// ErrConnection is returned for connection related errors
	ErrConnection = errors.New("gogm: connection error")
)
