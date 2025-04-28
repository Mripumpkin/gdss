package pkg

import (
	"fmt"
	"runtime"
	"time"
)

const (
	defaultStackSize = 4096
)

func GetCurrentGoroutineStack() string {
	var buf [defaultStackSize]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}

// generateTraceID generates a unique trace ID for distributed tracing (placeholder).
func GenerateTraceID() string {
	return "trace-" + fmt.Sprintf("%d", time.Now().UnixNano())
}
