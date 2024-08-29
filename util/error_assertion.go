package util

import "github.com/stretchr/testify/assert"

func ErrorContains(contains string) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
		return assert.Contains(t, err.Error(), contains, msgAndArgs...)
	}
}

func ErrorIs(target error) assert.ErrorAssertionFunc {
	return func(t assert.TestingT, err error, msgAndArgs ...interface{}) bool {
		return assert.ErrorIs(t, err, target, msgAndArgs...)
	}
}
