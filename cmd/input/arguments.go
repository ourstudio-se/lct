package input

import (
	"github.com/spf13/pflag"
)

const (
	noCacheArgName   = "no-cache"
	cacheFileArgName = "cache-file"
)

func DefineInputArgs(flags *pflag.FlagSet) {
	flags.Bool(noCacheArgName, false, "bypass cache")
	flags.String(cacheFileArgName, "", "path to cache file")
}
