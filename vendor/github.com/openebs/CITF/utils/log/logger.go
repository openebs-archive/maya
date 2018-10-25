package log

import (
	"fmt"
	"io"
	"os"

	"github.com/golang/glog"
)

// DebugEnabled specifies if this package print debug information
var DebugEnabled = false

// Logger is a struct which will help to call CITF specific logging functions
type Logger struct{}

// WritefDebugMessage formats according to a format specifier and writes to w only when DebugEnabled is true.
// A newline is always appended. It returns the number of bytes written and any write error encountered.
func (logger Logger) WritefDebugMessage(w io.Writer, format string, a ...interface{}) (n int, err error) {
	if DebugEnabled {
		return fmt.Fprintf(w, format+"\n", a...)
	}
	return
}

// PrintfDebugMessage formats according to a format specifier and writes to standard output only when DebugEnabled is true.
// //  A newline is always appended. It returns the number of bytes written and any write error encountered.
func (logger Logger) PrintfDebugMessage(format string, a ...interface{}) (n int, err error) {
	return logger.WritefDebugMessage(os.Stdout, format, a...)
}

// WritelnDebugMessage formats using the default formats for its operands and writes to w only when DebugEnabled.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (logger Logger) WritelnDebugMessage(w io.Writer, a ...interface{}) (n int, err error) {
	if DebugEnabled {
		return fmt.Fprintln(w, a...)
	}
	return
}

// PrintlnDebugMessage formats using the default formats for its operands and writes to standard output only when DebugEnabled.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (logger Logger) PrintlnDebugMessage(a ...interface{}) (n int, err error) {
	return logger.WritelnDebugMessage(os.Stdout, a...)
}

// LogError logs error using `glog.Error` only when err is not nil.
// Please follow convensions for error message e.g. start with lowercase, don't end with period etc.
func (logger Logger) LogError(err error, message string) {
	if err != nil {
		glog.Error(message+":", err)
	}
}

// LogNonError logs info using `glog.Info` only when err is nil.
func (logger Logger) LogNonError(err error, message string) {
	if err == nil {
		glog.Info(message)
	}
}

// LogErrorf formats according to a format specifier and logs error using `glog.Error` only when err is not nil.
// Please follow convensions for error message e.g. start with lowercase, don't end with period etc.
func (logger Logger) LogErrorf(err error, message string, a ...interface{}) {
	if err != nil {
		glog.Error(fmt.Sprintf(message, a...)+":", err)
	}
}

// LogNonErrorf logs info using `glog.Infof` only when err is nil.
// formatting is taken care by `glog.Infof`
func (logger Logger) LogNonErrorf(err error, message string, a ...interface{}) {
	if err == nil {
		glog.Infof(message, a...)
	}
}

// LogFatal logs error using `glog.Error` only when err is not nil.
// Please follow convensions for error message e.g. start with lowercase, don't end with period etc.
func (logger Logger) LogFatal(err error, message string) {
	if err != nil {
		glog.Fatal(message+":", err)
	}
}

// LogFatalf formats according to a format specifier and logs error using `glog.Error` only when err is not nil.
// Please follow convensions for error message e.g. start with lowercase, don't end with period etc.
func (logger Logger) LogFatalf(err error, message string, a ...interface{}) {
	if err != nil {
		glog.Fatal(fmt.Sprintf(message, a...)+":", err)
	}
}

// PrintError wtires error message to os.StdErr only when err is not nil.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
// Please follow convensions for error message e.g. start with lowercase, don't end with period etc.
func (logger Logger) PrintError(err error, message string) (n int, errr error) {
	if err != nil {
		return fmt.Fprintln(os.Stderr, message+":", err)
	}
	return
}

// PrintNonError wtites info message to os.Stdout only when err is nil.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (logger Logger) PrintNonError(err error, message string) (n int, errr error) {
	if err == nil {
		return fmt.Println(message)
	}
	return
}

