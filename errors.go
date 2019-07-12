package gogm

import "fmt"

type InvalidDecoratorConfigError struct{
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
	return fmt.Sprintf("issue: %s. occured on field '%s'", i.Issue, i.Field)
}

type InvalidStructConfigError struct{
	issue string
}

func NewInvalidStructConfigError(issue string) *InvalidStructConfigError{
	return &InvalidStructConfigError{
		issue: issue,
	}
}

func (i *InvalidStructConfigError) Error() string{
	return i.issue
}