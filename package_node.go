package composergraph

import (
	"strings"
)

type PackageNode struct {
	repositoryID    string
	name            string
	composerData    *composerJSON
	version         string
	sourceReference string
	inEdges         []*Edge
	outEdges        []*Edge
	attributes      map[string]string
}

func NewPackageNode(name string, composerData *composerJSON, attributes map[string]string) *PackageNode {
	if attributes == nil {
		attributes = make(map[string]string)
	}

	return &PackageNode{name: name, composerData: composerData, attributes: attributes}
}

func (n *PackageNode) SetRepositoryID(repositoryID string) {
	n.repositoryID = repositoryID
}

func (n *PackageNode) IsPhpExtension() bool {
	return strings.HasPrefix(strings.ToLower(n.QualifiedName()), "ext-")
}

func (n *PackageNode) IsPhpRuntime() bool {
	return strings.ToLower(n.QualifiedName()) == "php"
}

func (n *PackageNode) QualifiedName() string {
	if !n.HasAttribute("dir") {
		return n.name
	}

	repoID := "packagist"
	if n.repositoryID != "" {
		repoID = n.repositoryID
	}

	return strings.Join([]string{repoID, "__", n.name}, ".")
}

func (n *PackageNode) SetAttribute(key string, value string) {
	n.attributes[key] = value
}

func (n *PackageNode) HasAttribute(key string) bool {
	_, ok := n.attributes[key]

	return ok
}

func (n *PackageNode) Attribute(key string) (string, bool) {
	value, ok := n.attributes[key]

	return value, ok
}

func (n *PackageNode) SetVersion(version string) {
	n.version = version
}

func (n *PackageNode) Version() string {
	return n.version
}

func (n *PackageNode) data() *composerJSON {
	return n.composerData
}

func (n *PackageNode) Name() string {
	return n.name
}

func (n *PackageNode) SetSourceReference(sourceReference string) {
	n.sourceReference = sourceReference
}

func (n *PackageNode) SourceReference() string {
	return n.sourceReference
}

func (n *PackageNode) InEdges() []*Edge {
	return n.inEdges
}

func (n *PackageNode) OutEdges() []*Edge {
	return n.outEdges
}

func (n *PackageNode) AddInEdge(edge *Edge) {
	n.inEdges = append(n.inEdges, edge)
}

func (n *PackageNode) AddOutEdge(edge *Edge) {
	n.outEdges = append(n.outEdges, edge)
}

func (n *PackageNode) Replaces(packageName string) bool {
	if n.composerData == nil || n.composerData.Replace == nil {
		return false
	}

	for k := range n.composerData.Replace {
		if strings.ToLower(k) == packageName {
			return true
		}
	}

	return false
}
