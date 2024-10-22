package cmd

import (
	"github.com/spf13/cobra"
	"github.com/thelicato/dqcs/pkg/runners"
)

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Run the DQCS host component",
	Run: func(cmd *cobra.Command, args []string) {
		runners.RunHost()
	},
}

func init() {
	rootCmd.AddCommand(hostCmd)
}
