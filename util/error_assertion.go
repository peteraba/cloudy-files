package util

import "github.com/stretchr/testify/assert"

// ErrorContains returns an assertion to check if the error contains the given string.
func ErrorContains(contains string) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
		return assert.Contains(t, err.Error(), contains, msgAndArgs...)
	}
}

// ErrorIs returns an assertion to check if the error is the target error.
func ErrorIs(target error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
		return assert.ErrorIs(t, err, target, msgAndArgs...)
	}
}
