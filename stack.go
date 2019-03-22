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
// StackError holds full stacks (collections of frames), and not single frames like in the original proposal.
// This means connecting the message and the exact frame that produced it is no longer possible, i.e.
// "foo (fooFunc): bar (barFunc): ..."
// is no longer possible, instead have to settle with
// "foo: bar ... (fooFunc, barFunc, ...)"
//
// On the plus side, it means frames is no longer a need to wrap every error, as long as it is wrapped close to the
// causal error all frames are captured. It also limits memory allocation and performance cost through only calling
// runtime.Callers once an producing one single StackError
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
