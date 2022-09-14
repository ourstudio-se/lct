package input

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ourstudio-se/lct/internal/deps"
	"github.com/ourstudio-se/lct/internal/json"
	"github.com/spf13/cobra"
)

const defaultCacheFile = "/lct/gomod.cache"

type (
	JsonDeserializer func([]byte) (*deps.DependencyNode, error)
	JsonSerializer   func(node *deps.DependencyNode) ([]byte, error)
)

func ParseCacheArgs(cmd *cobra.Command) (deps.CacheReader, deps.CacheWriter, func(), error) {
	noCache, err := cmd.Flags().GetBool(noCacheArgName)
	if err != nil {
		return nil, nil, nil, err
	}

	if noCache {
		return deps.NoOpReader(), deps.NoOpWriter(), func() {}, nil
	}

	cacheFile, err := cmd.Flags().GetString(cacheFileArgName)
	if err != nil {
		return nil, nil, nil, err
	}
	if cacheFile == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return nil, nil, nil, err
		}
		cacheFile = cacheDir + defaultCacheFile
	}

	var handle *os.File
	if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
		dir := filepath.Dir(cacheFile)
		_ = os.MkdirAll(dir, os.ModePerm)
	}

	handle, err = os.OpenFile(cacheFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open cache file: %w", err)
	}

	deserializer := JsonDeserializer(func(b []byte) (*deps.DependencyNode, error) {
		graph, err := json.Unmarshal(b)
		if err != nil {
			return nil, err
		}

		return graph, nil
	})

	serializer := JsonSerializer(func(d *deps.DependencyNode) ([]byte, error) {
		if err := handle.Truncate(0); err != nil {
			return nil, err
		}

		if _, err := handle.Seek(0, 0); err != nil {
			return nil, err
		}

		return json.Marshal(d)
	})

	return deps.NewCacheReader(handle, deserializer),
		deps.NewCacheWriter(handle, serializer),
		func() {
			_ = handle.Close()
		},
		nil
}

func (fn JsonDeserializer) Deserialize(b []byte) (*deps.DependencyNode, error) {
	return fn(b)
}

func (fn JsonSerializer) Serialize(d *deps.DependencyNode) ([]byte, error) {
	return fn(d)
}
