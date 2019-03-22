package xerrors_test

import (
	"reflect"
	"testing"

	"github.com/JavierZunzunegui/xerrors"
)

const (
	expectUnrelated uint8 = iota
	expectContained
	expectSimilar
	expectEqual
)

func TestContains(t *testing.T) {
	for _, scenario := range compareScenarios() {
		scenario := scenario

		t.Run(scenario.name, func(t *testing.T) {
			if scenario.expectations >= expectContained {
				if !xerrors.Contains(scenario.err1, scenario.err2) {
					t.Fatal("expect the error to be contained")
				}
			} else {
				if xerrors.Contains(scenario.err1, scenario.err2) {
					t.Fatal("expected the error to not be contained")
				}
			}

			if scenario.expectations >= expectSimilar {
				if !xerrors.Contains(scenario.err2, scenario.err1) {
					t.Fatal("expected similar errors to be contained also in reverse order")
				}
			}

			if !xerrors.Contains(scenario.err1, scenario.err1) || !xerrors.Contains(scenario.err2, scenario.err2) {
				t.Fatal("expected every error to contain itself")
			}
		})
	}
}

func TestSimilar(t *testing.T) {
	for _, scenario := range compareScenarios() {
		scenario := scenario

		t.Run(scenario.name, func(t *testing.T) {
			if scenario.expectations >= expectSimilar {
				if !xerrors.Similar(scenario.err1, scenario.err2) {
					t.Fatal("expected the errors to be similar")
				}
				if !xerrors.Similar(scenario.err2, scenario.err1) {
					t.Fatal("expected the errors to be similar also in reverse order")
				}
				if !xerrors.Contains(scenario.err2, scenario.err1) {
					t.Fatal("expected similar errors to be contained also in reverse order")
				}
			} else {
				if xerrors.Similar(scenario.err1, scenario.err2) {
					t.Fatal("expected the errors to not be similar")
				}
				if xerrors.Similar(scenario.err2, scenario.err1) {
					t.Fatal("expected the errors to not be similar also in reverse order")
				}
			}

			if !xerrors.Similar(scenario.err1, scenario.err1) || !xerrors.Similar(scenario.err2, scenario.err2) {
				t.Fatal("expected every error to be similar to itself")
			}
		})
	}
}

func TestDeepEqual(t *testing.T) {
	for _, scenario := range compareScenarios() {
		scenario := scenario

		t.Run(scenario.name, func(t *testing.T) {
			if scenario.expectations >= expectEqual {
				if !reflect.DeepEqual(scenario.err1, scenario.err2) {
					t.Fatal("expected the errors to be equal")
				}
				if !reflect.DeepEqual(scenario.err2, scenario.err1) {
					t.Fatal("expected the errors to be equal also in reverse order")
				}
			} else {
				if reflect.DeepEqual(scenario.err1, scenario.err2) {
					t.Fatal("expected the errors to not be equal")
				}
				if reflect.DeepEqual(scenario.err2, scenario.err1) {
					t.Fatal("expected the errors to not be equal also in reverse order")
				}
			}

			if !reflect.DeepEqual(scenario.err1, scenario.err1) || !reflect.DeepEqual(scenario.err2, scenario.err2) {
				t.Fatal("expected every error to be equal to itself")
			}
		})
	}
}

type barError struct{}

func (barError) Error() string { return "bar" }

func compareScenarios() []struct {
	name         string
	err1         error
	err2         error
	expectations uint8
} {
	return []struct {
		name         string
		err1         error
		err2         error
		expectations uint8
	}{
		{
			name:         "equalSentinel",
			err1:         xerrors.New("foo"),
			err2:         xerrors.New("foo"),
			expectations: expectEqual,
		},
		{
			name:         "differentSentinel",
			err1:         xerrors.New("foo"),
			err2:         xerrors.New("bar"),
			expectations: expectUnrelated,
		},
		{
			name:         "sentinelSameMsgDifferentType",
			err1:         barError{},
			err2:         xerrors.New("bar"),
			expectations: expectUnrelated,
		},
		{
			name:         "containedSentinelInPayload",
			err1:         xerrors.Wrap(xerrors.New("bar"), xerrors.New("foo")),
			err2:         xerrors.New("foo"),
			expectations: expectContained,
		},
		{
			name:         "containedSentinelInNext",
			err1:         xerrors.Wrap(xerrors.New("bar"), xerrors.New("foo")),
			err2:         xerrors.New("bar"),
			expectations: expectContained,
		},
		{
			name:         "unrelatedSentinel",
			err1:         xerrors.Wrap(xerrors.New("bar"), xerrors.New("foo")),
			err2:         xerrors.New("foobar"),
			expectations: expectUnrelated,
		},
		{
			name:         "equalWrapped",
			err1:         xerrors.Wrap(xerrors.New("bar"), xerrors.New("foo")),
			err2:         xerrors.Wrap(xerrors.New("bar"), xerrors.New("foo")),
			expectations: expectSimilar,
		},
		{
			name:         "equalWrappedNoStack",
			err1:         xerrors.WrapWithOpts(xerrors.New("bar"), xerrors.New("foo"), xerrors.StackOpts{}),
			err2:         xerrors.WrapWithOpts(xerrors.New("bar"), xerrors.New("foo"), xerrors.StackOpts{}),
			expectations: expectEqual,
		},
		{
			name: "containedWrapped",
			err1: xerrors.Wrap(
				xerrors.Wrap(xerrors.New("bar"), xerrors.New("foobar")),
				xerrors.New("foo"),
			),
			err2:         xerrors.Wrap(xerrors.New("bar"), xerrors.New("foo")),
			expectations: expectContained,
		},
		{
			name:         "unrelatedWrapped",
			err1:         xerrors.Wrap(xerrors.New("bar"), xerrors.New("foo")),
			err2:         xerrors.Wrap(xerrors.New("bar"), xerrors.New("foobar")),
			expectations: expectUnrelated,
		},
		{
			name:         "wrappedOutOfOrder",
			err1:         xerrors.Wrap(xerrors.New("bar"), xerrors.New("foo")),
			err2:         xerrors.Wrap(xerrors.New("foo"), xerrors.New("bar")),
			expectations: expectUnrelated,
		},
	}
}
