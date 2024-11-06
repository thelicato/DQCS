package cmd

import (
	"github.com/spf13/cobra"
	"github.com/thelicato/dqcs/pkg/socket"
)

var socketPath string

var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Run the DQCS host component",
	Run: func(cmd *cobra.Command, args []string) {
		socket.RunHost(socketPath)
	},
}

//nolint:errcheck
func init() {
	rootCmd.AddCommand(hostCmd)
	hostCmd.Flags().StringVarP(&socketPath, "socket", "s", "", "Path to the socket file")
	hostCmd.MarkFlagRequired("socket")
}
