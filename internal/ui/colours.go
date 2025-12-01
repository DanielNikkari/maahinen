package ui

import (
	"fmt"
	"strings"
)

const Indent = "    "

// ANSI color codes
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"

	// Regular colors
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright colors
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
)

func Color(color, text string) string {
	return fmt.Sprintf("%s%s%s", color, text, Reset)
}

func PrintColor(color, text string) {
	fmt.Println(Color(color, text))
}

func BoldText(text string) string {
	return fmt.Sprintf("%s%s%s", Bold, text, Reset)
}

func UserPrompt() string {
	return Color(BrightCyan, "> ")
}

func AssistantPrompt() string {
	return Color(BrightMagenta, "Maahinen:\n")
}

func ActionAdjPrompt(actionAdj string) string {
	return Color(BrightYellow, actionAdj)
}

func PrintError(err error) {
	fmt.Println(Color(Red, fmt.Sprintf("Error: %v", err)))
}

func Indented(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = Indent + line
	}
	return strings.Join(lines, "\n")
}

func PrintIndented(text string) {
	fmt.Println(Indented(text))
}
