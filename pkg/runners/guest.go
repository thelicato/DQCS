package runners

import (
	"fmt"
	"os"

	"github.com/thelicato/dqcs/pkg/logger"
	"github.com/thelicato/dqcs/pkg/utils"
)

func RunGuest() {
	logger.Info("Running Guest component....")

	conn, err := utils.OpenVirtioPort("guest")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening virtio port: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	utils.HandleConnection(conn)
}
