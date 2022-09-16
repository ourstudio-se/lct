package cmd

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/ourstudio-se/lct/cmd/input"
	"github.com/ourstudio-se/lct/cmd/output"
	"github.com/ourstudio-se/lct/internal/deps"
	"github.com/ourstudio-se/lct/internal/npm"
	"github.com/spf13/cobra"
)

const (
	withoutDevDepsArgName  = "without-dev"
	npmjsHTTPClientTimeout = time.Second * 3
)

var (
	runNpmCmd = &cobra.Command{
		Use: "npm",
		RunE: func(cmd *cobra.Command, args []string) error {
			source, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return errors.New("invalid input")
			}

			r, w, close, err := input.ParseCacheArgs(cmd)
			if err != nil {
				return err
			}
			defer close()

			withoutDevDeps, err := cmd.Flags().GetBool(withoutDevDepsArgName)
			if err != nil {
				return err
			}

			graph, err := npm.Parse(string(source),
				npm.WithCache(r, w),
				npm.WithParallelization(deps.DefaultParallelizationLevel),
				npm.WithDevelopmentDependencies(!withoutDevDeps),
				npm.WithLicenseResolver(npm.NpmJsLicenseResolver(
					deps.WithHTTPClient(&http.Client{
						Timeout: npmjsHTTPClientTimeout,
					}),
					deps.WithBaseContext(cmd.Context()))))
			if err != nil {
				return err
			}

			return output.Render(cmd, graph)
		},
	}
)

func init() {
	rootCmd.AddCommand(runNpmCmd)

	runNpmCmd.PersistentFlags().Bool(withoutDevDepsArgName, false, "exclude development dependencies")
}
