package xerrors_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/JavierZunzunegui/xerrors"
)

// Example_formatter shows how a single error can be converted to a string in many different ways
func Example_formatter() {
	err := xerrors.Wrap(
		func() error {
			return xerrors.Wrap(
				func() error { return xerrors.Wrap(nil, xerrors.New("cause")) }(),
				xerrors.New("bar"),
			)
		}(),
		xerrors.New("foo"),
	)

	// the printers have no special meaning, just doing several different implementations to showcase the functionality

	fmt.Println("default:")
	fmt.Println(err.Error())
	fmt.Println("")
	fmt.Println("short stacks:")
	fmt.Println(shortStackPrinter.String(err))
	fmt.Println("")
	fmt.Println("default and short stacks:")
	fmt.Println(defaultAndShortStackPrinter.String(err))
	fmt.Println("")
	fmt.Println("JSON:")
	fmt.Println(jsonPrinter.String(err))
	fmt.Println("")
	fmt.Println("reverse default:")
	fmt.Println(reverseColonPrinter.String(err))

	//Output:
	//default:
	//foo: bar: cause
	//
	//short stacks:
	//xerrors_test.Example_formatter.func1.1 - xerrors_test.Example_formatter.func1 - xerrors_test.Example_formatter - testing.runExample - testing.runExamples - testing.(*M).Run - main.main - runtime.main - runtime.goexit
	//
	//default and short stacks:
	//foo: bar: cause (xerrors_test.Example_formatter.func1.1 - xerrors_test.Example_formatter.func1 - xerrors_test.Example_formatter - testing.runExample - testing.runExamples - testing.(*M).Run - main.main - runtime.main - runtime.goexit)
	//
	//JSON:
	//["foo","bar","cause"]
	//
	//reverse default:
	//cause: bar: foo
}

var shortStackPrinter = xerrors.NewPrinter(func() xerrors.Formatter { return &shortStackFormatter{} })

type shortStackFormatter struct {
	currentErr *xerrors.WrappingError
	firstEntry bool
}

func (s *shortStackFormatter) Init(wErr *xerrors.WrappingError) {
	s.currentErr = wErr
	s.firstEntry = true
}

func isStackError(err error) bool {
	_, ok := err.(*xerrors.StackError)
	return ok
}

func (s *shortStackFormatter) Next() error {
	for ; s.currentErr != nil && !isStackError(s.currentErr.Payload()); s.currentErr = s.currentErr.Next() {
	}

	if s.currentErr == nil {
		return nil
	}

	out := s.currentErr.Payload()
	s.currentErr = s.currentErr.Next()

	return out
}

func (s *shortStackFormatter) CustomFormat(err error, buf *bytes.Buffer) bool {
	stackErr := err.(*xerrors.StackError)
	frames := stackErr.Frames()

	frame, ok := frames.Next()
	shortFormatFrame(frame, buf)

	for ok {
		frame, ok = frames.Next()
		buf.WriteString(" - ")
		shortFormatFrame(frame, buf)
	}

	return true
}

func shortFormatFrame(frame runtime.Frame, buf *bytes.Buffer) {
	if frame.Function != "" {
		f := frame.Function
		if i := strings.LastIndex(f, "/"); i != -1 {
			f = f[i+1:]
		}
		buf.WriteString(f)
	} else {
		buf.WriteString("<unknown>")
	}
}

func (s *shortStackFormatter) Append(w *bytes.Buffer, msg []byte) {
	if s.firstEntry {
		s.firstEntry = false
	} else {
		w.WriteString(" - ")
	}

	w.Write(msg)
}

var defaultAndShortStackPrinter = xerrors.NewPrinter(func() xerrors.Formatter { return &defaultAndShortStackFormatter{} })

type defaultAndShortStackFormatter struct {
	currentErr      *xerrors.WrappingError
	firstEntry      bool
	currentStackErr *xerrors.WrappingError
	firstStackEntry bool
	isStack         bool
}

func (s *defaultAndShortStackFormatter) Init(wErr *xerrors.WrappingError) {
	s.currentErr = wErr
	s.firstEntry = true
	s.currentStackErr = wErr
	s.firstStackEntry = true
	s.isStack = false
}

