package output

import (
	"io"

	"github.com/ourstudio-se/lct/cmd/output/graph"
	"github.com/ourstudio-se/lct/cmd/output/json"
	"github.com/ourstudio-se/lct/cmd/output/list"
	"github.com/ourstudio-se/lct/cmd/output/verify"
	"github.com/ourstudio-se/lct/internal/deps"
	"github.com/spf13/cobra"
)

func Render(cmd *cobra.Command, node *deps.DependencyNode) error {
	renderer := parseOutputType(cmd)
	return renderer(cmd.OutOrStdout(), node)
}

func parseOutputType(cmd *cobra.Command) func(io.Writer, *deps.DependencyNode) error {
	isOutputGraph, err := cmd.Flags().GetBool(GraphArg)
	if err == nil && isOutputGraph {
		return graph.NewRenderer()
	}

	isOutputJSON, err := cmd.Flags().GetBool(JsonArg)
	if err == nil && isOutputJSON {
		return json.NewRenderer()
	}

	verificationConfig, err := cmd.Flags().GetString(VerifyArg)
	if err == nil && verificationConfig != "" {
		return verify.NewRenderer(verify.WithYAML(verificationConfig))
	}

	return list.NewRenderer()
}
