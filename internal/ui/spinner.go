package ui

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

var Spinners = map[string][]string{
	"dots":     {"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
	"wizard":   {"🧙", "🧙‍♂️", "✨", "🪄", "✨", "🧙‍♂️"},
	"moon":     {"🌑", "🌒", "🌓", "🌔", "🌕", "🌖", "🌗", "🌘"},
	"bounce":   {"⠁", "⠂", "⠄", "⠂"},
	"arrows":   {"←", "↖", "↑", "↗", "→", "↘", "↓", "↙"},
	"thinking": {"🤔", "💭", "🧠", "💡", "🧠", "💭"},
	"classic":  {"|", "/", "-", "\\"},
}

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
	if frames, ok := Spinners[style]; ok {
		return frames
	}
	return Spinners["dots"]
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

func ListSpinners() []string {
	spinners := make([]string, 0, len(Spinners))
	for name := range Spinners {
		spinners = append(spinners, name)
	}
	sort.Strings(spinners)
	return spinners
}
