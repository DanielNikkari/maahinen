package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileReadTool struct {
	workDir string
}

func NewFileReadTool(workDir string) *FileReadTool {
	return &FileReadTool{workDir: workDir}
}

func (t *FileReadTool) Name() string        { return "file_read" }
func (t *FileReadTool) Description() string { return "Read the contents of a file" }

func (t *FileReadTool) Execute(ctc context.Context, args map[string]any) (Result, error) {
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

type FileWriteTool struct {
	workDir string
}

func NewFileWriteTool(workDir string) *FileWriteTool {
	return &FileWriteTool{workDir: workDir}
}

func (t *FileWriteTool) Name() string        { return "file_write" }
func (t *FileWriteTool) Description() string { return "Create or overwrite a file with content" }

func (t *FileWriteTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
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

type FileEditTool struct {
	workDir string
}

func NewFileEditTool(workDir string) *FileEditTool {
	return &FileEditTool{workDir: workDir}
}

func (t *FileEditTool) Name() string        { return "file_edit" }
func (t *FileEditTool) Description() string { return "Edit a file by replacing a specific string" }

func (t *FileEditTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
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

type FileListTool struct {
	workDir string
}

func NewFileListTool(workDir string) *FileListTool {
	return &FileListTool{workDir: workDir}
}

func (t *FileListTool) Name() string        { return "file_list" }
func (t *FileListTool) Description() string { return "List files and directories in a path" }

func (t *FileListTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
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
