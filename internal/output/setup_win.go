//go:build windows

package output

import (
	"fmt"
	"golang.org/x/sys/windows"
)

func Setup() {
	handle, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil {
		fmt.Printf("Failed to retrieve current console handle: %s\n", err)
	}
	var mode uint32
	err = windows.GetConsoleMode(handle, &mode)
	if err != nil {
		fmt.Printf("Failed to get console mode: %s\n", err)
	}
	mode |= 4
	err = windows.SetConsoleMode(handle, mode)
	if err != nil {
		fmt.Printf("Failed to set console mode: %s\n", err)
	}
}
