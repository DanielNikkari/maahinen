package tui

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/DanielNikkari/maahinen/internal/llm"
	"github.com/DanielNikkari/maahinen/internal/ollama"
	"github.com/DanielNikkari/maahinen/internal/tools"
	tea "github.com/charmbracelet/bubbletea"
)

// ToolConfirmation represents a pending tool confirmation request
type ToolConfirmation struct {
	ID        string
	Name      string
	Arguments map[string]any
	Response  chan bool // true = confirmed, false = denied
}

// TUIAgent wraps the agent functionality for TUI integration
type TUIAgent struct {
	client   *llm.Client
	messages []llm.Message
	tools    *tools.Registry
	program  *tea.Program
	model    *Model
	logFile  *os.File

	// Tool confirmation
	autoConfirm      bool
	pendingConfirm   *ToolConfirmation
	pendingConfirmMu sync.Mutex

	// Spinner style
	spinnerStyle string
}

// NewTUIAgent creates a new TUI-integrated agent
func NewTUIAgent(client *llm.Client, registry *tools.Registry) *TUIAgent {
	// Register tools in registry
	for _, tool := range registry.All() {
		client.RegisterTool(tool.Definition())
	}

	// Set up tool call logging
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Warning: could not create logs directory: %v", err)
	}

	logPath := filepath.Join(logDir, fmt.Sprintf("tools_%s.log", time.Now().Format("2006-01-02")))
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: could not open tool log file: %v", err)
	}

	return &TUIAgent{
		client:       client,
		tools:        registry,
		logFile:      logFile,
		autoConfirm:  false, // Default to confirm tools for safety
		spinnerStyle: "dots",
		messages: []llm.Message{
			{
				Role: llm.RoleSystem,
				Content: `You are Maahinen, a helpful coding assistant. You help users with programming tasks,
answer questions about code, and assist with debugging. Be concise and practical.
You should aim to take action, for example, when user asks you for example write code
you should utilize your tools to complete the user request. If a tool call fails, first try to fix the issue by recalling the tool with 
corrected arguments.`,
			},
		},
	}
}

// SetAutoConfirm sets whether tools should be auto-confirmed
func (a *TUIAgent) SetAutoConfirm(auto bool) {
	a.autoConfirm = auto
}

// SetProgram sets the tea.Program for sending messages
func (a *TUIAgent) SetProgram(p *tea.Program, m *Model) {
	a.program = p
	a.model = m
	m.SetModel(a.client.Model())
	m.SetAutoConfirmTools(a.autoConfirm)

	// Set up the message callback
	m.SetOnSendMessage(func(content string) {
		go a.handleUserMessage(content)
	})

	// Set up tool confirmation callback
	m.SetOnToolConfirm(func(confirmed bool) {
		a.handleToolConfirmation(confirmed)
	})

	// Set up auto-confirm toggle callback
	m.SetOnAutoConfirmToggle(func(enabled bool) {
		a.autoConfirm = enabled
	})

	// Set up prune callback
	m.SetOnPrune(func() {
		a.pruneContext()
	})
}

// handleToolConfirmation handles user's tool confirmation response
func (a *TUIAgent) handleToolConfirmation(confirmed bool) {
	a.pendingConfirmMu.Lock()
	pending := a.pendingConfirm
	a.pendingConfirmMu.Unlock()

	if pending != nil && pending.Response != nil {
		pending.Response <- confirmed
	}
}

// handleUserMessage processes a user message
func (a *TUIAgent) handleUserMessage(content string) {
	// Check for commands first
	if strings.HasPrefix(content, "/") {
		a.handleCommand(content)
		return
	}

	// Check for exit commands
	if content == "exit" || content == "quit" {
		a.program.Send(tea.Quit())
		return
	}

	// Add user message to history
	a.messages = append(a.messages, llm.Message{
		Role:    llm.RoleUser,
		Content: content,
	})

	// Process with LLM
	a.processResponse()
}

