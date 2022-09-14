package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/ourstudio-se/lct/cmd/input"
	"github.com/ourstudio-se/lct/cmd/output"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: "lct",
	}
)

func Execute() {
	input.DefineInputArgs(rootCmd.PersistentFlags())
	output.DefineOutputArgs(rootCmd.PersistentFlags())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
