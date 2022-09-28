package npm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ourstudio-se/lct/internal/deps"
	"github.com/schollz/progressbar/v3"
)

const PackageTypeName string = "npm"

type (
	ParseOptionSet struct {
		cacheReader                 deps.CacheReader
		cacheWriter                 deps.CacheWriter
		resolver                    deps.LicenseResolverFunc
		withDevelopmentDependencies bool
		numOfWorkers                uint64
	}
	ParseOption func(*ParseOptionSet)
)

func WithCache(r deps.CacheReader, w deps.CacheWriter) ParseOption {
	return func(set *ParseOptionSet) {
		set.cacheReader = r
		set.cacheWriter = w
	}
}

func WithLicenseResolver(resolver deps.LicenseResolverFunc) ParseOption {
	return func(set *ParseOptionSet) {
		set.resolver = resolver
	}
}

func WithParallelization(numOfWorkers uint64) ParseOption {
	return func(set *ParseOptionSet) {
		set.numOfWorkers = numOfWorkers
	}
}

func WithDevelopmentDependencies(withDevDeps bool) ParseOption {
	return func(set *ParseOptionSet) {
		set.withDevelopmentDependencies = withDevDeps
	}
}

func Parse(source string, opts ...ParseOption) (*deps.DependencyNode, error) {
	optionSet := ParseOptionSet{
		cacheReader:                 deps.NoOpReader(),
		cacheWriter:                 deps.NoOpWriter(),
		resolver:                    deps.NoOpLicenseResolver(),
		withDevelopmentDependencies: false,
		numOfWorkers:                1,
	}

	for _, opt := range opts {
		opt(&optionSet)
	}

	cached, err := optionSet.cacheReader()
	if err != nil {
		return nil, err
	}

	graph, err := parseGraph([]byte(source), optionSet.withDevelopmentDependencies)
	if err != nil {
		return nil, err
	}

	worker := func(jobs <-chan *deps.DependencyNode, results chan<- struct{}) {
		for node := range jobs {
			c, ok := cached[node.DisplayName()]
			if !ok {
				licenses, _ := optionSet.resolver.Resolve(context.TODO(), node)
				node.Licenses = licenses
			} else {
				node.Licenses = c.Licenses
			}

			results <- struct{}{}
		}
	}

	jobs := make(chan *deps.DependencyNode, graph.Size())
	results := make(chan struct{}, graph.Size())

	for i := uint64(0); i < optionSet.numOfWorkers; i++ {
		go worker(jobs, results)
	}

	graph.Walk(func(node *deps.DependencyNode, level int) {
		jobs <- node
	})
	close(jobs)

	bar := progressbar.Default(int64(graph.Size()))

	for x := uint64(0); x < graph.Size(); x++ {
		<-results
		_ = bar.Add(1)
	}

	if err := optionSet.cacheWriter(cached, graph); err != nil {
		return nil, err
	}

	return graph, nil
}

type (
	graphNode struct {
		Version      string                `json:"version"`
		Requires     map[string]string     `json:"requires"`
		Dependencies map[string]*graphNode `json:"dependencies"`
		IsDevDep     bool                  `json:"dev"`
	}

	sourceGraph struct {
		Name         string                `json:"name"`
		Dependencies map[string]*graphNode `json:"dependencies"`
	}
)

func parseGraph(source []byte, withDevelopmentDependencies bool) (*deps.DependencyNode, error) {
	var sg sourceGraph
	if err := json.Unmarshal(source, &sg); err != nil {
		return nil, err
	}

	graph := deps.NewGraph(sg.Name, PackageTypeName)
	buildDependencyGraph(withDevelopmentDependencies, graph, sg.Dependencies)
	return graph, nil
}

func buildDependencyGraph(withDevelopmentDependencies bool, parent *deps.DependencyNode, dmap map[string]*graphNode) {
	if len(dmap) == 0 {
		return
	}

	skippable := make(map[string]struct{})

	for _, d := range dmap {
		if len(d.Dependencies) == 0 {
			d.Dependencies = make(map[string]*graphNode)
		}

		if len(d.Requires) > 0 {
			for rp, rv := range d.Requires {
				pkg := fmt.Sprintf("%s@%s", rp, rv)
				skippable[pkg] = struct{}{}
				d.Dependencies[rp] = &graphNode{
					Version: rv,
				}
			}
		}
	}

	for packageName, d := range dmap {
		pkg := fmt.Sprintf("%s@%s", packageName, d.Version)

		if _, skip := skippable[pkg]; skip {
			continue
		}

		if d.IsDevDep && !withDevelopmentDependencies {
			continue
		}

		pm := deps.New(packageName, d.Version)
		parent.Add(pm)

		buildDependencyGraph(withDevelopmentDependencies, pm, d.Dependencies)
	}
}
