package xerrors_test

import (
	"fmt"

	"github.com/JavierZunzunegui/xerrors"
)

// Example_basic shows the most basic functionality in xerrors:
// using Wrap to wrap errors onto other errors,
// calling Error() on wrapped errors to access their default string form, and
// using Find to access errors in the wrapped chain (in this case via type matching)
func Example_basic() {
	err := xerrors.Wrap(foo(), xerrors.New("in_basic_example"))
	fmt.Println(err)

	bErr := xerrors.Find(err, IsBarError)
	_ = bErr.(*BarError)
	fmt.Println(bErr)

	// Output:
	// in_basic_example: foo: bar-abc: some error
	// bar-abc
}

func errFunc() error {
	return xerrors.Wrap(nil, xerrors.New("some error"))
}

func bar() error {
	if err := errFunc(); err != nil {
		return xerrors.Wrap(err, &BarError{"abc"})
	}
	panic("assume non error bar code")
}

func foo() error {
	if err := bar(); err != nil { // always error
		return xerrors.Wrap(err, xerrors.New("foo"))
	}
	panic("assume non error foo code")
}

type BarError struct {
	msg string
}

func (err *BarError) Error() string { return "bar-" + err.msg }

func IsBarError(err error) bool {
	_, ok := err.(*BarError)
	return ok
}
