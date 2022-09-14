package json

import (
	"fmt"
	"io"

	"github.com/ourstudio-se/lct/internal/deps"
	"github.com/ourstudio-se/lct/internal/json"
)

func NewRenderer() func(io.Writer, *deps.DependencyNode) error {
	return func(w io.Writer, graph *deps.DependencyNode) error {
		b, err := json.Marshal(graph)
		if err != nil {
			return err
		}

		fmt.Fprintln(w, string(b))
		return nil
	}
}
