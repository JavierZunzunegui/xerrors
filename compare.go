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

	return reflect.DeepEqual(reflect.TypeOf(err1), reflect.TypeOf(err2)) && err1.Error() == err2.Error()
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
	for wErr1, wErr2 = wErr1.Find(IsNotStackError), wErr2.Find(IsNotStackError); wErr1 != nil && wErr2 != nil; wErr1, wErr2 = wErr1.next.Find(IsNotStackError), wErr2.next.Find(IsNotStackError) {
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
	for wErr2 = wErr2.Find(IsNotStackError); wErr2 != nil; wErr2 = wErr2.next.Find(IsNotStackError) {
		wErr1 = wErr1.Find(equalFunc(wErr2.payload))
		if wErr1 == nil {
			return false
		}

		wErr1 = wErr1.next
	}

	return true
}
