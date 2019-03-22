package xerrors

import "reflect"

func find(wErr *WrappingError, f func(error) bool) *WrappingError {
	for ; wErr != nil; wErr = wErr.next {
		if f(wErr.payload) {
			return wErr
		}
	}

	return nil
}

// Find finds the first error in the wrapping chain that passes the given function evaluation, or returns nil.
// The arguments to f will be non-WrappingError errors (the payloads within err).
// If err is not a WrappingError it returns itself if it passes the check, otherwise nil.
// If err is nil, nil is returned.
// For f's doing type equality or interface advisability use FindTyped instead.
//
// [PROPOSAL NOTES]
//
// I have discussed this pattern before in the original proposal and the first iteration of this proposal, but was
// calling it First not Find. I changed the name after receiving feedback the naming was confusing.
//
// if myErr, ok := err.(MyErrorInterface); ok {...} to be migrated to use this (for interfaces, not type matches)
func Find(err error, f func(error) bool) error {
	if err == nil || f == nil {
		return nil
	}

	wErr, ok := err.(*WrappingError)
	if !ok {
		if !f(err) {
			return nil
		}
		return err
	}

	wErr = find(wErr, f)
	if wErr == nil {
		return nil
	}

	return wErr.payload
}

// FindTyped finds the first error in the wrapping chain that is the same type to target.
// The target will be compared to non-WrappingError errors (the payloads within err).
// If the target is a WrappingError or nil, FindTyped will always return nil.
// If err is not a WrappingError is returns itself if it is the same type as the target, otherwise nil.
// If err is nil, nil is returned.
// The expected use of FindTyped is:
//
// if myErr, ok := xerrors.FindTyped(err, (*MyError)(nil)).(*MyError); !ok {
//   // use myErr
// }
//
// FindType does NOT support finding based on interface implementation, use Find for that.
//
// [PROPOSAL NOTES]
//
// if myErr, ok := err.(*MyError); ok {...} to be migrated to use this (for type matches, not interfaces)
//
// FindTyped is functionally equivalent to As in the original proposal.
// I actually think an alternative where FindTyped is removed and only Find is used may be better if some other patterns
// around errors are introduced: For every public error type or interface (e.g. MyError) also create:
// - func IsMyError(err error) bool {...}
// - func FindMyError(error) MyError {...}
// See https://github.com/JavierZunzunegui/Go2_error_values_feedback for reference (Last there = Find in this proposal)
func FindTyped(err error, target error) error {
	if err == nil || target == nil {
		return nil
	}

	tTarget := reflect.TypeOf(target)

	wErr, ok := err.(*WrappingError)
	if !ok {
		if reflect.TypeOf(err) != tTarget {
			return nil
		}
		return err
	}

	wErr = find(wErr, func(e error) bool {
		return reflect.TypeOf(e) == tTarget
	})
	if wErr == nil {
		return nil
	}

	return wErr.payload
}

// Cause retrieves the causal payload error, the first error that originated this chain.
//
// [PROPOSAL NOTES]
//
// I added this because it is very popular in the current error wrapped implementations, but I think it is not needed,
// should be removed and the pattern be discouraged.
func Cause(err error) error {
	wErr, ok := err.(*WrappingError)
	if !ok {
		return err
	}

	for ; wErr.next != nil; wErr = wErr.next {
	}

	return wErr.payload
}
