package formatter

import (
	"bytes"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"mvdan.cc/gofumpt/format"
)

type declCollector struct {
	blankVarSpecs  []dst.Spec
	constSpecs     []dst.Spec
	constructors   map[string][]*dst.FuncDecl
	functions      []dst.Decl
	imports        []dst.Decl
	initFuncs      []*dst.FuncDecl
	iotaConstDecls []*dst.GenDecl
	mainFunc       *dst.FuncDecl
	methodsByType  map[string][]*dst.FuncDecl
	orphanMethods  []*dst.FuncDecl
	typeDecls      []*dst.GenDecl
	typeNames      map[string]bool
	varSpecs       []dst.Spec
}

func (c *declCollector) collect(f *dst.File) {
	c.collectTypeNames(f)

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *dst.GenDecl:
			c.collectGenDecl(d)
		case *dst.FuncDecl:
			c.collectFuncDecl(d)
		}
	}
}

func (c *declCollector) collectFuncDecl(d *dst.FuncDecl) {
	switch {
	case d.Recv != nil:
		recvType := getReceiverTypeName(d)
		if recvType == "" || !c.typeNames[recvType] {
			c.orphanMethods = append(c.orphanMethods, d)
		} else {
			c.methodsByType[recvType] = append(c.methodsByType[recvType], d)
		}
	case d.Name.Name == "init":
		c.initFuncs = append(c.initFuncs, d)
	case d.Name.Name == "main":
		c.mainFunc = d
	case strings.HasPrefix(d.Name.Name, "New"):
		if typeName := findConstructorType(d, c.typeNames); typeName != "" {
			c.constructors[typeName] = append(c.constructors[typeName], d)
		} else {
			c.functions = append(c.functions, d)
		}
	default:
		c.functions = append(c.functions, d)
	}
}

func (c *declCollector) collectGenDecl(d *dst.GenDecl) {
	switch d.Tok {
	case token.IMPORT:
		c.imports = append(c.imports, d)
	case token.CONST:
		if hasIota(d) {
			c.iotaConstDecls = append(c.iotaConstDecls, d)
		} else {
			c.constSpecs = append(c.constSpecs, d.Specs...)
		}
	case token.VAR:
		for _, spec := range d.Specs {
			if isBlankVarSpec(spec) {
				c.blankVarSpecs = append(c.blankVarSpecs, spec)
			} else {
				c.varSpecs = append(c.varSpecs, spec)
			}
		}
	case token.TYPE:
		c.typeDecls = append(c.typeDecls, d)
	}
}

func (c *declCollector) collectTypeNames(f *dst.File) {
	for _, decl := range f.Decls {
		gd, ok := decl.(*dst.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			if ts, ok := spec.(*dst.TypeSpec); ok {
				c.typeNames[ts.Name.Name] = true
			}
		}
	}
}

func (c *declCollector) sort() {
	sortSpecsByExportabilityThenName(c.constSpecs)
	sortSpecsByExportabilityThenName(c.varSpecs)

	for typeName := range c.constructors {
		sortFuncDeclsByName(c.constructors[typeName])
	}

	for typeName := range c.methodsByType {
		sortFuncDeclsByExportabilityThenLayer(c.methodsByType[typeName])
	}

	sortFuncDeclsByExportabilityThenLayer(c.orphanMethods)

	sortDeclsByExportabilityThenLayer(c.functions)
}

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

	return os.WriteFile(filePath, formatted, 0644)
}

func reorderDeclarations(f *dst.File) []dst.Decl {
	c := newDeclCollector()
	c.collect(f)
	c.sort()

	var result []dst.Decl

	result = append(result, c.imports...)
	result = appendInitFuncs(result, c.initFuncs)
	result = appendConstBlock(result, c.constSpecs)
	result = appendIotaConstBlocks(result, c.iotaConstDecls)
	result = appendVarBlock(result, c.blankVarSpecs, c.varSpecs)
	result = appendTypesWithMethods(result, c.typeDecls, c.constructors, c.methodsByType)
	result = appendOrphanMethods(result, c.orphanMethods)
	result = appendFunctions(result, c.functions)
	result = appendMainFunc(result, c.mainFunc)

	return result
}

func appendConstBlock(result []dst.Decl, constSpecs []dst.Spec) []dst.Decl {
	if len(constSpecs) == 0 {
		return result
	}
	constDecl := mergeSpecsIntoBlock(token.CONST, constSpecs)
	if len(result) > 0 {
		constDecl.Decs.Before = dst.EmptyLine
	}

	return append(result, constDecl)
}

