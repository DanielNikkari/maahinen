package ui

import (
	"fmt"
	"time"
)

func Typewriter(text string, delay time.Duration) {
	for _, char := range text {
		fmt.Print(string(char))
		time.Sleep(delay)
	}
	fmt.Println()
}

func TypewriterColored(text string, color string, delay time.Duration) {
	fmt.Print(color)
	for _, char := range text {
		fmt.Print(string(char))
		time.Sleep(delay)
	}
	fmt.Print(Reset)
	fmt.Println()
}

func TypewriterFast(text string) {
	Typewriter(text, 30*time.Millisecond)
}

func TypewriterSlow(text string) {
	Typewriter(text, 80*time.Millisecond)
}
