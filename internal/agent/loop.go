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
}

func NewAgent(client *llm.Client, registry *tools.Registry) *Agent {
	client.RegisterTool(llm.BashToolDefinition())
	client.RegisterTool(llm.FileReadToolDefinition())
	client.RegisterTool(llm.FileWriteToolDefinition())
	client.RegisterTool(llm.FileEditToolDefinition())
	client.RegisterTool(llm.FileListToolDefinition())

	return &Agent{
		client: client,
		tools:  registry,
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
	fmt.Printf("🧙 %s (%s)\n", ui.Color(ui.BrightMagenta, "Maahinen"), a.client.Model())
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
		spinner := ui.NewSpinner("wizard")
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
		if tc, ok := llm.ParseToolCallFromContent(resp.Content); ok {
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
	tool, ok := a.tools.Get(tc.Function.Name)
	if !ok {
		return fmt.Errorf("unknown tool: %s", tc.Function.Name)
	}

	fmt.Printf("%s Running: %s\n", ui.Color(ui.Yellow, "⚡"), tc.Function.Name)
	if cmd, ok := tc.Function.Arguments["command"].(string); ok {
		fmt.Printf("   %s\n", ui.Color(ui.Dim, cmd))
	}

	result, err := tool.Execute(context.Background(), tc.Function.Arguments)
	if err != nil {
		return err
	}

	if result.Success {
		fmt.Printf("%s\n", ui.Color(ui.BrightGreen, "✓ Success"))
	} else {
		fmt.Printf("%s %s\n", ui.Color(ui.Red, "✗ Failed:"), result.Error)
	}
	if result.Output != "" {
		fmt.Printf("%s\n", result.Output)
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
