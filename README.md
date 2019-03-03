**This repo is a counter-proposal to the [Go2 error values](https://github.com/golang/go/issues/29934) proposal only.**
That proposal is referred throughout as the 'original proposal'.

This code is not otherwise meant to be imported or used.

Please review the original proposal, this assumes knowledge of it and of the discussions around it, from which this takes many ideas.

The debate is in https://github.com/golang/go/issues/30350.

# Packages

- `xerrors/`: 
[![GoDoc](https://godoc.org/github.com/JavierZunzunegui/Go2_error_values_counter_proposal/xerrors?status.svg)](https://godoc.org/github.com/JavierZunzunegui/Go2_error_values_counter_proposal/xerrors)
the proposed code changes
- `xerrors/xserialiserexamples/`: 
[![GoDoc](https://godoc.org/github.com/JavierZunzunegui/Go2_error_values_counter_proposal/xerrors/xserialiserexamples?status.svg)](https://godoc.org/github.com/JavierZunzunegui/Go2_error_values_counter_proposal/xerrors/xserialiserexamples)
contains some examples used to demonstrate how these changes may be used, but is not meant to be merged.

# Overview

This introduces wrapped errors in a similar manner to the original proposal but with significant differences:
- separates the responsibility of carrying error data (the error interface) and that of turning errors into a string (a new serializer interface).
- discourages the use of string-like errors, favours using custom error types.
- easy use of stack frames alongside custom errors.
- efficient error serialisation.
- easy to navigate the wrapped error chain including finding wrapped errors, equality comparison and cause error.
- abandons any intention to automatically have existing errors upgraded - users will need to update code to start wrapping errors.

In more detail:

### Wrapping

Identical to the original proposal, an new interface `Wrapper` extending `error` with an `Unwrap() error` method a helper `func Unwrap(error) error`.

### Error() string

The meaning of error's `Error() string` is conceptually different to that of the original proposal.
In the original, it returns both the target error's message as well as that of all the wrapped errors.
In this one, `Error() string` only returns the message for the target error and not that of the wrapped ones.
To print a full error message a `Serializer` is required, see below.

### Serializer

A new interface, it is responsible for the key logic used in serializing errors, including wrapped ones.
In some regards it is similar to the `Formatter` and `Printer` interfaces of the original proposal, 
but unlike `Formatter` it is not implemented by the error.
It has 4 methods:
- `Keep`: defines what errors this serializers will print and which ignore
- `CustomFormat`: allows a way to serialize individual errors in a way other than the default `Error() string`
- `Append`: combines the messages of the different wrapped errors
- `Reset`: implementation detail to reduce memory allocations

### Printer

A struct that wraps a `Serializer`.
It mainly exists to remove some otherwise repetitive logic from `Serializer` and optimise errors serialization, particularly with regards to heap allocation via `sync.Pool`.

### String, DetailString, Bytes, DetailedBytes functions

Methods using the default serializer (the popular "{err1}: {err2}" format) for use in %s/%v formatting. 
The Detailed forms additionally print frame information if available.

Note this produce very efficient results, particularly regarding memory allocation. [See benchmark results](https://github.com/JavierZunzunegui/Go2_error_values_counter_proposal/blob/master/xerrors/benchmark.md).

### Wrapping, NewWrapping and FrameError

Embedding `Wrapping` in errors provide a very easy way for custom errors to implement `Wrapper`. 
`NewWrapping` transparently provides frames support to custom errors which use embedded `Wrapping`.
Frames support is almost identical (and largely copied) to that in the original proposal except it is not a property of the error ('detail') but a separate error, `FrameError`.

### Last

`Last(error, func(error) bool) error` is used to navigate the error wrapping chain and identify errors of interest.
It serves the same purposes as `As` and `Is` in the original proposal.
The comparison between these has already been discussed in detail in the original's discussion.

### Similar and Contains functions

Replacement for `reflect.DeepEqual` comparison of errors.
Both ignore wrapped `FrameError`.


# Migration & Tooling

### Preliminaries: Error() to String(error); reflect.Equal to Similar

While the introduction of this new error package does not break anything in itself, it's use by imported libraries will not interact well with some code written before these changes.
Two things in particular need to be addressed: the meaning of `error`'s `Error` method, and comparing errors.

`Error() string` currently represents the whole error message, as it does in the original proposal.
In this one it does not, it requires an external `Serializer` to do so, or the new standard `String` method.
All printf-like methods in the standard library must be updated to use `String(err)` instead of `err.Error()` for error's %s representation, 
but even then if user's call `Error()` explicitly this will not have the expected behaviour once the libraries they depend on stand returning wrapped errors.
Therefore **before** the main changes may be landed, a mocked `String` method should be added to the errors package and vet encourage users to use it instead of `Error()`.
In other words, `fmt.Println(err.Error())` will eventually become problematic and should be changed to `fmt.Println(errors.String(err))` (or simply `fmt.Println(err)`). 
If not obvious, the initial implementation of `String` will be `func String(err error) string {return err.Error()}`.

Error comparison has been brought up recently in the main proposal's discussion. 
Because `reflect.DeepEqual` will not work well with wrapped errors and frames, these need to be migrated to use the new `errors.Similar` or `errors.Contains`. 
Again, this should be done before the main migration, initially with a mock implementation being just a wrapper of `reflect.DeepEqual`.
The vet tool should also be updated to push this change.

### Main migration

Once the preliminary changes are considered broadly implemented, the actual proposed errors package can be added to the standard library.
Libraries which did not follow the preliminary steps could find themselves printing incomplete error messages or getting false negatives when comparing errors for equality. 

# Other Remarks

### fmt.Errorf and Wrapf

A stated objective of the original proposal is to automatically make `fmt.Errorf` 'just work' in the wrapped form, and integrate it with frame just as easily.
This proposal takes the opposite approach - `fmt.Errorf` remains unchanged, does not use the wrapping pattern and should be considered obsolete (and it's use discouraged by vet).

Similarly, I have not added a default `Wrapf` method because I think it should be seen as an anti-pattern. 
Instead, typed errors are encouraged so that `Serializers` may filter them out or customize their output.
Either way, `Wrapf` is really just `Wrap(err, fmt.Sprintf(...))` so it is hardly a blocker.

### Code generation and Last

The `Last` pattern has been discussed in detail in the original proposal. 
One aspect brought up is that to get a typed error one has write some custom helper function `func(error) bool` and type-assert the output value.
While this can be done through `go generate`, I consider it unnecessary and not part of this proposal, 
but nonetheless if you want to review that you can read [this feedback to the original proposal](https://github.com/JavierZunzunegui/Go2_error_values_feedback).