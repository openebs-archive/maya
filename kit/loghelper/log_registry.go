package loghelper

// This is an adaptation of Hashicorp's Nomad library
import (
	"sync"
)

// LogHandler interface is used for clients that want to subscribe
// to logs, for example to stream them over an IPC mechanism
type LogHandler interface {
	HandleLog(string)
}

// LogRegistrar can be used as a log sink.
// It maintains a circular buffer of logs, and a set of handlers to
// which it can stream the logs to.
//
// LogRegistrar implements:
//  1. io.Writer interface
// LogRegistrar composes:
//  1. LogHandler interface
type LogRegistrar struct {
	// embedded mutex
	sync.Mutex

	// logs holds names of log handlers
	logs []string

	index int

	// registry of handlers
	registry map[LogHandler]struct{}
}

// NewLogRegistrar creates a LogRegistrar with the given buffer capacity
func NewLogRegistrar(buf int) *LogRegistrar {
	return &LogRegistrar{
		logs:     make([]string, buf),
		index:    0,
		registry: make(map[LogHandler]struct{}),
	}
}

// RegisterHandler adds a log handler to receive logs, and sends
// the last buffered logs to the handler
func (l *LogRegistrar) RegisterHandler(lh LogHandler) {
	l.Lock()
	defer l.Unlock()

	// Do nothing if already registered
	if _, ok := l.registry[lh]; ok {
		return
	}

	// Register
	l.registry[lh] = struct{}{}

	// Send the old logs
	if l.logs[l.index] != "" {
		for i := l.index; i < len(l.logs); i++ {
			lh.HandleLog(l.logs[i])
		}
	}
	for i := 0; i < l.index; i++ {
		lh.HandleLog(l.logs[i])
	}
}

// DeregisterHandler removes a LogHandler and prevents more invocations
func (l *LogRegistrar) DeregisterHandler(lh LogHandler) {
	l.Lock()
	defer l.Unlock()
	delete(l.registry, lh)
}

// Write is used to accumulate new logs
func (l *LogRegistrar) Write(p []byte) (n int, err error) {
	l.Lock()
	defer l.Unlock()

	// Strip off newlines at the end if there are any since we store
	// individual log lines in the agent.
	n = len(p)
	if p[n-1] == '\n' {
		p = p[:n-1]
	}

	l.logs[l.index] = string(p)
	l.index = (l.index + 1) % len(l.logs)

	for lh, _ := range l.registry {
		lh.HandleLog(string(p))
	}
	return
}
