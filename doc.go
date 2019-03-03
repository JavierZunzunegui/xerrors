// Package xerrors is a proposed new errors package introducing error wrapping support.
// This is meant as a counter-proposal to https://github.com/golang/go/issues/29934.
//
// The Wrapper interface expects to be implemented by all errors and introduces a common error wrapping mechanism.
//
// The Error() method of the built-in error interface now represents only the message for the target error.
// It must not print the message of it's wrapped error (recursively, errors).
// An error is only responsible for establishing the default string representation of its non-wrapped contents.
//
// The Serializer interface holds methods used to turn errors into a string, including all wrapped inner errors.
// A default Serializer implementation is provided, serializing errors in the popular  "%s: %s: %s: ..." format.
//
// The Printer uses Serializer to turn an error into a string, including all it's wrapped inner errors.
//
// The String and DetailString methods give easy access to the default Serializer.
// They are meant to be used when printing errors in "%s" and "%v" format.
// String(err) is the new representation of what was previously written as err.Error().
//
// Method New is preserved as a default string error initialisation, without error wrapping.
// It is to be used for sentinel errors only.
//
// Method Wrap is a default string error initialisation with wrapping support.
// It additionally ads some frame information.
//
// Wrapping and NewWrapping provide an easy way for custom errors to implement Wrapper and have frame information.
//
// Method Last is used to navigate the wrapped error chain and fetch any error of interest within it.
//
// Method Similar is an alternative to reflect.DeepEqual for error comparison which omits frames from the comparison.
//
// Method Contains validates if one error is contained within another, including all wrapped errors but omitting frames.
package xerrors
