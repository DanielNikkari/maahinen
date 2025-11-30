package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PickerOption struct {
	Name        string
	Description string
	Extra       string
}

func PickModel(options []PickerOption) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	for i, opt := range options {
		fmt.Printf("  %s %s", Color(BrightCyan, fmt.Sprintf("%d)", i+1)), opt.Name)
		if opt.Extra != "" {
			fmt.Printf(" %s", Color(Dim, "("+opt.Extra+")"))
		}
		fmt.Println()
		if opt.Description != "" {
			fmt.Printf("     %s\n", Color(Dim, opt.Description))
		}
		fmt.Println()
	}

	fmt.Printf("  %s Enter a custom model name\n\n", Color(BrightCyan, fmt.Sprintf("%d)", len(options)+1)))

	for {
		fmt.Print(Color(Yellow, "Select (1-"+strconv.Itoa(len(options)+1)+"): "))
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		input = strings.TrimSpace(input)

		num, err := strconv.Atoi(input)
		if err == nil {
			if num >= 1 && num <= len(options) {
				return options[num-1].Name, nil
			}
			if num == len(options)+1 {
				// Custom input
				fmt.Print(Color(Yellow, "Enter model name: "))
				custom, err := reader.ReadString('\n')
				if err != nil {
					return "", err
				}
				custom = strings.TrimSpace(custom)
				if custom != "" {
					return custom, nil
				}
				fmt.Println(Color(Red, "Model name cannot be empty"))
				continue
			}
		}

		fmt.Println(Color(Red, "Invalid selection"))
	}
}
