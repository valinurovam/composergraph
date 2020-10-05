package composergraph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valinurovam/composergraph"
)

func TestGraph_ImplicitCreationOfRootPackage(t *testing.T) {
	graph := composergraph.NewGraph(nil)
	packages := graph.Packages()

	assert.Len(t, packages, 1)
	assert.True(t, graph.IsRootPackage(packages["__root"]))
}