// PrintErrorf formats according to a format specifier and wtires error message to os.StdErr only when err is not nil.
// A newline is always appended. It returns the number of bytes written and any write error encountered.
// Please follow convensions for error message e.g. start with lowercase, don't end with period etc.
func (logger Logger) PrintErrorf(err error, message string, a ...interface{}) (n int, errr error) {
	if err != nil {
		a = append(a, err)
		return fmt.Fprintf(os.Stderr, message+":%+v\n", a...)
	}
	return
}

// PrintNonErrorf formats according to a format specifier and wtites info message to os.Stdout only when err is nil.
// A newline is always appended. It returns the number of bytes written and any write error encountered.
func (logger Logger) PrintNonErrorf(err error, message string, a ...interface{}) (n int, errr error) {
	if err == nil {
		return fmt.Printf(message+"\n", a...)
	}
	return
}

// WritefDebugMessageIfError formats according to a format specifier and writes to w only when DebugEnabled is true and err is not nil.
//  A newline is always appended. It returns the number of bytes written and any write error encountered.
func (logger Logger) WritefDebugMessageIfError(err error, w io.Writer, format string, a ...interface{}) (n int, errr error) {
	if err != nil {
		a = append(a, err)
		return logger.WritefDebugMessage(w, format+": %+v", a...)
	}
	return
}

// PrintfDebugMessageIfError formats according to a format specifier and writes to standard output only when DebugEnabled is true and err is not nil.
// It returns the number of bytes written and any write error encountered.
func (logger Logger) PrintfDebugMessageIfError(err error, format string, a ...interface{}) (n int, errr error) {
	return logger.WritefDebugMessageIfError(err, os.Stderr, format, a...)
}

// WritelnDebugMessageIfError formats using the default formats for its operands and writes to w only when DebugEnabled and err is not nil.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (logger Logger) WritelnDebugMessageIfError(err error, w io.Writer, a ...interface{}) (n int, errr error) {
	if err != nil {
		return logger.WritefDebugMessage(w, fmt.Sprintln(a...)+": %+v", err)
	}
	return
}

// PrintlnDebugMessageIfError formats using the default formats for its operands and writes to standard output only when DebugEnabled and err is not nil.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (logger Logger) PrintlnDebugMessageIfError(err error, a ...interface{}) (n int, errr error) {
	return logger.WritelnDebugMessageIfError(err, os.Stderr, a...)
}

// WritefDebugMessageIfNotError formats according to a format specifier and writes to w only when DebugEnabled is true and err is nil.
//  A newline is always appended. It returns the number of bytes written and any write error encountered.
func (logger Logger) WritefDebugMessageIfNotError(err error, w io.Writer, format string, a ...interface{}) (n int, errr error) {
	if err == nil {
		return logger.WritefDebugMessage(w, format+": %+v", a...)
	}
	return
}

// PrintfDebugMessageIfNotError formats according to a format specifier and writes to standard output only when DebugEnabled is true and err is nil.
// It returns the number of bytes written and any write error encountered.
func (logger Logger) PrintfDebugMessageIfNotError(err error, format string, a ...interface{}) (n int, errr error) {
	return logger.WritefDebugMessageIfNotError(err, os.Stderr, format, a...)
}

// WritelnDebugMessageIfNotError formats using the default formats for its operands and writes to w only when DebugEnabled and err is nil.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (logger Logger) WritelnDebugMessageIfNotError(err error, w io.Writer, a ...interface{}) (n int, errr error) {
	if err == nil {
		return logger.WritelnDebugMessage(w, a...)
	}
	return
}

// PrintlnDebugMessageIfNotError formats using the default formats for its operands and writes to standard output only when DebugEnabled and err is nil.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (logger Logger) PrintlnDebugMessageIfNotError(err error, a ...interface{}) (n int, errr error) {
	return logger.WritelnDebugMessageIfNotError(err, os.Stderr, a...)
}
