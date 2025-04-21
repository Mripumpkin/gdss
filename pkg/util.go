package pkg

import "runtime"

const (
	defaultStackSize = 4096
)

func GetCurrentGoroutineStack() string {
	var buf [defaultStackSize]byte
	n := runtime.Stack(buf[:], false)
	return string(buf[:n])
}
