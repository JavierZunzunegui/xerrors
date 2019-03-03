package xerrors

import (
	"bytes"
	"sync"
)

var (
	defaultBufferPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

// NewPrinter initialises an error printer.
func NewPrinter(fFactory func() Serializer) *Printer {
	return &Printer{
		pool: sync.Pool{
			New: func() interface{} {
				return &printerAlloc{
					buf: bytes.Buffer{},
					s:   fFactory(),
				}
			},
		},
	}
}

// an auxiliary structure used for efficient use of sync.Pool within Printers
type printerAlloc struct {
	buf bytes.Buffer
	s   Serializer
}

// Printer is an error printer.
// It is thin wrapper around a Serializer factory, this fully defines the output of the printer.
type Printer struct {
	pool sync.Pool
}

// String provides the error's string format given the Serializer defining this Printer.
// Use this as a primary means to convert errors to strings with a non-default Serializer.
// For default, use err.Error() directly.
// It is safe to be called concurrently.
func (p *Printer) String(err error) string {
	buf := defaultBufferPool.Get().(*bytes.Buffer)

	p.Write(buf, err)

	out := buf.String()

	buf.Reset()
	defaultBufferPool.Put(buf)

	return out
}

// Write is the buffer equivalent of Printer.String.
// It is safe to be called concurrently (though not on the same buffer).
func (p *Printer) Write(w *bytes.Buffer, err error) {
	alloc := p.pool.Get().(*printerAlloc)

	wErr, ok := err.(*WrappingError)
	if !ok {
		wErr = &WrappingError{
			payload: err,
		}
	}

	alloc.s.Init(wErr)

	p.write(w, alloc.s, &alloc.buf)

	p.pool.Put(alloc)
}

func (*Printer) write(w *bytes.Buffer, s Serializer, auxiliary *bytes.Buffer) {
	var ok bool
	var bufErr BufferError

	for err := s.Next(); err != nil; err = s.Next() {
		if ok = s.CustomFormat(err, auxiliary); !ok {
			if bufErr, ok = err.(BufferError); ok {
				bufErr.ErrorToBuffer(auxiliary)
			} else {
				auxiliary.WriteString(err.Error())
			}
		}

		s.Append(w, auxiliary.Bytes())
		auxiliary.Reset()
	}
}

// BufferError is an optional interface that errors may implement for efficiency purposes.
// If an error's Error() method results in a string allocation for the return statement, it would benefit from this.
// Any error implementing BufferError is advised to implement Error() using BufferErrorToString as:
//
// func (err xError) Error() string {return xerrors.BufferErrorToString(err)}
//
// This guarantees that Error() a ErrorToBuffer(.) have matching implementations, and does so efficiently.
type BufferError interface {
	// ErrorToBuffer is a buffering equivalent to Error().
	// The same string as is returned by Error() is to be written to the buffer.
	ErrorToBuffer(*bytes.Buffer)
}

// BufferErrorToString is an auxiliary method to facilitate the Error() method in BufferError implementations.
func BufferErrorToString(buffErr BufferError) string {
	buf := defaultBufferPool.Get().(*bytes.Buffer)

	buffErr.ErrorToBuffer(buf)

	out := buf.String()

	buf.Reset()
	defaultBufferPool.Put(buf)

	return out
}
