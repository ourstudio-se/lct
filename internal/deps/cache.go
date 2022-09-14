package deps

import (
	"io"
)

type (
	Serializer interface {
		Serialize(*DependencyNode) ([]byte, error)
	}

	Deserializer interface {
		Deserialize([]byte) (*DependencyNode, error)
	}

	DependencyCache map[string]*DependencyNode

	CacheReader func() (DependencyCache, error)
	CacheWriter func(DependencyCache, *DependencyNode) error
)

func NewCacheReader(r io.Reader, deserializer Deserializer) CacheReader {
	return func() (DependencyCache, error) {
		raw, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		set := make(DependencyCache)

		if len(raw) == 0 {
			return set, nil
		}

		root, err := deserializer.Deserialize(raw)
		if err != nil {
			return nil, err
		}

		var nodes []*DependencyNode
		root.Walk(func(node *DependencyNode, level int) {
			if level == 0 {
				return
			}

			nodes = append(nodes, node)
		})

		for _, node := range nodes {
			set[node.DisplayName()] = node
		}

		return set, nil
	}
}

func NewCacheWriter(w io.Writer, serializer Serializer) CacheWriter {
	return func(cache DependencyCache, root *DependencyNode) error {
		root.Walk(func(node *DependencyNode, level int) {
			if node.IsRootNode() {
				return
			}

			if _, ok := cache[node.DisplayName()]; !ok {
				cache[node.DisplayName()] = &DependencyNode{
					PackageName:    node.PackageName,
					PackageVersion: node.PackageVersion,
					Licenses:       node.Licenses,
				}
			}
		})

		next := New(root.PackageName, root.PackageVersion)
		for _, node := range cache {
			next.Add(node)
		}

		raw, err := serializer.Serialize(next)
		if err != nil {
			return err
		}

		_, err = w.Write(raw)
		return err
	}
}

func NoOpReader() CacheReader {
	return func() (DependencyCache, error) {
		return nil, nil
	}
}

func NoOpWriter() CacheWriter {
	return func(_ DependencyCache, _ *DependencyNode) error {
		return nil
	}
}
