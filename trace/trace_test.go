package trace

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	tracer := New(&buf)
	if tracer == nil {
		t.Error("Return from New should not be nil")
	} else {
		msg := "Hello trace package."
		tracer.Trace(msg)
		if buf.String() != msg+"\n" {
			t.Errorf("Trace should not write '%s'.", buf.String())
		}
	}
}
