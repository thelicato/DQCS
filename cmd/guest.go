package cmd

import (
	"github.com/spf13/cobra"
	"github.com/thelicato/dqcs/pkg/socket"
)

var guestCmd = &cobra.Command{
	Use:   "guest",
	Short: "Run the DQCS host component",
	Run: func(cmd *cobra.Command, args []string) {
		socket.RunGuest()
	},
}

func init() {
	rootCmd.AddCommand(guestCmd)
}
