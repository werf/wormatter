//go:build ai_tests

package formatter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/werf/wormatter/pkg/formatter"
)

func TestAI_VarOrderPreservesInitDependencies(t *testing.T) {
	inputPath := filepath.Join("testdata", "var_order_input.go")
	expectedPath := filepath.Join("testdata", "var_order_expected.go")
	actualPath := filepath.Join("testdata", "var_order_actual.go")

	inputBytes, err := os.ReadFile(inputPath)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(actualPath, inputBytes, 0o644))
	defer os.Remove(actualPath)

	require.NoError(t, formatter.FormatFile(actualPath, formatter.Options{}))

	actualBytes, err := os.ReadFile(actualPath)
	require.NoError(t, err)

	expectedBytes, err := os.ReadFile(expectedPath)
	require.NoError(t, err)

	assert.Equal(t, string(expectedBytes), string(actualBytes))
}
