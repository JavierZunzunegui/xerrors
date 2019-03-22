// Package xerrors is a proposal for adding error wrapping to go.
// It is a counter-proposal to Go2 error values: https://github.com/golang/go/issues/29934.
//
// The debate is in TODO (go issue).
//
// Compared to the original proposal, this one:
//   - has no requirement on error types (no `Unwrap() error` or equivalent)
//   - allows for custom error conversion to string in a more powerful manner
//   - has no automatic migration to wrapping form (code is not immediately using wrapping)
//   - uses stacks and not frames, and integrates them seamlessly with formatter
//   - compile-time safe implementation with fewer gotchas
//   - can compare errors without requiring modification of reflect.DeepEqual
//
// If new to this proposal, I suggest you look at the examples first as the are intended to showcase the new functionality.
package xerrors