func (s *defaultAndShortStackFormatter) Next() error {
	if s.currentErr != nil {
		for ; s.currentErr != nil && isStackError(s.currentErr.Payload()); s.currentErr = s.currentErr.Next() {
		}

		if s.currentErr != nil {
			out := s.currentErr.Payload()
			s.currentErr = s.currentErr.Next()

			return out
		}
	}

	s.isStack = true

	for ; s.currentStackErr != nil && !isStackError(s.currentStackErr.Payload()); s.currentStackErr = s.currentStackErr.Next() {
	}

	if s.currentStackErr == nil {
		return nil
	}

	out := s.currentStackErr.Payload()
	s.currentStackErr = s.currentStackErr.Next()

	return out
}

func (s *defaultAndShortStackFormatter) CustomFormat(err error, buf *bytes.Buffer) bool {
	stackErr, ok := err.(*xerrors.StackError)
	if !ok {
		return false
	}

	frames := stackErr.Frames()

	frame, ok := frames.Next()
	shortFormatFrame(frame, buf)

	for ok {
		frame, ok = frames.Next()
		buf.WriteString(" - ")
		shortFormatFrame(frame, buf)
	}

	return true
}

func (s *defaultAndShortStackFormatter) Append(w *bytes.Buffer, msg []byte) {
	if s.firstEntry {
		s.firstEntry = false
	} else if !s.isStack {
		w.WriteString(": ")
	} else if s.firstStackEntry {
		s.firstStackEntry = false
		w.WriteString(" (")
	} else {
		w.Truncate(w.Len() - 1)
		w.WriteString(" - ")
	}

	w.Write(msg)

	if s.isStack {
		w.WriteString(")")
	}
}

var jsonPrinter = xerrors.NewPrinter(func() xerrors.Formatter { return &jsonFormatter{} })

type jsonFormatter struct {
	currentErr *xerrors.WrappingError
	firstEntry bool
}

func (j *jsonFormatter) Init(wErr *xerrors.WrappingError) {
	j.currentErr = wErr
	j.firstEntry = true
}

func (j *jsonFormatter) Next() error {
	for ; j.currentErr != nil && isStackError(j.currentErr.Payload()); j.currentErr = j.currentErr.Next() {
	}

	if j.currentErr == nil {
		return nil
	}

	out := j.currentErr.Payload()
	j.currentErr = j.currentErr.Next()

	return out
}

func (j *jsonFormatter) CustomFormat(err error, buf *bytes.Buffer) bool {
	b, err := json.Marshal(err.Error())
	if err != nil {
		buf.WriteString(`"<unmarhsable>"`)
	} else {
		buf.Write(b)
	}

	return true
}

func (j *jsonFormatter) Append(w *bytes.Buffer, msg []byte) {
	if j.firstEntry {
		w.WriteString("[")
		j.firstEntry = false
	} else {
		w.Truncate(w.Len() - 1)
		w.WriteString(",")
	}

	w.Write(msg)

	w.WriteString("]")
}

var reverseColonPrinter = xerrors.NewPrinter(func() xerrors.Formatter { return &reverseColonFormatter{} })

type reverseColonFormatter struct {
	payloads []error
	current  int
}

func (r *reverseColonFormatter) Init(wErr *xerrors.WrappingError) {
	r.payloads = make([]error, 0)
	for ; wErr != nil; wErr = wErr.Next() {
		if !isStackError(wErr.Payload()) {
			r.payloads = append(r.payloads, wErr.Payload())
		}
	}
	r.current = len(r.payloads) - 1
}

func (r *reverseColonFormatter) Next() error {
	if r.current == -1 {
		return nil
	}

	return r.payloads[r.current]
}

func (r *reverseColonFormatter) CustomFormat(error, *bytes.Buffer) bool {
	return false
}

func (r *reverseColonFormatter) Append(w *bytes.Buffer, msg []byte) {
	if r.current != len(r.payloads)-1 {
		w.WriteString(": ")
	}

	r.current--

	w.Write(msg)
}
