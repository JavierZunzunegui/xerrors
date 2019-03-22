package xerrors

// New produces a unwrapped string error without any frame information.
// Use it to produce untyped sentinel.
//
// [PROPOSAL NOTES]
//
// has not been changed
func New(msg string) error {
	return &stringError{msg}
}

type stringError struct {
	msg string
}

func (err *stringError) Error() string {
	return err.msg
}
