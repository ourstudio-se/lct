package deps

import (
	"fmt"
)

const LicenseUnknown = "#Unkonwn"

type DependencyNode struct {
	PackageName    string
	PackageVersion string
	Licenses       []string
	isRoot         bool
	size           uint64
	parent         *DependencyNode
	Children       []*DependencyNode
}

func NewGraph(source string) *DependencyNode {
	return &DependencyNode{
		PackageName:    source,
		PackageVersion: "",
		Licenses:       nil,
		isRoot:         true,
		size:           1,
	}
}

func New(name, version string) *DependencyNode {
	return &DependencyNode{
		PackageName:    name,
		PackageVersion: version,
		isRoot:         false,
		size:           1,
	}
}

func (n *DependencyNode) DisplayName() string {
	if n.PackageVersion != "" {
		return fmt.Sprintf("%s@%s", n.PackageName, n.PackageVersion)
	}

	return n.PackageName
}

func (n *DependencyNode) IsRootNode() bool {
	return n.isRoot
}

func (n *DependencyNode) Size() uint64 {
	return n.size
}

func (n *DependencyNode) Add(node *DependencyNode) {
	n.Children = append(n.Children, node)
	node.parent = n

	increaseSize(n)
}

func (n *DependencyNode) Walk(fn func(*DependencyNode, int)) {
	if n == nil {
		return
	}

	collectNode(n, 0, fn)
}

func increaseSize(n *DependencyNode) {
	if n == nil {
		return
	}

	n.size++
	increaseSize(n.parent)
}

func collectNode(n *DependencyNode, level int, fn func(*DependencyNode, int)) {
	fn(n, level)

	for _, child := range n.Children {
		collectNode(child, level+1, fn)
	}
}
