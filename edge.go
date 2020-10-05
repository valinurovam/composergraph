package composergraph

import "strings"

type Edge struct {
	sourcePackage     *PackageNode
	destPackage       *PackageNode
	versionConstraint string
}

func (d *Edge) VersionConstraint() string {
	return d.versionConstraint
}

func (d *Edge) DestPackage() *PackageNode {
	return d.destPackage
}

func (d *Edge) SourcePackage() *PackageNode {
	return d.sourcePackage
}

func (d *Edge) IsDevDependency() bool {
	if d.sourcePackage.data() == nil || d.sourcePackage.data().RequireDev == nil {
		return false
	}

	for k := range d.sourcePackage.data().RequireDev {
		if strings.ToLower(k) == d.destPackage.Name() {
			return true
		}
	}

	return false
}
