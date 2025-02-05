package utils

import (
	"fmt"
	"strings"
)

// CombineErrors combines multiple errors into a single error.
// It skips nil errors and returns nil if all errors are nil.
func CombineErrors(errs ...error) error {
	var errorMessages []string

	for _, err := range errs {
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
		}
	}

	if len(errorMessages) == 0 {
		return nil
	}

	return fmt.Errorf("multiple errors occurred: %s", strings.Join(errorMessages, "; "))
}
