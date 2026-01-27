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

// collectOriginalFieldOrder collects the original (unsorted) field order for each struct.
// This is needed for converting positional literals to keyed literals.
func collectOriginalFieldOrder(f *dst.File) map[string][]string {
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

		structDefs[ts.Name.Name] = getFieldNamesFromStructType(st)

		return true
	})

	return structDefs
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

func reorderStructLiterals(f *dst.File, structDefs map[string][]string) {
	dst.Inspect(f, func(n dst.Node) bool {
		cl, ok := n.(*dst.CompositeLit)
		if !ok {
			return true
		}

		// Process this literal and all nested children for reordering
		reorderCompositeLitRecursive(cl, nil, structDefs)

		// Don't let dst.Inspect descend into children - we handle them
		return false
	})
}

func reorderCompositeLitRecursive(cl *dst.CompositeLit, inheritedFieldOrder []string, structDefs map[string][]string) {
	// Determine field order for THIS literal
	fieldOrder := resolveSortedFieldOrder(cl.Type, inheritedFieldOrder, structDefs)

	// Reorder if we know the field order
	if len(fieldOrder) > 0 {
		reorderCompositeLitFields(cl, fieldOrder)
	}

	// Determine field order to pass to children (from element type)
	childFieldOrder := getElementSortedFieldOrder(cl.Type, structDefs)

	// Process all child elements
	for _, elt := range cl.Elts {
		reorderElementRecursive(elt, childFieldOrder, structDefs)
	}
}

func reorderElementRecursive(elt dst.Expr, inheritedFieldOrder []string, structDefs map[string][]string) {
	switch e := elt.(type) {
	case *dst.CompositeLit:
		reorderCompositeLitRecursive(e, inheritedFieldOrder, structDefs)
	case *dst.KeyValueExpr:
		// Value might be a composite literal (map values, struct fields)
		if child, ok := e.Value.(*dst.CompositeLit); ok {
			reorderCompositeLitRecursive(child, inheritedFieldOrder, structDefs)
		}
	}
}

func resolveSortedFieldOrder(t dst.Expr, inherited []string, structDefs map[string][]string) []string {
	if t == nil {
		return inherited
	}

	// Anonymous struct type - get field names from the (now sorted) struct
	if st, ok := t.(*dst.StructType); ok {
		return getFieldNamesFromStructType(st)
	}

	// Named type
	if typeName := extractTypeName(t); typeName != "" {
		if order, exists := structDefs[typeName]; exists {
			return order
		}
	}

	return nil
}

func getElementSortedFieldOrder(t dst.Expr, structDefs map[string][]string) []string {
	if t == nil {
		return nil
	}

	// Slice/array: []T or [N]T
	if at, ok := t.(*dst.ArrayType); ok {
		return resolveSortedFieldOrder(at.Elt, nil, structDefs)
	}

	// Map: map[K]V - return value type's field order
	if mt, ok := t.(*dst.MapType); ok {
		return resolveSortedFieldOrder(mt.Value, nil, structDefs)
	}

	return nil
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

	// Capture the original first element's decoration
	var originalFirstBefore dst.SpaceType
	if len(cl.Elts) > 0 {
		if kv, ok := cl.Elts[0].(*dst.KeyValueExpr); ok {
			originalFirstBefore = kv.Decs.Before
		}
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

	// Preserve original decoration style
	for i, elt := range newElts {
		if kv, ok := elt.(*dst.KeyValueExpr); ok {
			if i == 0 {
				kv.Decs.Before = originalFirstBefore
			} else {
				kv.Decs.Before = dst.None
			}
		}
	}

	cl.Elts = newElts
}

func isPositionalLiteral(cl *dst.CompositeLit) bool {
	if len(cl.Elts) == 0 {
		return false
	}

	for _, elt := range cl.Elts {
		if _, ok := elt.(*dst.KeyValueExpr); ok {
			return false
		}
	}

	return true
}

func getFieldNamesFromStructType(st *dst.StructType) []string {
	if st == nil || st.Fields == nil {
		return nil
	}

	var names []string
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			// Embedded field - use type name
			names = append(names, getFieldTypeName(field))
		} else {
			// Named field(s)
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
		}
	}

	return names
}

func convertToKeyedLiteral(cl *dst.CompositeLit, fieldNames []string) {
	if len(fieldNames) == 0 || len(cl.Elts) == 0 {
		return
	}

	newElts := make([]dst.Expr, 0, len(cl.Elts))
	for i, elt := range cl.Elts {
		if i >= len(fieldNames) {
			break
		}

		kv := &dst.KeyValueExpr{
			Key:   dst.NewIdent(fieldNames[i]),
			Value: elt,
		}
		newElts = append(newElts, kv)
	}

	cl.Elts = newElts
}

func convertPositionalToKeyed(f *dst.File, structDefs map[string][]string) {
	dst.Inspect(f, func(n dst.Node) bool {
		cl, ok := n.(*dst.CompositeLit)
		if !ok {
			return true
		}

		// Process this literal and all nested children
		processCompositeLit(cl, nil, structDefs)

		// Don't let dst.Inspect descend into children - we handle them
		return false
	})
}

func processCompositeLit(cl *dst.CompositeLit, inheritedFieldNames []string, structDefs map[string][]string) {
	// Determine field names for THIS literal
	fieldNames := resolveFieldNames(cl.Type, inheritedFieldNames, structDefs)

	// Convert if positional and we know the field names
	if len(fieldNames) > 0 && isPositionalLiteral(cl) {
		convertToKeyedLiteral(cl, fieldNames)
	}

	// Determine field names to pass to children (from element type)
	childFieldNames := getElementFieldNames(cl.Type, structDefs)

	// Process all child elements
	for _, elt := range cl.Elts {
		processElement(elt, childFieldNames, structDefs)
	}
}

func processElement(elt dst.Expr, inheritedFieldNames []string, structDefs map[string][]string) {
	switch e := elt.(type) {
	case *dst.CompositeLit:
		processCompositeLit(e, inheritedFieldNames, structDefs)
	case *dst.KeyValueExpr:
		// Value might be a composite literal (map values, struct fields)
		if child, ok := e.Value.(*dst.CompositeLit); ok {
			processCompositeLit(child, inheritedFieldNames, structDefs)
		}
	}
}

func resolveFieldNames(t dst.Expr, inherited []string, structDefs map[string][]string) []string {
	if t == nil {
		return inherited
	}

	// Anonymous struct type
	if st, ok := t.(*dst.StructType); ok {
		return getFieldNamesFromStructType(st)
	}

	// Named type
	if typeName := extractTypeName(t); typeName != "" {
		if names, exists := structDefs[typeName]; exists {
			return names
		}
	}

	return nil
}

func getElementFieldNames(t dst.Expr, structDefs map[string][]string) []string {
	if t == nil {
		return nil
	}

	// Slice/array: []T or [N]T
	if at, ok := t.(*dst.ArrayType); ok {
		return resolveFieldNames(at.Elt, nil, structDefs)
	}

	// Map: map[K]V - return value type's field names
	if mt, ok := t.(*dst.MapType); ok {
		return resolveFieldNames(mt.Value, nil, structDefs)
	}

	return nil
}
