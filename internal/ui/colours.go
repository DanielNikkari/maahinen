package ui

import "fmt"

// ANSI color codes
const (
	Reset = "\033[0m"
	Dim   = "\033[2m"

	// Regular colors
	Red    = "\033[31m"
	Yellow = "\033[33m"

	// Bright colors
	BrightGreen = "\033[92m"
	BrightCyan  = "\033[96m"
)

// Color wraps text with ANSI color codes
func Color(color, text string) string {
	return fmt.Sprintf("%s%s%s", color, text, Reset)
}

// PrintColor prints colored text to stdout
func PrintColor(color, text string) {
	fmt.Println(Color(color, text))
}