func appendTypesWithMethods(result []dst.Decl, typeDecls []*dst.GenDecl, constructors, methodsByType map[string][]*dst.FuncDecl) []dst.Decl {
	splitTypes := splitAndGroupTypeDecls(typeDecls)

	for i, typeDecl := range splitTypes {
		if i == 0 && len(result) > 0 {
			setDeclSpacing(typeDecl, dst.EmptyLine)
		}
		result = append(result, typeDecl)

		gd, ok := typeDecl.(*dst.GenDecl)
		if !ok || len(gd.Specs) != 1 {
			continue
		}
		ts, ok := gd.Specs[0].(*dst.TypeSpec)
		if !ok {
			continue
		}

		typeName := ts.Name.Name
		for _, c := range constructors[typeName] {
			c.Decs.Before = dst.EmptyLine
			result = append(result, c)
		}
		for _, m := range methodsByType[typeName] {
			m.Decs.Before = dst.EmptyLine
			result = append(result, m)
		}
	}

	return result
}

func appendVarBlock(result []dst.Decl, blankVarSpecs, varSpecs []dst.Spec) []dst.Decl {
	allVarSpecs := append(blankVarSpecs, varSpecs...)
	if len(allVarSpecs) == 0 {
		return result
	}
	varDecl := mergeSpecsIntoBlock(token.VAR, allVarSpecs)
	if len(result) > 0 {
		varDecl.Decs.Before = dst.EmptyLine
	}

	return append(result, varDecl)
}

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

