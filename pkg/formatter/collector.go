package formatter

import (
	"go/token"
	"strings"

	"github.com/dave/dst"
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

func newDeclCollector() *declCollector {
	return &declCollector{
		constructors:  make(map[string][]*dst.FuncDecl),
		methodsByType: make(map[string][]*dst.FuncDecl),
		typeNames:     make(map[string]bool),
	}
}
