package cmd

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/ourstudio-se/lct/cmd/input"
	"github.com/ourstudio-se/lct/cmd/output"
	"github.com/ourstudio-se/lct/internal/gomod"
	"github.com/spf13/cobra"
)

var (
	defaultParallelizationNum = 50

	runGoModCmd = &cobra.Command{
		Use: "gomod",
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

			graph, err := gomod.Parse(string(source),
				gomod.WithCache(r, w),
				gomod.WithParallelization(uint64(defaultParallelizationNum)),
				gomod.WithLicenseResolver(gomod.GoDevLicenseResolver(
					gomod.WithHTTPClient(&http.Client{
						Timeout: time.Second * 2,
					}),
					gomod.WithBaseContext(cmd.Context()))))
			if err != nil {
				return err
			}

			return output.Render(cmd, graph)
		},
	}
)

func init() {
	rootCmd.AddCommand(runGoModCmd)
}
