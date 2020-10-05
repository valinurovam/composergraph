package composergraph_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/valinurovam/composergraph"
)

func TestAnalyzer_UpperCasePhp(t *testing.T) {
	a := composergraph.NewAnalyzer()
	composerData := `
{
    "require": {
        "PHP": ">= 5.2",
        "ExT-foo": "dev-master"
    }
}
`
	dgraph, err := a.AnalyzeComposerData([]byte(composerData), nil, "")
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, dgraph.Packages(), 3)
}

func TestAnalyzer_ErrWhenLockFileIsMissing(t *testing.T) {
	a := composergraph.NewAnalyzer()
	composerData := `
{
    "require": {
        "test/foo": "1.*"
    }
}
`
	_, err := a.AnalyzeComposerData([]byte(composerData), nil, "")

	assert.Error(t, err)
}

func TestAnalyzer_RedundantPackage(t *testing.T) {
	a := composergraph.NewAnalyzer()

	redundantJSON, _ := ioutil.ReadFile("test/fixture/redundant_package.json")
	redundantLock, _ := ioutil.ReadFile("test/fixture/redundant_package.lock")

	expected, _ := ioutil.ReadFile("test/fixture/redundant_package_graph.txt")

	graph, err := a.AnalyzeComposerData(redundantJSON, redundantLock, "")
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, string(expected), dumpGraph(graph))
}

func TestAnalyzer_DuplicatePackageBug(t *testing.T) {
	a := composergraph.NewAnalyzer()

	json, _ := ioutil.ReadFile("test/fixture/prophecy_composer.json")
	lock, _ := ioutil.ReadFile("test/fixture/prophecy_composer.lock")

	expected, _ := ioutil.ReadFile("test/fixture/prophecy_graph.txt")

	graph, err := a.AnalyzeComposerData(json, lock, "")
	if err != nil {
		t.Fatal(err)
	}

	assert.EqualValues(t, string(expected), dumpGraph(graph))
}

func TestAnalyzer_Analyze(t *testing.T) {
	testFiles := []string{
		"no_deps",
		"regular_deps",
		"open_source_lib",
		"dev_stability",
		"aggregate_package",
		"unavailable_dev_package",
		"dep_only_on_php_and_ext",
		"mixed_case",
	}

	for _, testFile := range testFiles {
		testFile := testFile
		t.Run(testFile, func(t *testing.T) {
			t.Parallel()
			graphPath := "test/fixture/" + testFile + "_graph.txt"
			if _, err := os.Stat(graphPath); os.IsNotExist(err) {
				t.Fatalf("the dependency graph %s does not exist", graphPath)
			}

			depGraph, err := analyze(testFile)
			if err != nil {
				t.Fatal(err)
			}

			expected, err := ioutil.ReadFile(graphPath)
			if err != nil {
				t.Fatal(err)
			}

			assert.EqualValues(t, string(expected), dumpGraph(depGraph))
		})
	}
}

func dumpGraph(graph *composergraph.Graph) string {
	var packages []*composergraph.PackageNode

	for _, p := range graph.Packages() {
		packages = append(packages, p)
	}

	sort.Slice(packages, func(i, j int) bool {
		nodeA := packages[i]
		nodeB := packages[j]

		if graph.IsRootPackage(nodeA) {
			return true
		}

		if graph.IsRootPackage(nodeB) {
			return false
		}

		return strings.Compare(nodeA.Name(), nodeB.Name()) < 0
	})

	var txt strings.Builder

	for _, p := range packages {
		if txt.Len() > 0 {
			txt.WriteString("\n\n")
		}

		name := p.Name()
		if graph.IsRootPackage(p) {
			name += " (Root)"
		}

		txt.WriteString(name)
		txt.WriteString("\n")
		txt.WriteString(strings.Repeat("=", len(name)))
		txt.WriteString("\n")
		txt.WriteString("Version: ")

		if p.Version() != "" {
			txt.WriteString(p.Version())
		} else {
			txt.WriteString("<null>")
		}

		txt.WriteString("\n")

		if p.SourceReference() != "" {
			txt.WriteString("Source-Reference: ")
			txt.WriteString(p.SourceReference())
			txt.WriteString("\n")
		}

		if len(p.OutEdges()) > 0 {
			edges := p.OutEdges()
			sort.Slice(edges, func(i, j int) bool {
				return strings.Compare(edges[i].DestPackage().Name(), edges[j].DestPackage().Name()) < 0
			})

			for _, edge := range edges {
				txt.WriteString("-> ")
				txt.WriteString(edge.DestPackage().Name())
				txt.WriteString("\n")
			}
		}
	}

	return txt.String()
}

func analyze(configName string) (*composergraph.Graph, error) {
	basePath := "test/fixture/" + configName
	configPath := basePath + "_composer.json"
	lockPath := basePath + "_composer.lock"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "the root config %s does not exist", configPath)
	}

	a := composergraph.NewAnalyzer()

	if _, err := os.Stat(lockPath); err == nil {
		configJSON, _ := ioutil.ReadFile(configPath)
		lock, _ := ioutil.ReadFile(lockPath)

		return a.AnalyzeComposerData(configJSON, lock, "")
	}

	dir, err := ioutil.TempDir("", "composerTests")
	if err != nil {
		return nil, err
	}

	defer os.RemoveAll(dir)

	cmdCp := exec.Command("cp", configPath, path.Join(dir, "composer.json"))
	if err := cmdCp.Run(); err != nil {
		return nil, err
	}

	cmdComposer := exec.Command("composer", "install", "--dev")
	cmdComposer.Dir = dir

	if err := cmdComposer.Run(); err != nil {
		return nil, err
	}

	return a.Analyze(dir)
}
