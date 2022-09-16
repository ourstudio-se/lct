package gomod

import (
	"context"
	"errors"
	"strings"

	"github.com/ourstudio-se/lct/internal/deps"
	"github.com/schollz/progressbar/v3"
)

const (
	htmlLicenseAttrID    = "data-test-id"
	htmlLicenseAttrValue = "UnitHeader-licenses"
)

type (
	ParseOptionSet struct {
		cacheReader  deps.CacheReader
		cacheWriter  deps.CacheWriter
		resolver     deps.LicenseResolverFunc
		numOfWorkers uint64
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

func Parse(source string, opts ...ParseOption) (*deps.DependencyNode, error) {
	optionSet := ParseOptionSet{
		cacheReader:  deps.NoOpReader(),
		cacheWriter:  deps.NoOpWriter(),
		resolver:     deps.NoOpLicenseResolver(),
		numOfWorkers: 1,
	}

	for _, opt := range opts {
		opt(&optionSet)
	}

	cached, err := optionSet.cacheReader()
	if err != nil {
		return nil, err
	}

	graph, err := parseGraph(source)
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

func parseGraph(source string) (*deps.DependencyNode, error) {
	source = strings.TrimSpace(source)
	lines := strings.Split(source, "\n")

	var sourceModule string
	dependencyMap := make(map[string][]string)
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, errors.New("invalid data")
		}

		if sourceModule == "" {
			sourceModule = fields[0]
		}

		dependencyMap[fields[0]] = append(dependencyMap[fields[0]], fields[1])
	}

	graph := deps.NewGraph(sourceModule)
	block := make(map[string]struct{})
	buildDependencyGraph(graph, dependencyMap, block)
	return graph, nil
}

func buildDependencyGraph(parent *deps.DependencyNode, dmap map[string][]string, block map[string]struct{}) {
	list, ok := dmap[parent.DisplayName()]
	if !ok {
		return
	}

	for _, x := range list {
		packageName, packageVersion := parseDependencyModule(x)
		pm := deps.New(packageName, packageVersion)
		parent.Add(pm)

		if _, ok := block[pm.DisplayName()]; !ok {
			block[pm.DisplayName()] = struct{}{}
			buildDependencyGraph(pm, dmap, block)
		}
	}
}

func parseDependencyModule(raw string) (string, string) {
	parts := strings.Split(raw, "@")
	return parts[0], parts[1]
}
