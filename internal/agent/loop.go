package agent

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"maahinen/internal/llm"
	"maahinen/internal/ui"
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
}

func NewAgent(client *llm.Client) *Agent {
	return &Agent{
		client: client,
		messages: []llm.Message{
			{
				Role:    llm.RoleSystem,
				Content: "You are Maahinen, a helpful coding assistant. You help users with programming tasks, answer questions about code, and assist with debugging. Be concise and practical.",
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

		spinner := ui.NewSpinner("wizard")
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		actionAdj := actionAdjectives[r.Intn(len(actionAdjectives))]
		spinner.Start(ui.ActionAdjPrompt(actionAdj))

		resp, err := a.client.Chat(a.messages)

		spinner.Stop()

		if err != nil {
			fmt.Printf("\nError: %v\n\n", err)
			a.messages = a.messages[:len(a.messages)-1]
			continue
		}

		fmt.Printf("%s%s\n\n", ui.AssistantPrompt(), resp.Content)

		a.messages = append(a.messages, *resp)
	}
}
