package main

import (
	"runtime/debug"
	"strings"
)

// default values overridden by -ldflags during builds
var (
	version   = "dev"
	commit    = ""
	buildDate = ""
)

func displayVersion() string {
	ver := strings.TrimSpace(version)
	sha := strings.TrimSpace(commit)
	date := strings.TrimSpace(buildDate)

	if ver == "" || ver == "(devel)" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				ver = info.Main.Version
			}
		}
	}

	if ver == "" {
		ver = "dev"
	}

	parts := []string{ver}
	if sha != "" && sha != "unknown" {
		parts = append(parts, sha)
	}
	if date != "" {
		parts = append(parts, date)
	}

	return strings.Join(parts, " | ")
}
