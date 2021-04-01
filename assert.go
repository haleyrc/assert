package assert

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

type EmptyReporter interface {
	IsEmpty() bool
}

type ZeroReporter interface {
	IsZero() bool
}

type ErrorPredicate func(err error) (string, bool)

type Assert struct {
	failed bool
	t      *testing.T
}

func New(t *testing.T) *Assert {
	return &Assert{t: t}
}

func (a *Assert) Equal(got, want interface{}) *Assert {
	a.t.Helper()
	a.failed = false

	switch w := want.(type) {
	case string:
		g, ok := got.(string)
		if !ok {
			a.failed = true
			a.t.Errorf("mismatched types.\nwanted:\n\t(%T)(%v)\ngot:\n\t(%T)(%v)", want, want, got, got)
			return a
		}
		if g != w {
			a.failed = true
		}
	default:
		if !reflect.DeepEqual(got, want) {
			a.failed = true
		}
	}

	if a.failed {
		a.t.Errorf(
			"expected values to be equal.\nwanted:\n\t%T(%v)\ngot:\n\t%T(%v)",
			want,
			want,
			got,
			got,
		)
	}

	return a
}

func (a *Assert) Error(err error, regex string) *Assert {
	a.t.Helper()
	a.failed = false
	return a.ErrorSatisfies(err, errMatches(regex))
}

func (a *Assert) ErrorSatisfies(err error, pred ErrorPredicate) *Assert {
	a.t.Helper()
	a.failed = false

	if err == nil {
		a.failed = true
		a.t.Errorf("expected error to be not nil, but got %v", err)
		return a
	}

	if msg, ok := pred(err); !ok {
		a.failed = true
		a.t.Error(msg)
	}

	return a
}

func (a *Assert) False(label string, pred bool) *Assert {
	a.t.Helper()
	a.failed = false

	if pred {
		a.failed = true
		a.t.Errorf("expected %q to be false, but was actually true", label)
	}

	return a
}

func (a *Assert) Fatal() {
	a.t.Helper()

	failed := a.failed
	a.failed = false

	if failed {
		a.t.FailNow()
	}
}

func (a *Assert) NotEmpty(label string, value interface{}) *Assert {
	a.t.Helper()
	a.failed = false

	switch v := value.(type) {
	case EmptyReporter:
		a.failed = v.IsEmpty()
	case ZeroReporter:
		a.failed = v.IsZero()
	case string:
		a.failed = v == ""
	default:
		err := fmt.Errorf("assert/NotEmpty: unhandled type %T for value %v", value, value)
		panic(err)
	}

	if a.failed {
		a.t.Errorf("expected %s to not be empty, but it was", label)
	}

	return a
}

func (a *Assert) OK(err error) *Assert {
	a.t.Helper()
	a.failed = false

	if err != nil {
		a.failed = true
		a.t.Errorf("expected error to be nil, but got %v", err)
	}

	return a
}

func (a *Assert) True(label string, pred bool) *Assert {
	a.t.Helper()
	a.failed = false

	if !pred {
		a.failed = true
		a.t.Errorf("expected %q to be true, but was actually false", label)
	}

	return a
}

func errMatches(regex string) ErrorPredicate {
	return func(err error) (string, bool) {
		matched, matchErr := regexp.MatchString(regex, err.Error())
		if matchErr != nil {
			return matchErr.Error(), false
		}
		return fmt.Sprintf("expected %q to match %q but didn't", err.Error(), regex), matched
	}
}
