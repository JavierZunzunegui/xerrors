package xerrors_test

import (
	"fmt"
	"sync"

	"github.com/JavierZunzunegui/xerrors"
)

// Example_wrapWithOpts shows how WrapWithOpts can be used to collect stacks of errors sent between goroutines.
// By identifying when an error has changed goroutine we can reflect the errors flow through multiple stacks.
func Example_wrapWithOpts() {
	c := make(chan error, 1)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		// we know error is generated in another goroutine
		err := <-c

		// not forcing the wrap! Will only show one stack, missing
		fmt.Println("standard Wrap, prints only one stack:")
		err1Stack := xerrors.Wrap(err, xerrors.New("error from second stack"))
		fmt.Println(shortStackPrinter.String(err1Stack))
		fmt.Println("")

		// using WrapWithOpts, forcing a re-Wrap with a stack even if it already has one
		err2Stacks := xerrors.WrapWithOpts(err, xerrors.New("error from second stack"), xerrors.StackOpts{Depth: 10})

		fmt.Println("used WrapWithOpts, prints both stacks:")
		fmt.Println(shortStackPrinter.String(err2Stacks))
		wg.Done()
	}()

	c <- foo()

	wg.Wait()

	// Output:
	// standard Wrap, prints only one stack:
	// (xerrors_test.errFunc - xerrors_test.bar - xerrors_test.foo - xerrors_test.Example_wrapWithOpts - testing.runExample - testing.runExamples - testing.(*M).Run - main.main - runtime.main - runtime.goexit)
	//
	// used WrapWithOpts, prints both stacks:
	// (xerrors_test.Example_wrapWithOpts.func1 - runtime.goexit) (xerrors_test.errFunc - xerrors_test.bar - xerrors_test.foo - xerrors_test.Example_wrapWithOpts - testing.runExample - testing.runExamples - testing.(*M).Run - main.main - runtime.main - runtime.goexit)
}
