package cmd

// 由 Makefile -ldflags 在构建时注入。
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)
