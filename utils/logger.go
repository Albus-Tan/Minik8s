package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

var lck sync.Mutex

type ComponentLogger interface {
	Flags() int
	SetFlags(flag int)
	Prefix() string
	SetPrefix(prefix string)
	Writer() io.Writer
	Print(v ...any)
	Printf(format string, v ...any)
	Println(v ...any)
	Fatal(v ...any)
	Fatalf(format string, v ...any)
	Fatalln(v ...any)
	Panic(v ...any)
	Panicf(format string, v ...any)
	Panicln(v ...any)
	Output(calldepth int, s string) error
}

// componentLogger is for components log to same Output
type componentLogger struct {
	l    *log.Logger
	lock *sync.Mutex
}

func NewComponentLogger(componentName string) ComponentLogger {
	prefix := formatPrefix(componentName)
	return &componentLogger{
		l:    log.New(os.Stderr, prefix, log.LstdFlags),
		lock: &lck,
	}
}

func formatPrefix(componentName string) string {
	return fmt.Sprintf("[%v] ", componentName)
}

//// SetOutput sets the output destination for the standard logger.
//func (c *componentLogger) SetOutput(w io.Writer) {
//	c.l.SetOutput(w)
//}

// Flags returns the output flags for the standard logger.
// The flag bits are Ldate, Ltime, and so on.
func (c *componentLogger) Flags() int {
	return c.l.Flags()
}

// SetFlags sets the output flags for the standard logger.
// The flag bits are Ldate, Ltime, and so on.
func (c *componentLogger) SetFlags(flag int) {
	c.l.SetFlags(flag)
}

// Prefix returns the output prefix for the standard logger.
func (c *componentLogger) Prefix() string {
	return c.l.Prefix()
}

// SetPrefix sets the output prefix for the standard logger.
func (c *componentLogger) SetPrefix(prefix string) {
	c.l.SetPrefix(prefix)
}

// Writer returns the output destination for the standard logger.
func (c *componentLogger) Writer() io.Writer {
	return c.l.Writer()
}

// These functions write to the standard logger.

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func (c *componentLogger) Print(v ...any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.l.Output(2, fmt.Sprint(v...))
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (c *componentLogger) Printf(format string, v ...any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.l.Output(2, fmt.Sprintf(format, v...))
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func (c *componentLogger) Println(v ...any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.l.Output(2, fmt.Sprintln(v...))
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func (c *componentLogger) Fatal(v ...any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.l.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func (c *componentLogger) Fatalf(format string, v ...any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.l.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func (c *componentLogger) Fatalln(v ...any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.l.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic is equivalent to Print() followed by a call to panic().
func (c *componentLogger) Panic(v ...any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	s := fmt.Sprint(v...)
	c.l.Output(2, s)
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func (c *componentLogger) Panicf(format string, v ...any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	s := fmt.Sprintf(format, v...)
	c.l.Output(2, s)
	panic(s)
}

// Panicln is equivalent to Println() followed by a call to panic().
func (c *componentLogger) Panicln(v ...any) {
	c.lock.Lock()
	defer c.lock.Unlock()

	s := fmt.Sprintln(v...)
	c.l.Output(2, s)
	panic(s)
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is the count of the number of
// frames to skip when computing the file name and line number
// if Llongfile or Lshortfile is set; a value of 1 will print the details
// for the caller of Output.
func (c *componentLogger) Output(calldepth int, s string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.l.Output(calldepth+1, s) // +1 for this frame.
}
