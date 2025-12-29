package formatter

import (
	"sort"

	"github.com/dave/dst"
)

func reorderStructFields(f *dst.File) {
	dst.Inspect(f, func(n dst.Node) bool {
		if st, ok := n.(*dst.StructType); ok {
			reorderFields(st)
		}

		return true
	})
}

func reorderFields(st *dst.StructType) {
	if st.Fields == nil || len(st.Fields.List) == 0 {
		return
	}

	var embedded, public, private []*dst.Field

	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			embedded = append(embedded, field)
		} else if isExported(field.Names[0].Name) {
			public = append(public, field)
		} else {
			private = append(private, field)
		}
	}

	sortFieldsByTypeName(embedded)
	sortFieldsByName(public)
	sortFieldsByName(private)

	st.Fields.List = assembleFieldList(embedded, public, private)
}

func assembleFieldList(embedded, public, private []*dst.Field) []*dst.Field {
	var result []*dst.Field

	for _, f := range embedded {
		f.Decs.Before = dst.NewLine
		result = append(result, f)
	}

	if len(public) > 0 && len(embedded) > 0 {
		public[0].Decs.Before = dst.EmptyLine
	}
	for _, f := range public {
		if f.Decs.Before != dst.EmptyLine {
			f.Decs.Before = dst.NewLine
		}
		result = append(result, f)
	}

	if len(private) > 0 && (len(embedded) > 0 || len(public) > 0) {
		private[0].Decs.Before = dst.EmptyLine
	}
	for _, f := range private {
		if f.Decs.Before != dst.EmptyLine {
			f.Decs.Before = dst.NewLine
		}
		result = append(result, f)
	}

	return result
}

func collectStructDefinitions(f *dst.File) map[string][]string {
	structDefs := make(map[string][]string)

	dst.Inspect(f, func(n dst.Node) bool {
		ts, ok := n.(*dst.TypeSpec)
		if !ok {
			return true
		}
		st, ok := ts.Type.(*dst.StructType)
		if !ok {
			return true
		}

		structDefs[ts.Name.Name] = computeFieldOrder(st)

		return true
	})

	return structDefs
}

func computeFieldOrder(st *dst.StructType) []string {
	var embedded, public, private []string

	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			embedded = append(embedded, getFieldTypeName(field))
		} else {
			name := field.Names[0].Name
			if isExported(name) {
				public = append(public, name)
			} else {
				private = append(private, name)
			}
		}
	}

	sort.Strings(embedded)
	sort.Strings(public)
	sort.Strings(private)

	result := make([]string, 0, len(embedded)+len(public)+len(private))
	result = append(result, embedded...)
	result = append(result, public...)
	result = append(result, private...)

	return result
}

func reorderStructLiterals(f *dst.File, structDefs map[string][]string) {
	dst.Inspect(f, func(n dst.Node) bool {
		cl, ok := n.(*dst.CompositeLit)
		if !ok {
			return true
		}

		typeName := extractTypeName(cl.Type)
		if typeName == "" {
			return true
		}

		fieldOrder, exists := structDefs[typeName]
		if !exists {
			return true
		}

		reorderCompositeLitFields(cl, fieldOrder)

		return true
	})
}

func reorderCompositeLitFields(cl *dst.CompositeLit, fieldOrder []string) {
	if len(cl.Elts) == 0 {
		return
	}

	keyedElts := make(map[string]*dst.KeyValueExpr)
	var nonKeyed []dst.Expr

	for _, elt := range cl.Elts {
		if kv, ok := elt.(*dst.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*dst.Ident); ok {
				keyedElts[ident.Name] = kv
			}
		} else {
			nonKeyed = append(nonKeyed, elt)
		}
	}

	if len(keyedElts) == 0 {
		return
	}

	var newElts []dst.Expr
	for _, fieldName := range fieldOrder {
		if kv, exists := keyedElts[fieldName]; exists {
			newElts = append(newElts, kv)
			delete(keyedElts, fieldName)
		}
	}

	for _, kv := range keyedElts {
		newElts = append(newElts, kv)
	}

	newElts = append(newElts, nonKeyed...)

	for i, elt := range newElts {
		if kv, ok := elt.(*dst.KeyValueExpr); ok {
			if i == 0 {
				kv.Decs.Before = dst.NewLine
			} else {
				kv.Decs.Before = dst.None
			}
		}
	}

	cl.Elts = newElts
}
