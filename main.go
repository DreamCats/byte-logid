package main

import "github.com/DreamCats/logid/cmd"

// 通过 ldflags 注入
var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, gitCommit, buildDate)
	cmd.Execute()
}
