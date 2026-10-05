package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// decodeAndValidate is a reusable test helper available to all test files in this package
func decodeAndValidate(t *testing.T, body string, data any) {
	t.Helper()

	err := json.Unmarshal([]byte(body), &data)
	require.NoError(t, err, "failed to unmarshal response body")

	if Validate != nil {
		err = Validate.Struct(data)
		require.NoError(t, err, "response body failed validation rules")
	}
}
