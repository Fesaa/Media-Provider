package utils

import (
	"errors"
	"fmt"
	"strings"
)

type DetailedError struct {
	OriginalError error
	Context       string
	ErrorChain    []string
	RootCause     string
}

func (de *DetailedError) Error() string {
	var builder strings.Builder

	builder.WriteString("\n=== DETAILED ERROR REPORT ===\n")

	if de.Context != "" {
		builder.WriteString(fmt.Sprintf("Context: %s\n", de.Context))
	}

	builder.WriteString(fmt.Sprintf("Root Cause: %s\n", de.RootCause))

	if len(de.ErrorChain) > 1 {
		builder.WriteString("\nError Chain (most recent first):\n")
		for i, errMsg := range de.ErrorChain {
			builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, errMsg))
		}
	}

	builder.WriteString("================================")

	return builder.String()
}

func (de *DetailedError) Unwrap() error {
	return de.OriginalError
}

func NewDetailedError(err error) *DetailedError {
	if err == nil {
		return nil
	}

	// Extract error chain
	var errorChain []string
	var rootCause string

	current := err
	for {
		errorChain = append(errorChain, current.Error())
		rootCause = current.Error() // Keep updating, last one will be root

		if unwrapped := errors.Unwrap(current); unwrapped != nil {
			current = unwrapped
		} else {
			break
		}
	}

	return &DetailedError{
		OriginalError: err,
		ErrorChain:    errorChain,
		RootCause:     rootCause,
	}
}
