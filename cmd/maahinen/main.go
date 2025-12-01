package main

import (
	"fmt"
	"os"

	"github.com/DanielNikkari/maahinen/internal/agent"
	"github.com/DanielNikkari/maahinen/internal/llm"
	"github.com/DanielNikkari/maahinen/internal/setup"
	"github.com/DanielNikkari/maahinen/internal/tools"
	"github.com/DanielNikkari/maahinen/internal/ui"
)

func main() {
	model, err := setup.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.Color(ui.Red, fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}

	// Get Ollama URL from env or use default
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	// Create LLM client
	client := llm.NewClient(ollamaURL, model)

	// Create tool registry
	registry := tools.NewRegistry()
	registry.Register(tools.NewBashTool(""))
	registry.Register(tools.NewReadTool(""))
	registry.Register(tools.NewWriteTool(""))
	registry.Register(tools.NewEditTool(""))
	registry.Register(tools.NewListTool(""))

	// Create an Agent and run the agent loop
	a := agent.NewAgent(client, registry)
	if err := a.Run(); err != nil {
		ui.PrintError(err)
		os.Exit(1)
	}
}
