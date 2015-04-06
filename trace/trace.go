// Copyright 2015 mparaiso <mparaiso@online.fr>
// License MIT

package trace

import (
	"fmt"
	"io"
)

//Tracer describres an object capable of tracing events through code
type Tracer interface {
	Trace(...interface{})
}

type tracer struct {
	out io.Writer
}

func (t *tracer) Trace(a ...interface{}) {
	t.out.Write([]byte(fmt.Sprint(a...)))
	t.out.Write([]byte("\n"))
}
func New(w io.Writer) Tracer {
	return &tracer{out: w}
}
