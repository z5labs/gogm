package gogm

import "fmt"

type InvalidStructConfigError struct{
	Field string
	Issue string
}

func NewInvalidStructConfigError(issue, field string) *InvalidStructConfigError{
	return &InvalidStructConfigError{
		Issue: issue,
		Field: field,
	}
}

func (i *InvalidStructConfigError) Error() string {
	return fmt.Sprintf("issue: %s. occured on field '%s'", i.Issue, i.Field)
}