func mergeSpecsIntoBlock(tok token.Token, specs []dst.Spec) *dst.GenDecl {
	gd := &dst.GenDecl{
		Tok:   tok,
		Specs: specs,
	}
	if len(specs) > 1 {
		gd.Lparen = true
		addEmptyLinesBetweenSpecGroups(specs)
	}

	return gd
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

func splitAndGroupTypeDecls(typeDecls []*dst.GenDecl) []dst.Decl {
	var simpleTypes, funcInterfaces, nonFuncInterfaces, structs []dst.Decl

	for _, gd := range typeDecls {
		if len(gd.Specs) == 0 {
			continue
		}
		if len(gd.Specs) == 1 {
			categorizeType(gd, gd.Specs[0].(*dst.TypeSpec), &simpleTypes, &funcInterfaces, &nonFuncInterfaces, &structs)
			continue
		}
		for i, spec := range gd.Specs {
			ts := spec.(*dst.TypeSpec)
			newGd := &dst.GenDecl{
				Tok:   token.TYPE,
				Specs: []dst.Spec{spec},
			}
			if i == 0 {
				newGd.Decs = gd.Decs
			}
			categorizeType(newGd, ts, &simpleTypes, &funcInterfaces, &nonFuncInterfaces, &structs)
		}
	}

	return combineTypeGroups(simpleTypes, funcInterfaces, nonFuncInterfaces, structs)
}

func addEmptyLinesBetweenSpecGroups(specs []dst.Spec) {
	var lastGroup int
	for i, spec := range specs {
		vs, ok := spec.(*dst.ValueSpec)
		if !ok {
			continue
		}
		currentGroup := getSpecExportGroup(vs)
		if i == 0 {
			vs.Decs.Before = dst.NewLine
		} else if currentGroup != lastGroup {
			vs.Decs.Before = dst.EmptyLine
		} else {
			vs.Decs.Before = dst.NewLine
		}
		vs.Decs.After = dst.None
		lastGroup = currentGroup
	}
}

func categorizeType(gd *dst.GenDecl, ts *dst.TypeSpec, simpleTypes, funcInterfaces, nonFuncInterfaces, structs *[]dst.Decl) {
	switch t := ts.Type.(type) {
	case *dst.StructType:
		*structs = append(*structs, gd)
	case *dst.InterfaceType:
		if isFuncInterface(t) {
			*funcInterfaces = append(*funcInterfaces, gd)
		} else {
			*nonFuncInterfaces = append(*nonFuncInterfaces, gd)
		}
	default:
		*simpleTypes = append(*simpleTypes, gd)
	}
}

func combineTypeGroups(simpleTypes, funcInterfaces, nonFuncInterfaces, structs []dst.Decl) []dst.Decl {
	var result []dst.Decl
	result = appendTypeGroup(result, simpleTypes)
	result = appendTypeGroup(result, funcInterfaces)
	result = appendTypeGroup(result, nonFuncInterfaces)
	result = appendTypeGroup(result, structs)

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

func sortDeclsByExportabilityThenLayer(decls []dst.Decl) {
	exported, unexported := lo.FilterReject(decls, func(d dst.Decl, _ int) bool {
		if fn, ok := d.(*dst.FuncDecl); ok {
			return isExported(fn.Name.Name)
		}

		return true
	})
	sortDeclsByLayer(exported)
	sortDeclsByLayer(unexported)
	copy(decls, append(exported, unexported...))
}

func sortFieldsByTypeName(fields []*dst.Field) {
	sort.SliceStable(fields, func(i, j int) bool {
		return getFieldTypeName(fields[i]) < getFieldTypeName(fields[j])
	})
}

func sortFuncDeclsByExportabilityThenLayer(funcs []*dst.FuncDecl) {
	exported, unexported := lo.FilterReject(funcs, func(fn *dst.FuncDecl, _ int) bool {
		return isExported(fn.Name.Name)
	})
	sortFuncsByLayer(exported)
	sortFuncsByLayer(unexported)
	copy(funcs, append(exported, unexported...))
}

func sortSpecsByExportabilityThenName(specs []dst.Spec) {
	sort.SliceStable(specs, func(i, j int) bool {
		nameI := getSpecFirstName(specs[i])
		nameJ := getSpecFirstName(specs[j])
		groupI := getExportGroup(nameI)
		groupJ := getExportGroup(nameJ)
		if groupI != groupJ {
			return groupI < groupJ
		}

		return nameI < nameJ
	})
}

func appendFunctions(result []dst.Decl, functions []dst.Decl) []dst.Decl {
	for i, fn := range functions {
		if i == 0 && len(result) > 0 || i > 0 {
			setDeclSpacing(fn, dst.EmptyLine)
		}
		result = append(result, fn)
	}

	return result
}

func appendTypeGroup(result, group []dst.Decl) []dst.Decl {
	for i, d := range group {
		if i > 0 || (i == 0 && len(result) > 0) {
			setDeclSpacing(d, dst.EmptyLine)
		}
		result = append(result, d)
	}

	return result
}

func findConstructorType(fn *dst.FuncDecl, typeNames map[string]bool) string {
	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return ""
	}

	for _, result := range fn.Type.Results.List {
		typeName := extractTypeName(result.Type)
		if typeName == "" || !typeNames[typeName] {
			continue
		}
		if matchesConstructorPattern(fn.Name.Name, typeName) {
			return typeName
		}
	}

	return ""
}

func getExportGroup(name string) int {
	switch {
	case name == "_":
		return 0
	case isExported(name):
		return 1
	default:
		return 2
	}
}

func getFieldTypeName(field *dst.Field) string {
	return extractTypeName(field.Type)
}

func getReceiverTypeName(fn *dst.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}

	return extractTypeName(fn.Recv.List[0].Type)
}

func getSpecExportGroup(vs *dst.ValueSpec) int {
	if len(vs.Names) == 0 {
		return 0
	}
	name := vs.Names[0].Name
	switch {
	case name == "_":
		return 0
	case isExported(name):
		return 1
	default:
		return 2
	}
}

func hasIota(d *dst.GenDecl) bool {
	for _, spec := range d.Specs {
		vs, ok := spec.(*dst.ValueSpec)
		if !ok {
			continue
		}
		for _, val := range vs.Values {
			if containsIota(val) {
				return true
			}
		}
	}

	return false
}

