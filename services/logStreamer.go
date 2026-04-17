package services

import (
	"sync"
)

type LogStreamer struct {
	mu          sync.Mutex
	subscribers map[string][]chan string // deploymentID -> list of subscriber channels
	logs        map[string][]string      // deploymentID -> stored log lines (for replay)
	completed   map[string]bool          // deploymentID -> whether build is done
}

func NewLogStreamer() *LogStreamer {
	return &LogStreamer{
		subscribers: make(map[string][]chan string),
		logs:        make(map[string][]string),
		completed:   make(map[string]bool),
	}
}

func (l *LogStreamer) Subscribe(deploymentID string) chan string {
	// adding a mutex lock
	l.mu.Lock()
	defer l.mu.Unlock() // unlock once the process completes

	// create a buffered channel
	ch := make(chan string, 500)

	// add the channel in subscribers map under its deploymentID
	l.subscribers[deploymentID] = append(l.subscribers[deploymentID], ch)

	// replay any existing logs
	for _, line := range l.logs[deploymentID] {
		ch <- line
	}

	// return the channel
	return ch
}

func (l *LogStreamer) Unsubscribe(deploymentID string, ch chan string) {
	// adding a mutex lock
	l.mu.Lock()
	defer l.mu.Unlock() // unlock once the process completes

	// remove the channel from the subscribers list
	channels := l.subscribers[deploymentID]
	found := false
	for i, sub := range channels {
		if sub == ch {
			// rebuilding the list excluding this subscriber channel
			l.subscribers[deploymentID] = append(channels[:i], channels[i+1:]...)
			found = true
			break
		}
	}

	// only close if we found it (not already closed by Complete)
	if found {
		close(ch)
	}
}

func (l *LogStreamer) Publish(deploymentID string, message string) {
	// adding a mutex lock
	l.mu.Lock()
	defer l.mu.Unlock()

	// adding the message to the logs
	l.logs[deploymentID] = append(l.logs[deploymentID], message)

	// send the message to all subscriber channels for this deploymentID
	for _, subCh := range l.subscribers[deploymentID] {
		select {
		case subCh <- message:
		default:
			continue
		}
	}
}

func (l *LogStreamer) Complete(deploymentID string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// mark the deploymentID as completed
	l.completed[deploymentID] = true

	// close all the subscriber channels for this deploymentID
	channels := l.subscribers[deploymentID]

	for _, subCh := range channels {
		close(subCh)
	}

	// clear the subscribers list for this deploymentID
	delete(l.subscribers, deploymentID)
}

func (l *LogStreamer) IsCompleted(deploymentID string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.completed[deploymentID]
}

func (l *LogStreamer) GetLogs(deploymentID string) []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.logs[deploymentID]
}
