package formatter

import (
	"bytes"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dave/dst/decorator"
	"mvdan.cc/gofumpt/format"
)

func FormatDirectory(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".go") {
			if err := FormatFile(path); err != nil {
				return err
			}
		}

		return nil
	})
}

func FormatFile(filePath string) error {
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

	return os.WriteFile(filePath, formatted, 0o644)
}
