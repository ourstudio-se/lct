package json

import (
	"encoding/json"

	"github.com/ourstudio-se/lct/internal/deps"
)

type serializableDependencyNode struct {
	PackageName    string                       `json:"package_name"`
	PackageVersion string                       `json:"package_version"`
	Licenses       []string                     `json:"licenses"`
	Dependencies   []serializableDependencyNode `json:"dependencies,omitempty"`
}

func Marshal(node *deps.DependencyNode) ([]byte, error) {
	snode := asSerializableNode(node)
	return json.MarshalIndent(snode, "", "  ")
}

func Unmarshal(b []byte) (*deps.DependencyNode, error) {
	var snode serializableDependencyNode
	if err := json.Unmarshal(b, &snode); err != nil {
		return nil, err
	}

	return fromSerializableNode(snode, 0), nil
}

func asSerializableNode(node *deps.DependencyNode) serializableDependencyNode {
	var dependencies []serializableDependencyNode

	for _, child := range node.Children {
		dependencies = append(dependencies, asSerializableNode(child))
	}

	licenses := node.Licenses
	if len(licenses) == 0 {
		licenses = []string{}
	}

	return serializableDependencyNode{
		PackageName:    node.PackageName,
		PackageVersion: node.PackageVersion,
		Licenses:       licenses,
		Dependencies:   dependencies,
	}
}

func fromSerializableNode(snode serializableDependencyNode, level int) *deps.DependencyNode {
	var node *deps.DependencyNode
	if level == 0 {
		node = deps.NewGraph(snode.PackageName)
	} else {
		node = deps.New(snode.PackageName, snode.PackageVersion)
	}
	node.Licenses = snode.Licenses

	for _, dependency := range snode.Dependencies {
		node.Add(fromSerializableNode(dependency, level+1))
	}

	return node
}
