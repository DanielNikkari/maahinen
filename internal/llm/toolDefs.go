package llm

func BashToolDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolDefinition{
			Name:        "bash",
			Description: "Execute a bash command on the system. Use this to run shell commands, check files, install packages, etc.",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"command": {
						Type:        "string",
						Description: "The bash command to execute",
					},
				},
				Required: []string{"command"},
			},
		},
	}
}
