package ollama

import (
	"fmt"
	"net/http"
	"os/exec"
	"time"
)

const defaultURL = "http://localhost:11434"

func IsRunningAt(baseURL string) bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func IsRunning() bool {
	return IsRunningAt(defaultURL)
}

func Start() error {
	cmd := exec.Command("ollama", "serve")
	if err := cmd.Start(); err != nil {
		return err
	}
	for range 60 {
		if IsRunning() {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("ollama failed to start within 30 seconds")
}
