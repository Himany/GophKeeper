package main

import (
	"github.com/Himany/GophKeeper/internal/client/cmd"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, buildTime)

	cmd.Execute()
}
