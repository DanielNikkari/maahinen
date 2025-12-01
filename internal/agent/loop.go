package agent

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/DanielNikkari/maahinen/internal/llm"
	"github.com/DanielNikkari/maahinen/internal/tools"
	"github.com/DanielNikkari/maahinen/internal/ui"
)

var actionAdjectives = [...]string{
	"otherworldly meddling",
	"underground magic",
	"spirit-work",
	"mischief from the earth-folk",
	"fairy-like interference",
	"elfish enchantment",
	"gnomic trickery",
	"maahinen tomfoolery",
}

type Agent struct {
	client   *llm.Client
	messages []llm.Message
	tools    *tools.Registry
	spinner  string
}

func NewAgent(client *llm.Client, registry *tools.Registry) *Agent {
	// Register tools in registry
	for _, tool := range registry.All() {
		client.RegisterTool(tool.Definition())
	}

	return &Agent{
		client:  client,
		tools:   registry,
		spinner: "wizard",
		messages: []llm.Message{
			{
				Role: llm.RoleSystem,
				Content: `You are Maahinen, a helpful coding assistant. You help users with programming tasks, 
				answer questions about code, and assist with debugging. Be concise and practical.
				You should aim to take action, for example, when user asks you for example write code
				you should utilize your tools to complete the user request.`,
			},
		},
	}
}

func (a *Agent) Run() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("🔮 %s (%s)\n", ui.Color(ui.BrightMagenta, "Maahinen"), a.client.Model())
	fmt.Println("Type 'exit' or 'quit' to end the session.")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println()

	for {
		fmt.Print(ui.UserPrompt())
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		fmt.Println()

		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			fmt.Println()
			ui.PrintColor(ui.BrightMagenta, "🧙 Goodbye!")
			return nil
		}

		if a.handleCommand(input) {
			continue
		}

		a.messages = append(a.messages, llm.Message{
			Role:    llm.RoleUser,
			Content: input,
		})

		if err := a.processResponse(); err != nil {
			ui.PrintError(err)
			a.messages = a.messages[:len(a.messages)-1]
		}
	}
}

func (a *Agent) processResponse() error {
	for {
		spinner := ui.NewSpinner(a.spinner)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		actionAdj := actionAdjectives[r.Intn(len(actionAdjectives))]
		spinner.Start(ui.ActionAdjPrompt(actionAdj))

		resp, err := a.client.Chat(a.messages)

		spinner.Stop()

		if err != nil {
			return err
		}

		a.messages = append(a.messages, *resp)

		// Check for native tool calls
		if resp.HasToolCalls() {
			for _, tc := range resp.ToolCalls {
				if err := a.executeTool(tc); err != nil {
					return err
				}
			}
			continue
		}

		// Check for JSON tool calls
		if tc, ok := tools.ParseToolCallFromContent(resp.Content); ok {
			if err := a.executeTool(*tc); err != nil {
				return err
			}
			continue
		}

		// Regular text response
		if resp.Content != "" {
			fmt.Printf("%s%s\n\n", ui.AssistantPrompt(), resp.Content)
		}
		return nil
	}
}

func (a *Agent) executeTool(tc llm.ToolCall) error {
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

	tool, ok := a.tools.Get(toolName)
	if !ok {
		availableTools := strings.Join(a.tools.List(), ", ")
		errMsg := fmt.Sprintf("Unknown tool '%s'. Available tools: %s", tc.Function.Name, availableTools)

		fmt.Printf("%s%s %s%s\n\n", ui.Indent, ui.Color(ui.Yellow, "⚠"), "Unknown tool ", tc.Function.Name)

		a.messages = append(a.messages, llm.Message{
			Role:    llm.RoleTool,
			Content: errMsg,
		})
		return nil
	}

	argsStr := formatArgs(tc.Function.Arguments)
	fmt.Printf("%s%s%s(%s)\n", ui.Indent, ui.Color(ui.Yellow, "⚡"), ui.Color(ui.Yellow, toolName), argsStr)

	result, err := tool.Execute(context.Background(), tc.Function.Arguments)
	if err != nil {
		return err
	}

	if result.Success {
		fmt.Printf("%s%s\n", ui.Indent, ui.Color(ui.BrightGreen, "✓ Success"))
	} else {
		fmt.Printf("%s%s %s\n", ui.Indent, ui.Color(ui.Red, "✗ Failed:"), result.Error)
	}
	if result.Output != "" {
		fmt.Printf("%s\n", ui.Indented(ui.Indented(result.Output)))
	}
	fmt.Println()

	toolOutput := result.Output
	if toolOutput == "" && result.Success {
		toolOutput = "Command executed successfully (no output)"
	} else if !result.Success {
		toolOutput = fmt.Sprintf("Command failed: %s\nOutput: %s", result.Error, result.Output)
	}

	// Add tool result to messages
	a.messages = append(a.messages, llm.Message{
		Role:    llm.RoleTool,
		Content: toolOutput,
	})

	return nil
}

func formatArgs(args map[string]any) string {
	if len(args) == 0 {
		return ""
	}

	var parts []string
	for key, value := range args {
		strVal := fmt.Sprintf("%v", value)
		if len(strVal) > 50 {
			strVal = strVal[:47] + "..."
		}
		parts = append(parts, fmt.Sprintf("%s=%q", key, strVal))
	}

	return strings.Join(parts, ", ")
}
