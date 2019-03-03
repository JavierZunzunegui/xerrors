package xerrors_test

import (
	"regexp"
	"testing"

	"github.com/JavierZunzunegui/xerrors"
)

// this test is very fragile, it's the nature of frames lines. Beware of formatting the below

func errFunc1() error {
	// line = 14
	return errFunc2()
}

func errFunc2() error {
	// line = 19
	return xerrors.Wrap(nil, xerrors.New("foo"))
}

func TestStackError_Error(t *testing.T) {
	// line = 24
	err := errFunc1()

	msg := xerrors.Find(err, xerrors.IsStackError).Error()

	const (
		thisPkg  = "github[.]com[/]JavierZunzunegui[/]xerrors_test"
		thisFile = ".*[/]github[.]com[/]JavierZunzunegui[/]xerrors[/]stack_test[.]go"
	)

	const expectedRegex = "" +
		thisPkg + "[.]errFunc2" + ":" + thisFile + ":" + "19" +
		" - " +
		thisPkg + "[.]errFunc1" + ":" + thisFile + ":" + "14" +
		" - " +
		thisPkg + "[.]TestStackError_Error" + ":" + thisFile + ":" + "24" +
		" - " +
		".*"

	if !regexp.MustCompile(expectedRegex).MatchString(msg) {
		t.Fatal("mismatched StackError.Error() output and expected regex")
	}
}
