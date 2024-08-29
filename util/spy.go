package util

import (
	"fmt"
	"reflect"
	"sync"
)

// Spy is a test spy that can be used to record method calls and their arguments.
type Spy struct {
	mutex   *sync.RWMutex
	Methods map[string][]InAndOut
}

// InAndOut represents the input arguments and return value of a method.
type InAndOut struct {
	Skip  int
	Calls int
	In    []interface{}
	Out   error
}

// NewSpy creates a new Spy.
func NewSpy() *Spy {
	return &Spy{
		mutex:   &sync.RWMutex{},
		Methods: make(map[string][]InAndOut),
	}
}

// Register registers a method with its arguments and return value.
func (s *Spy) Register(method string, skip int, err error, args ...interface{}) *Spy {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.Methods[method]; !ok {
		s.Methods[method] = []InAndOut{}
	}

	s.Methods[method] = append(s.Methods[method], InAndOut{Skip: skip, In: args, Out: err, Calls: 0})

	return s
}

// Anything is a special type that can be used to match any argument.
type Anything struct{}

// Any is a special value that can be used to match any argument.
var Any = Anything{} //nolint:gochecknoglobals // This is a special value, not intended to be modified.

// GetError returns a pre-defined error for a given method and arguments.
func (s *Spy) GetError(method string, args ...interface{}) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if _, ok := s.Methods[method]; !ok {
		return nil
	}

	for i, inAndOut := range s.Methods[method] {
		if s.isMatch(inAndOut, args...) {
			if inAndOut.Calls < inAndOut.Skip {
				s.Methods[method][i].Calls++

				return nil
			}

			return inAndOut.Out
		}
	}

	return nil
}

func (s *Spy) isMatch(inAndOut InAndOut, args ...interface{}) bool {
	if len(inAndOut.In) != len(args) {
		panic(fmt.Sprintf("invalid number of arguments. expected %d, got %d", len(inAndOut.In), len(args)))
	}

	match := true

	for i, arg := range args {
		if inAndOut.In[i] == Any {
			continue
		}

		if !reflect.DeepEqual(arg, inAndOut.In[i]) {
			match = false

			break
		}
	}

	return match
}

func (s *Spy) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Methods = make(map[string][]InAndOut)
}
