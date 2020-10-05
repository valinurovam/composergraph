package composergraph

import (
	"io/ioutil"
	"os"
	os_path "path"
	"strings"

	"github.com/pkg/errors"
)

type Analyzer struct {
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(path string) (*Graph, error) {
	path = os_path.Clean(path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	jsonFilePath := os_path.Join(path, "composer.json")
	jsonLockPath := os_path.Join(path, "composer.lock")

	if _, err := os.Stat(jsonFilePath); os.IsNotExist(err) {
		depGraph := NewGraph(nil)
		depGraph.RootPackage().SetAttribute("dir", path)

		return depGraph, err
	}

	composerJSONData, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read composer.json composerData")
	}

	composerLockData, err := ioutil.ReadFile(jsonLockPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrap(err, "unable to read composer.lock composerData")
	}

	return a.AnalyzeComposerData(composerJSONData, composerLockData, path)
}

func (a *Analyzer) AnalyzeComposerData(composerJsonData []byte, composerLockData []byte, path string) (*Graph, error) {
	rootComposerJSON, err := parseComposerJSON(composerJsonData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse composer.json")
	}

	if rootComposerJSON == nil {
		return nil, errors.New("unexpected nil rootComposerJSON")
	}

	if rootComposerJSON.Name == "" {
		rootComposerJSON.Name = "__root"
	}

	// If there is no composer.lock file, then either the project has no
	// dependencies, or the dependencies were not installed.
	if composerLockData == nil {
		if a.hasDependencies(rootComposerJSON) {
			return nil, errors.New("missing lock file")
		}

		depGraph := NewGraph(NewPackageNode(rootComposerJSON.Name, rootComposerJSON, nil))
		depGraph.RootPackage().SetAttribute("dir", path)

		// Connect built-in dependencies for example on the PHP version, or
		// on PHP extensions. For these, composer does not create a composer.lock.

		if rootComposerJSON.Require != nil {
			for name, version := range rootComposerJSON.Require {
				a.connect(depGraph, rootComposerJSON.Name, name, version)
			}
		}

		if rootComposerJSON.RequireDev != nil {
			for name, version := range rootComposerJSON.RequireDev {
				a.connect(depGraph, rootComposerJSON.Name, name, version)
			}
		}

		return depGraph, nil
	}

	depGraph := NewGraph(NewPackageNode(rootComposerJSON.Name, rootComposerJSON, nil))
	depGraph.RootPackage().SetAttribute("dir", path)

	vendorPath := os_path.Join(path, "vendor")
	if rootComposerJSON.Config != nil && rootComposerJSON.Config.VendorDir != "" {
		vendorPath = os_path.Join(path, rootComposerJSON.Config.VendorDir)
	}

	rootComposerLock, err := parseComposerLock(composerLockData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse composer.lock")
	}

	// Add regular packages.
	if rootComposerLock.Packages != nil {
		if err := a.addPackages(depGraph, rootComposerLock.Packages, vendorPath); err != nil {
			return nil, err
		}
	}

	// Add development packages.
	if rootComposerLock.PackagesDev != nil {
		if err := a.addPackages(depGraph, rootComposerLock.PackagesDev, vendorPath); err != nil {
			return nil, err
		}
	}

	for _, node := range depGraph.Packages() {
		packageData := node.data()

		if packageData != nil && packageData.Require != nil {
			for name, version := range packageData.Require {
				a.connect(depGraph, packageData.Name, name, version)
			}
		}

		if packageData != nil && packageData.RequireDev != nil {
			for name, version := range packageData.RequireDev {
				a.connect(depGraph, packageData.Name, name, version)
			}
		}
	}

	return depGraph, nil
}

func (a *Analyzer) addPackages(depGraph *Graph, packages []*composerJSON, vendorPath string) error {
	for _, packageData := range packages {
		if depGraph.IsRootPackageName(packageData.Name) || depGraph.HasPackage(packageData.Name) {
			continue
		}

		p, err := depGraph.CreatePackage(packageData.Name, packageData)
		if err != nil {
			return errors.Wrap(err, "unable to create package")
		}

		p.SetAttribute("dir", os_path.Join(vendorPath, packageData.Name))
		a.processLockedData(depGraph, packageData)
	}

	return nil
}

func (a *Analyzer) processLockedData(depGraph *Graph, lockedData *composerJSON) {
	packageName := ""
	if lockedData.Name != "" {
		packageName = lockedData.Name
	} else if lockedData.Package != "" {
		packageName = lockedData.Package
	}

	if packageName == "" {
		return
	}

	p := depGraph.Package(packageName)
	if p == nil {
		return
	}

	p.SetVersion(lockedData.Version)

	if lockedData.Source != nil &&
		lockedData.Source.Reference != "" &&
		lockedData.Version != lockedData.Source.Reference {
		p.SetSourceReference(lockedData.Source.Reference)
	}
}

func (a *Analyzer) hasDependencies(json *composerJSON) bool {
	if a.hasUserlandDependency(json.Require) {
		return true
	}

	if a.hasUserlandDependency(json.RequireDev) {
		return true
	}

	return false
}

func (a *Analyzer) hasUserlandDependency(require map[string]string) bool {
	if len(require) == 0 {
		return false
	}

	for name := range require {
		name := strings.ToLower(name)
		if name == "php" {
			continue
		}

		if strings.HasPrefix(name, "ext-") {
			continue
		}

		return true
	}

	return false
}

func (a *Analyzer) connect(depGraph *Graph, srcName, dstName, version string) {
	// If the dest package is available, just connect it.
	if depGraph.HasPackage(dstName) {
		depGraph.Connect(srcName, dstName, version)

		return
	}

	// If the dest package is not available, let's check to see if there is
	// some aggregate package that replaces our dest package, and connect to
	// this package.
	aggregatePackage := depGraph.GetAggregatePackageContaining(dstName)
	if aggregatePackage != nil {
		depGraph.Connect(srcName, aggregatePackage.Name(), version)

		return
	}

	// If we reach this, we have stumbled upon a package that is only available
	// if the source package is installed with dev dependencies. We still add
	// the connection, but we will not have any data about the dest package.
	depGraph.Connect(srcName, dstName, version)
}
