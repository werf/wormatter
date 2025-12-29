package formatter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/werf/wormatter/pkg/formatter"
)

func TestFormatter(t *testing.T) {
	testdataDir := "testdata"

	inputPath := filepath.Join(testdataDir, "input.go")
	expectedPath := filepath.Join(testdataDir, "expected.go")
	actualPath := filepath.Join(testdataDir, "actual.go")

	inputBytes, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("failed to read input file: %v", err)
	}

	if err := os.WriteFile(actualPath, inputBytes, 0o644); err != nil {
		t.Fatalf("failed to write actual file: %v", err)
	}
	defer os.Remove(actualPath)

	if err := formatter.FormatFile(actualPath, formatter.Options{}); err != nil {
		t.Fatalf("formatter failed: %v", err)
	}

	actualBytes, err := os.ReadFile(actualPath)
	if err != nil {
		t.Fatalf("failed to read actual file: %v", err)
	}

	expectedBytes, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read expected file: %v", err)
	}

	if string(actualBytes) != string(expectedBytes) {
		t.Errorf("formatted output does not match expected.\n\nActual:\n%s\n\nExpected:\n%s", actualBytes, expectedBytes)
	}
}
