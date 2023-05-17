package assertive_test

import (
	"fmt"
	"testing"

	. "cdop.pt/go/open/assertive"
)

func TestWantFailed(t *testing.T) {
	ft := &fakeT{logs: []string{}}

	// act on a fake T
	Want(ft, 11 == 12)

	// assert with a real T, this library tests itself
	Need(t, len(ft.logs) == 1)
	Want(t, ft.logs[0] == "\tWant(ft, 11 == 12)")
	Want(t, ft.stop == false)
}

func TestWantSucceeded(t *testing.T) {
	ft := &fakeT{logs: []string{}}

	Want(ft, 13 == 13)

	Want(t, len(ft.logs) == 0)
	Want(t, ft.stop == false)
}

func TestNeedFailed(t *testing.T) {
	ft := &fakeT{logs: []string{}}

	Need(ft, 21 == 22)

	Need(t, len(ft.logs) == 1)
	Want(t, ft.logs[0] == "\tNeed(ft, 21 == 22)")
	Want(t, ft.stop == true)
}

func TestNeedSucceeded(t *testing.T) {
	ft := &fakeT{logs: []string{}}

	Need(ft, 23 == 23)

	Want(t, len(ft.logs) == 0)
	Want(t, ft.stop == false)
}

type fakeT struct {
	logs []string
	stop bool
}

func (t *fakeT) Error(args ...any) {
	t.logs = append(t.logs, fmt.Sprint(args...))
}

func (t *fakeT) Fatal(args ...any) {
	t.logs = append(t.logs, fmt.Sprint(args...))
	t.stop = true
}

func (t *fakeT) Helper() {}
