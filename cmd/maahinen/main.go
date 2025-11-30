package main

import (
	"fmt"
	"os"

	"maahinen/internal/agent"
	"maahinen/internal/llm"
	"maahinen/internal/setup"
)

func main() {
	model, err := setup.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Get Ollama URL from env or use default
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	client := llm.NewClient(ollamaURL, model)

	// Create an Agent and run the agent loop
	a := agent.NewAgent(client)
	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
