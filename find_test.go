package xerrors_test

import (
	"reflect"
	"testing"

	"github.com/JavierZunzunegui/xerrors"
)

type fooError struct{}

func (fooError) Error() string { return "foo" }

func (fooError) Foo() {}

func TestFind(t *testing.T) {
	scenarios := []struct {
		name        string
		err         error
		f           func(error) bool
		expectedOut error
	}{
		{
			name:        "nil",
			err:         nil,
			f:           func(error) bool { panic("not to be called") },
			expectedOut: nil,
		},
		{
			name:        "nonWrappedFindAny",
			err:         xerrors.New("msg"),
			f:           func(error) bool { return true },
			expectedOut: xerrors.New("msg"),
		},
		{
			name:        "nonWrappedFindNone",
			err:         xerrors.New("msg"),
			f:           func(error) bool { return false },
			expectedOut: nil,
		},
		{
			name:        "basicWrappedFindAny",
			err:         xerrors.WrapWithOpts(nil, xerrors.New("msg"), xerrors.StackOpts{}),
			f:           func(error) bool { return true },
			expectedOut: xerrors.New("msg"),
		},
		{
			name:        "basicWrappedFindNone",
			err:         xerrors.WrapWithOpts(nil, xerrors.New("msg"), xerrors.StackOpts{}),
			f:           func(error) bool { return false },
			expectedOut: nil,
		},
		{
			name: "wrappedFindNotStack",
			err: xerrors.Wrap(
				xerrors.New("msg"),
				xerrors.New("wrapper"),
			),
			f: func(err error) bool {
				_, ok := err.(*xerrors.StackError)
				return !ok
			},
			expectedOut: xerrors.New("wrapper"),
		},
		{
			name: "wrappedFindAny",
			err: xerrors.Wrap(
				xerrors.New("msg"),
				xerrors.New("wrapper"),
			),
			f:           func(error) bool { return true },
			expectedOut: &xerrors.StackError{},
		},
		{
			name: "wrappedFindNone",
			err: xerrors.WrapWithOpts(
				xerrors.New("msg"),
				xerrors.New("wrapper"),
				xerrors.StackOpts{},
			),
			f:           func(error) bool { return false },
			expectedOut: nil,
		},
		{
			name: "wrappedFooFindAny",
			err: xerrors.WrapWithOpts(
				fooError{},
				xerrors.New("wrapper"),
				xerrors.StackOpts{},
			),
			f:           func(error) bool { return true },
			expectedOut: xerrors.New("wrapper"),
		},
		{
			name: "wrappedFooFindNone",
			err: xerrors.WrapWithOpts(
				fooError{},
				xerrors.New("wrapper"),
				xerrors.StackOpts{},
			),
			f:           func(error) bool { return false },
			expectedOut: nil,
		},
		{
			name: "wrappedFooFindSpecificTyped",
			err: xerrors.WrapWithOpts(
				fooError{},
				xerrors.New("wrapper"),
				xerrors.StackOpts{},
			),
			f:           func(err error) bool { _, ok := err.(fooError); return ok },
			expectedOut: fooError{},
		},
		{
			name: "wrappedFooFindNotSpecificTyped",
			err: xerrors.WrapWithOpts(
				fooError{},
				xerrors.New("wrapper"),
				xerrors.StackOpts{},
			),
			f:           func(err error) bool { _, ok := err.(fooError); return !ok },
			expectedOut: xerrors.New("wrapper"),
		},
		{
			name: "fooWrappingFindAny",
			err: xerrors.WrapWithOpts(
				xerrors.New("msg"),
				fooError{},
				xerrors.StackOpts{},
			),
			f:           func(error) bool { return true },
			expectedOut: fooError{},
		},
		{
			name: "fooWrappingFindNone",
			err: xerrors.WrapWithOpts(
				xerrors.WrapWithOpts(nil, xerrors.New("msg"), xerrors.StackOpts{}),
				fooError{},
				xerrors.StackOpts{},
			),
			f:           func(error) bool { return false },
			expectedOut: nil,
		},
		{
			name: "fooWrappingFindSpecificTyped",
			err: xerrors.WrapWithOpts(
				xerrors.WrapWithOpts(nil, xerrors.New("msg"), xerrors.StackOpts{}),
				fooError{},
				xerrors.StackOpts{},
			),
			f:           func(err error) bool { _, ok := err.(fooError); return ok },
			expectedOut: fooError{},
		},
		{
			name: "fooWrappingFindNotSpecificTyped",
			err: xerrors.WrapWithOpts(
				xerrors.WrapWithOpts(nil, xerrors.New("msg"), xerrors.StackOpts{}),
				fooError{},
				xerrors.StackOpts{},
			),
			f:           func(err error) bool { _, ok := err.(fooError); return !ok },
			expectedOut: xerrors.New("msg"),
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			out := xerrors.Find(scenario.err, scenario.f)
			if _, ok := scenario.expectedOut.(*xerrors.StackError); ok {
				// stack error comparison is difficult, just doing type matching
				if _, ok = out.(*xerrors.StackError); !ok {
					t.Fatal("mismatched outputs, expecting to produce a StackError")
				}
			} else if !reflect.DeepEqual(out, scenario.expectedOut) {
				t.Fatalf("mismatched outputs, expected %q got %q", scenario.expectedOut, out)
			}
		})
	}
}

