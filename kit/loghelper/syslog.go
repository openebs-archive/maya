package loghelper

import (
	"bytes"
	"github.com/hashicorp/go-syslog"
	"github.com/hashicorp/logutils"
)

// levelPriority is used to map a log level to a
// syslog priority level
var levelPriority = map[string]gsyslog.Priority{
	"TRACE": gsyslog.LOG_DEBUG,
	"DEBUG": gsyslog.LOG_INFO,
	"INFO":  gsyslog.LOG_NOTICE,
	"WARN":  gsyslog.LOG_WARNING,
	"ERR":   gsyslog.LOG_ERR,
	"CRIT":  gsyslog.LOG_CRIT,
}

// SyslogWriter is used to cleaup log messages before
// writing them to a Syslogger.
//
// Implements the io.Writer interface.
type SyslogWriter struct {
	GSyslog gsyslog.Syslogger
	LFilter *logutils.LevelFilter
}

// Write is used to implement io.Writer
func (s *SyslogWriter) Write(p []byte) (int, error) {
	// Skip syslog if the log level doesn't apply
	if !s.LFilter.Check(p) {
		return 0, nil
	}

	// Segregate log level from actual log message
	var level string
	afterLevel := p
	x := bytes.IndexByte(p, '[')
	if x >= 0 {
		y := bytes.IndexByte(p[x:], ']')
		if y >= 0 {
			level = string(p[x+1 : x+y])
			afterLevel = p[x+y+2:]
		}
	}

	// Each log level will be handled by a specific syslog priority
	priority, ok := levelPriority[level]
	if !ok {
		priority = gsyslog.LOG_NOTICE
	}

	// Attempt the write
	err := s.GSyslog.WriteLevel(priority, afterLevel)
	return len(p), err
}
