package xerrors

// [PROPOSAL NOTES]
//
// I just picked an arbitrary value for now, the real thing needs more thought
const defaultDepth = 10

// WrappingError provides the error wrapping functionality, and is exclusively the only error type doing so.
// Each WrappingError holds a (non-WrappingError, non-nil) error as the payload, and points to the next WrappingError.
// The causal error (lowest in the wrapping chain) points to nil.
type WrappingError struct {
	payload error
	next    *WrappingError
}

func isWrappingError(err error) bool {
	_, ok := err.(*WrappingError)
	return ok
}

// Error serializes the target according to the default colon serializer, omitting frames and joining with ": ".
// For example, for "wrapper-2" -> "wrapper-1" -> StackError -> "cause" the result would be:
// "wrapper-2: wrapper-1: cause"
func (wErr *WrappingError) Error() string {
	return defaultPrinter.String(wErr)
}

// Payload is a getter to payload error.
// It is non-nil and never a WrappingError.
func (wErr *WrappingError) Payload() error {
	return wErr.payload
}

// Next is a getter for the next WrappingError in the chain.
// It returns nil for the last error in the chain.
func (wErr *WrappingError) Next() *WrappingError {
	return wErr.next
}

var defaultStackOpts = StackOpts{
	Skip:  0 + 1,
	Depth: defaultDepth,
}

// StackOpts defines how stacks are recorded
type StackOpts struct {
	Skip  uint8
	Depth uint8
}

// Wrap produces a WrappingError out of two errors and is the standard way users should produce these.
//
// The intended use is:
//   if err != nil {
//     return Wrap(err, SomeNewError)
//   }
//   ...
//
// For new errors, simply call return Wrap(nil, SomeCausalError)
//
// If the input errors aren't already wrapped, it will also add a default stack to the output error via StackError.
// If err or payload are non-null, the output is a WrappingError. Double nil input returns nil but is discouraged.
// The order of wrapping is payload wraps err. Payload is discouraged from being a WrappingError itself.
//
// If added, the stack starts from Wrap, Wrap not included.
func Wrap(err, payload error) error {
	if payload == nil {
		if err == nil {
			// avoid doing this
			return nil
		}

		if _, ok := err.(*WrappingError); ok {
			return err
		}

		return frameWrap(&WrappingError{payload: err}, defaultStackOpts)
	}

	if err == nil {
		if _, ok := payload.(*WrappingError); ok {
			// avoid doing this, payload should not be a WrappingError
			return payload
		}

		return frameWrap(&WrappingError{payload: payload}, defaultStackOpts)
	}

	out := merge(err, payload)

	if isWrappingError(err) || isWrappingError(payload) {
		// if already wrapped not attempting to add a stack
		return out
	}

	return frameWrap(out, defaultStackOpts)
}

// WrapWithOpts is similar to Wrap except with regards to adding stacks.
// It adds a stack with the given opts regardless of the values of err and payload, including if they are WrappingError,
// except for the double nil input which returns nil (and is discouraged).
// If opts.Depth is 0 the StackError is not added, but the returned error is still a WrappingError.
//
// Use WrapWithOpts:
// - when not wanting frames at all (wrap causal error with Depth=0)
// - when an error changes goroutine and capturing an additional stack is desired
func WrapWithOpts(err error, payload error, opts StackOpts) error {
	opts.Skip++

	if payload == nil {
		if err == nil {
			// avoid doing this
			return nil
		}

		wErr, ok := err.(*WrappingError)
		if !ok {
			wErr = &WrappingError{payload: err}
		}

		return frameWrap(wErr, opts)
	}

	if err == nil {
		wErr, ok := payload.(*WrappingError)
		if !ok {
			wErr = &WrappingError{payload: payload}
		}

		return frameWrap(wErr, opts)
	}

	out := merge(err, payload)

	return frameWrap(out, opts)
}

func merge(err, payload error) *WrappingError {
	out := &WrappingError{}

	// used to build up the out error,
	current := out

	if pErr, ok := payload.(*WrappingError); ok {
		// avoid doing this, payload should not be WrappingError as it causes many allocations
		current.payload = pErr.payload

		for pErr = pErr.next; pErr != nil; current, pErr = current.next, pErr.next {
			current.next = &WrappingError{
				payload: pErr.payload,
			}
		}
	} else {
		current.payload = payload
	}

	if eErr, ok := err.(*WrappingError); ok {
		current.next = eErr
	} else {
		current.next = &WrappingError{
			payload: err,
		}
	}

	return out
}

func frameWrap(wErr *WrappingError, opts StackOpts) *WrappingError {
	if opts.Depth == 0 {
		return wErr
	}

	opts.Skip++

	return &WrappingError{
		payload: newStackError(opts),
		next:    wErr,
	}
}
