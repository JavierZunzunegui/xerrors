package xerrors_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/JavierZunzunegui/xerrors"
)

type encodeScenario struct {
	name           string
	err            *xerrors.WrappingError
	expectedOutput string
}

func TestWrappingError_Error(t *testing.T) {
	scenarios := encodeScenarios()

	t.Run("individual", func(t *testing.T) {
		for _, scenario := range scenarios {
			scenario := scenario

			t.Run(scenario.name, func(t *testing.T) {
				if out, expectedOut := scenario.err.Error(), scenario.expectedOutput; out != expectedOut {
					t.Fatalf("expected %q got %q", expectedOut, out)
				}
			})
		}
	})

	t.Run("stress", func(t *testing.T) {
		const stressReps = 10000

		for _, scenario := range scenarios {
			scenario := scenario

			t.Run(scenario.name, func(t *testing.T) {
				wg := sync.WaitGroup{}
				var errorCount int32

				wg.Add(stressReps)
				for i := 0; i < stressReps; i++ {
					go func() {
						if out, expectedOut := scenario.err.Error(), scenario.expectedOutput; out != expectedOut {
							atomic.AddInt32(&errorCount, 1)
						}
						wg.Done()
					}()
				}

				wg.Wait()

				if errorCount != 0 {
					t.Fatalf("expecting no async related errors, found %d/%d", errorCount, stressReps)
				}
			})
		}
	})
}

func BenchmarkWrappingError_Error(b *testing.B) {
	scenarios := encodeScenarios()

	for _, scenario := range scenarios {
		scenario := scenario

		b.Run(scenario.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// not bothering to check the output, already covered in the tests
				_ = scenario.err.Error()
			}
		})
	}
}

func encodeScenarios() []encodeScenario {
	return []encodeScenario{
		{
			name:           "nonWrapped",
			err:            xerrors.Wrap(nil, xerrors.New("msg")).(*xerrors.WrappingError),
			expectedOutput: "msg",
		},
		{
			name: "singleWrapped",
			err: xerrors.Wrap(
				xerrors.New("cause_msg"),
				xerrors.New("wrapping_msg"),
			).(*xerrors.WrappingError),
			expectedOutput: "wrapping_msg: cause_msg",
		},
		{
			name: "doubleWrapped",
			err: xerrors.Wrap(
				xerrors.Wrap(
					xerrors.New("cause_msg"),
					xerrors.New("wrapping_msg_1"),
				),
				xerrors.New("wrapping_msg_2"),
			).(*xerrors.WrappingError),
			expectedOutput: "wrapping_msg_2: wrapping_msg_1: cause_msg",
		},
		{
			name: "doubleWrapped",
			err: xerrors.Wrap(
				xerrors.Wrap(
					xerrors.New("cause_msg"),
					xerrors.New("wrapping_msg_1"),
				),
				xerrors.New("wrapping_msg_2"),
			).(*xerrors.WrappingError),
			expectedOutput: "wrapping_msg_2: wrapping_msg_1: cause_msg",
		},
	}
}
