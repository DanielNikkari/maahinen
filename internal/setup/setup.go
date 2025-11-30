package setup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/DanielNikkari/maahinen/internal/ollama"
	"github.com/DanielNikkari/maahinen/internal/ui"
)

const defaultURL = "http://localhost:11434"

func Run() (string, error) {
	fmt.Println("🧙 Maahinen Setup")
	fmt.Println("=================")
	fmt.Println()

	var selectedModel string

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = defaultURL
	}

	if ollama.IsRunningAt(ollamaURL) {
		ui.PrintColor(ui.BrightGreen, "✓ Ollama server is running")
	} else {
		// Install Ollama if not yet installed
		if !ollama.IsInstalled() {
			fmt.Println("Ollama is not installed.")
			if !confirm("Would you like to install it now?") {
				return "", fmt.Errorf("ollama is rquired to run Maahinen")
			}
			if err := ollama.Install(); err != nil {
				return "", fmt.Errorf("failed to install Ollama: %w", err)
			}
			ui.PrintColor(ui.BrightGreen, "✓ Ollama installed succesfully!")
		}

		// Check if Ollama is running
		if !ollama.IsRunning() {
			fmt.Println("Starting Ollama server...")
			if err := ollama.Start(); err != nil {
				return "", fmt.Errorf("failed to start Ollama: %w", err)
			}
			ui.PrintColor(ui.BrightGreen, "✓ Ollama server started")
		} else {
			ui.PrintColor(ui.BrightGreen, "✓ Ollama server is running")
		}
	}

	// Check for models
	hasModels, err := ollama.HasModels(ollamaURL)
	if err != nil {
		return "", fmt.Errorf("failed to check models: %w", err)
	}

	if !hasModels {
		fmt.Println()
		fmt.Println("No models installed yet.")
		if selectedModel, err = pickAndPullModel(ollamaURL); err != nil {
			return "", err
		}
	} else {
		models, _ := ollama.ListModels(ollamaURL)
		selectedModel = models[0].Name // Use first available model, TODO: change this later to let the user select the model
		ui.PrintColor(ui.BrightGreen, fmt.Sprintf("✓ Using model: %s", selectedModel))
	}

	ui.PrintColor(ui.BrightGreen, "✓ Setup complete! Maahinen is ready!")
	return selectedModel, nil
}

func pickAndPullModel(ollamaURL string) (string, error) {
	recommended := ollama.GetRecommendedModels()

	options := make([]ui.PickerOption, len(recommended))
	for i, m := range recommended {
		options[i] = ui.PickerOption{
			Name:        m.Name,
			Description: m.Description,
			Extra:       m.Size,
		}
	}

	for {
		selected, err := ui.PickModel(options)
		if err != nil {
			return "", err
		}

		fmt.Printf("\n%s Downloading %s...\n", ui.Color(ui.Yellow, "⚡"), selected)

		err = ollama.PullModel(ollamaURL, selected, func(p ollama.PullProgress) {
			if p.Total > 0 {
				pct := float64(p.Completed) / float64(p.Total) * 100
				fmt.Printf("\r   %s: %.1f%%", p.Status, pct)
			} else if p.Status != "" {
				fmt.Printf("\r   %s...          ", p.Status)
			}
		})

		fmt.Println()

		if err != nil {
			fmt.Printf("%s Model '%s' not found. Please try again.\n\n", ui.Color(ui.Red, "✗"), selected)
			continue
		}

		fmt.Println(ui.Color(ui.BrightGreen, fmt.Sprintf("✓ %s downloaded successfully!", selected)))
		return selected, nil
	}
}

func confirm(question string) bool {
	fmt.Printf("%s [y/N]: ", question)
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