// handleCommand processes slash commands
func (a *TUIAgent) handleCommand(input string) {
	parts := strings.Split(input, "/")
	if len(parts) < 2 {
		return
	}

	switch parts[1] {
	case "model":
		a.handleModelCommand(parts[2:])
	case "spinner":
		a.handleSpinnerCommand(parts[2:])
	case "autoconfirm":
		a.autoConfirm = !a.autoConfirm
		a.model.SetAutoConfirmTools(a.autoConfirm)
		status := "disabled"
		if a.autoConfirm {
			status = "enabled"
		}
		a.program.Send(ResponseMsg{
			Role:    "system",
			Content: fmt.Sprintf("Auto-confirm tools: %s", status),
		})
	case "help":
		a.handleHelpCommand()
	case "prune":
		a.handlePruneCommand()
	default:
		a.program.Send(ResponseMsg{
			Role:    "system",
			Content: fmt.Sprintf("Unknown command: %s", input),
		})
	}
}

func (a *TUIAgent) handleModelCommand(args []string) {
	if len(args) == 0 {
		a.program.Send(ResponseMsg{
			Role:    "system",
			Content: fmt.Sprintf("Current model: %s", a.client.Model()),
		})
		return
	}

	ollamaURL := a.client.BaseURL()

	switch args[0] {
	case "list":
		models, err := ollama.ListModels(ollamaURL)
		if err != nil {
			a.program.Send(ResponseMsg{
				Role:    "system",
				Content: fmt.Sprintf("Error listing models: %v", err),
			})
			return
		}

		var sb strings.Builder
		sb.WriteString("Installed models:\n")
		for _, m := range models {
			if m.Name == a.client.Model() {
				sb.WriteString(fmt.Sprintf("  * %s (current)\n", m.Name))
			} else {
				sb.WriteString(fmt.Sprintf("    %s\n", m.Name))
			}
		}
		a.program.Send(ResponseMsg{
			Role:    "system",
			Content: sb.String(),
		})
	default:
		modelName := args[0]
		// Check if model is already installed
		models, err := ollama.ListModels(ollamaURL)
		if err != nil {
			a.program.Send(ResponseMsg{
				Role:    "system",
				Content: fmt.Sprintf("Error: %v", err),
			})
			return
		}

		found := false
		for _, m := range models {
			if m.Name == modelName {
				found = true
				break
			}
		}

		if found {
			// Model already installed, switch to it
			a.client.SetModel(modelName)
			a.program.Send(ModelChangedMsg{Model: modelName})
			a.program.Send(ResponseMsg{
				Role:    "system",
				Content: fmt.Sprintf("Switched to model: %s", modelName),
			})
		} else {
			// Model not installed, try to pull it from Ollama
			a.program.Send(ResponseMsg{
				Role:    "system",
				Content: fmt.Sprintf("Pulling %s: starting...", modelName),
			})

			// Pull model with progress updates (update in place)
			err := ollama.PullModel(ollamaURL, modelName, func(progress ollama.PullProgress) {
				if progress.Status != "" {
					var msg string
					if progress.Total > 0 && progress.Completed > 0 {
						percent := float64(progress.Completed) / float64(progress.Total) * 100
						msg = fmt.Sprintf("Pulling %s: %s (%.1f%%)", modelName, progress.Status, percent)
					} else {
						msg = fmt.Sprintf("Pulling %s: %s", modelName, progress.Status)
					}
					// Update the last message in place
					a.program.Send(UpdateLastMessageMsg{Content: msg})
				}
			})

			if err != nil {
				// Update last message with error
				a.program.Send(UpdateLastMessageMsg{
					Content: fmt.Sprintf("Failed to pull model '%s': %v", modelName, err),
				})
				return
			}

			// Successfully pulled, update the message and switch
			a.client.SetModel(modelName)
			a.program.Send(ModelChangedMsg{Model: modelName})
			a.program.Send(UpdateLastMessageMsg{
				Content: fmt.Sprintf("Successfully pulled and switched to model: %s", modelName),
			})
		}
	}
}

