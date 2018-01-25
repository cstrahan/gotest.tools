/*Package assert provides assertions and checks for comparing expected values to
actual values. When an assertion or check fails a helpful error message is
printed.

Assert and Check

Assert() and Check() both accept a Comparison, and fail the test when the
comparison fails. The one difference is that Assert() will end the test execution
immediately (using t.FailNow()) whereas Check() will fail the test (using t.Fail()),
return the value of the comparison, then proceed with the rest of the test case.

Example Usage

The example below shows assert used with some common types.


	import (
	    "testing"

	    "github.com/gotestyourself/gotestyourself/assert"
	    is "github.com/gotestyourself/gotestyourself/assert/cmp"
	)

	func TestEverything(t *testing.T) {
	    // booleans
	    assert.Assert(t, ok)
	    assert.Assert(t, !missing)

	    // primitives
	    assert.Equal(t, count, 1)
	    assert.Equal(t, msg, "the message")
	    assert.Assert(t, total != 10) // NotEqual

	    // errors
	    assert.NilError(t, closer.Close())
	    assert.Assert(t, is.Error(err, "the exact error message"))
	    assert.Assert(t, is.ErrorContains(err, "includes this"))
	    assert.Assert(t, os.IsNotExist(err), "got %+v", err)

	    // complex types
	    assert.Assert(t, is.Len(items, 3))
	    assert.Assert(t, len(sequence) != 0) // NotEmpty
	    assert.Assert(t, is.Contains(mapping, "key"))
	    assert.Assert(t, is.DeepEqual(result, myStruct{Name: "title"}))

	    // pointers and interface
	    assert.Assert(t, is.Nil(ref))
	    assert.Assert(t, ref != nil) // NotNil
	}

Comparisons

https://godoc.org/github.com/gotestyourself/gotestyourself/assert/cmp provides
many common comparisons. Additional comparisons can be written to compare
values in other ways.

Below is an example of a custom comparison using a regex pattern:

	func RegexP(value string, pattern string) func() (bool, string) {
	    return func() (bool, string) {
	        re := regexp.MustCompile(pattern)
	        msg := fmt.Sprintf("%q did not match pattern %q", value, pattern)
	        return re.MatchString(value), msg
	    }
	}

*/
package assert

import (
	"fmt"

	"github.com/gotestyourself/gotestyourself/assert/cmp"
	"github.com/gotestyourself/gotestyourself/internal/format"
	"github.com/gotestyourself/gotestyourself/internal/source"
)

// BoolOrComparison can be a bool, or cmp.Comparison. Other types will panic.
type BoolOrComparison interface{}

// TestingT is the subset of testing.T used by the assert package.
type TestingT interface {
	FailNow()
	Fail()
	Log(args ...interface{})
}

type helperT interface {
	Helper()
}

const failureMessage = "assertion failed: "

func assert(
	t TestingT,
	failer func(),
	argsFilter astExprListFilter,
	comparison BoolOrComparison,
	msgAndArgs ...interface{},
) bool {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	var success bool
	switch check := comparison.(type) {
	case bool:
		if check {
			return true
		}
		logFailureFromBool(t, msgAndArgs...)

	// Undocumented legacy comparison without Result type
	case func() (success bool, message string):
		success = runCompareFunc(t, check, msgAndArgs...)

	case cmp.Comparison:
		success = runComparison(t, argsFilter, check, msgAndArgs...)

	case func() cmp.Result:
		success = runComparison(t, argsFilter, check, msgAndArgs...)

	default:
		panic(fmt.Sprintf("comparison arg must be bool or Comparison, not %T", comparison))
	}

	if success {
		return true
	}
	failer()
	return false
}

func runCompareFunc(
	t TestingT,
	f func() (success bool, message string),
	msgAndArgs ...interface{},
) bool {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	if success, message := f(); !success {
		t.Log(format.WithCustomMessage(failureMessage+message, msgAndArgs...))
		return false
	}
	return true
}

func logFailureFromBool(t TestingT, msgAndArgs ...interface{}) {
	const stackIndex = 3 // Assert()/Check(), assert(), formatFailureFromBool()
	const comparisonArgPos = 1
	source, err := source.FormattedCallExprArg(stackIndex, comparisonArgPos)
	if err != nil {
		t.Log(err.Error())
	}

	msg := " is false"
	t.Log(format.WithCustomMessage(failureMessage+source+msg, msgAndArgs...))
}

// Assert performs a comparison, marks the test as having failed if the comparison
// returns false, and stops execution immediately.
func Assert(t TestingT, comparison BoolOrComparison, msgAndArgs ...interface{}) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	assert(t, t.FailNow, filterExprArgsFromComparison, comparison, msgAndArgs...)
}

// Check performs a comparison and marks the test as having failed if the comparison
// returns false. Returns the result of the comparison.
func Check(t TestingT, comparison BoolOrComparison, msgAndArgs ...interface{}) bool {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	return assert(t, t.Fail, filterExprArgsFromComparison, comparison, msgAndArgs...)
}

// NilError fails the test immediately if the last arg is a non-nil error.
// This is equivalent to Assert(t, cmp.NilError(err)).
func NilError(t TestingT, err error, msgAndArgs ...interface{}) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	assert(t, t.FailNow, filterExprExcludeFirst, cmp.NilError(err), msgAndArgs...)
}

// Equal uses the == operator to assert two values are equal and fails the test
// if they are not equal. This is equivalent to Assert(t, cmp.Equal(x, y)).
func Equal(t TestingT, x, y interface{}, msgAndArgs ...interface{}) {
	if ht, ok := t.(helperT); ok {
		ht.Helper()
	}
	assert(t, t.FailNow, filterExprExcludeFirst, cmp.Equal(x, y), msgAndArgs...)
}
