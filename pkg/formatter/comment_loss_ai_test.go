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

func TestAI_CommentPreservationDuringMerge(t *testing.T) {
	runFormatterTest(t, "comment_loss")
}

func TestAI_FreeFloatingCommentDetection(t *testing.T) {
	actualPath := filepath.Join("testdata", "floating_actual.go")
	content := "package main\n\n// Section: Database vars\n\nvar dbHost = \"localhost\"\n\nfunc main() {}\n"

	require.NoError(t, os.WriteFile(actualPath, []byte(content), 0o644))
	defer os.Remove(actualPath)

	err := formatter.FormatFile(actualPath, formatter.Options{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "free-floating comments")
}

func TestAI_NormalDocCommentNoError(t *testing.T) {
	actualPath := filepath.Join("testdata", "doc_comment_actual.go")
	content := "package main\n\n// dbHost is the database host.\nvar dbHost = \"localhost\"\n\nfunc main() {}\n"

	require.NoError(t, os.WriteFile(actualPath, []byte(content), 0o644))
	defer os.Remove(actualPath)

	require.NoError(t, formatter.FormatFile(actualPath, formatter.Options{}))
}

func TestAI_EncodingTaggedStructFieldsNotReordered(t *testing.T) {
	runFormatterTest(t, "encoding_tags")
}

func TestAI_AnonymousStructFieldsNotReordered(t *testing.T) {
	runFormatterTest(t, "anon_struct")
}

func TestAI_InlineCommentAttachmentDuringReorder(t *testing.T) {
	runFormatterTest(t, "inline_comments")
}

func runFormatterTest(t *testing.T, name string) {
	t.Helper()

	inputPath := filepath.Join("testdata", name+"_input.go")
	expectedPath := filepath.Join("testdata", name+"_expected.go")
	actualPath := filepath.Join("testdata", name+"_actual.go")

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
