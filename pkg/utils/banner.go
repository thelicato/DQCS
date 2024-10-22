package utils

import "fmt"

func Banner(version string) {
	banner := `
	░█▀▄░▄▀▄░█▀▀░█▀▀
	░█░█░█\█░█░░░▀▀█
	░▀▀░░░▀\░▀▀▀░▀▀▀

v%s - https://github.com/thelicato/dqcs

`

	fmt.Printf(banner, version)
}