package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func completionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generate the autocompletion script for the specified shell",
	}
}

var rootCmd = &cobra.Command{
	Use:   "dqcs",
	Short: "Dummy QEMU Clipboard Sharing for those who don't like SPICE",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Read the help page using the -h flag to see the commands")
	},
}

func init() {
	completion := completionCmd()
	completion.Hidden = true
	rootCmd.AddCommand(completion)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
