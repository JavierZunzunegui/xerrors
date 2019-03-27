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
		frames: frames[:d],
	}
}

func isNotStackError(err error) bool {
	_, ok := err.(*StackError)
	return !ok
}

// StackError holds a stack - a collection of frames capturing the program state at the time of creating its creation.
// Do not initialise a StackError directly, use Wrap or WrapWithOpts.
//
// [PROPOSAL NOTES]
//
// In this proposal, a single StackError is normally required per wrapped error, normally towards the bottom of the
// error chain.
// This is made trivially easy by the Wrap method.
// The alternative would be to wrap errors with single frames, wrapping each new error in the chain with a new frame.
// The frame option is the direction taken in the original proposal, which this one deviates from.
//
// The main advantage of stacks over frames users don't have to wrap all errors for them to get frame information.
//
// The main disadvantage that we can't tell with wrapped errors belongs with which frame, i.e.:
// "foo (fooFunc): bar (barFunc): ..."
// is only possible with frame wrapping, with stacks we have to settle with
// "foo: bar ... (fooFunc, barFunc, ...)"
//
// Note however this difference is otherwise irrelevant in the proposal, a frame StackError could be replaced with a
// FrameError trivially easily if the stack option was not favoured
type StackError struct {
	frames []uintptr
}

// ErrorToBuffer provides the default formatting of StackErrors and makes it implement BufferError.
// The format is "{frame_format[0]} - {frame_format[1]} - ... - {frame_format[N-1]}" for a stack N frames deep.
// Each frame format is "package.function_name:file_path:line_number"
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

// Error is the string format of StackError.ErrorToBuffer
func (err *StackError) Error() string {
	return BufferErrorToString(err)
}

// Frames exports access to all data held by the StackError.
// It is intended to be used by custom Formatters that wish to convert StackErrors to strings in a specific manner.
//
// [PROPOSAL NOTES]
//
// See Formatter's CustomFormat notes
func (err *StackError) Frames() *runtime.Frames {
	return runtime.CallersFrames(err.frames)
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
