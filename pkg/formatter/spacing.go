package formatter

import (
	"strings"

	"github.com/dave/dst"
)

func addSpaceBeforeComments(f *dst.File) {
	dst.Inspect(f, func(n dst.Node) bool {
		block, ok := n.(*dst.BlockStmt)
		if !ok || len(block.List) < 2 {
			return true
		}
		for i, stmt := range block.List {
			if i == 0 {
				continue
			}
			if hasLineComment(stmt) && stmt.Decorations().Before != dst.EmptyLine {
				stmt.Decorations().Before = dst.EmptyLine
			}
		}

		return true
	})
}

func removeBlankLinesBetweenCases(f *dst.File) {
	dst.Inspect(f, func(n dst.Node) bool {
		switch stmt := n.(type) {
		case *dst.SwitchStmt:
			if stmt.Body != nil {
				normalizeCaseSpacing(stmt.Body.List)
			}
		case *dst.TypeSwitchStmt:
			if stmt.Body != nil {
				normalizeCaseSpacing(stmt.Body.List)
			}
		case *dst.SelectStmt:
			if stmt.Body != nil {
				normalizeCaseSpacing(stmt.Body.List)
			}
		}

		return true
	})
}

func addSpaceBeforeReturns(f *dst.File) {
	dst.Inspect(f, func(n dst.Node) bool {
		block, ok := n.(*dst.BlockStmt)
		if !ok || len(block.List) < 2 {
			return true
		}
		for i, stmt := range block.List {
			if i == 0 {
				continue
			}
			if _, isReturn := stmt.(*dst.ReturnStmt); isReturn {
				if stmt.Decorations().Before != dst.EmptyLine {
					stmt.Decorations().Before = dst.EmptyLine
				}
			}
		}

		return true
	})
}

func expandOneLineFunctions(f *dst.File) {
	dst.Inspect(f, func(n dst.Node) bool {
		fn, ok := n.(*dst.FuncDecl)
		if !ok || fn.Body == nil || len(fn.Body.List) == 0 {
			return true
		}
		fn.Body.List[0].Decorations().Before = dst.NewLine

		return true
	})
}

func hasLineComment(stmt dst.Stmt) bool {
	decs := stmt.Decorations()

	return len(decs.Start) > 0 && strings.HasPrefix(decs.Start[0], "//")
}

func normalizeCaseSpacing(stmts []dst.Stmt) {
	for _, stmt := range stmts {
		switch cc := stmt.(type) {
		case *dst.CaseClause:
			cc.Decs.Before = dst.NewLine
			cc.Decs.After = dst.None
			if len(cc.Body) > 0 {
				cc.Body[0].Decorations().Before = dst.NewLine
			}
		case *dst.CommClause:
			cc.Decs.Before = dst.NewLine
			cc.Decs.After = dst.None
			if len(cc.Body) > 0 {
				cc.Body[0].Decorations().Before = dst.NewLine
			}
		}
	}
}

func normalizeSpacing(f *dst.File) {
	dst.Inspect(f, func(n dst.Node) bool {
		if n == nil {
			return false
		}
		switch d := n.(type) {
		case *dst.GenDecl:
			if d.Decs.Before > dst.EmptyLine {
				d.Decs.Before = dst.EmptyLine
			}
		case *dst.FuncDecl:
			if d.Decs.Before > dst.EmptyLine {
				d.Decs.Before = dst.EmptyLine
			}
		case *dst.Field:
			if d.Decs.Before > dst.EmptyLine {
				d.Decs.Before = dst.EmptyLine
			}
		}

		return true
	})
}
