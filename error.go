package xerrors

// New produces a unwrapped string error without any frame information.
// Use it to produce untyped sentinel.
func New(msg string) error {
	return &stringError{msg}
}

type stringError struct {
	msg string
}

func (err *stringError) Error() string {
	return err.msg
}
