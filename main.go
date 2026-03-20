package main

import commands "github.com/DreamCats/byte-logid/cmd"

// 通过 ldflags 注入
var (
	version   = "dev"
	gitCommit = "unknown"
	buildDate = "unknown"
)

func main() {
	commands.SetVersionInfo(version, gitCommit, buildDate)
	commands.Execute()
}
