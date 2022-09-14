package output

import (
	"github.com/spf13/pflag"
)

const (
	GraphArg  = "graph"
	JsonArg   = "json"
	VerifyArg = "verify-with"
)

func DefineOutputArgs(flags *pflag.FlagSet) {
	flags.Bool(GraphArg, false, "output graph")
	flags.Bool(JsonArg, false, "output graph as JSON")
	flags.String(VerifyArg, "", "verify using YAML config")
}