type valueError struct{}

func (valueError) Error() string { return "value" }

type ptrError struct{ s string }

func (err *ptrError) Error() string { return err.s }

func TestFindTyped(t *testing.T) {
	stringTarget := xerrors.New("")

	scenarios := []struct {
		name        string
		err         error
		target      error
		expectedOut error
	}{
		{
			name:        "nil argument",
			err:         nil,
			target:      stringTarget,
			expectedOut: nil,
		},
		{
			name:        "nil target",
			err:         xerrors.New("msg"),
			target:      nil,
			expectedOut: nil,
		},
		{
			name:        "foundSentinel",
			err:         xerrors.New("msg"),
			target:      stringTarget,
			expectedOut: xerrors.New("msg"),
		},
		{
			name:        "notFoundSentinel",
			err:         xerrors.New("msg"),
			target:      valueError{},
			expectedOut: nil,
		},
		{
			name:        "foundWrappedWrapper",
			err:         xerrors.Wrap(xerrors.New("msg"), xerrors.New("wrap")),
			target:      stringTarget,
			expectedOut: xerrors.New("wrap"),
		},
		{
			name:        "foundWrappedCause",
			err:         xerrors.Wrap(xerrors.New("msg"), valueError{}),
			target:      stringTarget,
			expectedOut: xerrors.New("msg"),
		},
		{
			name:        "notFoundWrapped",
			err:         xerrors.Wrap(xerrors.New("msg"), &ptrError{}),
			target:      valueError{},
			expectedOut: nil,
		},
		{
			name:        "foundTypedNilReceiver",
			err:         &ptrError{"foo"},
			target:      (*ptrError)(nil),
			expectedOut: &ptrError{"foo"},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			out := xerrors.FindTyped(scenario.err, scenario.target)
			if !reflect.DeepEqual(out, scenario.expectedOut) {
				t.Fatalf("mismatched outputs, expected %q got %q", scenario.expectedOut, out)
			}
		})
	}
}

func TestCause(t *testing.T) {
	scenarios := []struct {
		name        string
		err         error
		expectedOut error
	}{
		{
			name:        "nil",
			err:         nil,
			expectedOut: nil,
		},
		{
			name:        "nonWrapped",
			err:         xerrors.New("msg"),
			expectedOut: xerrors.New("msg"),
		},
		{
			name:        "nilWrapped",
			err:         xerrors.Wrap(nil, xerrors.New("msg")),
			expectedOut: xerrors.New("msg"),
		},
		{
			name:        "wrapped",
			err:         xerrors.Wrap(xerrors.New("msg"), xerrors.New("wrapper")),
			expectedOut: xerrors.New("msg"),
		},
		{
			name: "doubleWrapped",
			err: xerrors.Wrap(
				xerrors.Wrap(
					xerrors.New("msg"),
					xerrors.New("wrapper_1"),
				),
				xerrors.New("wrapper_2"),
			),
			expectedOut: xerrors.New("msg"),
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.name, func(t *testing.T) {
			out := xerrors.Cause(scenario.err)
			if !reflect.DeepEqual(out, scenario.expectedOut) {
				t.Fatalf("mismatched outputs, expected %q got %q", scenario.expectedOut, out)
			}
		})
	}
}
