package formatter

import (
	"go/token"
	"sort"

	"github.com/dave/dst"
	"github.com/samber/lo"
)

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

func sortFuncDeclsByExportabilityThenLayer(funcs []*dst.FuncDecl) {
	exported, unexported := lo.FilterReject(funcs, func(fn *dst.FuncDecl, _ int) bool {
		return isExported(fn.Name.Name)
	})
	sortFuncsByLayer(exported)
	sortFuncsByLayer(unexported)
	copy(funcs, append(exported, unexported...))
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

	var result []dst.Decl
	result = appendTypeGroup(result, simpleTypes)
	result = appendTypeGroup(result, funcInterfaces)
	result = appendTypeGroup(result, nonFuncInterfaces)
	result = appendTypeGroup(result, structs)

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

func sortFieldsByName(fields []*dst.Field) {
	sort.SliceStable(fields, func(i, j int) bool {
		return fields[i].Names[0].Name < fields[j].Names[0].Name
	})
}

func sortFieldsByTypeName(fields []*dst.Field) {
	sort.SliceStable(fields, func(i, j int) bool {
		return getFieldTypeName(fields[i]) < getFieldTypeName(fields[j])
	})
}

func sortFuncDeclsByName(funcs []*dst.FuncDecl) {
	sort.SliceStable(funcs, func(i, j int) bool {
		return funcs[i].Name.Name < funcs[j].Name.Name
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
