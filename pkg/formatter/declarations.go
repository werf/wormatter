package formatter

import (
	"go/token"

	"github.com/dave/dst"
)

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

func appendFunctions(result, functions []dst.Decl) []dst.Decl {
	for i, fn := range functions {
		if i == 0 && len(result) > 0 || i > 0 {
			setDeclSpacing(fn, dst.EmptyLine)
		}
		result = append(result, fn)
	}

	return result
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

func setDeclSpacing(decl dst.Decl, spacing dst.SpaceType) {
	switch d := decl.(type) {
	case *dst.GenDecl:
		d.Decs.Before = spacing
	case *dst.FuncDecl:
		d.Decs.Before = spacing
	}
}
