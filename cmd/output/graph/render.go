package graph

import (
	"fmt"
	"io"
	"strings"

	"github.com/jedib0t/go-pretty/v6/list"
	"github.com/ourstudio-se/lct/internal/deps"
)

func NewRenderer() func(io.Writer, *deps.DependencyNode) error {
	return func(w io.Writer, graph *deps.DependencyNode) error {
		lw := list.NewWriter()
		lw.SetStyle(list.StyleConnectedRounded)

		graph.Walk(func(node *deps.DependencyNode, level int) {
			for i := 0; i < level; i++ {
				lw.Indent()
			}

			if !node.IsRootNode() {
				licenses := strings.Join(node.Licenses, ", ")
				lw.AppendItem(fmt.Sprintf("%s@%s %s", node.PackageName, node.PackageVersion, licenses))
			} else {
				lw.AppendItem(node.PackageName)
			}

			for i := 0; i < level; i++ {
				lw.UnIndent()
			}
		})

		fmt.Fprintln(w, lw.Render())
		return nil
	}
}
