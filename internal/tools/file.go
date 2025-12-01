package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DanielNikkari/maahinen/internal/llm"
)

type ReadTool struct {
	workDir string
}

func NewReadTool(workDir string) *ReadTool {
	return &ReadTool{workDir: workDir}
}

func (t *ReadTool) Name() string        { return "read" }
func (t *ReadTool) Description() string { return "Read the contents of a file" }

func (t *ReadTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return Result{Success: false, Error: "missing 'path' argument"}, nil
	}

	if !filepath.IsAbs(path) && t.workDir != "" {
		path = filepath.Join(t.workDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Result{Success: false, Error: err.Error()}, nil
	}

	return Result{Success: true, Output: string(content)}, nil
}

func ReadToolDefinition() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolDefinition{
			Name:        "read",
			Description: "Read the contents of a file",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
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

type WriteTool struct {
	workDir string
}

func NewWriteTool(workDir string) *WriteTool {
	return &WriteTool{workDir: workDir}
}

func (t *WriteTool) Name() string        { return "write" }
func (t *WriteTool) Description() string { return "Create or overwrite a file with content" }

func (t *WriteTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return Result{Success: false, Error: "missing 'path' argument"}, nil
	}

	content, ok := args["content"].(string)
	if !ok {
		return Result{Success: false, Error: "missing 'content' argument"}, nil
	}

	if !filepath.IsAbs(path) && t.workDir != "" {
		path = filepath.Join(t.workDir, path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return Result{Success: false, Error: err.Error()}, nil
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return Result{Success: false, Error: err.Error()}, nil
	}

	return Result{Success: true, Output: fmt.Sprintf("File written: %s", path)}, nil
}

func WriteToolDefinition() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolDefinition{
			Name:        "write",
			Description: "Create or overwrite a file with content",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
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

type EditTool struct {
	workDir string
}

func NewEditTool(workDir string) *EditTool {
	return &EditTool{workDir: workDir}
}

func (t *EditTool) Name() string        { return "edit" }
func (t *EditTool) Description() string { return "Edit a file by replacing a specific string" }

func (t *EditTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		return Result{Success: false, Error: "missing 'path' argument"}, nil
	}

	oldStr, ok := args["old_string"].(string)
	if !ok || oldStr == "" {
		return Result{Success: false, Error: "missing 'old_string' argument"}, nil
	}

	newStr, _ := args["new_string"].(string) // Can be empty (deletion)

	if !filepath.IsAbs(path) && t.workDir != "" {
		path = filepath.Join(t.workDir, path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Result{Success: false, Error: err.Error()}, nil
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, oldStr) {
		return Result{Success: false, Error: "old_string not found in file"}, nil
	}

	newContent := strings.Replace(contentStr, oldStr, newStr, 1)

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return Result{Success: false, Error: err.Error()}, nil
	}

	return Result{Success: true, Output: fmt.Sprintf("File edited: %s", path)}, nil
}

func EditToolDefinition() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolDefinition{
			Name:        "edit",
			Description: "Edit a file by finding and replacing a specific string",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
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

type ListTool struct {
	workDir string
}

func NewListTool(workDir string) *ListTool {
	return &ListTool{workDir: workDir}
}

func (t *ListTool) Name() string        { return "list" }
func (t *ListTool) Description() string { return "List files and directories in a path" }

func (t *ListTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
	path, ok := args["path"].(string)
	if !ok || path == "" {
		path = "."
	}

	if !filepath.IsAbs(path) && t.workDir != "" {
		path = filepath.Join(t.workDir, path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return Result{Success: false, Error: err.Error()}, nil
	}

	var lines []string
	for _, entry := range entries {
		info, _ := entry.Info()
		if entry.IsDir() {
			lines = append(lines, fmt.Sprintf("[DIR]  %s/", entry.Name()))
		} else {
			size := ""
			if info != nil {
				size = fmt.Sprintf("(%d bytes)", info.Size())
			}
			lines = append(lines, fmt.Sprintf("[FILE] %s %s", entry.Name(), size))
		}
	}

	return Result{Success: true, Output: strings.Join(lines, "\n")}, nil
}

func ListToolDefinition() llm.Tool {
	return llm.Tool{
		Type: "function",
		Function: llm.ToolDefinition{
			Name:        "list",
			Description: "List files and directories in a path",
			Parameters: llm.Parameters{
				Type: "object",
				Properties: map[string]llm.Property{
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

func (t *ReadTool) Definition() llm.Tool {
	return ReadToolDefinition()
}

func (t *WriteTool) Definition() llm.Tool {
	return WriteToolDefinition()
}

func (t *EditTool) Definition() llm.Tool {
	return EditToolDefinition()
}

func (t *ListTool) Definition() llm.Tool {
	return ListToolDefinition()
}
