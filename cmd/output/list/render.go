package list

import (
	"fmt"
	"io"
	"sort"

	"github.com/fatih/color"
	"github.com/ourstudio-se/lct/internal/deps"
	"github.com/rodaine/table"
)

func NewRenderer() func(io.Writer, *deps.DependencyNode) error {
	return func(w io.Writer, graph *deps.DependencyNode) error {
		licenseSet := make(map[string]struct{})
		graph.Walk(func(node *deps.DependencyNode, level int) {
			if len(node.Licenses) == 0 {
				return
			}

			for _, l := range node.Licenses {
				licenseSet[l] = struct{}{}
			}
		})

		licenses := make([]string, 0, len(licenseSet))
		for l := range licenseSet {
			licenses = append(licenses, l)
		}

		sort.Strings(licenses)

		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		fmt.Fprintln(w)

		tbl := table.New("License", "Link").
			WithWriter(w).
			WithHeaderFormatter(headerFmt).
			WithFirstColumnFormatter(columnFmt)

		for _, license := range licenses {
			link := fmt.Sprintf("https://opensource.org/licenses/%s", license)
			if license == deps.LicenseUnknown {
				link = ""
			}

			tbl.AddRow(license, link)
		}

		tbl.Print()
		fmt.Fprintln(w)

		return nil
	}
}
