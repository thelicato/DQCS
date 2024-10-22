package runners

import (
	"fmt"
	"os"

	"github.com/thelicato/dqcs/pkg/logger"
	"github.com/thelicato/dqcs/pkg/utils"
)

func RunHost() {
	logger.Info("Running Host component....")
	conn, err := utils.OpenVirtioPort("host")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening virtio port: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	utils.HandleConnection(conn)
}
