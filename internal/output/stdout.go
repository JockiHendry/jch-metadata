package output

import "fmt"

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
