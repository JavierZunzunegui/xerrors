package xerrors

import (
	"bytes"
	"runtime"
	"strconv"
)

func newStackError(opts StackOpts) *StackError {
	frames := make([]uintptr, int(opts.Depth))
	d := runtime.Callers(int(opts.Skip+2), frames)

	return &StackError{
		frames:      frames[:d],
		isSameStack: opts.IsSameStack,
	}
}

func IsNotStackError(err error) bool {
	_, ok := err.(*StackError)
	return !ok
}

func IsStackError(err error) bool {
	_, ok := err.(*StackError)
	return ok
}

type StackError struct {
	frames      []uintptr
	isSameStack bool
}

func (err *StackError) ErrorToBuffer(buf *bytes.Buffer) {
	frames := err.Frames()

	frame, ok := frames.Next()
	formatFrame(frame, buf)

	for ok {
		frame, ok = frames.Next()
		buf.WriteString(" - ")
		formatFrame(frame, buf)
	}
}

func (err *StackError) Error() string {
	return BufferErrorToString(err)
}

func (err *StackError) Frames() *runtime.Frames {
	return runtime.CallersFrames(err.frames)
}

func (err *StackError) IsSameStack() bool {
	return err.isSameStack
}

func formatFrame(frame runtime.Frame, buf *bytes.Buffer) {
	if frame.Function != "" {
		buf.WriteString(frame.Function)
	}
	if frame.Function != "" && frame.File != "" {
		buf.WriteString(":")
	}
	if frame.File != "" {
		buf.WriteString(frame.File)
		buf.WriteString(":")
		buf.WriteString(strconv.Itoa(frame.Line))
	}
}
