package ui

import (
	"fmt"
	"sync"
	"time"
)

type Spinner struct {
	mu       sync.Mutex
	message  string
	stopChan chan struct{}
	stopped  bool
}

func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
	}
}

func (s *Spinner) Start() {
	s.stopChan = make(chan struct{})
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	go func() {
		i := 0
		for {
			select {
			case <-s.stopChan:
				fmt.Print("\r\033[K")
				return
			default:
				s.mu.Lock()
				msg := s.message
				s.mu.Unlock()
				fmt.Printf("\r\033[K\033[36m%s\033[0m %s", frames[i%len(frames)], msg)
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop(finalMsg string) {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	s.mu.Unlock()
	close(s.stopChan)
	if finalMsg != "" {
		fmt.Println(finalMsg)
	}
}

func (s *Spinner) UpdateMessage(msg string) {
	s.mu.Lock()
	s.message = msg
	s.mu.Unlock()
}
