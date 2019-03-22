package xerrors

import (
	"reflect"
)

// for two non-WrapperError errors, equal is true if they have matching types and Error() output
func equal(err1, err2 error) bool {
	if err1 == err2 {
		// shortcut
		return true
	}

	return reflect.TypeOf(err1) == reflect.TypeOf(err2) && err1.Error() == err2.Error()
}

// for a non-WrapperError error, equalFunc returns a function that is true if its argument is of the same type and has
// the same Error() output as err
func equalFunc(err error) func(error) bool {
	t := reflect.TypeOf(err)
	msg := err.Error()

	return func(err2 error) bool {
		return reflect.DeepEqual(reflect.TypeOf(err2), t) && err2.Error() == msg
	}
}

// Similar compares to errors and validates if they are logically identical.
// This involves checking all error types and Error() outputs are identical, but ignores wrapped StackErrors.
// It is a replacement for reflect.DeepEqual(err1, err2) as the frame information will cause false negatives.
//
// [PROPOSAL NOTES]
//
// reflect.DeepEqual(err1, err2) to be migrated to use this
func Similar(err1, err2 error) bool {
	wErr1, ok1 := err1.(*WrappingError)
	wErr2, ok2 := err2.(*WrappingError)

	if !ok1 || !ok2 {
		if ok2 {
			return wErr2.next == nil && equal(err1, wErr2.payload)
		}

		if ok1 {
			return wErr1.next == nil && equal(err2, wErr1.payload)
		}

		return equal(err1, err2)
	}

	return similar(wErr1, wErr2)
}

// similar is the WrappingError-only form of Similar
func similar(wErr1, wErr2 *WrappingError) bool {
	for wErr1, wErr2 = find(wErr1, isNotStackError), find(wErr2, isNotStackError); wErr1 != nil && wErr2 != nil; wErr1, wErr2 = find(wErr1.next, isNotStackError), find(wErr2.next, isNotStackError) {
		if !equal(wErr1.payload, wErr2.payload) {
			return false
		}
	}

	if wErr1 != nil || wErr2 != nil {
		return false
	}

	return true
}

// Contains checks if err2 is logically contained within err1.
// This involves checking all wrapped error types and Error() outputs in err2 appear in err1 in identical order.
// It ignores wrapped FrameErrors altogether.
//
// [PROPOSAL NOTES]
//
// if err1 == err2 {...} comparisons to be migrated to use this
func Contains(err1, err2 error) bool {
	wErr2, ok2 := err2.(*WrappingError)
	if !ok2 {
		return Find(err1, equalFunc(err2)) != nil
	}

	wErr1, ok1 := err1.(*WrappingError)
	if !ok1 {
		if wErr2.next != nil {
			return false
		}
		return equal(err2, wErr2.payload)
	}

	return contains(wErr1, wErr2)
}

// contains is the WrappingError-only form of Contains
func contains(wErr1, wErr2 *WrappingError) bool {
	for wErr2 = find(wErr2, isNotStackError); wErr2 != nil; wErr2 = find(wErr2.next, isNotStackError) {
		wErr1 = find(wErr1, equalFunc(wErr2.payload))
		if wErr1 == nil {
			return false
		}

		wErr1 = wErr1.next
	}

	return true
}
