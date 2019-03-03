package xerrors

type WrappingError struct {
	payload error
	next    *WrappingError
}

func (wErr *WrappingError) Error() string {
	return defaultPrinter.String(wErr)
}

func (wErr *WrappingError) Payload() error {
	return wErr.payload
}

func (wErr *WrappingError) Next() *WrappingError {
	return wErr.next
}

func (wErr *WrappingError) Find(f func(error) bool) *WrappingError {
	for ; wErr != nil; wErr = wErr.next {
		if f(wErr.payload) {
			return wErr
		}
	}

	return nil
}

func Find(err error, f func(error) bool) error {
	wErr, ok := err.(*WrappingError)
	if !ok {
		if err == nil || !f(err) {
			return nil
		}
		return err
	}

	wErr = wErr.Find(f)
	if wErr == nil {
		return nil
	}

	return wErr.payload
}

func Cause(err error) error {
	wErr, ok := err.(*WrappingError)
	if !ok {
		return err
	}

	for ; wErr.next != nil; wErr = wErr.next {
	}

	return wErr.payload
}

var defaultFrameOpts = StackOpts{
	Skip:        0 + 1,
	Depth:       10,
	IsSameStack: false,
}

type StackOpts struct {
	Skip        uint8
	Depth       uint8
	IsSameStack bool
}

// payload must not be empty, must not be type *WrappingError
// err may or may not be empty, may or may not be *WrappingError
func Wrap(err, payload error) error {
	if err == nil {
		return frameWrap(&WrappingError{payload: payload}, defaultFrameOpts)
	}

	if next, ok := err.(*WrappingError); ok {
		return &WrappingError{
			payload: payload,
			next:    next,
		}
	}

	wErr := &WrappingError{
		payload: payload,
		next: &WrappingError{
			payload: err,
		},
	}

	return frameWrap(wErr, defaultFrameOpts)
}

// payload must not be empty
// err may or may not be empty
func WrapWithOpts(err error, payload error, opts StackOpts) error {
	opts.Skip++

	if err == nil {
		return frameWrap(&WrappingError{payload: payload}, opts)
	}

	next, ok := err.(*WrappingError)
	if !ok {
		next = &WrappingError{
			payload: err,
		}
	}

	return frameWrap(
		&WrappingError{
			payload: payload,
			next:    next,
		},
		opts,
	)
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

/*

type wrapOptions struct {
	omitFrame bool
	skip      uint8
}

// WrapOptionFunc represent optional arguments to NewWrapping or Wrap methods.
type WrapOptionFunc = func(wrapOptions) wrapOptions

// OmitFrame stops frames from being included in NewWrapping or Wrap methods.
func OmitFrame() WrapOptionFunc {
	return func(opts wrapOptions) wrapOptions {
		opts.omitFrame = true
		return opts
	}
}

// SkipNFrames can be used to have the frame reported in NewWrapping or Wrap be something other than the calling one.
func SkipNFrames(skip uint8) WrapOptionFunc {
	return func(opts wrapOptions) wrapOptions {
		opts.skip += skip
		return opts
	}
}

func newWrapping(err error, wrapOpts wrapOptions, opts ...WrapOptionFunc) Wrapping {
	for _, opt := range opts {
		wrapOpts = opt(wrapOpts)
	}

	if wrapOpts.omitFrame {
		return Wrapping{err: err}
	}

	return Wrapping{err: newFrameError(wrapOpts.skip+1, err)}
}

*/
