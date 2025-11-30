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

func FileReadToolDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolDefinition{
			Name:        "file_read",
			Description: "Read the contents of a file",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"path": {
						Type:        "string",
						Description: "Path to the file to read",
					},
				},
				Required: []string{"path"},
			},
		},
	}
}

func FileWriteToolDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolDefinition{
			Name:        "file_write",
			Description: "Create or overwrite a file with content",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"path": {
						Type:        "string",
						Description: "Path to the file to write",
					},
					"content": {
						Type:        "string",
						Description: "Content to write to the file",
					},
				},
				Required: []string{"path", "content"},
			},
		},
	}
}

func FileEditToolDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolDefinition{
			Name:        "file_edit",
			Description: "Edit a file by finding and replacing a specific string",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"path": {
						Type:        "string",
						Description: "Path to the file to edit",
					},
					"old_string": {
						Type:        "string",
						Description: "The exact string to find and replace",
					},
					"new_string": {
						Type:        "string",
						Description: "The string to replace it with (empty to delete)",
					},
				},
				Required: []string{"path", "old_string", "new_string"},
			},
		},
	}
}

func FileListToolDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolDefinition{
			Name:        "file_list",
			Description: "List files and directories in a path",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"path": {
						Type:        "string",
						Description: "Path to the directory to list (defaults to current directory)",
					},
				},
				Required: []string{},
			},
		},
	}
}