func (a *TUIAgent) handleSpinnerCommand(args []string) {
	if len(args) == 0 {
		a.program.Send(ResponseMsg{
			Role:    "system",
			Content: fmt.Sprintf("Current spinner: %s", a.spinnerStyle),
		})
		return
	}

	switch args[0] {
	case "list":
		spinners := GetAvailableSpinners()
		var sb strings.Builder
		sb.WriteString("Available spinners:\n")
		for _, s := range spinners {
			if s == a.spinnerStyle {
				sb.WriteString(fmt.Sprintf("  * %s (current)\n", s))
			} else {
				sb.WriteString(fmt.Sprintf("    %s\n", s))
			}
		}
		a.program.Send(ResponseMsg{
			Role:    "system",
			Content: sb.String(),
		})
	default:
		spinnerName := args[0]
		spinners := GetAvailableSpinners()
		if !slices.Contains(spinners, spinnerName) {
			a.program.Send(ResponseMsg{
				Role:    "system",
				Content: fmt.Sprintf("Spinner '%s' not found. Use /spinner/list to see available spinners.", spinnerName),
			})
			return
		}
		a.spinnerStyle = spinnerName
		a.model.SetSpinnerStyle(spinnerName)
		a.program.Send(ResponseMsg{
			Role:    "system",
			Content: fmt.Sprintf("Switched to spinner: %s", spinnerName),
		})
	}
}

func (a *TUIAgent) handleHelpCommand() {
	help := `Available commands:
/model           Show current model
/model/list      List installed models
/model/{name}    Switch to model (pulls if needed)
/spinner         Show current spinner
/spinner/list    List available spinners
/spinner/{name}  Switch to spinner
/prune           Clear message history and context
/autoconfirm     Toggle auto-confirm for tools
/help            Show this help
exit, quit       Exit Maahinen`

	a.program.Send(ResponseMsg{
		Role:    "system",
		Content: help,
	})
}

func (a *TUIAgent) handlePruneCommand() {
	a.pruneContext()
	a.program.Send(ResponseMsg{
		Role:    "system",
		Content: "Context cleared. Message history has been pruned.",
	})
}

// pruneContext clears the message history while keeping the system prompt
func (a *TUIAgent) pruneContext() {
	// Keep only the system message
	if len(a.messages) > 0 && a.messages[0].Role == llm.RoleSystem {
		a.messages = a.messages[:1]
	} else {
		a.messages = []llm.Message{}
	}

	// Clear the UI
	a.model.ClearMessages()
}

// processResponse handles LLM response processing with streaming
func (a *TUIAgent) processResponse() {
	for {
		var resp *llm.Message
		var streamErr error

		// Use streaming to show response as it's generated
		resp, streamErr = a.client.ChatStream(a.messages, func(chunk string, done bool, fullMessage *llm.Message) {
			if !done && chunk != "" {
				// Send each chunk to the TUI for display
				a.program.Send(StreamChunkMsg{
					Content: chunk,
					Done:    false,
				})
			}
		})

		if streamErr != nil {
			a.program.Send(ErrorMsg{Error: streamErr})
			return
		}

		a.messages = append(a.messages, *resp)

		// Check for native tool calls
		if resp.HasToolCalls() {
			// Signal end of streaming before handling tools
			a.program.Send(StreamChunkMsg{Content: "", Done: true})

			for _, tc := range resp.ToolCalls {
				_, err := a.executeTool(tc)
				if err != nil {
					a.program.Send(ErrorMsg{Error: err})
					return
				}
			}
			continue // Continue the conversation with tool results
		}

		// Check for JSON tool calls in content
		if tc, ok := tools.ParseToolCallFromContent(resp.Content); ok {
			// Signal end of streaming before handling tools
			a.program.Send(StreamChunkMsg{Content: "", Done: true})

			_, err := a.executeTool(*tc)
			if err != nil {
				a.program.Send(ErrorMsg{Error: err})
				return
			}
			continue
		}

		// Regular text response - signal completion
		if resp.Content != "" {
			a.program.Send(StreamChunkMsg{Content: "", Done: true})
		}
		return
	}
}

