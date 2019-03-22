package xerrors

import (
	"bytes"
	"sync"
)

// NewPrinter initialises an error printer, defined by the Formatter factory provided.
//
// [PROPOSAL NOTES]
//
// See Formatter
func NewPrinter(fFactory func() Formatter) *Printer {
	return &Printer{
		pool: sync.Pool{
			New: func() interface{} {
				return &printerAlloc{
					f: fFactory(),
				}
			},
		},
	}
}

// an auxiliary structure used for efficient use of sync.Pool within Printers
type printerAlloc struct {
	w         bytes.Buffer // used by String only
	auxiliary bytes.Buffer // required for byte form of single error (payload)
	f         Formatter    // defines a printer
}

// Printer is an error printer.
// It is thin wrapper around a Formatter factory, this fully defines the output of the printer.
type Printer struct {
	pool sync.Pool
}

// String provides the error's string format given the Serializer defining this Printer.
// Use this as a primary means to convert errors to strings with a non-default Serializer.
// For default, use err.Error() directly.
// It is safe to be called concurrently.
func (p *Printer) String(err error) string {
	alloc := p.pool.Get().(*printerAlloc)

	p.write(alloc.f, &alloc.w, &alloc.auxiliary, err)

	out := alloc.w.String()

	alloc.w.Reset()
	p.pool.Put(alloc)

	return out
}

// Write is the buffer equivalent of Printer.String.
// It is safe to be called concurrently (though not on the same buffer).
func (p *Printer) Write(w *bytes.Buffer, err error) {
	alloc := p.pool.Get().(*printerAlloc)

	p.write(alloc.f, w, &alloc.auxiliary, err)

	p.pool.Put(alloc)
}

func (p *Printer) write(f Formatter, w, auxiliary *bytes.Buffer, err error) {
	wErr, ok := err.(*WrappingError)
	if !ok {
		wErr = &WrappingError{
			payload: err,
		}
	}

	f.Init(wErr)

	var bufErr BufferError

	for err := f.Next(); err != nil; err = f.Next() {
		if ok = f.CustomFormat(err, auxiliary); !ok {
			if bufErr, ok = err.(BufferError); ok {
				bufErr.ErrorToBuffer(auxiliary)
			} else {
				auxiliary.WriteString(err.Error())
			}
		}

		f.Append(w, auxiliary.Bytes())
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

var (
	bufferErrorToStringPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

// BufferErrorToString is an auxiliary method to facilitate the Error() method in BufferError implementations.
func BufferErrorToString(buffErr BufferError) string {
	buf := bufferErrorToStringPool.Get().(*bytes.Buffer)

	buffErr.ErrorToBuffer(buf)

	out := buf.String()

	buf.Reset()
	bufferErrorToStringPool.Put(buf)

	return out
}
