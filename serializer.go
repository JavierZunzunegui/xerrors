package xerrors

import (
	"bytes"
)

// Serializer defines how errors will be serialised.
// Serializers are stateful.
type Serializer interface {
	// Any information from the provided error is stored in the Serializer (normally the error itself).
	// Any remaining state in the serializer from previous uses is cleared.
	// If required, it may do any initial pre-processing of the error and store it in its state.
	Init(*WrappingError)

	// Next defines the order in which errors are serialized.
	// It returns non-WrappingError errors, expected to be contained in the original error as payloads.
	// Serialization will stop when it returns nil.
	Next() error

	// CustomFormat allows the Serializer to serialise the error in some way other that the error's Error() method.
	// If the error has a custom serialisation, it should be written to the buffer and return true.
	// If it doesn't, nothing should be written to the buffer and the it should return false.
	CustomFormat(error, *bytes.Buffer) bool

	// Append writes the provided bytes to the buffer, along with any custom prefix and/or suffix.
	// The input bytes will be the content of CustomFormat's buffer, or Error().
	Append(*bytes.Buffer, []byte)
}

var (
	colonSeparator = []byte(": ")
)

type colonSerializer struct {
	currentErr *WrappingError
	firstEntry bool
}

func (s *colonSerializer) Init(wErr *WrappingError) {
	s.currentErr = wErr
	s.firstEntry = true
}

func (s *colonSerializer) Next() error {
	wErr := s.currentErr.Find(s.keep)
	if wErr == nil {
		s.currentErr = nil
		return nil
	}
	s.currentErr = wErr.next
	return wErr.payload
}

func (s *colonSerializer) keep(err error) bool {
	return IsNotStackError(err)
}

func (s *colonSerializer) CustomFormat(error, *bytes.Buffer) bool {
	return false
}

func (s *colonSerializer) Append(w *bytes.Buffer, msg []byte) {
	if s.firstEntry {
		// move?
		s.firstEntry = false
	} else {
		w.Write(colonSeparator)
	}

	w.Write(msg)
}

// NewColonBasicSerializer provides a formatter that appends messages with ': ' and omits frames.
// It is the serializer used by the %s representation of errors.
func NewColonBasicSerializer() Serializer {
	return &colonSerializer{}
}

var (
	defaultPrinter = NewPrinter(NewColonBasicSerializer)
)
