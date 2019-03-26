package xerrors

import (
	"bytes"
)

// Formatter defines how errors will be converted to a string.
// They are stateful.
//
// [PROPOSAL NOTES]
//
// The motivation for a Formatter arises from the different needs on how wrapped errors are converted to strings, and
// the different context on which they arise.
// Some examples:
//   - want all errors to be logged in a particular format (e.g. JSON) for better consumption by a log aggregator.
//   - only want to display some specific errors, such as only errors approved for user display in a HTTP response.
//   - want the stack to be logged in a particular way.
//   - ...
//
// The key thing here is this requirements are not known at the time the error is generated (and may in fact want to use
// more than one Formatter sometimes) so they must belong to an external component, here called Formatter.
type Formatter interface {
	// Any information from the provided error is stored in the Formatter (normally the error itself).
	// Any remaining state in the Formatter from previous uses is cleared.
	// If required, it may do any initial pre-processing of the error and store it in its state.
	Init(*WrappingError)

	// Next defines the order in which errors are appended to the output.
	// It returns non-WrappingError errors, expected to be contained in the original error as payloads.
	// When it returns nil no more appending is done
	Next() error

	// CustomFormat allows to convert the error to a string in some way other that the error's Error() method.
	// If the error has a custom format, it should be written to the buffer and return true.
	// If it doesn't, nothing should be written to the buffer and the it should return false.
	//
	// [PROPOSAL NOTES]
	//
	// CustomFormat is expected to return true/false by either type asserting the err to a specific error type or types
	// or checking if it implements a specific interface or interfaces.
	// This means the Formatter may have to have some explicit knowledge of the errors it handles.
	// Some highlights in this regards:
	//   - Some Formatters may want to reject altogether errors that don't match certain criteria, use Formatter.Next for these (e.g. converting errors to strings that may be visible to users).
	//   - Similarly, Formatters may want to exclude errors that do match certain criteria (e.g. error types containing sensitive information unsafe for general logging).
	//   - Some Formatters may want to keep all errors and yet apply some custom logic to errors of unrecognised type, by relying on the output of Error() (e.g. applying JSON encoding).
	//
	// To facilitate a scalable approach and easier integration of errors in 3rd party libraries with custom Formatters,
	// it would be advisable to encourage some minimal standards amongst errors such as having certain method(s):
	// ErrorData() []interface{}
	// and/or
	// StringErrorData() []string
	// and/or
	// KeyValueErrorData() [][2]interface{}
	// and/or
	// ...
	// Also all errors (even private ones) should have some public method through which custom Formatters can, if they
	// need to, identify them and access their raw data. This is particularly significant for 3rd party error types.
	CustomFormat(error, *bytes.Buffer) bool

	// Append writes the provided bytes to the buffer, along with any custom prefix and/or suffix.
	// The input bytes will be the content of CustomFormat's buffer, or Error().
	Append(*bytes.Buffer, []byte)
}

type colonFormatter struct {
	currentErr *WrappingError
	firstEntry bool
}

func (s *colonFormatter) Init(wErr *WrappingError) {
	s.currentErr = wErr
	s.firstEntry = true
}

func (s *colonFormatter) Next() error {
	wErr := find(s.currentErr, isNotStackError)
	if wErr == nil {
		s.currentErr = nil
		return nil
	}
	s.currentErr = wErr.next
	return wErr.payload
}

func (s *colonFormatter) CustomFormat(error, *bytes.Buffer) bool {
	return false
}

func (s *colonFormatter) Append(w *bytes.Buffer, msg []byte) {
	if s.firstEntry {
		s.firstEntry = false
	} else {
		w.WriteString(": ")
	}

	w.Write(msg)
}

// NewColonFormatter provides a formatter that appends messages with ': ' and omits frames.
// It is the Formatter used by the %s representation of errors.
func NewColonFormatter() Formatter {
	return &colonFormatter{}
}

var (
	defaultPrinter = NewPrinter(NewColonFormatter)
)
