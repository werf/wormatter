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
	"regexp"
	"strings"

	"github.com/dave/dst"
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

	if err := checkFreeFloatingComments(f, filePath); err != nil {
		return err
	}

	collapseFuncSignatures(f)
	originalFieldOrder := collectOriginalFieldOrder(f)
	convertPositionalToKeyed(f, originalFieldOrder)
	reorderStructFields(f)
	sortedFieldOrder := collectStructDefinitions(f)
	reorderStructLiterals(f, sortedFieldOrder)
	f.Decls = reorderDeclarations(f)
	normalizeSpacing(f)
	expandOneLineFunctions(f)
	addSpaceBeforeReturns(f)
	addSpaceBeforeComments(f)
	removeBlankLinesBetweenCases(f)

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

func checkFreeFloatingComments(f *dst.File, filePath string) error {
	for _, decl := range f.Decls {
		gd, ok := decl.(*dst.GenDecl)
		if !ok || (gd.Tok != token.CONST && gd.Tok != token.VAR) {
			continue
		}
		if hasFreeFloatingComment(gd.Decs.Start) {
			return fmt.Errorf("%s: file has free-floating comments, cannot safely reorder declarations", filePath)
		}
	}

	return nil
}

func matchesAnyPattern(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, path); matched {
			return true
		}
	}

	return false
}
