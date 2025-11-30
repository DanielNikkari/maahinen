package agent

import (
	"fmt"
	"strings"

	"github.com/DanielNikkari/maahinen/internal/ollama"
	"github.com/DanielNikkari/maahinen/internal/ui"
)

func (a *Agent) handleCommand(input string) bool {
	if !strings.HasPrefix(input, "/") {
		return false
	}

	parts := strings.Split(input, "/")
	if len(parts) < 2 {
		return false
	}

	switch parts[1] {
	case "model":
		a.handleModelCommand(parts[2:])
		return true
	case "help":
		a.handleHelpCommand()
		return true
	default:
		fmt.Printf("%s Unknown command: %s\n\n", ui.Color(ui.Red, "✗"), input)
		return true
	}
}

func (a *Agent) handleModelCommand(args []string) {
	if len(args) == 0 {
		fmt.Printf("%s Current model: %s\n\n", ui.Color(ui.Cyan, "📦"), a.client.Model())
		return
	}

	ollamaURL := a.client.BaseURL()

	switch args[0] {
	case "list":
		models, err := ollama.ListModels(ollamaURL)
		if err != nil {
			ui.PrintError(err)
			return
		}

		fmt.Printf("%s Installed models:\n", ui.Color(ui.Cyan, "📦"))
		for _, m := range models {
			if m.Name == a.client.Model() {
				fmt.Printf("   %s %s (current)\n", ui.Color(ui.BrightGreen, "●"), m.Name)
			} else {
				fmt.Printf("   %s %s\n", ui.Color(ui.Dim, "○"), m.Name)
			}
		}
		fmt.Println()
	default:
		modelName := args[0]
		a.selectOrInstallModel(modelName, ollamaURL)
	}
}

func (a *Agent) selectOrInstallModel(modelName, ollamaURL string) {
	models, err := ollama.ListModels(ollamaURL)
	if err != nil {
		ui.PrintError(err)
		return
	}

	for _, m := range models {
		if m.Name == modelName {
			a.client.SetModel(modelName)
			fmt.Printf("%s Switched to model: %s\n\n", ui.Color(ui.BrightGreen, "✓"), modelName)
			return
		}
	}

	fmt.Printf("%s Model not found. Downloading %s...\n", ui.Color(ui.Yellow, "⚡"), modelName)

	err = ollama.PullModel(ollamaURL, modelName, func(p ollama.PullProgress) {
		if p.Total > 0 {
			pct := float64(p.Completed) / float64(p.Total) * 100
			fmt.Printf("\r   %s: %.1f%%", p.Status, pct)
		} else if p.Status != "" {
			fmt.Printf("\r   %s...          ", p.Status)
		}
	})
	fmt.Println()

	if err != nil {
		ui.PrintError(err)
		return
	}

	a.client.SetModel(modelName)
	fmt.Printf("%s Downloaded and switched to: %s\n\n", ui.Color(ui.BrightGreen, "✓"), modelName)
}

func (a *Agent) handleHelpCommand() {
	fmt.Println(ui.Color(ui.Cyan, "Available commands:"))
	fmt.Println("   /model           Show current model")
	fmt.Println("   /model/list      List installed models")
	fmt.Println("   /model/{name}    Switch to or install a model, e.g., `/model/qwen2.5-coder:7b`")
	fmt.Println("   /help            Show this help")
	fmt.Println("   exit, quit       Exit Maahinen")
	fmt.Println()
}
