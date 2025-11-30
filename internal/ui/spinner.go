package ui

import (
	"fmt"
	"sync"
	"time"
)

type Spinner struct {
	frames   []string
	interval time.Duration
	stop     chan struct{}
	wg       sync.WaitGroup
}

func NewSpinner(style string) *Spinner {
	frames := getFrames(style)
	return &Spinner{
		frames:   frames,
		interval: 100 * time.Millisecond,
		stop:     make(chan struct{}),
	}
}

func getFrames(style string) []string {
	switch style {
	case "dots":
		return []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	case "wizard":
		return []string{"🧙", "🧙‍♂️", "✨", "🪄", "✨", "🧙‍♂️"}
	case "moon":
		return []string{"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"}
	case "bounce":
		return []string{"⠁", "⠂", "⠄", "⠂"}
	case "arrows":
		return []string{"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"}
	case "thinking":
		return []string{"🤔", "💭", "🧠", "💡", "🧠", "💭"}
	default: // classic
		return []string{"|", "/", "-", "\\"}
	}
}

func (s *Spinner) Start(message string) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Print("\r\033[K") // Clear line
				return
			default:
				fmt.Printf("\r%s %s", s.frames[i], message)
				i = (i + 1) % len(s.frames)
				time.Sleep(s.interval)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	close(s.stop)
	s.wg.Wait()
}
