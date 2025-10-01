package main

import (
	"fmt"
	"strings"
)

var (
	version   = "dev"
	commit    = ""
	buildDate = ""
)

func displayVersion() string {
	base := version
	if base == "" {
		base = "dev"
	}
	details := make([]string, 0, 2)
	if commit != "" {
		details = append(details, fmt.Sprintf("commit %s", commit))
	}
	if buildDate != "" {
		details = append(details, fmt.Sprintf("built %s", buildDate))
	}
	if len(details) == 0 {
		return base
	}
	return fmt.Sprintf("%s [%s]", base, strings.Join(details, ", "))
}
