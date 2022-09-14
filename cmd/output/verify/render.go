package verify

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ourstudio-se/lct/internal/deps"
	"gopkg.in/yaml.v3"
)

type (
	OptionSet struct {
		cfg Config
	}
	Option func(*OptionSet) error
)

func WithYAML(yamlFile string) Option {
	return func(set *OptionSet) error {
		if _, err := os.Stat(yamlFile); os.IsNotExist(err) {
			return err
		}

		raw, err := os.ReadFile(yamlFile)
		if err != nil {
			return err
		}

		var cfg Config
		if err := yaml.Unmarshal(raw, &cfg); err != nil {
			return err
		}

		set.cfg = cfg
		return nil
	}
}

func NewRenderer(opts ...Option) func(io.Writer, *deps.DependencyNode) error {
	return func(writer io.Writer, node *deps.DependencyNode) error {
		set := OptionSet{
			cfg: Config{},
		}

		for _, opt := range opts {
			if err := opt(&set); err != nil {
				return err
			}
		}

		licenseSet := make(map[string]map[string]struct{})
		node.Walk(func(node *deps.DependencyNode, _ int) {
			if node.IsRootNode() {
				return
			}

			whitelisted := false
			for _, pkgsrc := range set.cfg.Rules.WhitelistedPackageSources {
				if strings.HasPrefix(node.PackageName, pkgsrc) {
					whitelisted = true
				}
			}

			if whitelisted {
				return
			}

			for _, license := range node.Licenses {
				if len(licenseSet[license]) == 0 {
					licenseSet[license] = make(map[string]struct{})
				}

				licenseSet[license][node.DisplayName()] = struct{}{}
			}
		})

		for license, usage := range licenseSet {
			found := false
			for _, valid := range set.cfg.Rules.AllowedLicenses {
				if valid == license {
					found = true
				}
			}

			if !found {
				var nodes []string
				for nodeName := range usage {
					nodes = append(nodes, nodeName)
				}

				nodeList := strings.Join(nodes, "\n\t")
				fmt.Fprintln(writer, fmt.Errorf("disallowed license found: %s\n  in\n\t%s", license, nodeList))
				os.Exit(1)
			}
		}

		return nil
	}
}
