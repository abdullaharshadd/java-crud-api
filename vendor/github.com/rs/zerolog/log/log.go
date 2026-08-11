package log

import (
	"fmt"
	"os"
)

type Event struct {
	err error
	msg string
}

func (e *Event) Err(err error) *Event {
	e.err = err
	return e
}

func (e *Event) Msg(msg string) {
	if e.err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, e.err)
	} else {
		fmt.Fprintln(os.Stderr, msg)
	}
}

func (e *Event) Str(key, val string) *Event {
	return e
}

func Info() *Event  { return &Event{} }
func Error() *Event { return &Event{} }

func Fatal() *fatalEvent { return &fatalEvent{} }

type fatalEvent struct {
	err error
}

func (f *fatalEvent) Err(err error) *fatalEvent {
	f.err = err
	return f
}

func (f *fatalEvent) Msg(msg string) {
	if f.err != nil {
		fmt.Fprintf(os.Stderr, "FATAL %s: %v\n", msg, f.err)
	} else {
		fmt.Fprintln(os.Stderr, "FATAL "+msg)
	}
	os.Exit(1)
}