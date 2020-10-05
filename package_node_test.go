package composergraph_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valinurovam/composergraph"
)

func TestPackageNode_IsPhpRuntime(t *testing.T) {
	testData := []struct {
		name     string
		expected bool
	}{
		{"php", true},
		{"Php", true},
		{"PHP", true},
		{"php-foo", false},
		{"php/asdf", false},
		{"asdf", false},
		{"ext-asdf", false},
	}

	for _, data := range testData {
		data := data
		t.Run(data.name, func(t *testing.T) {
			node := composergraph.NewPackageNode(data.name, nil, nil)
			assert.Equal(t, data.expected, node.IsPhpRuntime())
		})
	}
}

func TestPackageNode_IsPhpExtension(t *testing.T) {
	testData := []struct {
		name     string
		expected bool
	}{
		{"ext-foo", true},
		{"Ext-asdf", true},
		{"EXT-bar", true},
		{"ext/foo", false},
		{"php", false},
		{"asdf", false},
	}

	for _, data := range testData {
		data := data
		t.Run(data.name, func(t *testing.T) {
			node := composergraph.NewPackageNode(data.name, nil, nil)
			assert.Equal(t, data.expected, node.IsPhpExtension())
		})
	}
}