func isFuncInterface(iface *dst.InterfaceType) bool {
	return iface.Methods != nil && len(iface.Methods.List) == 1 && isFuncType(iface.Methods.List[0].Type)
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

func sortDeclsByLayer(decls []dst.Decl) {
	funcs := lo.FilterMap(decls, func(d dst.Decl, _ int) (*dst.FuncDecl, bool) {
		fn, ok := d.(*dst.FuncDecl)

		return fn, ok
	})

	if len(funcs) <= 1 {
		return
	}

	funcNames := lo.SliceToMap(funcs, func(fn *dst.FuncDecl) (string, bool) {
		return fn.Name.Name, true
	})

	callGraph := buildCallGraph(funcs, funcNames)
	layers := assignLayers(callGraph, funcNames)

	sort.SliceStable(decls, func(i, j int) bool {
		fnI, okI := decls[i].(*dst.FuncDecl)
		fnJ, okJ := decls[j].(*dst.FuncDecl)
		if !okI || !okJ {
			return false
		}
		layerI, layerJ := layers[fnI.Name.Name], layers[fnJ.Name.Name]
		if layerI != layerJ {
			return layerI > layerJ
		}

		return fnI.Name.Name < fnJ.Name.Name
	})
}

func sortFuncsByLayer(funcs []*dst.FuncDecl) {
	if len(funcs) <= 1 {
		return
	}

	funcNames := lo.SliceToMap(funcs, func(fn *dst.FuncDecl) (string, bool) {
		return fn.Name.Name, true
	})

	callGraph := buildCallGraph(funcs, funcNames)
	layers := assignLayers(callGraph, funcNames)

	sort.SliceStable(funcs, func(i, j int) bool {
		layerI, layerJ := layers[funcs[i].Name.Name], layers[funcs[j].Name.Name]
		if layerI != layerJ {
			return layerI > layerJ
		}

		return funcs[i].Name.Name < funcs[j].Name.Name
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

func appendInitFuncs(result []dst.Decl, initFuncs []*dst.FuncDecl) []dst.Decl {
	for _, initFn := range initFuncs {
		initFn.Decs.Before = dst.EmptyLine
		result = append(result, initFn)
	}

	return result
}

func appendIotaConstBlocks(result []dst.Decl, iotaConstDecls []*dst.GenDecl) []dst.Decl {
	for _, constDecl := range iotaConstDecls {
		if len(result) > 0 {
			constDecl.Decs.Before = dst.EmptyLine
		}
		result = append(result, constDecl)
	}

	return result
}

func appendMainFunc(result []dst.Decl, mainFunc *dst.FuncDecl) []dst.Decl {
	if mainFunc == nil {
		return result
	}
	mainFunc.Decs.Before = dst.EmptyLine

	return append(result, mainFunc)
}

func appendOrphanMethods(result []dst.Decl, orphanMethods []*dst.FuncDecl) []dst.Decl {
	for _, m := range orphanMethods {
		if len(result) > 0 {
			m.Decs.Before = dst.EmptyLine
		}
		result = append(result, m)
	}

	return result
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

func assignLayers(callGraph map[string][]string, funcNames map[string]bool) map[string]int {
	g := simple.NewDirectedGraph()
	nameToID := make(map[string]int64)
	idToName := make(map[int64]string)

	var nextID int64
	for name := range funcNames {
		nameToID[name] = nextID
		idToName[nextID] = name
		g.AddNode(simple.Node(nextID))
		nextID++
	}

	for caller, callees := range callGraph {
		for _, callee := range callees {
			g.SetEdge(g.NewEdge(simple.Node(nameToID[caller]), simple.Node(nameToID[callee])))
		}
	}

	sccs := topo.TarjanSCC(g)

	sccID := make(map[string]int)
	for i, scc := range sccs {
		for _, node := range scc {
			sccID[idToName[node.ID()]] = i
		}
	}

	sccGraph := make(map[int][]int)
	for i := range sccs {
		sccGraph[i] = []int{}
	}
	for caller, callees := range callGraph {
		callerSCC := sccID[caller]
		for _, callee := range callees {
			calleeSCC := sccID[callee]
			if callerSCC != calleeSCC {
				sccGraph[callerSCC] = append(sccGraph[callerSCC], calleeSCC)
			}
		}
	}

	sccLayers := make(map[int]int)
	var computeSCCLayer func(scc int) int
	computeSCCLayer = func(scc int) int {
		if layer, ok := sccLayers[scc]; ok {
			return layer
		}
		maxChildLayer := -1
		for _, child := range sccGraph[scc] {
			childLayer := computeSCCLayer(child)
			if childLayer > maxChildLayer {
				maxChildLayer = childLayer
			}
		}
		sccLayers[scc] = maxChildLayer + 1

		return sccLayers[scc]
	}

	for i := range sccs {
		computeSCCLayer(i)
	}

	layers := make(map[string]int)
	for name := range funcNames {
		layers[name] = sccLayers[sccID[name]]
	}

	return layers
}

func buildCallGraph(funcs []*dst.FuncDecl, localFuncs map[string]bool) map[string][]string {
	graph := make(map[string][]string)

	for _, fn := range funcs {
		name := fn.Name.Name
		graph[name] = []string{}

		if fn.Body == nil {
			continue
		}

		dst.Inspect(fn.Body, func(n dst.Node) bool {
			call, ok := n.(*dst.CallExpr)
			if !ok {
				return true
			}
			ident, ok := call.Fun.(*dst.Ident)
			if !ok {
				return true
			}
			if localFuncs[ident.Name] && ident.Name != name {
				graph[name] = append(graph[name], ident.Name)
			}

			return true
		})
	}

	return graph
}

func detectGoVersion(filePath string) string {
	dir := filepath.Dir(filePath)
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(goModPath); err == nil {
			if mf, err := modfile.Parse(goModPath, data, nil); err == nil && mf.Go != nil {
				return "go" + mf.Go.Version
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func containsIota(expr dst.Expr) bool {
	switch e := expr.(type) {
	case *dst.Ident:
		return e.Name == "iota"
	case *dst.BinaryExpr:
		return containsIota(e.X) || containsIota(e.Y)
	case *dst.UnaryExpr:
		return containsIota(e.X)
	case *dst.ParenExpr:
		return containsIota(e.X)
	case *dst.CallExpr:
		for _, arg := range e.Args {
			if containsIota(arg) {
				return true
			}
		}
	}

	return false
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

func extractTypeName(expr dst.Expr) string {
	switch t := expr.(type) {
	case *dst.Ident:
		return t.Name
	case *dst.StarExpr:
		return extractTypeName(t.X)
	case *dst.SelectorExpr:
		return t.Sel.Name
	case *dst.IndexExpr:
		return extractTypeName(t.X)
	case *dst.IndexListExpr:
		return extractTypeName(t.X)
	}

	return ""
}

func getSpecFirstName(spec dst.Spec) string {
	switch s := spec.(type) {
	case *dst.ValueSpec:
		if len(s.Names) > 0 {
			return s.Names[0].Name
		}
	case *dst.TypeSpec:
		return s.Name.Name
	}

	return ""
}

func isBlankVarSpec(spec dst.Spec) bool {
	vs, ok := spec.(*dst.ValueSpec)
	if !ok {
		return false
	}

	return lo.ContainsBy(vs.Names, func(name *dst.Ident) bool {
		return name.Name == "_"
	})
}

func isExported(name string) bool {
	return len(name) > 0 && unicode.IsUpper(rune(name[0]))
}

func isFuncType(expr dst.Expr) bool {
	_, ok := expr.(*dst.FuncType)

	return ok
}

func isGeneratedFile(f *dst.File) bool {
	if len(f.Decs.Start) == 0 {
		return false
	}
	firstComment := f.Decs.Start[0]

	return strings.HasPrefix(firstComment, "// Code generated") ||
		strings.HasPrefix(firstComment, "// DO NOT EDIT") ||
		strings.HasPrefix(firstComment, "// GENERATED") ||
		strings.HasPrefix(firstComment, "// Autogenerated") ||
		strings.HasPrefix(firstComment, "// auto-generated") ||
		strings.HasPrefix(firstComment, "// Automatically generated")
}

func matchesConstructorPattern(funcName, typeName string) bool {
	suffix := strings.TrimPrefix(funcName, "New")
	if suffix == typeName {
		return true
	}
	if strings.HasPrefix(suffix, typeName) && len(suffix) > len(typeName) {
		nextChar := rune(suffix[len(typeName)])

		return !unicode.IsLower(nextChar)
	}

	return false
}

func newDeclCollector() *declCollector {
	return &declCollector{
		constructors:  make(map[string][]*dst.FuncDecl),
		methodsByType: make(map[string][]*dst.FuncDecl),
		typeNames:     make(map[string]bool),
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

func setDeclSpacing(decl dst.Decl, spacing dst.SpaceType) {
	switch d := decl.(type) {
	case *dst.GenDecl:
		d.Decs.Before = spacing
	case *dst.FuncDecl:
		d.Decs.Before = spacing
	}
}

func sortFieldsByName(fields []*dst.Field) {
	sort.SliceStable(fields, func(i, j int) bool {
		return fields[i].Names[0].Name < fields[j].Names[0].Name
	})
}

func sortFuncDeclsByName(funcs []*dst.FuncDecl) {
	sort.SliceStable(funcs, func(i, j int) bool {
		return funcs[i].Name.Name < funcs[j].Name.Name
	})
}
