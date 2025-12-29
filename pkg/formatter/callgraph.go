package formatter

import (
	"github.com/dave/dst"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

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
