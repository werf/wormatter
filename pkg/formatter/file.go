package formatter

import (
	"bytes"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/dst/decorator"
	"mvdan.cc/gofumpt/format"
)

var ErrNeedsFormatting = errors.New("file needs formatting")

type Options struct {
	CheckOnly       bool
	ExcludePatterns []string
}

func FormatDirectory(dir string, opts Options) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			if matchesAnyPattern(path, opts.ExcludePatterns) {
				return nil
			}
			if err := FormatFile(path, opts); err != nil {
				return err
			}
		}

		return nil
	})
}

func FormatFile(filePath string, opts Options) error {
	if matchesAnyPattern(filePath, opts.ExcludePatterns) {
		return nil
	}

	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	if isGeneratedFile(f) {
		return nil
	}

	structDefs := collectStructDefinitions(f)
	reorderStructFields(f)
	reorderStructLiterals(f, structDefs)
	f.Decls = reorderDeclarations(f)
	normalizeSpacing(f)
	expandOneLineFunctions(f)
	addSpaceBeforeReturns(f)

	var buf bytes.Buffer
	if err := decorator.Fprint(&buf, f); err != nil {
		return err
	}

	formatted, err := format.Source(buf.Bytes(), format.Options{
		LangVersion: detectGoVersion(filePath),
		ExtraRules:  true,
	})
	if err != nil {
		return err
	}

	formatted, err = formatImports(filePath, formatted)
	if err != nil {
		return err
	}

	if opts.CheckOnly {
		original, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if !bytes.Equal(original, formatted) {
			return fmt.Errorf("%s: %w", filePath, ErrNeedsFormatting)
		}

		return nil
	}

	return os.WriteFile(filePath, formatted, 0o644)
}

func matchesAnyPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
			return true
		}
	}

	return false
}
