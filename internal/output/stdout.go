package output

import (
	"fmt"
	"strings"
)

func Printf(indented bool, format string, a ...any) {
	if indented {
		fmt.Print("  ")
	}
	fmt.Printf(format, a...)
}

func Println(indented bool, a ...any) {
	if indented {
		fmt.Print("  ")
	}
	fmt.Println(a...)
}

func PrintForm(indented bool, label string, value any, labelWidth int) {
	if indented {
		fmt.Print("  ")
	}
	fmt.Printf("\033[2m%-*s:\033[22m %s\n", labelWidth, label, value)
}

func PrintHeader(indented bool, format string, a ...any) {
	if indented {
		fmt.Print("  ")
	}
	fmt.Printf("\033[4m"+format+"\033[24m\n", a...)
}

func PrintHexDump(indented bool, value []byte) {
	for i := 0; i < len(value); i += 32 {
		if indented {
			fmt.Print("  ")
		}
		for j := i; j < i+32; j++ {
			if j >= len(value) {
				fmt.Printf("   ")
			} else {
				fmt.Printf("%02X ", value[j])
			}
		}
		fmt.Printf("   ")
		for j := i; j < i+32; j++ {
			if j >= len(value) {
				fmt.Printf("   ")
			} else {
				if value[j] >= 33 && value[j] <= 126 {
					fmt.Printf("%c", value[j])
				} else {
					fmt.Printf(".")
				}

			}
		}
		fmt.Println()
	}
}

func PrintMultiline(indented bool, value string) {
	lines := strings.Split(value, "\n")
	for _, line := range lines {
		if indented {
			fmt.Printf("  ")
		}
		fmt.Printf("%.120s", line)
		if len(line) > 120 {
			fmt.Printf("...")
		}
		fmt.Println()
	}
}
