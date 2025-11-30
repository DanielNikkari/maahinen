package main

import (
	"fmt"
	"os"

	"maahinen/internal/llm"
	"maahinen/internal/setup"
)

func main() {
	if err := setup.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Get Ollama URL from env or use default
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	client := llm.NewClient(ollamaURL, "llama3.2:3b")

	messages := []llm.Message{
		{Role: llm.RoleUser, Content: "Say hello in Finnish!"},
	}

	fmt.Println("Sending message...")
	resp, err := client.Chat(messages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Chat error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Response: %s\n", resp.Content)
}
