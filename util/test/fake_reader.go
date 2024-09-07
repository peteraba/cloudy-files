package test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// MustReader creates a new io.Reader from any interface that can be JSON marshaled.
func MustReader(t *testing.T, data interface{}) io.Reader {
	t.Helper()

	result, err := json.Marshal(data)
	require.NoError(t, err)

	return bytes.NewReader(result)
}
