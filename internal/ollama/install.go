package ollama

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func IsInstalled() bool {
	_, err := exec.LookPath("ollama")
	return err == nil
}

func Install() error {
	switch runtime.GOOS {
	case "darwin":
		return installMacOS()
	case "linux":
		return installLinux()
	default:
		return fmt.Errorf("automatic install not supported on %s system, please install manually from https://ollama.com", runtime.GOOS)
	}

}

func installMacOS() error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("homebrew not found, please install brew first of install Ollama manually from https://ollama.com")
	}

	fmt.Println("Installing Ollama via Homebrew...")
	cmd := exec.Command("brew", "install", "ollama")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installLinux() error {
	if _, err := exec.LookPath("curl"); err != nil {
		return fmt.Errorf("curl not found, please install curl first or install Ollama manually")
	}

	fmt.Println("Installing Ollama via official install script...")
	cmd := exec.Command("sh", "-c", "curl -fsSl https://ollama.com/install.sh | sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
