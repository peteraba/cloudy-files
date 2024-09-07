package test

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// FakeDisplay is a fake display for test.
type FakeDisplay struct {
	mutex              *sync.Mutex
	out                []string
	containsAssertions []string
	t                  *testing.T
}

// NewFakeDisplay creates a new FakeDisplay instance.
func NewFakeDisplay(t *testing.T) *FakeDisplay {
	t.Helper()

	return &FakeDisplay{
		mutex:              &sync.Mutex{},
		out:                make([]string, 0),
		containsAssertions: nil,
		t:                  t,
	}
}

// QueueContainsAssertion queues an assertion for later execution.
func (f *FakeDisplay) QueueContainsAssertion(contains string) {
	f.containsAssertions = append(f.containsAssertions, contains)
}

// Assert executes the queue assertions.
func (f *FakeDisplay) Assert() {
	str := f.String()

	for _, contains := range f.containsAssertions {
		assert.Contains(f.t, str, contains)
	}
}

// Println prints the arguments to the output.
func (f *FakeDisplay) Println(args ...interface{}) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.out = append(f.out, fmt.Sprintln(args...))
}

// String returns the output as a string.
func (f *FakeDisplay) String() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	return strings.Join(f.out, "")
}

// ExitWithHelp exits the application with a help message.
func (f *FakeDisplay) ExitWithHelp(msg, help string) {
	f.Println("msg:", msg)
	f.Println("---")
	f.Println(help)

	f.Assert()

	// This will exit the test (or any goroutine it's called from)
	f.t.SkipNow()
}

// Exit exits the application after displaying a message and an error.
func (f *FakeDisplay) Exit(msg string, err error) {
	f.Println("msg:", msg)
	f.Println("error:", err)

	f.Assert()

	// This will exit the test (or any goroutine it's called from)
	f.t.SkipNow()
}