// processResponseNonStreaming handles LLM response processing without streaming (kept for reference)
func (a *TUIAgent) processResponseNonStreaming() {
	for {
		resp, err := a.client.Chat(a.messages)
		if err != nil {
			a.program.Send(ErrorMsg{Error: err})
			return
		}

		a.messages = append(a.messages, *resp)

		// Check for native tool calls
		if resp.HasToolCalls() {
			for _, tc := range resp.ToolCalls {
				_, err := a.executeTool(tc)
				if err != nil {
					a.program.Send(ErrorMsg{Error: err})
					return
				}
			}
			continue
		}

		// Check for JSON tool calls in content
		if tc, ok := tools.ParseToolCallFromContent(resp.Content); ok {
			_, err := a.executeTool(*tc)
			if err != nil {
				a.program.Send(ErrorMsg{Error: err})
				return
			}
			continue
		}

		// Regular text response
		if resp.Content != "" {
			a.program.Send(ResponseMsg{
				Role:    "assistant",
				Content: resp.Content,
			})
		}
		return
	}
}

// executeTool executes a tool call and logs it
// Returns (wasExecuted, error)
func (a *TUIAgent) executeTool(tc llm.ToolCall) (bool, error) {
	toolName := tc.Function.Name

	// Map common tool name mistakes
	switch toolName {
	case "go", "python", "shell", "sh", "cmd", "command", "terminal":
		toolName = "bash"
	case "file_read", "read_file":
		toolName = "read"
	case "file_write", "write_file", "create_file":
		toolName = "write"
	case "file_edit", "edit_file":
		toolName = "edit"
	case "file_list", "list_files", "ls", "dir":
		toolName = "list"
	}

	// Generate a unique ID for this tool call
	toolID := fmt.Sprintf("%s_%d", toolName, time.Now().UnixNano())

	// Request confirmation if needed
	if !a.autoConfirm {
		confirmed := a.requestToolConfirmation(toolID, toolName, tc.Function.Arguments)
		if !confirmed {
			a.logToolCall(toolID, toolName, tc.Function.Arguments, "denied by user")
			// Send cancelled message to TUI (for display in tool panel)
			a.program.Send(ToolCancelledMsg{
				ID:        toolID,
				Name:      toolName,
				Arguments: tc.Function.Arguments,
			})
			// Add cancelled tool call to message history (dimmed)
			argsOneLine := formatToolArgsOneLine(tc.Function.Arguments)
			a.program.Send(ResponseMsg{
				Role:    "toolcall_cancelled",
				Content: fmt.Sprintf("%s(%s) - cancelled", toolName, argsOneLine),
			})
			// Add denial message to conversation
			a.messages = append(a.messages, llm.Message{
				Role:    llm.RoleTool,
				Content: "Tool execution was denied by the user.",
			})
			return false, nil
		}
	}

	// Send tool call to TUI (for display in tool panel)
	a.program.Send(ToolCallMsg{
		ID:        toolID,
		Name:      toolName,
		Arguments: tc.Function.Arguments,
	})

	// Add tool call as one-liner to message history
	argsOneLine := formatToolArgsOneLine(tc.Function.Arguments)
	a.program.Send(ResponseMsg{
		Role:    "toolcall",
		Content: fmt.Sprintf("%s(%s)", toolName, argsOneLine),
	})

	// Log the tool call
	a.logToolCall(toolID, toolName, tc.Function.Arguments, "started")

	tool, ok := a.tools.Get(toolName)
	if !ok {
		availableTools := strings.Join(a.tools.List(), ", ")
		errMsg := fmt.Sprintf("Unknown tool '%s'. Available tools: %s", tc.Function.Name, availableTools)

		a.program.Send(ToolResultMsg{
			ID:      toolID,
			Name:    toolName,
			Success: false,
			Error:   errMsg,
		})

		a.logToolCall(toolID, toolName, nil, "error: unknown tool")

		a.messages = append(a.messages, llm.Message{
			Role:    llm.RoleTool,
			Content: errMsg,
		})
		return true, nil
	}

	// Execute the tool
	result, err := tool.Execute(context.Background(), tc.Function.Arguments)
	if err != nil {
		a.program.Send(ToolResultMsg{
			ID:      toolID,
			Name:    toolName,
			Success: false,
			Error:   err.Error(),
		})
		a.logToolCall(toolID, toolName, nil, fmt.Sprintf("execution error: %v", err))
		return true, err
	}

	// Send result to TUI
	a.program.Send(ToolResultMsg{
		ID:      toolID,
		Name:    toolName,
		Success: result.Success,
		Output:  result.Output,
		Error:   result.Error,
	})

	// Log result
	status := "success"
	if !result.Success {
		status = fmt.Sprintf("failed: %s", result.Error)
	}
	a.logToolCall(toolID, toolName, nil, status)

	// Update tool call in message history based on result
	if !result.Success {
		argsOneLine := formatToolArgsOneLine(tc.Function.Arguments)
		failedContent := fmt.Sprintf("%s(%s) - failed: %s", toolName, argsOneLine, result.Error)
		a.model.UpdateToolCallStatus(toolName, "toolcall_failed", failedContent)
	}

	// Add tool result to messages
	toolOutput := result.Output
	if toolOutput == "" && result.Success {
		toolOutput = "Command executed successfully (no output)"
	} else if !result.Success {
		toolOutput = fmt.Sprintf("Command failed: %s\nOutput: %s", result.Error, result.Output)
	}

	a.messages = append(a.messages, llm.Message{
		Role:    llm.RoleTool,
		Content: toolOutput,
	})

	return true, nil
}

