package common

import (
	"fmt"
	"os"
	"strings"
)

func Version() {
	data, err := os.ReadFile("version.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading version.txt: %v\n", err)
		os.Exit(1)
	}

	version := strings.TrimSpace(string(data))
	if version == "" {
		fmt.Fprintf(os.Stderr, "Error: version.txt is empty\n")
		os.Exit(1)
	}

	// 设置环境变量 VERSION
	err = os.Setenv("VERSION", version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting VERSION environment variable: %v\n", err)
		os.Exit(1)
	}
}
