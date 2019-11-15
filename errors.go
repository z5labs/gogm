package gogm

import (
	"errors"
	"fmt"
)

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

var ErrNotFound = errors.New("gogm: data not found")
var ErrInternal = errors.New("gogm: internal error")
var ErrValidation = errors.New("gogm: struct validation error")
var ErrInvalidParams = errors.New("gogm: invalid params")
var ErrConfiguration = errors.New("gogm: configuration was malformed")
var ErrTransaction = errors.New("gogm: transaction error")
var ErrConnection = errors.New("gogm: connection error")
