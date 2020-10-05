package composergraph

import (
	"fmt"
	"strings"
)

type Graph struct {
	packages    map[string]*PackageNode
	rootPackage *PackageNode
}

func NewGraph(rootPackage *PackageNode) *Graph {
	g := &Graph{
		packages: make(map[string]*PackageNode),
	}
	if rootPackage != nil {
		g.packages[strings.ToLower(rootPackage.Name())] = rootPackage
		g.rootPackage = rootPackage
	} else {
		g.rootPackage = g.getOrCreatePackage("__root")
	}

	return g
}

func (g *Graph) getOrCreatePackage(packageName string) *PackageNode {
	if p, ok := g.packages[strings.ToLower(packageName)]; ok {
		return p
	}

	p := &PackageNode{name: packageName}
	g.packages[strings.ToLower(packageName)] = p
	return p
}

func (g *Graph) RootPackage() *PackageNode {
	return g.rootPackage
}

func (g *Graph) IsRootPackage(node *PackageNode) bool {
	return g.rootPackage == node
}

func (g *Graph) IsRootPackageName(name string) bool {
	return strings.EqualFold(g.rootPackage.Name(), name)
}

func (g *Graph) Packages() map[string]*PackageNode {
	return g.packages
}

func (g *Graph) Package(name string) *PackageNode {
	return g.packages[strings.ToLower(name)]
}

func (g *Graph) HasPackage(name string) bool {
	_, ok := g.packages[strings.ToLower(name)]

	return ok
}

func (g *Graph) CreatePackage(name string, composerData *composerJSON) (*PackageNode, error) {
	if _, ok := g.packages[strings.ToLower(name)]; ok {
		return nil, fmt.Errorf("the package '%s' already exists", name)
	}

	p := NewPackageNode(name, composerData, nil)
	g.packages[strings.ToLower(name)] = p

	return p, nil
}

func (g *Graph) Connect(packageA string, packageB string, versionConstraint string) {
	nodeA := g.getOrCreatePackage(packageA)
	nodeB := g.getOrCreatePackage(packageB)

	// Do not connect the same package
	if nodeA == nodeB {
		return
	}

	// Do not add duplicate connections.
	for _, edge := range nodeA.OutEdges() {
		if edge.DestPackage() == nodeB {
			return
		}
	}

	edge := &Edge{nodeA, nodeB, versionConstraint}
	nodeA.AddOutEdge(edge)
	nodeB.AddInEdge(edge)
}

func (g *Graph) GetAggregatePackageContaining(packageName string) *PackageNode {
	for _, p := range g.packages {
		if p.Replaces(packageName) {
			return p
		}
	}

	return nil
}
