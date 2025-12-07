package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DanielNikkari/maahinen/internal/config"
	"github.com/DanielNikkari/maahinen/internal/llm"
	"github.com/DanielNikkari/maahinen/internal/setup"
	"github.com/DanielNikkari/maahinen/internal/tools"
	"github.com/DanielNikkari/maahinen/internal/tui"
	"github.com/DanielNikkari/maahinen/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.Color(ui.Red, fmt.Sprintf("Error loading config: %v", err)))
		os.Exit(1)
	}

	selectedModel, err := setup.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.Color(ui.Red, fmt.Sprintf("Error: %v", err)))
		os.Exit(1)
	}

	// Get Ollama URL from env, config, or use default
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = cfg.Ollama.BaseURL
	}
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	// Use selected model from setup, or fall back to config default
	modelToUse := selectedModel
	if modelToUse == "" {
		modelToUse = cfg.Ollama.DefaultModel
	}

	// Create LLM client
	client := llm.NewClient(ollamaURL, modelToUse)

	// Create tool registry
	registry := tools.NewRegistry()
	registry.Register(tools.NewBashTool(""))
	registry.Register(tools.NewReadTool(""))
	registry.Register(tools.NewWriteTool(""))
	registry.Register(tools.NewEditTool(""))
	registry.Register(tools.NewListTool(""))

	// Set up debug logging
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Warning: could not create logs directory: %v", err)
	}
	f, err := tea.LogToFile("logs/debug.log", "debug")
	if err != nil {
		log.Printf("Warning: could not open debug log: %v", err)
	} else {
		defer f.Close()
	}

	// Create the TUI agent with config
	agent := tui.NewTUIAgent(client, registry, cfg)
	defer agent.Close()

	// Create the TUI program and model
	program, model, err := tui.StartProgram()
	if err != nil {
		fmt.Fprintln(os.Stderr, ui.Color(ui.Red, fmt.Sprintf("Error creating TUI: %v", err)))
		os.Exit(1)
	}

	// Apply UI config settings
	model.SetSpinnerStyle(cfg.UI.SpinnerStyle)
	model.SetAutoConfirmTools(cfg.Agent.AutoConfirm)

	// Connect agent to the TUI
	agent.SetProgram(program, model)

	// Run the TUI
	if _, err := program.Run(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Color(ui.Red, fmt.Sprintf("Error running TUI: %v", err)))
		os.Exit(1)
	}
}
