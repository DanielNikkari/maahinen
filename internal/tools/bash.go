package tools

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

type BashTool struct {
	workDir string
	timeout time.Duration
}

func NewBashTool(workDir string) *BashTool {
	return &BashTool{
		workDir: workDir,
		timeout: 30 * time.Second,
	}
}

func (b *BashTool) Name() string {
	return "bash"
}

func (b *BashTool) Description() string {
	return "Execute a bash command and return the output"
}

func (b *BashTool) Execute(ctx context.Context, args map[string]any) (Result, error) {
	command, ok := args["command"].(string)
	if !ok || command == "" {
		return Result{
			Success: false,
			Error:   "missing or invalid 'command' argument",
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	if b.workDir != "" {
		cmd.Dir = b.workDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := strings.TrimSpace(stdout.String())
	errOutput := strings.TrimSpace(stderr.String())

	if err != nil {
		combinedOutput := output
		if errOutput != "" {
			if combinedOutput != "" {
				combinedOutput += "\n"
			}
			combinedOutput += errOutput
		}

		return Result{
			Success: false,
			Output:  combinedOutput,
			Error:   err.Error(),
		}, nil
	}

	if errOutput != "" && output != "" {
		output = output + "\n" + errOutput
	} else if errOutput != "" {
		output = errOutput
	}

	return Result{
		Success: true,
		Output:  output,
	}, nil
}

func (b *BashTool) SetTimeout(d time.Duration) {
	b.timeout = d
}

func (b *BashTool) SetWorkDir(dir string) {
	b.workDir = dir
}
