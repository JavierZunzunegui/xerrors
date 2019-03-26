package xerrors_test

import (
	"fmt"

	"github.com/JavierZunzunegui/xerrors"
)

func Example_basic() {
	err := foo()
	fmt.Println(err)

	bErr := xerrors.Find(err, IsBarError)
	_ = bErr.(*BarError)
	fmt.Println(bErr)

	// Output:
	// foo: bar-abc: some error
	// bar-abc
}

func errFunc() error {
	return xerrors.New("some error")
}

func bar() error {
	if err := errFunc(); err != nil {
		return xerrors.Wrap(err, &BarError{"abc"})
	}
	panic("assume non-error foo code")
}

func foo() error {
	if err := bar(); err != nil { // always error
		return xerrors.Wrap(err, xerrors.New("foo"))
	}
	panic("assume non-error foo code")
}

type BarError struct {
	msg string
}

func (err *BarError) Error() string { return "bar-" + err.msg }

func IsBarError(err error) bool {
	_, ok := err.(*BarError)
	return ok
}
