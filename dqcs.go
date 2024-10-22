package main

import (
	"github.com/thelicato/dqcs/cmd"
	"github.com/thelicato/dqcs/pkg/utils"
)

func main() {
	utils.Banner(utils.Version)
	cmd.Execute()
}