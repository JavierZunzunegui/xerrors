package xerrors

import (
	"reflect"
	"testing"
)

func TestWrap(t *testing.T) {
	scenarios := []struct {
		name             string
		err              error
		payload          error
		expectedPayloads []error
	}{
		{
			name:             "nilErrAndPayload", // anti-pattern
			err:              nil,
			payload:          nil,
			expectedPayloads: nil,
		},
		{
			name:             "unwrappedErrAndNilPayload",
			err:              New("foo"),
			payload:          nil,
			expectedPayloads: []error{&StackError{}, New("foo")},
		},
		{
			name:             "nilErrAndUnwrappedPayload",
			err:              nil,
			payload:          New("foo"),
			expectedPayloads: []error{&StackError{}, New("foo")},
		},
		{
			name:             "unwrappedErrAndUnwrappedPayload",
			err:              New("bar"),
			payload:          New("foo"),
			expectedPayloads: []error{&StackError{}, New("foo"), New("bar")},
		},
		{
			name:             "wrappedErrAndNilPayload",
			err:              &WrappingError{payload: New("w-foo"), next: &WrappingError{payload: New("w-bar")}},
			payload:          nil,
			expectedPayloads: []error{New("w-foo"), New("w-bar")},
		},
		{
			name:             "wrappedErrAndUnwrappedPayload",
			err:              &WrappingError{payload: New("w-foo"), next: &WrappingError{payload: New("w-bar")}},
			payload:          New("foo"),
			expectedPayloads: []error{New("foo"), New("w-foo"), New("w-bar")},
		},
		{
			name:             "nilErrAndWrappedPayload", // anti-pattern
			err:              nil,
			payload:          &WrappingError{payload: New("w-foo"), next: &WrappingError{payload: New("w-bar")}},
			expectedPayloads: []error{New("w-foo"), New("w-bar")},
		},
		{
			name:             "unwrappedErrAndWrappedPayload", // anti-pattern
			err:              New("foo"),
			payload:          &WrappingError{payload: New("w-foo"), next: &WrappingError{payload: New("w-bar")}},
			expectedPayloads: []error{New("w-foo"), New("w-bar"), New("foo")},
		},
		{
			name:             "unwrappedErrAndWrappedPayload", // anti-pattern
			err:              &WrappingError{payload: New("foo-err"), next: &WrappingError{payload: New("bar-err")}},
			payload:          &WrappingError{payload: New("foo-payload"), next: &WrappingError{payload: New("bar-payload")}},
			expectedPayloads: []error{New("foo-payload"), New("bar-payload"), New("foo-err"), New("bar-err")},
		},
	}

	for _, scenario := range scenarios {
		scenario := scenario

		t.Run(scenario.name, func(t *testing.T) {
			out := Wrap(scenario.err, scenario.payload)

			if scenario.expectedPayloads == nil || out == nil {
				if scenario.expectedPayloads == nil && out != nil {
					t.Fatal("expected Wrap output to be the nil error")
				}

				if scenario.expectedPayloads != nil && out == nil {
					t.Fatal("expected Wrap output to not be the nil error")
				}

				return
			}

			outErr, ok := out.(*WrappingError)
			if !ok {
				t.Fatal("expected Wrap output to be type WrappingError")
			}

			payloads := make([]error, 0, len(scenario.expectedPayloads))
			for ; outErr != nil; outErr = outErr.Next() {
				payloads = append(payloads, outErr.Payload())
			}

			if len(scenario.expectedPayloads) != len(payloads) {
				t.Fatalf("expected Wrap error chain depth to be %d,m got %d", len(scenario.expectedPayloads), len(payloads))
			}

			for i := range scenario.expectedPayloads {
				expectedPayload := scenario.expectedPayloads[i]
				payload := payloads[i]

				_, isExpectedStack := expectedPayload.(*StackError)
				_, isStack := payload.(*StackError)

				if isExpectedStack || isStack {
					if isExpectedStack && !isStack {
						t.Fatalf("expected StackError, got type %T error, at depth %d", payload, i)
					}

					if isExpectedStack && !isStack {
						t.Fatalf("expected type %T error, got StackError, at depth %d", expectedPayload, i)
					}

					// not comparing stack errors in more detail
					continue
				}

				if reflect.TypeOf(expectedPayload) != reflect.TypeOf(payload) {
					t.Fatalf("expected type %T error, got ype %T error, at depth %d", expectedPayload, payload, i)
				}

				if expectedPayload.Error() != payload.Error() {
					t.Fatalf("unexpected error message, expected %q got %q", expectedPayload.Error(), payload.Error())
				}
			}
		})
	}
}
