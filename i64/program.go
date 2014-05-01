package i64

import (
	"bytes"
	"fmt"
	"io"
)

// Program is an amd64 program.
type Program []Instruction

// layOut translates Instruction into ins and resolves labels.
func (p Program) layOut() ([]ins, error) {
	// Collect labels.
	labels := make(map[string]int)
	for i := 0; i < len(p); i++ {
		if p[i].Op == LABEL {
			name := p[i].From.Name
			if _, ok := labels[name]; ok {
				return nil, fmt.Errorf("ins %d: label %q previously defined at ins %d", i, name, labels[name])
			}
			labels[name] = i
		}
	}

	// Use short jump if labels are close, otherwise mark it as a 16-bit jump.
	// Collect the jumps so we can update the call sites after codeblock offsets
	// are calculated.
	//
	// The amd64 spec limits Ops to being at most 15 bytes. However, this is an
	// ultra-conservative limit, we could do much better by measuring the upper bound
	// on known instructions..
	const shortJumpLimit = 17
	var jumps []int
	for i := 0; i < len(p); i++ {
		if l := &p[i].To; l.Type == Label {
			jumps = append(jumps, i)
			if p[i].Op == CALL {
				// TODO(crawshaw): Rel16 CALLs should be possible.
				l.Type = Rel32
				l.Value = int8(0)
			} else if i-labels[l.Name] < shortJumpLimit {
				l.Type = Rel8
				l.Value = int8(0)
			} else {
				// TODO(crawshaw): Support Rel32 jumps.
				l.Type = Rel16
				l.Value = int16(0)
			}
		}
	}

	// Lay out instructions.
	laidOut := make([]ins, len(p))
	for i := 0; i < len(p); i++ {
		if err := laidOut[i].make(&p[i]); err != nil {
			return nil, fmt.Errorf("%v: %v", p[i], err)
		}
	}

	// Calculate codeblock offsets.
	buf := new(bytes.Buffer)
	codeblock := 0
	for i := 0; i < len(p); i++ {
		laidOut[i].codeblock = codeblock
		buf.Reset()
		laidOut[i].writeTo(buf)
		codeblock += buf.Len()
		laidOut[i].codeblockEnd = codeblock
	}

	// Update jump locations now that we have real offsets.
	for _, jump := range jumps {
		target := &laidOut[labels[p[jump].To.Name]]
		val := target.codeblock - laidOut[jump].codeblockEnd
		switch p[jump].To.Type {
		case Rel8:
			p[jump].To.Value = int8(val)
		case Rel16:
			p[jump].To.Value = int16(val)
		case Rel32:
			p[jump].To.Value = int32(val)
		default:
			panic(fmt.Sprintf("unexpected jump type: %v", p[jump].To.Type))
		}
		laidOut[jump].make(&p[jump]) // remake
	}

	return laidOut, nil
}

// WriteTo writes the assembled bytes of a program to w.
func (p Program) WriteTo(w io.Writer) (n int64, err error) {
	laidOut, err := p.layOut()
	if err != nil {
		return 0, err
	}
	for _, c := range laidOut {
		n1, err := c.writeTo(w)
		n += n1
		if err != nil {
			return n, err
		}
	}
	return n, err
}

// Bytes returns the assembled bytes of a program.
func (p Program) Bytes() ([]uint8, error) {
	buf := new(bytes.Buffer)
	if _, err := p.WriteTo(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// PrintText writes a textual representation of the program to w.
func (p Program) PrintText(w io.Writer) error {
	laidOut, err := p.layOut()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	for i, ins := range laidOut {
		if p[i].Op == LABEL {
			fmt.Fprintf(w, "%s:\n", p[i].From.Name)
			continue
		}
		buf.Reset()
		ins.writeTo(buf)
		b := buf.Bytes()
		fmt.Fprintf(w, "%06x  %x", ins.codeblock, b)
		const bPad = "                     "
		if len(b)*2 < len(bPad) {
			fmt.Fprint(w, bPad[len(b)*2:])
		}
		fmt.Fprint(w, " | ")

		ins.printText(w)
		if i+1 < len(p) {
			io.WriteString(w, "\n")
		}
	}
	return nil
}
