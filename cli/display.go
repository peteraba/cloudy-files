package cli

import (
	"fmt"
	"os"
)

// Display is an interface for displaying output.
type Display interface {
	Println(args ...interface{})
	ExitWithHelp(msg, help string)
	Exit(msg string, err error)
}

// Stdout is the standard output.
type Stdout struct{}

// NewStdout creates a new Stdout instance.
func NewStdout() *Stdout {
	return &Stdout{}
}

// Println prints the arguments to the standard output.
func (s *Stdout) Println(args ...interface{}) {
	fmt.Println(args...) //nolint:forbidigo // This is meant to output to the standard output
}

// ExitWithHelp exits the application with a help message.
func (s *Stdout) ExitWithHelp(msg, help string) {
	s.Println("msg:", msg)
	s.Println("---")
	s.Println(help)

	os.Exit(1)
}

// Exit exits the application after displaying a message and an error.
func (s *Stdout) Exit(msg string, err error) {
	s.Println("msg:", msg)
	s.Println("error:", err)

	os.Exit(1)
}
