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

	return &Agent{
		client: client,
		tools:  registry,
		messages: []llm.Message{
			{
				Role: llm.RoleSystem,
				Content: `You are Maahinen, a helpful coding assistant.

IMPORTANT RULES FOR TOOL USAGE:
- Only use the bash tool when the use of the tool is actually needed to complete the request.
- For greetings, questions, explanations, or conversation, just respond with text. Do NOT use tools.
- When in doubt, respond with text first and ask if the user wants you to run a command.

Examples of when NOT to use tools:
- "Hello" → Just say your greetings back
- "How do I write a for loop?" → Explain with text
- "What is Go?" → Explain with text

Examples of when to use tools:
- "List files in this directory" → Use bash: ls
- "What's my Go version?" → Use bash: go version
- "Create a file called test.txt" → Use bash: echo "content" > test.txt

Be concise and practical.`,
			},
		},
	}
}

func (a *Agent) Run() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Printf("🧙 Maahinen (%s)\n", a.client.Model())
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
			ui.TypewriterColored("🧙 Goodbye!", ui.BrightMagenta, 30*time.Millisecond)
			return nil
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

		if resp.HasToolCalls() {
			for _, tc := range resp.ToolCalls {
				if err := a.executeTool(tc); err != nil {
					return err
				}
			}
			continue
		}

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

	// Add tool result to messages
	a.messages = append(a.messages, llm.Message{
		Role:    llm.RoleTool,
		Content: result.Output,
	})

	return nil
}