// requestToolConfirmation requests user confirmation for a tool call
func (a *TUIAgent) requestToolConfirmation(id, name string, args map[string]any) bool {
	responseChan := make(chan bool, 1)

	confirmation := &ToolConfirmation{
		ID:        id,
		Name:      name,
		Arguments: args,
		Response:  responseChan,
	}

	a.pendingConfirmMu.Lock()
	a.pendingConfirm = confirmation
	a.pendingConfirmMu.Unlock()

	// Send confirmation request to TUI
	a.program.Send(ToolConfirmRequestMsg{
		ID:        id,
		Name:      name,
		Arguments: args,
	})

	// Wait for response
	confirmed := <-responseChan

	a.pendingConfirmMu.Lock()
	a.pendingConfirm = nil
	a.pendingConfirmMu.Unlock()

	return confirmed
}

// logToolCall logs a tool call to the log file
func (a *TUIAgent) logToolCall(id, name string, args map[string]any, status string) {
	if a.logFile == nil {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	argsStr := ""
	if args != nil {
		var parts []string
		for k, v := range args {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
		argsStr = strings.Join(parts, ", ")
	}

	logLine := fmt.Sprintf("[%s] %s | tool=%s | args={%s} | status=%s\n",
		timestamp, id, name, argsStr, status)

	a.logFile.WriteString(logLine)
}

// Close cleans up resources
func (a *TUIAgent) Close() {
	if a.logFile != nil {
		a.logFile.Close()
	}
}

// formatToolArgsOneLine formats tool arguments as a short one-liner
func formatToolArgsOneLine(args map[string]any) string {
	if len(args) == 0 {
		return ""
	}

	var parts []string
	for k, v := range args {
		s := fmt.Sprintf("%s=%v", k, v)
		if len(s) > 25 {
			s = s[:22] + "..."
		}
		parts = append(parts, s)
	}

	result := strings.Join(parts, ", ")
	if len(result) > 60 {
		result = result[:57] + "..."
	}
	return result
}
